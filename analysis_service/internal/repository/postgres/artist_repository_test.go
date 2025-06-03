package postgres

import (
	"strings"
	"testing"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/artist"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/testutils"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestArtistRepo_Create_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)
	repo := NewArtistRepo(testutils.TestDB)

	a := &artist.Artist{Name: "Vlad"}

	err := repo.Create(a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	err = testutils.TestDB.QueryRow("SELECT COUNT(*) FROM artists WHERE name = $1", a.Name).Scan(&count)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}

func TestArtistRepo_Create_NameTooLong(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)
	repo := NewArtistRepo(testutils.TestDB)

	longName := "A" + strings.Repeat("verylongartistname", 10)

	a := &artist.Artist{Name: longName}

	err := repo.Create(a)

	if err == nil {
		t.Fatal("expected error due to name exceeding length constraint, got nil")
	}

	t.Logf("got expected error: %v", err)
}

func TestGetTopArtistsForUser_Repo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)

	testutils.TestDB.Exec(`INSERT INTO songs (song_id, name) VALUES (1, 'Track')`)
	testutils.TestDB.Exec(`INSERT INTO artists (artist_id, name) VALUES (1, 'Artist')`)
	testutils.TestDB.Exec(`INSERT INTO genres (genre_id, name) VALUES (1, 'Genre')`)
	_, err := testutils.TestDB.Exec(`
    INSERT INTO fact_listens (user_id, song_id, artist_id, genre_id)
    	VALUES ($1, $2, $3, $4)
			`, 1, 1, 1, 1)

	if err != nil {
		return
	}

	userID := 1
	limit := 5
	repo := NewArtistRepo(testutils.TestDB)

	artists, err := repo.GetTopArtistsForUser(userID, limit)
	require.NoError(t, err)
	require.NotNil(t, artists)
	require.LessOrEqual(t, len(artists), limit)

	for _, a := range artists {
		t.Logf("Artist ID: %d, Name: %s", a.ID, a.Name)
	}
}

func TestGetMostPopularArtists_Repo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)
	testutils.TestDB.Exec(`INSERT INTO songs (song_id, name) VALUES (1, 'Track')`)
	testutils.TestDB.Exec(`INSERT INTO artists (artist_id, name) VALUES (1, 'Artist')`)
	testutils.TestDB.Exec(`INSERT INTO genres (genre_id, name) VALUES (1, 'Genre')`)
	_, err := testutils.TestDB.Exec(`
    INSERT INTO fact_listens (user_id, song_id, artist_id, genre_id)
    	VALUES ($1, $2, $3, $4)
			`, 1, 1, 1, 1)
	if err != nil {
		return
	}

	repo := NewArtistRepo(testutils.TestDB)

	limit := 5
	artists, err := repo.GetMostPopularArtists(limit)
	require.NoError(t, err)
	require.NotNil(t, artists)
	require.LessOrEqual(t, len(artists), limit)

	for _, a := range artists {
		t.Logf("Artist ID: %d, Name: %s", a.ID, a.Name)
	}
}
