package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"music-lib/handlers"
	"net/http"
	"os"

	_ "music-lib/docs"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)


func config() map[string]string {
	err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }
	dConfig := make(map[string]string)
	dConfig["url"] = os.Getenv("DATABASE_URL")
	dConfig["port"] = os.Getenv("PORT")
	dConfig["externalAPI"] = os.Getenv("EXTERNAL_API_URL")
	dConfig["level"] = os.Getenv("LOGGER_LEVEL")
	return dConfig
}


func main() {
	dConfig := config()
    db, err := sql.Open("postgres", dConfig["url"])
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

	var logger *slog.Logger
	if dConfig["level"] == "debug" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else if dConfig["level"] == "info"{
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	} else {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
		logger.Info("Unknown logger level", "level", dConfig["level"])
	}
		
    r := chi.NewRouter()

    r.Get("/songs", handlers.GetSongsHandler(db, logger)) // get all information
    r.Get("/songs/text", handlers.GetSongTextHandler(db, logger)) // only text
    r.Post("/songs", handlers.AddSongHandler(db, dConfig["externalAPI"], logger)) // post a new one
    r.Put("/songs/{id}", handlers.UpdateSongHandler(db, logger)) // upate song
    r.Delete("/songs/{id}", handlers.DeleteSongHandler(db, logger)) // delete song
	r.Get("/swagger/*", httpSwagger.WrapHandler) // swagger docs

	logger.Info("Starting server", "port", dConfig["port"])
    http.ListenAndServe(fmt.Sprintf(":%s", dConfig["port"]), r)
}
