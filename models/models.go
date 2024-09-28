package models

type Song struct {
    ID          int    `db:"id"`
    Group       string `db:"group"`
    Title       string `db:"title"`
    ReleaseDate string `db:"release_date"`
    Text        string `db:"text"`
    Link        string `db:"link"`
}

type SongInfo struct {
    Group       string `db:"group"`
    Title       string `db:"title"`
    ReleaseDate string `db:"release_date"`
    Text        string `db:"text"`
    Link        string `db:"link"`
}