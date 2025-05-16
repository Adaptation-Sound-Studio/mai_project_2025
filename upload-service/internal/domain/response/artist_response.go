package response

import "upload-service/internal/domain"

type ArtistResponse struct {
	ArtistID int64          `json:"artist_id"`
	Name     string         `json:"name"`
	UserID   int64          `json:"user_id"`
	Songs    []domain.Song  `json:"songs"`
	Albums   []domain.Album `json:"albums"`
}

func MapArtistToResponse(artist *domain.Artist, songs []domain.Song, albums []domain.Album) ArtistResponse {
	return ArtistResponse{
		ArtistID: artist.ArtistID,
		Name:     artist.Name,
		UserID:   artist.UserID,
		Songs:    songs,
		Albums:   albums,
	}
}
