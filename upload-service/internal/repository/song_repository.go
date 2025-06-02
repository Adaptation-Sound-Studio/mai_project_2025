package repository

import (
	"context"
	"database/sql"
	"fmt"
	"upload-service/internal/domain/model"
	"upload-service/internal/domain/song"

	"github.com/lib/pq"
)

type songRepository struct {
	db *sql.DB
}

func NewSongRepository(db *sql.DB) song.Repository {
	return &songRepository{db: db}
}

func (r *songRepository) GetAllSongs(ctx context.Context) ([]model.Song, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT song_id, name, auditions, genre_id, date FROM songs ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []model.Song
	for rows.Next() {
		var s model.Song
		if err := rows.Scan(&s.SongID, &s.Name, &s.Auditions, &s.GenreID, &s.Date); err != nil {
			return nil, err
		}
		songs = append(songs, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return songs, nil
}

func (r *songRepository) GetSongByID(ctx context.Context, id int64) (*model.Song, error) {
	row := r.db.QueryRowContext(ctx, "SELECT song_id, name, name_on_minio, auditions, genre_id, date FROM songs WHERE song_id = $1", id)
	var s model.Song
	err := row.Scan(&s.SongID, &s.Name, &s.NameOfMinio, &s.Auditions, &s.GenreID, &s.Date)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *songRepository) CreateSongWithArtists(ctx context.Context, song *model.Song, artistIDs []int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	var songID int64
	err = tx.QueryRowContext(ctx,
		"INSERT INTO songs (name, genre_id, name_on_minio) VALUES ($1, $2, $3) RETURNING song_id",
		song.Name, song.GenreID, song.NameOfMinio,
	).Scan(&songID)

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if len(artistIDs) > 0 {
		query := "INSERT INTO song_artist (song_id, artist_id) VALUES "
		args := []interface{}{}
		argPos := 1

		for i, artistID := range artistIDs {
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf("($%d, $%d)", argPos, argPos+1)
			args = append(args, songID, artistID)
			argPos += 2
		}

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return songID, nil
}

func (r *songRepository) UpdateSong(ctx context.Context, song *model.Song) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE songs SET name = $1, genre_id = $2 WHERE song_id = $3",
		song.Name, song.GenreID, song.SongID,
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

func (r *songRepository) CheckArtistsExist(ctx context.Context, artistIDs []int64) ([]int64, error) {
	query := "SELECT artist_id FROM artists WHERE artist_id = ANY($1)"
	rows, err := r.db.QueryContext(ctx, query, pq.Array(artistIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var existing []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		existing = append(existing, id)
	}

	return existing, nil
}

func (r *songRepository) GetOneArtistBySongID(ctx context.Context, songID int64) (*model.Artist, error) {
	var artist model.Artist
	err := r.db.QueryRowContext(ctx, `
        SELECT a.artist_id, a.name, a.user_id
        FROM artists a
        JOIN song_artist sa ON a.artist_id = sa.artist_id
        WHERE sa.song_id = $1
        LIMIT 1
    `, songID).Scan(&artist.ArtistID, &artist.Name, &artist.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &artist, nil
}

func (r *songRepository) GetGenreBySongID(ctx context.Context, songID int64) (*model.Genre, error) {
	var genre model.Genre
	err := r.db.QueryRowContext(ctx, `
        SELECT g.genre_id, g.name
        FROM genres g
        JOIN songs s ON g.genre_id = s.genre_id
        WHERE s.song_id = $1
    `, songID).Scan(&genre.GenreID, &genre.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &genre, nil
}

func (r *songRepository) GetArtistsBySongID(ctx context.Context, songID int64) ([]model.Artist, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT a.artist_id, a.name, a.user_id 
		 FROM artists a
		 JOIN song_artist sa ON a.artist_id = sa.artist_id
		 WHERE sa.song_id = $1`, songID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artists []model.Artist
	for rows.Next() {
		var a model.Artist
		if err := rows.Scan(&a.ArtistID, &a.Name, &a.UserID); err != nil {
			return nil, err
		}
		artists = append(artists, a)
	}

	return artists, nil
}

func (r *songRepository) GetAlbumBySongID(ctx context.Context, songID int64) (*model.Album, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT al.album_id, al.name, al.auditions, al.artist_id, al.genre_id, al.date
		 FROM albums al
		 JOIN song_album sa ON al.album_id = sa.album_id
		 WHERE sa.song_id = $1`, songID)

	var a model.Album
	err := row.Scan(&a.AlbumID, &a.Name, &a.Auditions, &a.ArtistID, &a.GenreID, &a.Date)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}
