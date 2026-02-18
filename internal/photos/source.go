package photos

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/alexisprovost/immprune/internal/types"
)

func GetAssets(onlyVideos bool) ([]types.Asset, error) {
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("macOS only for now")
	}
	cmd := exec.Command("osascript", "-l", "JavaScript", "-e", photosJXAScript)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("unable to read Photos library via macOS Automation (check Photos/Automation permissions): %w", err)
	}

	var raw []struct {
		UUID             string `json:"uuid"`
		OriginalFilename string `json:"original_filename"`
		OriginalFilesize int64  `json:"original_filesize"`
		Date             string `json:"date"`
		IsMovie          bool   `json:"ismovie"`
		Path             string `json:"path,omitempty"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("invalid Photos response: %w", err)
	}

	var assets []types.Asset
	for _, r := range raw {
		dt, _ := time.Parse(time.RFC3339, r.Date)
		if onlyVideos && !r.IsMovie {
			continue
		}
		assets = append(assets, types.Asset{
			UUID:             r.UUID,
			OriginalFilename: strings.ToLower(r.OriginalFilename),
			OriginalFilesize: r.OriginalFilesize,
			Date:             dt,
			IsMovie:          r.IsMovie,
			LocalPath:        r.Path,
		})
	}
	return assets, nil
}

const photosJXAScript = `
const photos = Application("Photos");

function safeCall(fn, fallback) {
	try {
		const value = fn();
		return value === undefined || value === null ? fallback : value;
	} catch (e) {
		return fallback;
	}
}

const items = photos.mediaItems();
const result = [];

for (let i = 0; i < items.length; i++) {
	const item = items[i];
	const mediaType = String(safeCall(() => item.mediaType(), "")).toLowerCase();
	const dateValue = safeCall(() => item.date(), null);
	const isoDate = dateValue ? new Date(dateValue).toISOString() : "";

	result.push({
		uuid: String(safeCall(() => item.id(), "")),
		original_filename: String(safeCall(() => item.filename(), "")),
		original_filesize: 0,
		date: isoDate,
		ismovie: mediaType.indexOf("video") >= 0,
		path: ""
	});
}

JSON.stringify(result);
`
