package postgres

import (
	"database/sql"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/artist"
)

type ArtistRepo struct {
	DB *sql.DB
}

func NewArtistRepo(db *sql.DB) *ArtistRepo {
	return &ArtistRepo{DB: db}
}

func (r *ArtistRepo) Create(artist *artist.Artist) error {
	_, err := r.DB.Exec("INSERT INTO artists (name) VALUES ($1)", artist.Name)
	return err
}

func (r *ArtistRepo) GetTopArtistsForUser(userID int, limit int) ([]artist.Artist, error) {
	rows, err := r.DB.Query(`
		SELECT a.artist_id, a.name, COUNT(*) as listen_count
		FROM fact_listens fl
		JOIN artists a ON fl.artist_id = a.artist_id
		WHERE fl.user_id = $1
		GROUP BY a.artist_id, a.name
		ORDER BY listen_count DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artists []artist.Artist
	for rows.Next() {
		var art artist.Artist
		var count int
		if err := rows.Scan(&art.ID, &art.Name, &count); err != nil {
			return nil, err
		}
		artists = append(artists, art)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return artists, nil
}

func (r *ArtistRepo) GetMostPopularArtists(limit int) ([]artist.Artist, error) {
	rows, err := r.DB.Query(`
		SELECT a.artist_id, a.name, COUNT(*) AS listen_count
		FROM fact_listens fl
		JOIN artists a ON fl.artist_id = a.artist_id
		GROUP BY a.artist_id, a.name
		ORDER BY listen_count DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artists []artist.Artist
	for rows.Next() {
		var art artist.Artist
		var count int
		if err := rows.Scan(&art.ID, &art.Name, &count); err != nil {
			return nil, err
		}
		artists = append(artists, art)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return artists, nil
}
