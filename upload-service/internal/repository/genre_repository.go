package repository

import (
	"context"
	"database/sql"
	"upload-service/internal/domain"
)

type GenreRepository interface {
	GetAllGenres(ctx context.Context) ([]domain.Genre, error)
	CreateGenre(ctx context.Context, genre *domain.Genre) error
	UpdateGenre(ctx context.Context, genre *domain.Genre) error
}

type genreRepository struct {
	db *sql.DB
}

func NewGenreRepository(db *sql.DB) GenreRepository {
	return &genreRepository{db: db}
}

func (r *genreRepository) GetAllGenres(ctx context.Context) ([]domain.Genre, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT genre_id, name FROM genres")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []domain.Genre
	for rows.Next() {
		var g domain.Genre
		err = rows.Scan(&g.GenreID, &g.Name)
		if err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}

	return genres, nil
}

func (r *genreRepository) CreateGenre(ctx context.Context, genre *domain.Genre) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO genres (name) VALUES ($1)",
		genre.Name,
	)
	return err
}

func (r *genreRepository) UpdateGenre(ctx context.Context, genre *domain.Genre) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE genres SET name = $1 WHERE genre_id = $2",
		genre.Name, genre.GenreID,
	)
	return err
}
