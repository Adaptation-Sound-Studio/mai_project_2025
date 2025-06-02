package request

type RegisterArtistRequest struct {
	Name string `json:"name"`
}

type UpdateArtistRequest struct {
	Name string `json:"name"`
}
