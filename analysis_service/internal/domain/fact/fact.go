package fact

import "time"

type ListenFact struct {
	ID         int64     `json:"id,omitempty"`
	UserID     int64     `json:"user_id"`
	SongID     int64     `json:"song_id"`
	ArtistID   int64     `json:"artist_id"`
	AlbumID    int64     `json:"album_id"`
	GenreID    int64     `json:"genre_id"`
	ListenedAt time.Time `json:"listened_at"`
}
