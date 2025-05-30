package repository

import (
	"context"
	"fmt"
	"testing"
	"upload-service/internal/domain/model"
	"upload-service/internal/testutils"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAndGetAlbum(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)

	albumRepo := NewAlbumRepository(testDB)

	ctx := context.Background()

	var artistID, genreID int64

	err := testDB.QueryRow(`INSERT INTO artists (name, user_id) VALUES ($1, $2) RETURNING artist_id`, "Test Artist", 1234).Scan(&artistID)
	require.NoError(t, err)

	err = testDB.QueryRow(`INSERT INTO genres (name) VALUES ($1) RETURNING genre_id`, "Test Genre").Scan(&genreID)
	require.NoError(t, err)

	tx, err := testDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()

	album := &model.Album{
		Name:     "Integration Album",
		ArtistID: artistID,
		GenreID:  genreID,
	}

	id, err := albumRepo.CreateAlbum(ctx, tx, album)
	require.NoError(t, err)
	require.True(t, id > 0)

	require.NoError(t, tx.Commit())

	got, err := albumRepo.GetAlbumByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, album.Name, got.Name)
	require.Equal(t, album.ArtistID, got.ArtistID)
	require.Equal(t, album.GenreID, got.GenreID)
}

func TestUpdateAlbum(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewAlbumRepository(testDB)

	var artistID, genreID int64
	err := testDB.QueryRow(`INSERT INTO artists (name, user_id) VALUES ('Test Artist', 12345) RETURNING artist_id`).Scan(&artistID)
	require.NoError(t, err)

	err = testDB.QueryRow(`INSERT INTO genres (name) VALUES ('Test Genre') RETURNING genre_id`).Scan(&genreID)
	require.NoError(t, err)

	tx, err := testDB.BeginTx(ctx, nil)
	require.NoError(t, err)

	album := &model.Album{
		Name:     "Old Name",
		ArtistID: artistID,
		GenreID:  genreID,
	}
	id, err := repo.CreateAlbum(ctx, tx, album)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	err = repo.UpdateAlbum(ctx, &model.Album{
		AlbumID: id,
		Name:    "Updated Name",
		GenreID: genreID,
	})
	require.NoError(t, err)

	got, err := repo.GetAlbumByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "Updated Name", got.Name)
	require.Equal(t, genreID, got.GenreID)
}
func TestGetAllAlbums(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewAlbumRepository(testDB)

	var artistID, genreID int64
	err := testDB.QueryRow(`INSERT INTO artists (name, user_id) VALUES ('Test Artist', 1001) RETURNING artist_id`).Scan(&artistID)
	require.NoError(t, err)

	err = testDB.QueryRow(`INSERT INTO genres (name) VALUES ('Test Genre') RETURNING genre_id`).Scan(&genreID)
	require.NoError(t, err)

	tx, err := testDB.Begin()
	require.NoError(t, err)

	album := &model.Album{
		Name:     "Album One",
		ArtistID: artistID,
		GenreID:  genreID,
	}
	_, err = repo.CreateAlbum(ctx, tx, album)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	albums, err := repo.GetAllAlbums(ctx)
	require.NoError(t, err)
	require.Len(t, albums, 1)
	assert.Equal(t, "Album One", albums[0].Name)
	assert.Equal(t, artistID, albums[0].ArtistID)
	assert.Equal(t, genreID, albums[0].GenreID)
}

func TestGetSongsByAlbumID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewAlbumRepository(testDB)

	_, err := testDB.ExecContext(ctx, `INSERT INTO genres (genre_id, name) VALUES (1, 'Test Genre') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	_, err = testDB.ExecContext(ctx, `INSERT INTO artists (artist_id, name, user_id) VALUES (1, 'Test Artist', 100) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	var songID int64
	err = testDB.QueryRowContext(ctx, `
		INSERT INTO songs (name, auditions, genre_id, date, link)
		VALUES ('Test Song', 100, 1, CURRENT_DATE, 'http://example.com')
		RETURNING song_id`).Scan(&songID)
	require.NoError(t, err)

	tx, _ := testDB.BeginTx(ctx, nil)
	album := &model.Album{Name: "Test Album", ArtistID: 1, GenreID: 1}
	albumID, err := repo.CreateAlbum(ctx, tx, album)
	require.NoError(t, err)

	_, err = tx.ExecContext(ctx, `INSERT INTO song_album (song_id, album_id) VALUES ($1, $2)`, songID, albumID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	songs, err := repo.GetSongsByAlbumID(ctx, albumID)
	require.NoError(t, err)
	require.Len(t, songs, 1)
	assert.Equal(t, "Test Song", songs[0].Name)
	assert.Equal(t, songID, songs[0].SongID)
}

func TestCheckSongsExist(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewAlbumRepository(testDB)

	_, err := testDB.ExecContext(ctx, `INSERT INTO genres (genre_id, name) VALUES (1, 'Test Genre') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	tx, _ := testDB.BeginTx(ctx, nil)
	songIDs := make([]int64, 0)

	for i := 1; i <= 3; i++ {
		var id int64
		err := tx.QueryRowContext(ctx, `
			INSERT INTO songs (name, auditions, genre_id, date, link)
			VALUES ($1, 100, 1, CURRENT_DATE, $2)
			RETURNING song_id`, fmt.Sprintf("Song %d", i), fmt.Sprintf("http://example.com/%d", i),
		).Scan(&id)
		require.NoError(t, err)
		songIDs = append(songIDs, id)
	}
	require.NoError(t, tx.Commit())

	testIDs := append(songIDs, 999)

	tx2, _ := testDB.BeginTx(ctx, nil)
	found, err := repo.CheckSongsExist(ctx, tx2, testIDs)
	require.NoError(t, err)
	require.NoError(t, tx2.Commit())

	assert.ElementsMatch(t, songIDs, found)
}

func TestBatchInsertSongsToAlbum(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewAlbumRepository(testDB)

	tx1, _ := testDB.BeginTx(ctx, nil)

	var genreID int64
	err := tx1.QueryRowContext(ctx,
		"INSERT INTO genres (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING genre_id",
		"Test Genre",
	).Scan(&genreID)
	require.NoError(t, err)

	var artistID int64
	err = tx1.QueryRowContext(ctx,
		"INSERT INTO artists (name, user_id) VALUES ($1, $2) RETURNING artist_id",
		"Test Artist", 999,
	).Scan(&artistID)
	require.NoError(t, err)

	album := &model.Album{Name: "Batch Album", ArtistID: artistID, GenreID: 1}
	albumID, err := repo.CreateAlbum(ctx, tx1, album)
	require.NoError(t, err)

	songIDs := make([]int64, 0, 2)
	for i := 1; i <= 2; i++ {
		var id int64
		err := tx1.QueryRowContext(ctx,
			"INSERT INTO songs (name, auditions, genre_id, link) VALUES ($1, $2, $3, $4) RETURNING song_id",
			fmt.Sprintf("Song %d", i), 100*i, 1, fmt.Sprintf("http://link%d", i),
		).Scan(&id)
		require.NoError(t, err)
		songIDs = append(songIDs, id)
	}
	require.NoError(t, tx1.Commit())

	tx2, _ := testDB.BeginTx(ctx, nil)
	err = repo.BatchInsertSongsToAlbum(ctx, tx2, albumID, songIDs)
	require.NoError(t, err)
	require.NoError(t, tx2.Commit())

	rows, err := testDB.Query("SELECT song_id FROM song_album WHERE album_id = $1", albumID)
	require.NoError(t, err)
	defer rows.Close()

	var linkedIDs []int64
	for rows.Next() {
		var id int64
		require.NoError(t, rows.Scan(&id))
		linkedIDs = append(linkedIDs, id)
	}
	require.NoError(t, rows.Err())

	assert.ElementsMatch(t, songIDs, linkedIDs)
}
