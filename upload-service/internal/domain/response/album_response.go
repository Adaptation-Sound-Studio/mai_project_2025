package response

import (
	"time"
	"upload-service/internal/domain"
)

type AlbumResponse struct {
	AlbumID   int64         `json:"album_id"`
	Name      string        `json:"name"`
	Auditions int64         `json:"auditions"`
	ArtistID  int64         `json:"artist_id"`
	GenreID   int64         `json:"genre_id"`
	Date      time.Time     `json:"date"`
	Songs     []domain.Song `json:"songs"`
}

func MapAlbumToResponse(album *domain.Album, songs []domain.Song) AlbumResponse {
	return AlbumResponse{
		AlbumID:   album.AlbumID,
		Name:      album.Name,
		Auditions: album.Auditions,
		ArtistID:  album.ArtistID,
		GenreID:   album.GenreID,
		Date:      album.Date,
		Songs:     songs,
	}
}
