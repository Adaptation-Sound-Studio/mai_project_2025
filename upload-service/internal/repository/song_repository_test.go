package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"upload-service/internal/domain/model"
	"upload-service/internal/testutils"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAllSongs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewSongRepository(testDB)

	var genreID int64
	err := testDB.QueryRowContext(ctx,
		"INSERT INTO genres (name) VALUES ($1) RETURNING genre_id", "Rock",
	).Scan(&genreID)
	require.NoError(t, err)

	_, err = testDB.ExecContext(ctx,
		`INSERT INTO songs (name, auditions, genre_id, date, link) VALUES
		('Song One', 100, $1, '2024-01-01', 'http://one'),
		('Song Two', 200, $1, '2025-01-01', 'http://two')`, genreID,
	)
	require.NoError(t, err)

	songs, err := repo.GetAllSongs(ctx)
	require.NoError(t, err)
	require.Len(t, songs, 2)
	assert.Equal(t, "Song Two", songs[0].Name)
	assert.Equal(t, "Song One", songs[1].Name)
}

func TestGetSongByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewSongRepository(testDB)

	var genreID int64
	err := testDB.QueryRowContext(ctx,
		"INSERT INTO genres (name) VALUES ($1) RETURNING genre_id", "Jazz",
	).Scan(&genreID)
	require.NoError(t, err)

	var songID int64
	err = testDB.QueryRowContext(ctx,
		"INSERT INTO songs (name, auditions, genre_id, link) VALUES ($1, $2, $3, $4) RETURNING song_id",
		"My Song", 1, genreID, "http://song-link",
	).Scan(&songID)
	require.NoError(t, err)

	song, err := repo.GetSongByID(ctx, songID)
	require.NoError(t, err)
	require.NotNil(t, song)
	assert.Equal(t, "My Song", song.Name)
	assert.Equal(t, "http://song-link", song.Link)
}

func TestCreateSongWithArtists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewSongRepository(testDB)

	tx, _ := testDB.BeginTx(ctx, nil)
	var genreID int64
	err := tx.QueryRowContext(ctx,
		"INSERT INTO genres (name) VALUES ($1) ON CONFLICT DO NOTHING RETURNING genre_id",
		"Test Genre",
	).Scan(&genreID)
	if err == sql.ErrNoRows {

		_ = tx.QueryRow("SELECT genre_id FROM genres WHERE name = $1", "Test Genre").Scan(&genreID)
	}
	require.NoError(t, err)

	artistIDs := make([]int64, 0, 2)
	for i := 1; i <= 2; i++ {
		var id int64
		err := tx.QueryRowContext(ctx,
			"INSERT INTO artists (name, user_id) VALUES ($1, $2) RETURNING artist_id",
			fmt.Sprintf("Artist %d", i), 100+i,
		).Scan(&id)
		require.NoError(t, err)
		artistIDs = append(artistIDs, id)
	}
	require.NoError(t, tx.Commit())

	song := &model.Song{
		Name:    "My Song",
		GenreID: genreID,
		Link:    "https://link-to-song.com",
	}

	songID, err := repo.CreateSongWithArtists(ctx, song, artistIDs)
	require.NoError(t, err)
	require.True(t, songID > 0)

	stored, err := repo.GetSongByID(ctx, songID)
	require.NoError(t, err)
	assert.Equal(t, "My Song", stored.Name)

	artists, err := repo.GetArtistsBySongID(ctx, songID)
	require.NoError(t, err)
	assert.Len(t, artists, 2)
}

func TestUpdateSong(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewSongRepository(testDB)

	var genreID int64
	err := testDB.QueryRowContext(ctx,
		"INSERT INTO genres (name) VALUES ($1) RETURNING genre_id",
		"Test Genre",
	).Scan(&genreID)
	require.NoError(t, err)

	var songID int64
	err = testDB.QueryRowContext(ctx,
		"INSERT INTO songs (name, genre_id, link) VALUES ($1, $2, $3) RETURNING song_id",
		"Old Name", genreID, "http://old-link.com",
	).Scan(&songID)
	require.NoError(t, err)

	err = repo.UpdateSong(ctx, &model.Song{
		SongID:  songID,
		Name:    "New Name",
		GenreID: genreID,
		Link:    "http://new-link.com",
	})
	require.NoError(t, err)

	updated, err := repo.GetSongByID(ctx, songID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, "http://new-link.com", updated.Link)
}

func TestUpdateSong_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewSongRepository(testDB)

	err := repo.UpdateSong(ctx, &model.Song{
		SongID:  99999,
		Name:    "Ghost Song",
		GenreID: 1,
		Link:    "http://ghost",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "песня с ID 99999 не найдена")
}

func TestCheckArtistsExist(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewSongRepository(testDB)

	ids := []int64{}
	for i := 1; i <= 2; i++ {
		var id int64
		err := testDB.QueryRowContext(ctx,
			"INSERT INTO artists (name, user_id) VALUES ($1, $2) RETURNING artist_id",
			fmt.Sprintf("Artist %d", i), 100+i,
		).Scan(&id)
		require.NoError(t, err)
		ids = append(ids, id)
	}

	testIDs := append(ids, 9999)

	found, err := repo.CheckArtistsExist(ctx, testIDs)
	require.NoError(t, err)

	assert.ElementsMatch(t, ids, found)
}

func TestGetArtistsBySongID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewSongRepository(testDB)

	var genreID int64
	err := testDB.QueryRowContext(ctx,
		"INSERT INTO genres (name) VALUES ($1) RETURNING genre_id", "Genre for Song",
	).Scan(&genreID)
	require.NoError(t, err)

	var artistIDs []int64
	for i := 1; i <= 2; i++ {
		var id int64
		err := testDB.QueryRowContext(ctx,
			"INSERT INTO artists (name, user_id) VALUES ($1, $2) RETURNING artist_id",
			fmt.Sprintf("Artist %d", i), 200+i,
		).Scan(&id)
		require.NoError(t, err)
		artistIDs = append(artistIDs, id)
	}

	var songID int64
	err = testDB.QueryRowContext(ctx,
		"INSERT INTO songs (name, genre_id, link) VALUES ($1, $2, $3) RETURNING song_id",
		"Test Song", genreID, "http://link-to-song",
	).Scan(&songID)
	require.NoError(t, err)

	for _, artistID := range artistIDs {
		_, err := testDB.ExecContext(ctx,
			"INSERT INTO song_artist (song_id, artist_id) VALUES ($1, $2)",
			songID, artistID,
		)
		require.NoError(t, err)
	}

	artists, err := repo.GetArtistsBySongID(ctx, songID)
	require.NoError(t, err)
	require.Len(t, artists, 2)

	returnedIDs := []int64{artists[0].ArtistID, artists[1].ArtistID}
	assert.ElementsMatch(t, artistIDs, returnedIDs)
}

func TestGetAlbumBySongID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewSongRepository(testDB)

	var genreID int64
	err := testDB.QueryRowContext(ctx,
		"INSERT INTO genres (name) VALUES ($1) RETURNING genre_id", "Genre X",
	).Scan(&genreID)
	require.NoError(t, err)

	var artistID int64
	err = testDB.QueryRowContext(ctx,
		"INSERT INTO artists (name, user_id) VALUES ($1, $2) RETURNING artist_id", "Artist X", 700,
	).Scan(&artistID)
	require.NoError(t, err)

	var albumID int64
	err = testDB.QueryRowContext(ctx,
		"INSERT INTO albums (name, artist_id, genre_id) VALUES ($1, $2, $3) RETURNING album_id",
		"Album X", artistID, genreID,
	).Scan(&albumID)
	require.NoError(t, err)

	var songID int64
	err = testDB.QueryRowContext(ctx,
		"INSERT INTO songs (name, genre_id, link) VALUES ($1, $2, $3) RETURNING song_id",
		"Song X", genreID, "http://link-x",
	).Scan(&songID)
	require.NoError(t, err)

	_, err = testDB.ExecContext(ctx,
		"INSERT INTO song_album (song_id, album_id) VALUES ($1, $2)",
		songID, albumID,
	)
	require.NoError(t, err)

	album, err := repo.GetAlbumBySongID(ctx, songID)
	require.NoError(t, err)
	require.NotNil(t, album)
	assert.Equal(t, albumID, album.AlbumID)
	assert.Equal(t, "Album X", album.Name)
}

func TestBatchInsertSongsToAlbum_TooManySongs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewAlbumRepository(testDB)

	var genreID, artistID int64
	err := testDB.QueryRow("INSERT INTO genres (name) VALUES ('Test Genre') RETURNING genre_id").Scan(&genreID)
	require.NoError(t, err)
	err = testDB.QueryRow("INSERT INTO artists (name, user_id) VALUES ('Test Artist', 999) RETURNING artist_id").Scan(&artistID)
	require.NoError(t, err)

	tx1, _ := testDB.BeginTx(ctx, nil)
	album := &model.Album{Name: "Test Album", ArtistID: artistID, GenreID: genreID}
	albumID, err := repo.CreateAlbum(ctx, tx1, album)
	require.NoError(t, err)

	var songIDs []int64
	for i := 0; i < 51; i++ {
		var songID int64
		err := tx1.QueryRowContext(ctx,
			"INSERT INTO songs (name, genre_id, link) VALUES ($1, $2, $3) RETURNING song_id",
			fmt.Sprintf("Song %d", i+1), genreID, fmt.Sprintf("http://link/%d", i+1),
		).Scan(&songID)
		require.NoError(t, err)
		songIDs = append(songIDs, songID)
	}
	require.NoError(t, tx1.Commit())

	tx2, _ := testDB.BeginTx(ctx, nil)
	err = repo.BatchInsertSongsToAlbum(ctx, tx2, albumID, songIDs)
	require.Error(t, err)
	assert.Equal(t, "нельзя добавить более 50 песен в альбом", err.Error())
	_ = tx2.Rollback()
}
