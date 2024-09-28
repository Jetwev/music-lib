package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"music-lib/models"
	"music-lib/services"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// GetSongsHandler get all info about the song
// @Summary      Get all info about songs
// @Description  Get songs with optional filtering by group and song title, and pagination (songs per page)
// @Tags         songs
// @Param        group  query  string  false  "Filter by group"
// @Param        song   query  string  false  "Filter by song title"
// @Param        page   query  int     false  "Page number"   default(1)
// @Param        limit  query  int     false  "Limit per page" default(10)
// @Success      200    {array} models.Song
// @Failure      500    {string} string "Internal Server Error"
// @Router       /songs [get]
func GetSongsHandler(db *sql.DB, logger *slog.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        query := "SELECT * FROM songs"
        
        group := r.URL.Query().Get("group")
        if group != "" {
            query += fmt.Sprintf(" WHERE \"group\" = '%s'", group)
			logger.Debug("Filtering by group", "group", group)
        }
        song := r.URL.Query().Get("song")
        if song != "" {
            query += fmt.Sprintf(" AND title = '%s'", song)
			logger.Debug("Filtering by song title", "song", song)
        }
        
        page, _ := strconv.Atoi(r.URL.Query().Get("page"))
        limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
        if page == 0 {
            page = 1
        }
        if limit == 0 {
            limit = 10
        }
        offset := (page - 1) * limit
        query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

        rows, err := db.Query(query)
        if err != nil {
			logger.Error("Failed query", "error", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var songs []models.Song
        for rows.Next() {
            var song models.Song
            if err := rows.Scan(&song.ID, &song.Group, &song.Title, &song.ReleaseDate, &song.Text, &song.Link); err != nil {
				logger.Error("Failed to scan values", "error", err)
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            songs = append(songs, song)
        }

        json.NewEncoder(w).Encode(songs)
    }
}

// GetSongTextHandler get song text
// @Summary      Get song text by verses
// @Description  Get song text by song ID, paginated by verses
// @Tags         songs
// @Param        group  query  string  true  "Filter by group"
// @Param        song   query  string  true  "Filter by song title"
// @Param        page   query   int     false  "Page number"   default(1)
// @Param        limit  query   int     false  "Verses per page" default(4)
// @Success      200    {array} string
// @Failure      500    {string} string "Internal Server Error"
// @Router       /songs/text [get]
func GetSongTextHandler(db *sql.DB, logger *slog.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
		query := "SELECT text FROM songs"
		group := r.URL.Query().Get("group")
        if group != "" {
            query += fmt.Sprintf(" WHERE \"group\" = '%s'", group)
			logger.Debug("Filtering by group", "group", group)
        }else{
			logger.Error("Group name is missing")
			http.Error(w, "Group name is missing", http.StatusInternalServerError)
		}
        song := r.URL.Query().Get("song")
        if song != "" {
            query += fmt.Sprintf(" AND title = '%s'", song)
        }else{
			logger.Error("Song title is missing")
			http.Error(w, "Song title is missing", http.StatusInternalServerError)
		}

        page, _ := strconv.Atoi(r.URL.Query().Get("page"))
        limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

        if page == 0 {
            page = 1
        }
        if limit == 0 {
            limit = 4 
        }

        var text string
        err := db.QueryRow(query).Scan(&text)
        if err != nil {
			logger.Error("Failed query scanning", "error", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        verses := splitSongText(text, limit)
        json.NewEncoder(w).Encode(verses[page-1])
    }
}

func splitSongText(text string, count int) [][]string {
    verses := strings.Split(text, "\\n\\n")
	
    var pages [][]string
    for i := 0; i < len(verses); i += count {
        end := i + count
        if end > len(verses) {
            end = len(verses)
        }
        pages = append(pages, verses[i:end])
    }
    
    return pages
}

// AddSongHandler add new song to the db
// @Summary      Add a new song
// @Description  Add a new song to the database, fetch song details from external API
// @Tags         songs
// @Accept       json
// @Produce      json
// @Param        group  query  string  false  "Filter by group"
// @Param        song   query  string  false  "Filter by song title"
// @Success      201    {string} string "Created"
// @Failure      400    {string} string "Bad Request"
// @Failure      500    {string} string "Internal Server Error"
// @Router       /songs [post]
func AddSongHandler(db *sql.DB, externalAPI string, logger *slog.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
		group := r.URL.Query().Get("group")
		song := r.URL.Query().Get("song")

		if (group == "") && (song == ""){
			logger.Error("Empty data to search in external API")
			http.Error(w, "Empty data to search in external API", http.StatusBadRequest)
		} 

        songDetail, err := services.GetSongInfoFromAPI(group, song, externalAPI)
        if err != nil {
			logger.Error("Failed to get song details from external API")
            http.Error(w, "Failed to get song details from external API", http.StatusInternalServerError)
            return
        }

        query := "INSERT INTO songs (\"group\", title, release_date, text, link) VALUES ($1, $2, $3, $4, $5)"
        _, err = db.Exec(query, group, song, songDetail.ReleaseDate, songDetail.Text, songDetail.Link)
        if err != nil {
			logger.Error("Failed query", "error", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusCreated)
    }
}

// UpdateSongHandler update song info
// @Summary      Update song details
// @Description  Update song information by ID
// @Tags         songs
// @Accept       json
// @Param        id    path     string  true  "Song ID"
// @Param        song  body     models.SongInfo true "Song details"
// @Success      200    {string} string "OK"
// @Failure      400    {string} string "Bad Request"
// @Failure      500    {string} string "Internal Server Error"
// @Router       /songs/{id} [put]
func UpdateSongHandler(db *sql.DB, logger *slog.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := chi.URLParam(r, "id")
        var song models.SongInfo
        if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
			logger.Error("Failed json decoding", "error", err)
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        query := "UPDATE songs SET "
        var setFields []string

        // Dynamically add fields to be updated if they are provided
        if song.Group != "" {
            setFields = append(setFields, fmt.Sprintf("\"group\" = '%s'", song.Group))
        }
        if song.Title != "" {
            setFields = append(setFields, fmt.Sprintf("title = '%s'", song.Title))
        }
        if song.ReleaseDate != "" {
            if _, err := time.Parse("2006-01-02", song.ReleaseDate); err != nil {
                logger.Warn("Invalid release date format YYYY-MM-DD", "date", song.ReleaseDate)
            } else {
                setFields = append(setFields, fmt.Sprintf("release_date = '%s'", song.ReleaseDate))
            }
        }
        if song.Text != "" {
            setFields = append(setFields, fmt.Sprintf("text = '%s'", song.Text))
        }
        if song.Link != "" {
            setFields = append(setFields, fmt.Sprintf("link = '%s'", song.Link))
        }
        if len(setFields) == 0 {
            http.Error(w, "No fields to update", http.StatusBadRequest)
            return
        }
        
        query += strings.Join(setFields, ", ")
        if id != "" {
            query += fmt.Sprintf(" WHERE id = '%s'", id)
        } else {
            http.Error(w, "ID of the song is required for updating", http.StatusBadRequest)
            return
        }

        _, err := db.Exec(query)
        if err != nil {
			logger.Error("Failed query", "error", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    }
}

// DeleteSongHandler delete song from the db
// @Summary      Delete song
// @Description  Delete song by ID
// @Tags         songs
// @Param        id  path  string  true  "Song ID"
// @Success      200  {string} string "OK"
// @Failure      500  {string} string "Internal Server Error"
// @Router       /songs/{id} [delete]
func DeleteSongHandler(db *sql.DB, logger *slog.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := chi.URLParam(r, "id")
        _, err := db.Exec("DELETE FROM songs WHERE id = $1", id)
        if err != nil {
			logger.Error("Failed query", "error", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    }
}