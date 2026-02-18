package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/alexisprovost/immprune/internal/config"
	"github.com/alexisprovost/immprune/internal/immich"
	"github.com/alexisprovost/immprune/internal/photos"
	"github.com/alexisprovost/immprune/internal/types"
	"github.com/spf13/cobra"
)

var (
	onlyVideos bool
	afterDate  string
	limit      int
	outputFile string
)

var rootCmd = &cobra.Command{Use: "immprune", Short: "Safely prune iCloud Photos (already safe in Immich)"}

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Liste les fichiers sûrs à supprimer d'iCloud (nouveaux d'abord)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Load(); err != nil {
			fmt.Println("Config error:", err)
			os.Exit(1)
		}

		imm := immich.New(config.C.ImmichURL, config.C.ImmichKey)
		immichAssets, err := imm.GetAllAssets(onlyVideos)
		if err != nil {
			fmt.Println("Immich:", err)
			os.Exit(1)
		}

		immichSet := make(map[string]bool)
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
			if a.Checksum != "" {
				immichSet[a.Checksum] = true
			}
		}

		assets, err := photos.GetAssets(onlyVideos)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var safe []types.Asset
		after, _ := time.Parse("2006-01-02", afterDate)
		for _, a := range assets {
			if afterDate != "" && !a.Date.After(after) {
				continue
			}
			key := fmt.Sprintf("%s|%d|%s", a.OriginalFilename, a.OriginalFilesize, a.Date.Format("2006-01-02 15:04:05"))
			if immichSet[key] {
				safe = append(safe, a)
			}
		}

		sort.Slice(safe, func(i, j int) bool { return safe[i].Date.After(safe[j].Date) })
		if limit > 0 && len(safe) > limit {
			safe = safe[:limit]
		}

		f, _ := os.Create(outputFile)
		defer f.Close()
		fmt.Fprintf(f, "🚀 %s — %d fichiers sûrs à supprimer d'iCloud Photos\n", "immprune", len(safe))
		fmt.Fprintf(f, "Scan: %s\n\n", time.Now().Format(time.RFC3339))
		for _, a := range safe {
			fmt.Fprintf(f, "%s | %d MB | %s | %s | %s\n",
				a.Date.Format("2006-01-02"),
				a.OriginalFilesize/1024/1024,
				a.OriginalFilename,
				map[bool]string{true: "VIDEO", false: "PHOTO"}[a.IsMovie],
				a.LocalPath)
		}
		fmt.Printf("✅ Fichier prêt → %s (supprime du haut vers le bas)\n", outputFile)
	},
}

func main() {
	compareCmd.Flags().BoolVar(&onlyVideos, "only-videos", false, "Vidéos seulement")
	compareCmd.Flags().StringVar(&afterDate, "after", "", "Après YYYY-MM-DD")
	compareCmd.Flags().IntVar(&limit, "limit", 0, "Limite de résultats")
	compareCmd.Flags().StringVar(&outputFile, "output", "safe_to_delete.txt", "Fichier de sortie")

	rootCmd.AddCommand(compareCmd)
	rootCmd.Execute()
}
