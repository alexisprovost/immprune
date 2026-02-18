package immich

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/alexisprovost/immprune/internal/types"
)

type Client struct {
	URL string
	Key string
}

func New(url, key string) *Client { return &Client{URL: url, Key: key} }

func (c *Client) GetAllAssets(onlyVideos bool) ([]types.ImmichAsset, error) {
	var all []types.ImmichAsset
	page := 1
	for {
		body := map[string]interface{}{"page": page, "size": 1000, "withExif": true}
		if onlyVideos {
			body["type"] = "VIDEO"
		}
		b, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", c.URL+"/api/search/metadata", bytes.NewReader(b))
		req.Header.Set("x-api-key", c.Key)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result struct {
			Assets []types.ImmichAsset `json:"assets"`
		}
		json.NewDecoder(resp.Body).Decode(&result)

		all = append(all, result.Assets...)
		if len(result.Assets) < 1000 {
			break
		}
		page++
	}
	return all, nil
}
