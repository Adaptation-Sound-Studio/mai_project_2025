package request

type CreateAlbumRequest struct {
	Name    string  `json:"name"`
	GenreID int64   `json:"genre_id"`
	SongIDs []int64 `json:"song_ids"`
}

type UpdateAlbumRequest struct {
	Name    string `json:"name"`
	GenreID int64  `json:"genre_id"`
}
