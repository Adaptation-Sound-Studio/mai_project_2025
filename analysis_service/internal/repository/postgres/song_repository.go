package postgres

import (
	"database/sql"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/song"
)

type SongRepo struct {
	DB *sql.DB
}

func NewSongRepo(db *sql.DB) *SongRepo {
	return &SongRepo{DB: db}
}

func (r *SongRepo) Create(song *song.Song) error {
	_, err := r.DB.Exec("INSERT INTO songs (name) VALUES ($1)", song.Name)
	return err
}

func (r *SongRepo) GetPopularSongs(limit int) ([]song.PopularSong, error) {
	query := `
SELECT 
    s.song_id,
    s.name AS song_name,
    a.name AS artist_name,
    g.name AS genre_name,
    COUNT(fl.song_id) AS listens
FROM fact_listens fl
JOIN songs s ON fl.song_id = s.song_id
JOIN artists a ON fl.artist_id = a.artist_id
JOIN genres g ON fl.genre_id = g.genre_id
GROUP BY s.song_id, s.name, a.name, g.name
ORDER BY listens DESC
LIMIT $1;

`
	rows, err := r.DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []song.PopularSong
	for rows.Next() {
		var ps song.PopularSong
		err := rows.Scan(&ps.ID, &ps.Name, &ps.Artist, &ps.Genre, &ps.Listens)
		if err != nil {
			return nil, err
		}
		result = append(result, ps)
	}

	return result, nil
}
