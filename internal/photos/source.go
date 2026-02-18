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
		return nil, fmt.Errorf("macOS seulement pour l'instant (osxphotos). Utilise --folder sur Linux/Windows plus tard")
	}
	cmd := exec.Command("osxphotos", "query", "--json")
	if onlyVideos {
		cmd.Args = append(cmd.Args, "--movies")
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("installe osxphotos : pip install osxphotos")
	}

	var raw []struct {
		UUID             string `json:"uuid"`
		OriginalFilename string `json:"original_filename"`
		OriginalFilesize int64  `json:"original_filesize"`
		Date             string `json:"date"`
		IsMovie          bool   `json:"ismovie"`
		Path             string `json:"path,omitempty"`
	}
	json.Unmarshal(out, &raw)

	var assets []types.Asset
	for _, r := range raw {
		dt, _ := time.Parse(time.RFC3339, r.Date)
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
