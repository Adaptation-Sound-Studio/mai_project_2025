package request

type CreateGenreRequest struct {
	Name string `json:"name"`
}

type UpdateGenreRequest struct {
	Name string `json:"name"`
}
