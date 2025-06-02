package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"upload-service/internal/domain/album"
	"upload-service/internal/domain/model"

	"github.com/lib/pq"
)

type albumRepository struct {
	db *sql.DB
}

func NewAlbumRepository(db *sql.DB) album.Repository {
	return &albumRepository{db: db}
}

func (r *albumRepository) GetAllAlbums(ctx context.Context) ([]model.Album, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT album_id, name, auditions, artist_id, genre_id, date FROM albums")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums []model.Album
	for rows.Next() {
		var a model.Album
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

func (r *albumRepository) GetAlbumByID(ctx context.Context, id int64) (*model.Album, error) {
	row := r.db.QueryRowContext(ctx, "SELECT album_id, name, auditions, artist_id, genre_id, date FROM albums WHERE album_id = $1", id)
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

func (r *albumRepository) CreateAlbum(ctx context.Context, tx *sql.Tx, album *model.Album) (int64, error) {
	var albumID int64
	err := tx.QueryRowContext(ctx,
		"INSERT INTO albums (name, artist_id, genre_id) VALUES ($1, $2, $3) RETURNING album_id",
		album.Name, album.ArtistID, album.GenreID,
	).Scan(&albumID)
	if err != nil {
		return 0, err
	}
	return albumID, nil
}

func (r *albumRepository) UpdateAlbum(ctx context.Context, album *model.Album) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE albums SET name = $1, genre_id = $2 WHERE album_id = $3",
		album.Name, album.GenreID, album.AlbumID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("альбом с ID %d не найден", album.AlbumID)
	}

	return nil
}

func (r *albumRepository) GetSongsByAlbumID(ctx context.Context, albumID int64) ([]model.Song, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT s.song_id, s.name, s.auditions, s.genre_id, s.date
	 	FROM songs s
	 	JOIN song_album sa ON s.song_id = sa.song_id
	 	WHERE sa.album_id = $1`, albumID)
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

func (r *albumRepository) CheckSongsExist(ctx context.Context, tx *sql.Tx, songIDs []int64) ([]int64, error) {
	query := "SELECT song_id FROM songs WHERE song_id = ANY($1)"
	rows, err := tx.QueryContext(ctx, query, pq.Array(songIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var existingIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		existingIDs = append(existingIDs, id)
	}
	return existingIDs, rows.Err()
}

func (r *albumRepository) BatchInsertSongsToAlbum(ctx context.Context, tx *sql.Tx, albumID int64, songIDs []int64) error {
	if len(songIDs) == 0 {
		return nil
	}

	if len(songIDs) > 50 {
		return fmt.Errorf("нельзя добавить более 50 песен в альбом")
	}

	var placeholders []string
	var args []interface{}
	for i, songID := range songIDs {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		args = append(args, songID, albumID)
	}

	query := fmt.Sprintf("INSERT INTO song_album (song_id, album_id) VALUES %s", strings.Join(placeholders, ", "))
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *albumRepository) GetGenreByAlbumID(ctx context.Context, albumID int64) (*model.Genre, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT g.genre_id, g.name
		 FROM genres g
		 JOIN albums a ON g.genre_id = a.genre_id
		 WHERE a.album_id = $1`, albumID)

	var genre model.Genre
	err := row.Scan(&genre.GenreID, &genre.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &genre, nil
}

func (r *albumRepository) GetArtistByAlbumID(ctx context.Context, albumID int64) (*model.Artist, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT ar.artist_id, ar.name
		 FROM artists ar
		 JOIN albums al ON ar.artist_id = al.artist_id
		 WHERE al.album_id = $1`, albumID)

	var artist model.Artist
	err := row.Scan(&artist.ArtistID, &artist.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &artist, nil
}
