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

func TestGetAllArtists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewArtistRepository(testDB)

	_, err := testDB.ExecContext(ctx, `
		INSERT INTO artists (name, user_id) VALUES
		('Artist One', 1001),
		('Artist Two', 1002)
	`)
	require.NoError(t, err)

	artists, err := repo.GetAllArtists(ctx)
	require.NoError(t, err)
	require.Len(t, artists, 2)

	names := []string{artists[0].Name, artists[1].Name}
	assert.ElementsMatch(t, []string{"Artist One", "Artist Two"}, names)
}

func TestCreateAndGetArtistByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewArtistRepository(testDB)

	artist := &model.Artist{Name: "Test Artist"}
	userID := int64(1001)

	artistID, err := repo.CreateArtist(ctx, artist, userID)
	require.NoError(t, err)
	require.NotZero(t, artistID)

	got, err := repo.GetArtistByID(ctx, artistID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Test Artist", got.Name)
	assert.Equal(t, userID, got.UserID)
}

func TestCreateArtist_DuplicateUserID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewArtistRepository(testDB)

	artist1 := &model.Artist{Name: "Artist A"}
	id1, err := repo.CreateArtist(ctx, artist1, 123)
	require.NoError(t, err)
	require.NotZero(t, id1)

	artist2 := &model.Artist{Name: "Artist B"}
	_, err = repo.CreateArtist(ctx, artist2, 123)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key")
}

func TestGetArtistByID_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	repo := NewArtistRepository(testDB)

	artist, err := repo.GetArtistByID(context.Background(), 999)
	require.NoError(t, err)
	require.Nil(t, artist)
}

func TestGetArtistByUserID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewArtistRepository(testDB)

	artist := &model.Artist{Name: "Another Artist"}
	userID := int64(2002)

	artistID, err := repo.CreateArtist(ctx, artist, userID)
	require.NoError(t, err)
	require.NotZero(t, artistID)

	got, err := repo.GetArtistByUserID(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, artistID, got.ArtistID)
	assert.Equal(t, artist.Name, got.Name)
	assert.Equal(t, userID, got.UserID)
}

func TestGetArtistByUserID_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	repo := NewArtistRepository(testDB)

	artist, err := repo.GetArtistByUserID(context.Background(), 999)
	require.NoError(t, err)
	require.Nil(t, artist)
}

func TestUpdateArtist(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewArtistRepository(testDB)

	artist := &model.Artist{Name: "Initial Name"}
	userID := int64(3003)
	artistID, err := repo.CreateArtist(ctx, artist, userID)
	require.NoError(t, err)

	updated := &model.Artist{
		ArtistID: artistID,
		Name:     "Updated Name",
	}

	err = repo.UpdateArtist(ctx, updated)
	require.NoError(t, err)

	got, err := repo.GetArtistByID(ctx, artistID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Updated Name", got.Name)
	assert.Equal(t, userID, got.UserID)
}

func TestUpdateArtist_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewArtistRepository(testDB)

	err := repo.UpdateArtist(ctx, &model.Artist{
		ArtistID: 99999,
		Name:     "Ghost",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "артист с ID 99999 не найден")
}

func TestUpdateArtist_DBError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	badDB, err := sql.Open("postgres", "host=localhost port=5432 user=invalid password=invalid dbname=invalid sslmode=disable")
	require.NoError(t, err)
	defer badDB.Close()

	repo := NewArtistRepository(badDB)

	err = repo.UpdateArtist(context.Background(), &model.Artist{
		ArtistID: 1,
		Name:     "FailUpdate",
	})
	require.Error(t, err)
}

func TestGetSongsByArtistID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewArtistRepository(testDB)

	var genreID int64
	err := testDB.QueryRowContext(ctx,
		`INSERT INTO genres (name) VALUES ('Genre_01') RETURNING genre_id`,
	).Scan(&genreID)
	require.NoError(t, err)

	artist := &model.Artist{Name: "Artist_01"}
	userID := int64(4004)
	artistID, err := repo.CreateArtist(ctx, artist, userID)
	require.NoError(t, err)

	songIDs := make([]int64, 0, 2)
	for i := 1; i <= 2; i++ {
		var id int64
		err := testDB.QueryRowContext(ctx,
			`INSERT INTO songs (name, auditions, genre_id, name_on_minio) VALUES ($1, $2, $3, $4) RETURNING song_id`,
			fmt.Sprintf("Song%d", i), 100*i, genreID, fmt.Sprintf("http://song%d", i),
		).Scan(&id)
		require.NoError(t, err)
		songIDs = append(songIDs, id)

		_, err = testDB.ExecContext(ctx,
			`INSERT INTO song_artist (song_id, artist_id) VALUES ($1, $2)`, id, artistID)
		require.NoError(t, err)
	}

	songs, err := repo.GetSongsByArtistID(ctx, artistID)
	require.NoError(t, err)
	require.Len(t, songs, len(songIDs))

	names := make([]string, 0, len(songs))
	for _, s := range songs {
		names = append(names, s.Name)
	}

	assert.ElementsMatch(t, []string{"Song1", "Song2"}, names)
}

func TestGetSongsByArtistID_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewArtistRepository(testDB)

	artist := &model.Artist{Name: "Solo"}
	artistID, err := repo.CreateArtist(ctx, artist, 111)
	require.NoError(t, err)

	songs, err := repo.GetSongsByArtistID(ctx, artistID)
	require.NoError(t, err)
	assert.Empty(t, songs)
}

func TestGetAlbumsByArtistID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewArtistRepository(testDB)

	var genreID int64
	err := testDB.QueryRowContext(ctx,
		`INSERT INTO genres (name) VALUES ('Genre_1') RETURNING genre_id`,
	).Scan(&genreID)
	require.NoError(t, err)

	artist := &model.Artist{Name: "Artist_1"}
	userID := int64(5005)
	artistID, err := repo.CreateArtist(ctx, artist, userID)
	require.NoError(t, err)

	for i := 1; i <= 2; i++ {
		_, err := testDB.ExecContext(ctx,
			`INSERT INTO albums (name, artist_id, genre_id) VALUES ($1, $2, $3)`,
			fmt.Sprintf("Album%d", i), artistID, genreID,
		)
		require.NoError(t, err)
	}

	albums, err := repo.GetAlbumsByArtistID(ctx, artistID)
	require.NoError(t, err)
	require.Len(t, albums, 2)

	names := make([]string, 0, len(albums))
	for _, a := range albums {
		names = append(names, a.Name)
	}

	assert.ElementsMatch(t, []string{"Album1", "Album2"}, names)
}

func TestGetAlbumsByArtistID_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewArtistRepository(testDB)

	artist := &model.Artist{Name: "Albumless"}
	artistID, err := repo.CreateArtist(ctx, artist, 112)
	require.NoError(t, err)

	albums, err := repo.GetAlbumsByArtistID(ctx, artistID)
	require.NoError(t, err)
	assert.Empty(t, albums)
}
