package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alexisprovost/immprune/internal/config"
	"github.com/alexisprovost/immprune/internal/immich"
	"github.com/alexisprovost/immprune/internal/photos"
	"github.com/alexisprovost/immprune/internal/types"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	onlyVideos bool
	afterDate  string
	limit      int
	outputFile string
	uiMode     bool
)

type compareOptions struct {
	OnlyVideos  bool
	AfterDate   string
	Limit       int
	OutputFile  string
	UseBatch    bool
	StartYear   int
	EndYear     int
	BatchYears  int
	LimitPerSet int
}

var rootCmd = &cobra.Command{Use: "immprune", Short: "Safely prune iCloud Photos (already safe in Immich)"}

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "List files safe to delete from iCloud Photos (newest first)",
	Run: func(cmd *cobra.Command, args []string) {
		opts, err := collectCompareOptions(cmd)
		if err != nil {
			fmt.Println("❌ Input error:", err)
			os.Exit(1)
		}

		if err := config.Load(); err != nil {
			fmt.Println("Configuration error:", err)
			os.Exit(1)
		}

		imm := immich.New(config.C.ImmichURL, config.C.ImmichKey)
		stopImmich := startBrailleSpinner("🔭 Scanning Immich assets")
		immichAssets, err := imm.GetAllAssets(opts.OnlyVideos)
		stopImmich(err == nil, "Immich scan complete")
		if err != nil {
			fmt.Println("Immich:", err)
			os.Exit(1)
		}

		immichSet := make(map[string]bool)
		immichNameDateCount := make(map[string]int)
		for _, a := range immichAssets {
			name := strings.ToLower(a.OriginalFileName)
			size := a.FileSizeInByte
			if size == 0 {
				size = a.ExifInfo.FileSizeInByte
			}
			dateStr := ""
			if a.DateTimeOriginal != "" {
				t, _ := time.Parse(time.RFC3339, a.DateTimeOriginal)
				dateStr = t.Format("2006-01-02 15:04:05")
			}
			immichSet[fmt.Sprintf("%s|%d|%s", name, size, dateStr)] = true
			immichNameDateCount[fmt.Sprintf("%s|%s", name, dateStr)]++
			if a.Checksum != "" {
				immichSet[a.Checksum] = true
			}
		}

		stopPhotos := startBrailleSpinner("🧩 Reading Apple Photos library")
		assets, err := photos.GetAssets(opts.OnlyVideos)
		stopPhotos(err == nil, "Apple Photos scan complete")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var safe []types.Asset
		after, _ := time.Parse("2006-01-02", opts.AfterDate)
		total := len(assets)
		matchStarted := time.Now()
		for idx, a := range assets {
			if opts.AfterDate != "" && !a.Date.After(after) {
				renderInlineProgress("🧮 Matching", idx+1, total, displayName(a), matchStarted)
				continue
			}
			if opts.UseBatch {
				y := a.Date.Year()
				if y < opts.StartYear || y > opts.EndYear {
					renderInlineProgress("🧮 Matching", idx+1, total, displayName(a), matchStarted)
					continue
				}
			}
			dateKey := a.Date.Format("2006-01-02 15:04:05")
			strictKey := fmt.Sprintf("%s|%d|%s", a.OriginalFilename, a.OriginalFilesize, dateKey)
			fallbackKey := fmt.Sprintf("%s|%s", a.OriginalFilename, dateKey)

			if immichSet[strictKey] || (a.OriginalFilesize == 0 && immichNameDateCount[fallbackKey] == 1) {
				safe = append(safe, a)
			}
			renderInlineProgress("🧮 Matching", idx+1, total, displayName(a), matchStarted)
		}
		finishInlineProgress(fmt.Sprintf("🧮 Matching complete: %d safe candidates in %s", len(safe), time.Since(matchStarted).Round(time.Second)))

		sort.Slice(safe, func(i, j int) bool { return safe[i].Date.After(safe[j].Date) })

		f, _ := os.Create(opts.OutputFile)
		defer f.Close()

		if !opts.UseBatch && opts.Limit > 0 && len(safe) > opts.Limit {
			safe = safe[:opts.Limit]
		}

		fmt.Fprintf(f, "🛰️ %s — %d files are safe to delete from iCloud Photos\n", "immprune", len(safe))
		fmt.Fprintf(f, "Scan: %s\n\n", time.Now().Format(time.RFC3339))

		if opts.UseBatch {
			writeBatchedReport(f, safe, opts)
		} else {
			for _, a := range safe {
				fmt.Fprintf(f, "%s | %d MB | %s | %s | %s\n",
					a.Date.Format("2006-01-02"),
					a.OriginalFilesize/1024/1024,
					a.OriginalFilename,
					map[bool]string{true: "VIDEO", false: "PHOTO"}[a.IsMovie],
					a.LocalPath)
			}
		}
		fmt.Printf("🧾 Output file ready → %s (delete from top to bottom)\n", opts.OutputFile)
	},
}

func collectCompareOptions(cmd *cobra.Command) (compareOptions, error) {
	opts := compareOptions{
		OnlyVideos:  onlyVideos,
		AfterDate:   afterDate,
		Limit:       limit,
		OutputFile:  outputFile,
		UseBatch:    false,
		BatchYears:  1,
		LimitPerSet: limit,
	}

	if !uiMode {
		return opts, nil
	}

	useWizard := !cmd.Flags().Changed("after") && !cmd.Flags().Changed("limit") && !cmd.Flags().Changed("only-videos")
	if !useWizard {
		return opts, nil
	}

	fmt.Println("🪄 Smart compare wizard")

	videoSelect := promptui.Select{Label: "Content scope", Items: []string{"All photos and videos", "Videos only"}}
	_, videoChoice, err := videoSelect.Run()
	if err == nil {
		opts.OnlyVideos = videoChoice == "Videos only"
	}

	batchSelect := promptui.Select{Label: "Check style", Items: []string{"Single pass", "Year batches"}}
	_, batchChoice, err := batchSelect.Run()
	if err != nil {
		return opts, err
	}

	if batchChoice == "Year batches" {
		opts.UseBatch = true

		startYearPrompt := promptui.Prompt{Label: "Start year (Enter for default)", Default: "2018"}
		startYearRaw, err := startYearPrompt.Run()
		if err != nil {
			return opts, err
		}
		startYear, err := strconv.Atoi(strings.TrimSpace(startYearRaw))
		if err != nil {
			return opts, fmt.Errorf("invalid start year")
		}

		endYearPrompt := promptui.Prompt{Label: "End year (Enter for default)", Default: strconv.Itoa(time.Now().Year())}
		endYearRaw, err := endYearPrompt.Run()
		if err != nil {
			return opts, err
		}
		endYear, err := strconv.Atoi(strings.TrimSpace(endYearRaw))
		if err != nil {
			return opts, fmt.Errorf("invalid end year")
		}
		if endYear < startYear {
			return opts, fmt.Errorf("end year must be >= start year")
		}

		batchYearsPrompt := promptui.Prompt{Label: "Years per batch (Enter for default)", Default: "2"}
		batchRaw, err := batchYearsPrompt.Run()
		if err != nil {
			return opts, err
		}
		batchYears, err := strconv.Atoi(strings.TrimSpace(batchRaw))
		if err != nil || batchYears <= 0 {
			return opts, fmt.Errorf("invalid years per batch")
		}

		limitPrompt := promptui.Prompt{Label: "Limit per batch (0 = no limit, Enter for default)", Default: "0"}
		limitRaw, err := limitPrompt.Run()
		if err != nil {
			return opts, err
		}
		limitPerBatch, err := strconv.Atoi(strings.TrimSpace(limitRaw))
		if err != nil || limitPerBatch < 0 {
			return opts, fmt.Errorf("invalid batch limit")
		}

		opts.StartYear = startYear
		opts.EndYear = endYear
		opts.BatchYears = batchYears
		opts.LimitPerSet = limitPerBatch
	}

	outputPrompt := promptui.Prompt{Label: "Output file (Enter for default)", Default: opts.OutputFile}
	outputRaw, err := outputPrompt.Run()
	if err == nil && strings.TrimSpace(outputRaw) != "" {
		opts.OutputFile = strings.TrimSpace(outputRaw)
	}

	return opts, nil
}

func writeBatchedReport(f *os.File, safe []types.Asset, opts compareOptions) {
	batches := buildYearBatches(opts.StartYear, opts.EndYear, opts.BatchYears)
	fmt.Printf("📚 Writing %d batches\n", len(batches))

	for idx, b := range batches {
		start := b[0]
		end := b[1]
		var group []types.Asset
		fmt.Printf("🔹 Batch %d/%d: %d-%d\n", idx+1, len(batches), start, end)
		for _, a := range safe {
			y := a.Date.Year()
			if y >= start && y <= end {
				group = append(group, a)
			}
		}

		sort.Slice(group, func(i, j int) bool { return group[i].Date.After(group[j].Date) })
		if opts.LimitPerSet > 0 && len(group) > opts.LimitPerSet {
			group = group[:opts.LimitPerSet]
		}

		fmt.Fprintf(f, "📦 Batch %d-%d | %d candidates\n", start, end, len(group))
		batchStarted := time.Now()
		for i, a := range group {
			renderInlineProgress("📝 Writing batch", i+1, len(group), displayName(a), batchStarted)
			fmt.Fprintf(f, "%s | %d MB | %s | %s | %s\n",
				a.Date.Format("2006-01-02"),
				a.OriginalFilesize/1024/1024,
				a.OriginalFilename,
				map[bool]string{true: "VIDEO", false: "PHOTO"}[a.IsMovie],
				a.LocalPath)
		}
		if len(group) > 0 {
			finishInlineProgress(fmt.Sprintf("📝 Batch %d-%d written in %s", start, end, time.Since(batchStarted).Round(time.Second)))
		}
		fmt.Fprintln(f)
	}
}

func buildYearBatches(startYear, endYear, size int) [][2]int {
	var result [][2]int
	for from := endYear; from >= startYear; from -= size {
		to := from - size + 1
		if to < startYear {
			to = startYear
		}
		result = append(result, [2]int{to, from})
	}
	return result
}

func startBrailleSpinner(label string) func(success bool, doneMessage string) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	stop := make(chan struct{})
	done := make(chan struct{})
	var once sync.Once

	go func() {
		i := 0
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				fmt.Print("\r\033[K")
				close(done)
				return
			case <-ticker.C:
				fmt.Printf("\r%s %s", frames[i%len(frames)], label)
				i++
			}
		}
	}()

	return func(success bool, doneMessage string) {
		once.Do(func() {
			close(stop)
			<-done
			if success {
				fmt.Printf("✨ %s\n", doneMessage)
			} else {
				fmt.Printf("💥 %s failed\n", doneMessage)
			}
		})
	}
}

func renderInlineProgress(prefix string, current int, total int, item string, startedAt time.Time) {
	if total <= 0 {
		return
	}
	if len(item) > 42 {
		item = item[:39] + "..."
	}
	percent := (current * 100) / total
	elapsed := time.Since(startedAt)
	eta := "--:--"
	if current > 0 {
		remaining := total - current
		if remaining < 0 {
			remaining = 0
		}
		remainingDuration := time.Duration(float64(elapsed) * (float64(remaining) / float64(current)))
		eta = formatShortDuration(remainingDuration)
	}
	fmt.Printf("\r%s %3d%% (%d/%d) ETA %s  %s", prefix, percent, current, total, eta, item)
}

func finishInlineProgress(message string) {
	fmt.Print("\r\033[K")
	fmt.Println(message)
}

func displayName(a types.Asset) string {
	if a.OriginalFilename != "" {
		return a.OriginalFilename
	}
	if a.LocalPath != "" {
		return filepath.Base(a.LocalPath)
	}
	return "(unknown)"
}

func formatShortDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Round(time.Second)
	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func main() {
	compareCmd.Flags().BoolVar(&onlyVideos, "only-videos", false, "Videos only")
	compareCmd.Flags().StringVar(&afterDate, "after", "", "Only assets after YYYY-MM-DD")
	compareCmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results")
	compareCmd.Flags().StringVar(&outputFile, "output", "safe_to_delete.txt", "Output file path")
	compareCmd.Flags().BoolVar(&uiMode, "ui", true, "Enable interactive wizard UI")

	rootCmd.AddCommand(compareCmd)
	rootCmd.Execute()
}
