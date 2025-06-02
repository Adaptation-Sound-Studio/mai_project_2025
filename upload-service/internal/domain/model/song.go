package model

import "time"

type Song struct {
	SongID      int64     `json:"song_id"`
	Name        string    `json:"name"`
	NameOfMinio string    `json:"NameOfMinio"`
	Auditions   int64     `json:"auditions"`
	GenreID     int64     `json:"genre_id"`
	Date        time.Time `json:"date"`
}

type SongElastic struct {
	SongID      int64  `json:"song_id"`
	Name        string `json:"name"`
	NameOfMinio string `json:"NameOfMinio"`
	Auditions   int64  `json:"auditions"`
	Genre       string `json:"genre"`
	Artist      string `json:"artist"`
}
