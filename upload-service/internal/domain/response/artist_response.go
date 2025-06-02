package response

import (
	"upload-service/internal/domain/model"
)

type ArtistResponse struct {
	ArtistID int64         `json:"artist_id"`
	Name     string        `json:"name"`
	UserID   int64         `json:"user_id"`
	Songs    []model.Song  `json:"songs"`
	Albums   []model.Album `json:"albums"`
}

func MapArtistToResponse(artist *model.Artist, songs []model.Song, albums []model.Album) ArtistResponse {
	return ArtistResponse{
		ArtistID: artist.ArtistID,
		Name:     artist.Name,
		UserID:   artist.UserID,
		Songs:    songs,
		Albums:   albums,
	}
}
