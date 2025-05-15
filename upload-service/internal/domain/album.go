package domain

import "time"

type Album struct {
	AlbumID   int64     `json:"album_id"`
	Name      string    `json:"name"`
	Auditions int64     `json:"auditions"`
	ArtistID  int64     `json:"artist_id"`
	GenreID   int64     `json:"genre_id"`
	Date      time.Time `json:"date"`
}
