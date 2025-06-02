package album

import (
	"context"
	"database/sql"
	"upload-service/internal/domain/model"
)

type Repository interface {
	GetAllAlbums(ctx context.Context) ([]model.Album, error)
	GetAlbumByID(ctx context.Context, id int64) (*model.Album, error)
	CheckSongsExist(ctx context.Context, tx *sql.Tx, songIDs []int64) ([]int64, error)
	BatchInsertSongsToAlbum(ctx context.Context, tx *sql.Tx, albumID int64, songIDs []int64) error
	GetSongsByAlbumID(ctx context.Context, albumID int64) ([]model.Song, error)
	CreateAlbum(ctx context.Context, tx *sql.Tx, album *model.Album) (int64, error)
	UpdateAlbum(ctx context.Context, album *model.Album) error
	GetGenreByAlbumID(ctx context.Context, albumID int64) (*model.Genre, error)
	GetArtistByAlbumID(ctx context.Context, albumID int64) (*model.Artist, error)
	CheckSongsBelongToArtist(ctx context.Context, tx *sql.Tx, artistID int64, songIDs []int64) ([]int64, error)
}
