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

func (r *GenreRepo) GetTopGenresForUser(userID int, limit int) ([]genre.Genre, error) {
	rows, err := r.DB.Query(`
		SELECT g.genre_id, g.name, COUNT(*) AS listen_count
		FROM fact_listens fl
		JOIN genres g ON fl.genre_id = g.genre_id
		WHERE fl.user_id = $1
		GROUP BY g.genre_id, g.name
		ORDER BY listen_count DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []genre.Genre
	for rows.Next() {
		var g genre.Genre
		var count int
		if err := rows.Scan(&g.ID, &g.Name, &count); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return genres, nil
}

func (r *GenreRepo) GetMostPopularGenres(limit int) ([]genre.Genre, error) {
	rows, err := r.DB.Query(`
		SELECT g.genre_id, g.name, COUNT(*) AS listen_count
		FROM fact_listens fl
		JOIN genres g ON fl.genre_id = g.genre_id
		GROUP BY g.genre_id, g.name
		ORDER BY listen_count DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []genre.Genre
	for rows.Next() {
		var g genre.Genre
		var count int
		if err := rows.Scan(&g.ID, &g.Name, &count); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return genres, nil
}

func (r *GenreRepo) GetTopGenresAtNight(limit int) ([]genre.Genre, error) {
	rows, err := r.DB.Query(`
		SELECT g.genre_id, g.name, COUNT(*) AS listen_count
		FROM fact_listens fl
		JOIN genres g ON fl.genre_id = g.genre_id
		WHERE EXTRACT(HOUR FROM fl.listened_at) BETWEEN 0 AND 5
		GROUP BY g.genre_id, g.name
		ORDER BY listen_count DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []genre.Genre
	for rows.Next() {
		var g genre.Genre
		var count int
		if err := rows.Scan(&g.ID, &g.Name, &count); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return genres, nil
}

func (r *GenreRepo) GetTopGenresInMorning(limit int) ([]genre.Genre, error) {
	rows, err := r.DB.Query(`
		SELECT g.genre_id, g.name, COUNT(*) AS listen_count
		FROM fact_listens fl
		JOIN genres g ON fl.genre_id = g.genre_id
		WHERE EXTRACT(HOUR FROM fl.listened_at) BETWEEN 6 AND 11
		GROUP BY g.genre_id, g.name
		ORDER BY listen_count DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []genre.Genre
	for rows.Next() {
		var g genre.Genre
		var count int
		if err := rows.Scan(&g.ID, &g.Name, &count); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return genres, nil
}

func (r *GenreRepo) GetTopGenresInDay(limit int) ([]genre.Genre, error) {
	rows, err := r.DB.Query(`
		SELECT g.genre_id, g.name, COUNT(*) AS listen_count
		FROM fact_listens fl
		JOIN genres g ON fl.genre_id = g.genre_id
		WHERE EXTRACT(HOUR FROM fl.listened_at) BETWEEN 12 AND 17
		GROUP BY g.genre_id, g.name
		ORDER BY listen_count DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []genre.Genre
	for rows.Next() {
		var g genre.Genre
		var count int
		if err := rows.Scan(&g.ID, &g.Name, &count); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return genres, nil
}

func (r *GenreRepo) GetTopGenresInEvening(limit int) ([]genre.Genre, error) {
	rows, err := r.DB.Query(`
		SELECT g.genre_id, g.name, COUNT(*) AS listen_count
		FROM fact_listens fl
		JOIN genres g ON fl.genre_id = g.genre_id
		WHERE EXTRACT(HOUR FROM fl.listened_at) BETWEEN 18 AND 23
		GROUP BY g.genre_id, g.name
		ORDER BY listen_count DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []genre.Genre
	for rows.Next() {
		var g genre.Genre
		var count int
		if err := rows.Scan(&g.ID, &g.Name, &count); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return genres, nil
}
