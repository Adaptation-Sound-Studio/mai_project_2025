package model

import "time"

type Album struct {
	AlbumID   int64     `json:"album_id"`
	Name      string    `json:"name"`
	Auditions int64     `json:"auditions"`
	ArtistID  int64     `json:"artist_id"`
	GenreID   int64     `json:"genre_id"`
	Date      time.Time `json:"date"`
}

type AlbumElastic struct {
	AlbumID   int64     `json:"album_id"`
	Name      string    `json:"name"`
	Auditions int64     `json:"auditions"`
	Artist    string    `json:"artist"`
	Genre     string    `json:"genre"`
	Date      time.Time `json:"date"`
}
