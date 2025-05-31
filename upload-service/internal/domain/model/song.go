package model

import "time"

type Song struct {
	SongID    int64     `json:"song_id"`
	Name      string    `json:"name"`
	Auditions int64     `json:"auditions"`
	GenreID   int64     `json:"genre_id"`
	Date      time.Time `json:"date"`
}
