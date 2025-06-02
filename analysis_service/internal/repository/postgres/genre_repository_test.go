package postgres

import (
	"testing"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/genre"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestGenretRepo_Create_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)
	repo := NewGenreRepo(testutils.TestDB)

	g := &genre.Genre{Name: "Pop"}

	err := repo.Create(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	err = testutils.TestDB.QueryRow("SELECT COUNT(*) FROM genres WHERE name = $1", g.Name).Scan(&count)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}

func TestGenreRepo_Create_DuplicateName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)
	repo := NewGenreRepo(testutils.TestDB)

	first := &genre.Genre{Name: "Pop"}
	err := repo.Create(first)
	if err != nil {
		t.Fatalf("unexpected error on first insert: %v", err)
	}

	duplicate := &genre.Genre{Name: "Pop"}
	err = repo.Create(duplicate)
	if err == nil {
		t.Fatal("expected unique constraint error, got nil")
	}

	t.Logf("got expected error on duplicate insert: %v", err)
}

func TestGetTopGenresForUser(t *testing.T) {
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
	repo := NewGenreRepo(testutils.TestDB)

	userID := 1
	limit := 5

	genres, err := repo.GetTopGenresForUser(userID, limit)
	require.NoError(t, err)
	require.NotNil(t, genres)
	require.LessOrEqual(t, len(genres), limit)

	for _, g := range genres {
		t.Logf("Genre ID: %d, Name: %s", g.ID, g.Name)
	}
}

func TestGetMostPopularGenres(t *testing.T) {
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
	repo := NewGenreRepo(testutils.TestDB)

	limit := 5

	genres, err := repo.GetMostPopularGenres(limit)
	require.NoError(t, err)
	require.NotNil(t, genres)
	require.LessOrEqual(t, len(genres), limit)

	for _, g := range genres {
		t.Logf("Genre ID: %d, Name: %s", g.ID, g.Name)
	}
}

func TestGetTopGenresAtNight(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)
	testutils.TestDB.Exec(`INSERT INTO songs (song_id, name) VALUES (1, 'Track')`)
	testutils.TestDB.Exec(`INSERT INTO artists (artist_id, name) VALUES (1, 'Artist')`)
	testutils.TestDB.Exec(`INSERT INTO genres (genre_id, name) VALUES (1, 'Genre')`)
	_, err := testutils.TestDB.Exec(`
    INSERT INTO fact_listens (user_id, song_id, artist_id, genre_id, listende_at)
    	VALUES ($1, $2, $3, $4, $5)
			`, 1, 1, 1, 1, "2025-06-02 03:00:00")
	if err != nil {
		return
	}
	repo := NewGenreRepo(testutils.TestDB)

	limit := 5

	genres, err := repo.GetTopGenresAtNight(limit)
	require.NoError(t, err)
	require.NotNil(t, genres)
	require.LessOrEqual(t, len(genres), limit)

	for _, g := range genres {
		t.Logf("Night Genre ID: %d, Name: %s", g.ID, g.Name)
	}
}

func TestGetTopGenresInMorning(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)
	testutils.TestDB.Exec(`INSERT INTO songs (song_id, name) VALUES (1, 'Track')`)
	testutils.TestDB.Exec(`INSERT INTO artists (artist_id, name) VALUES (1, 'Artist')`)
	testutils.TestDB.Exec(`INSERT INTO genres (genre_id, name) VALUES (1, 'Genre')`)
	_, err := testutils.TestDB.Exec(`
    INSERT INTO fact_listens (user_id, song_id, artist_id, genre_id, listende_at)
    	VALUES ($1, $2, $3, $4, $5)
			`, 1, 1, 1, 1, "2025-06-02 07:00:00")
	if err != nil {
		return
	}
	repo := NewGenreRepo(testutils.TestDB)

	limit := 5

	genres, err := repo.GetTopGenresInMorning(limit)
	require.NoError(t, err)
	require.NotNil(t, genres)
	require.LessOrEqual(t, len(genres), limit)

	for _, g := range genres {
		t.Logf("Morning Genre ID: %d, Name: %s", g.ID, g.Name)
	}
}

func TestGetTopGenresInDay(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)
	testutils.TestDB.Exec(`INSERT INTO songs (song_id, name) VALUES (1, 'Track')`)
	testutils.TestDB.Exec(`INSERT INTO artists (artist_id, name) VALUES (1, 'Artist')`)
	testutils.TestDB.Exec(`INSERT INTO genres (genre_id, name) VALUES (1, 'Genre')`)
	_, err := testutils.TestDB.Exec(`
    INSERT INTO fact_listens (user_id, song_id, artist_id, genre_id, listende_at)
    	VALUES ($1, $2, $3, $4, $5)
			`, 1, 1, 1, 1, "2025-06-02 12:00:00")
	if err != nil {
		return
	}
	repo := NewGenreRepo(testutils.TestDB)

	limit := 5

	genres, err := repo.GetTopGenresInDay(limit)
	require.NoError(t, err)
	require.NotNil(t, genres)
	require.LessOrEqual(t, len(genres), limit)

	for _, g := range genres {
		t.Logf("Day Genre ID: %d, Name: %s", g.ID, g.Name)
	}
}

func TestGetTopGenresInEvening(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)
	testutils.TestDB.Exec(`INSERT INTO songs (song_id, name) VALUES (1, 'Track')`)
	testutils.TestDB.Exec(`INSERT INTO artists (artist_id, name) VALUES (1, 'Artist')`)
	testutils.TestDB.Exec(`INSERT INTO genres (genre_id, name) VALUES (1, 'Genre')`)
	_, err := testutils.TestDB.Exec(`
    INSERT INTO fact_listens (user_id, song_id, artist_id, genre_id, listende_at)
    	VALUES ($1, $2, $3, $4, $5)
			`, 1, 1, 1, 1, "2025-06-02 19:00:00")
	if err != nil {
		return
	}
	repo := NewGenreRepo(testutils.TestDB)

	limit := 5

	genres, err := repo.GetTopGenresInEvening(limit)
	require.NoError(t, err)
	require.NotNil(t, genres)
	require.LessOrEqual(t, len(genres), limit)

	for _, g := range genres {
		t.Logf("Evening Genre ID: %d, Name: %s", g.ID, g.Name)
	}
}
