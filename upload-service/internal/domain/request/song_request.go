package request

type CreateSongRequest struct {
	Name      string  `json:"name"`
	GenreID   int64   `json:"genre_id"`
	Link      string  `json:"link"`
	ArtistIDs []int64 `json:"artist_ids"`
}

type UpdateSongRequest struct {
	Name    string `json:"name"`
	GenreID int64  `json:"genre_id"`
	Link    string `json:"link"`
}
