package postgres

import (
	"database/sql"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/fact"
)

type FactRepo struct {
	DB *sql.DB
}

func NewFactRepo(db *sql.DB) *FactRepo {
	return &FactRepo{DB: db}
}

func (r *FactRepo) Insert(f *fact.ListenFact) error {
	_, err := r.DB.Exec(`
		INSERT INTO fact_listens (user_id, song_id, artist_id, genre_id, listened_at)
		VALUES ($1, $2, $3, $4, $5)
	`, f.UserID, f.SongID, f.ArtistID, f.GenreID, f.ListenedAt)
	return err
}
