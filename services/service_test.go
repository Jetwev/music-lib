package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSongInfoFromAPI(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/info" {
            http.Error(w, "Not Found", http.StatusNotFound)
            return
        }

        detail := SongDetail{
            ReleaseDate: "2024-09-26",
            Text:        "La la la la la",
            Link:        "https://www.youtube.com",
        }
        w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(detail)
    }))
    defer server.Close()
	
    detail, err := GetSongInfoFromAPI("Group", "Title", server.URL)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }

    if detail.ReleaseDate != "2024-09-26" {
        t.Errorf("Expected ReleaseDate to be '2024-09-26', got %v", detail.ReleaseDate)
    }

	if detail.Text != "La la la la la" {
        t.Errorf("Expected Text to be 'La la la la la', got %v", detail.ReleaseDate)
    }
	
	if detail.Link != "https://www.youtube.com" {
        t.Errorf("Expected Link to be 'https://www.youtube.com', got %v", detail.ReleaseDate)
    }
}
