package artist

import (
	"context"
	"upload-service/internal/domain/model"
)

type Repository interface {
	GetAllArtists(ctx context.Context) ([]model.Artist, error)
	GetArtistByID(ctx context.Context, id int64) (*model.Artist, error)
	GetArtistByUserID(ctx context.Context, userID int64) (*model.Artist, error)
	CreateArtist(ctx context.Context, artist *model.Artist, userID int64) (int64, error)
	UpdateArtist(ctx context.Context, artist *model.Artist) error
	GetSongsByArtistID(ctx context.Context, artistID int64) ([]model.Song, error)
	GetAlbumsByArtistID(ctx context.Context, artistID int64) ([]model.Album, error)
}
