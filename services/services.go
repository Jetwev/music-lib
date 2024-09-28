package services

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type SongDetail struct {
    ReleaseDate string `json:"releaseDate"`
    Text        string `json:"text"`
    Link        string `json:"link"`
}

func GetSongInfoFromAPI(group, song, externalAPI string) (*SongDetail, error) {
    url := fmt.Sprintf("%s/info?group=%s&song=%s", externalAPI, group, song)
    resp, err := http.Get(url)
	
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to get song information from external API")
    }

    var detail SongDetail
    if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
        return nil, err
    }

    return &detail, nil
}