package song

import (
	"context"
	"upload-service/internal/domain/model"
)

type Repository interface {
	GetAllSongs(ctx context.Context) ([]model.Song, error)
	GetSongByID(ctx context.Context, id int64) (*model.Song, error)
	CreateSongWithArtists(ctx context.Context, song *model.Song, artistIDs []int64) (int64, error)
	UpdateSong(ctx context.Context, song *model.Song) error
	CheckArtistsExist(ctx context.Context, artistIDs []int64) ([]int64, error)
	GetArtistsBySongID(ctx context.Context, songID int64) ([]model.Artist, error)
	GetAlbumBySongID(ctx context.Context, songID int64) (*model.Album, error)
	GetGenreBySongID(ctx context.Context, songID int64) (*model.Genre, error)
	GetOneArtistBySongID(ctx context.Context, songID int64) (*model.Artist, error)
	IncrementAuditions(ctx context.Context, songID int64) error
}
