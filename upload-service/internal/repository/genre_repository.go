package repository

import (
	"context"
	"database/sql"
	"fmt"
	"upload-service/internal/domain/genre"
	"upload-service/internal/domain/model"
)

type genreRepository struct {
	db *sql.DB
}

func NewGenreRepository(db *sql.DB) genre.Repository {
	return &genreRepository{db: db}
}

func (r *genreRepository) GetAllGenres(ctx context.Context) ([]model.Genre, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT genre_id, name FROM genres")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []model.Genre
	for rows.Next() {
		var g model.Genre
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

func (r *genreRepository) CreateGenre(ctx context.Context, genre *model.Genre) (int64, error) {
	var genreID int64
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO genres (name) VALUES ($1) RETURNING genre_id",
		genre.Name,
	).Scan(&genreID)
	if err != nil {
		return 0, err
	}
	return genreID, nil
}

func (r *genreRepository) UpdateGenre(ctx context.Context, genre *model.Genre) error {
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
