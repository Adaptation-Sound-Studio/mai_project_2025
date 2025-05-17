package domain

type SongArtist struct {
	SongArtistID int64 `json:"sa_id"`
	SongID       int64 `json:"song_id"`
	ArtistID     int64 `json:"artist_id"`
}
