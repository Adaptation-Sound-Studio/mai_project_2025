package repository

import (
	"context"
	"database/sql"
	"fmt"
	"upload-service/internal/domain"
)

type ArtistRepository interface {
	GetAllArtists(ctx context.Context) ([]domain.Artist, error)
	GetArtistByID(ctx context.Context, id int64) (*domain.Artist, error)
	GetArtistByUserID(ctx context.Context, userID int64) (*domain.Artist, error)
	CreateArtist(ctx context.Context, artist *domain.Artist, userID int64) (int64, error)
	UpdateArtist(ctx context.Context, artist *domain.Artist) error
	GetSongsByArtistID(ctx context.Context, artistID int64) ([]domain.Song, error)
	GetAlbumsByArtistID(ctx context.Context, artistID int64) ([]domain.Album, error)
}

type artistRepository struct {
	db *sql.DB
}

func NewArtistRepository(db *sql.DB) ArtistRepository {
	return &artistRepository{db: db}
}

func (r *artistRepository) GetAllArtists(ctx context.Context) ([]domain.Artist, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT artist_id, name, user_id FROM artists")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artists []domain.Artist
	for rows.Next() {
		var a domain.Artist
		if err := rows.Scan(&a.ArtistID, &a.Name, &a.UserID); err != nil {
			return nil, err
		}
		artists = append(artists, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return artists, nil
}

func (r *artistRepository) GetArtistByID(ctx context.Context, id int64) (*domain.Artist, error) {
	row := r.db.QueryRowContext(ctx, "SELECT artist_id, name, user_id FROM artists WHERE artist_id = $1", id)
	var a domain.Artist
	err := row.Scan(&a.ArtistID, &a.Name, &a.UserID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *artistRepository) GetArtistByUserID(ctx context.Context, userID int64) (*domain.Artist, error) {
	row := r.db.QueryRowContext(ctx, "SELECT artist_id, name, user_id FROM artists WHERE user_id = $1", userID)
	var a domain.Artist
	err := row.Scan(&a.ArtistID, &a.Name, &a.UserID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *artistRepository) CreateArtist(ctx context.Context, artist *domain.Artist, userID int64) (int64, error) {
	var artistID int64
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO artists (name, user_id) VALUES ($1, $2) RETURNING artist_id",
		artist.Name, userID,
	).Scan(&artistID)
	if err != nil {
		return 0, err
	}
	return artistID, nil
}

func (r *artistRepository) UpdateArtist(ctx context.Context, artist *domain.Artist) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE artists SET name = $1 WHERE artist_id = $2",
		artist.Name, artist.ArtistID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("артист с ID %d не найден", artist.ArtistID)
	}

	return nil
}

func (r *artistRepository) GetSongsByArtistID(ctx context.Context, artistID int64) ([]domain.Song, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT s.song_id, s.name, s.auditions, s.genre_id, s.date, s.link
		 FROM songs s
		 JOIN song_artist sa ON s.song_id = sa.song_id
		 WHERE sa.artist_id = $1`, artistID)
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

func (r *artistRepository) GetAlbumsByArtistID(ctx context.Context, artistID int64) ([]domain.Album, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT album_id, name, auditions, artist_id, genre_id, date
		 FROM albums
		 WHERE artist_id = $1`, artistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums []domain.Album
	for rows.Next() {
		var a domain.Album
		if err := rows.Scan(&a.AlbumID, &a.Name, &a.Auditions, &a.ArtistID, &a.GenreID, &a.Date); err != nil {
			return nil, err
		}
		albums = append(albums, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return albums, nil
}
