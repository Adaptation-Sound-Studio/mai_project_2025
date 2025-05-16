package postgres

import (
	"database/sql"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/genre"
)

type GenreRepo struct {
	DB *sql.DB
}

func NewGenreRepo(db *sql.DB) *GenreRepo {
	return &GenreRepo{DB: db}
}

func (r *GenreRepo) Create(genre *genre.Genre) error {
	_, err := r.DB.Exec("INSERT INTO genres (name) VALUES ($1)", genre.Name)
	return err
}
