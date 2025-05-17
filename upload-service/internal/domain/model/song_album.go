package domain

type SongAlbum struct {
	SongAlbumID int64 `json:"sa_id"`
	SongID      int64 `json:"song_id"`
	AlbumID     int64 `json:"album_id"`
}
