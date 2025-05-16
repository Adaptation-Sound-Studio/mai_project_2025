package repository

import (
	"context"
	"database/sql"
	"fmt"
	"upload-service/internal/domain"
)

type SongRepository interface {
	GetAllSongs(ctx context.Context) ([]domain.Song, error)
	GetSongByID(ctx context.Context, id int64) (*domain.Song, error)
	CreateSong(ctx context.Context, song *domain.Song) error
	UpdateSong(ctx context.Context, song *domain.Song) error
}

type songRepository struct {
	db *sql.DB
}

func NewSongRepository(db *sql.DB) SongRepository {
	return &songRepository{db: db}
}

func (r *songRepository) GetAllSongs(ctx context.Context) ([]domain.Song, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT song_id, name, auditions, genre_id, date, link FROM songs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []domain.Song
	for rows.Next() {
		var s domain.Song
		if err := rows.Scan(&s.SongID, &s.Name, &s.Auditions, &s.GenreID, &s.Date, &s.Link); err != nil {
			return nil, err
		}
		songs = append(songs, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return songs, nil
}

func (r *songRepository) GetSongByID(ctx context.Context, id int64) (*domain.Song, error) {
	row := r.db.QueryRowContext(ctx, "SELECT song_id, name, auditions, genre_id, date, link FROM songs WHERE song_id = $1", id)
	var s domain.Song
	err := row.Scan(&s.SongID, &s.Name, &s.Auditions, &s.GenreID, &s.Date, &s.Link)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *songRepository) CreateSong(ctx context.Context, song *domain.Song) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO songs (name, auditions, genre_id, date, link) VALUES ($1, $2, $3, $4, $5)",
		song.Name, song.Auditions, song.GenreID, song.Date, song.Link,
	)
	return err
}

func (r *songRepository) UpdateSong(ctx context.Context, song *domain.Song) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE songs SET name = $1, auditions = $2, genre_id = $3, date = $4, link = $5 WHERE song_id = $6",
		song.Name, song.Auditions, song.GenreID, song.Date, song.Link, song.SongID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("песня с ID %d не найдена", song.SongID)
	}

	return nil
}
