package repository

import (
	"context"
	"database/sql"
	"upload-service/internal/domain"
)

type ArtistRepository interface {
	GetAllArtists(ctx context.Context) ([]domain.Artist, error)
	GetArtistByID(ctx context.Context, id int64) (*domain.Artist, error)
	CreateArtist(ctx context.Context, artist *domain.Artist) error
	UpdateArtist(ctx context.Context, artist *domain.Artist) error
}

type artistRepository struct {
	db *sql.DB
}

func NewArtistRepository(db *sql.DB) ArtistRepository {
	return &artistRepository{db: db}
}

func (r *artistRepository) GetAllArtists(ctx context.Context) ([]domain.Artist, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT artist_id, name FROM artists")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artists []domain.Artist
	for rows.Next() {
		var a domain.Artist
		err = rows.Scan(&a.ArtistID, &a.Name)
		if err != nil {
			return nil, err
		}
		artists = append(artists, a)
	}

	return artists, nil
}

func (r *artistRepository) GetArtistByID(ctx context.Context, id int64) (*domain.Artist, error) {
	row := r.db.QueryRowContext(ctx, "SELECT artist_id, name FROM artists WHERE artist_id = $1", id)
	var a domain.Artist
	err := row.Scan(&a.ArtistID, &a.Name)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *artistRepository) CreateArtist(ctx context.Context, artist *domain.Artist) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO artists (name) VALUES ($1)",
		artist.Name,
	)
	return err
}

func (r *artistRepository) UpdateArtist(ctx context.Context, artist *domain.Artist) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE artists SET name = $1 WHERE artist_id = $2",
		artist.Name, artist.ArtistID,
	)
	return err
}
