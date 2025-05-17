package artist

import (
	"context"
	"upload-service/internal/domain"
)

type Repository interface {
	GetAllArtists(ctx context.Context) ([]Artist, error)
	GetArtistByID(ctx context.Context, id int64) (*Artist, error)
	GetArtistByUserID(ctx context.Context, userID int64) (*Artist, error)
	CreateArtist(ctx context.Context, artist *Artist, userID int64) (int64, error)
	UpdateArtist(ctx context.Context, artist *Artist) error
	GetSongsByArtistID(ctx context.Context, artistID int64) ([]domain.Song, error)
	GetAlbumsByArtistID(ctx context.Context, artistID int64) ([]domain.Album, error)
}
