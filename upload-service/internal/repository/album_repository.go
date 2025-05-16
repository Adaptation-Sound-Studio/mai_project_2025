package repository

import (
	"context"
	"database/sql"
	"fmt"
	"upload-service/internal/domain"
)

type AlbumRepository interface {
	GetAllAlbums(ctx context.Context) ([]domain.Album, error)
	GetAlbumByID(ctx context.Context, id int64) (*domain.Album, error)
	CreateAlbum(ctx context.Context, album *domain.Album) error
	UpdateAlbum(ctx context.Context, album *domain.Album) error
}

type albumRepository struct {
	db *sql.DB
}

func NewAlbumRepository(db *sql.DB) AlbumRepository {
	return &albumRepository{db: db}
}

func (r *albumRepository) GetAllAlbums(ctx context.Context) ([]domain.Album, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT album_id, name, auditions, artist_id, genre_id, date FROM albums")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums []domain.Album
	for rows.Next() {
		var a domain.Album
		if err := rows.Scan(&a.AlbumID, &a.Name, &a.Auditions, &a.ArtistID, &a.GenreID, &a.Date); err != nil {
			return nil, err
		}
		albums = append(albums, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return albums, nil
}

func (r *albumRepository) GetAlbumByID(ctx context.Context, id int64) (*domain.Album, error) {
	row := r.db.QueryRowContext(ctx, "SELECT album_id, name, auditions, artist_id, genre_id, date FROM albums WHERE album_id = $1", id)
	var a domain.Album
	err := row.Scan(&a.AlbumID, &a.Name, &a.Auditions, &a.ArtistID, &a.GenreID, &a.Date)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *albumRepository) CreateAlbum(ctx context.Context, album *domain.Album) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO albums (name, auditions, artist_id, genre_id, date) VALUES ($1, $2, $3, $4, $5)",
		album.Name, album.Auditions, album.ArtistID, album.GenreID, album.Date,
	)
	return err
}

func (r *albumRepository) UpdateAlbum(ctx context.Context, album *domain.Album) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE albums SET name = $1, auditions = $2, artist_id = $3, genre_id = $4, date = $5 WHERE album_id = $6",
		album.Name, album.Auditions, album.ArtistID, album.GenreID, album.Date, album.AlbumID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("альбом с ID %d не найден", album.AlbumID)
	}

	return nil
}
