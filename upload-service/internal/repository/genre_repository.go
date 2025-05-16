package repository

import (
	"context"
	"database/sql"
	"fmt"
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
		if err := rows.Scan(&g.GenreID, &g.Name); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}

	if err := rows.Err(); err != nil {
		return nil, err
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
	res, err := r.db.ExecContext(ctx,
		"UPDATE genres SET name = $1 WHERE genre_id = $2",
		genre.Name, genre.GenreID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("жанр с ID %d не найден", genre.GenreID)
	}

	return nil
}
