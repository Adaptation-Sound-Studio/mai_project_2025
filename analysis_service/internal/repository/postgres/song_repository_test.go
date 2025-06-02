package postgres

import (
	"strings"
	"testing"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/song"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/testutils"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestSongRepo_Create_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)
	repo := NewSongRepo(testutils.TestDB)

	s := &song.Song{Name: "My Track"}
	err := repo.Create(s)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	err = testutils.TestDB.QueryRow("SELECT COUNT(*) FROM songs WHERE name = $1", s.Name).Scan(&count)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 song, got %d", count)
	}
}

func TestSongRepo_Create_NameTooLong(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)
	repo := NewSongRepo(testutils.TestDB)

	longName := strings.Repeat("A", 300)

	s := &song.Song{Name: longName}
	err := repo.Create(s)

	if err == nil {
		t.Fatal("expected error due to long name, got nil")
	}
	t.Logf("got expected error: %v", err)
}

func TestSongRepo_GetPopularSongs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)
	testutils.TestDB.Exec(`INSERT INTO artists (name) VALUES ('Artist1')`)
	testutils.TestDB.Exec(`INSERT INTO genres (name) VALUES ('Rock')`)
	testutils.TestDB.Exec(`INSERT INTO songs (name) VALUES ('Track1')`)

	testutils.TestDB.Exec(`
	INSERT INTO fact_listens (user_id, song_id, artist_id, genre_id, listened_at)
	VALUES 
	(1, 1, 1, 1, CURRENT_TIMESTAMP),
	(1, 1, 1, 1, CURRENT_TIMESTAMP),
	(1, 1, 1, 1, CURRENT_TIMESTAMP)
	`)

	repo := NewSongRepo(testutils.TestDB)

	songs, err := repo.GetPopularSongs(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(songs) != 1 {
		t.Fatalf("expected 1 popular song, got %d", len(songs))
	}
	if songs[0].Listens != 3 {
		t.Fatalf("expected 3 listens, got %d", songs[0].Listens)
	}
}

func TestFactListens_Insert_InvalidArtistFK(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)

	testutils.TestDB.Exec(`INSERT INTO genres (genre_id, name) VALUES (1, 'Genre')`)
	testutils.TestDB.Exec(`INSERT INTO songs (song_id, name) VALUES (1, 'Track')`)

	_, err := testutils.TestDB.Exec(`
		INSERT INTO fact_listens (user_id, song_id, artist_id, genre_id)
		VALUES (1, 1, 999, 1)
	`)
	if err == nil {
		t.Fatal("expected foreign key constraint error, got nil")
	}

	t.Logf("got expected FK error: %v", err)
}

func TestGetTopSongsForUser(t *testing.T) {
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
	repo := NewSongRepo(testutils.TestDB)

	userID := 1
	limit := 1

	songs, err := repo.GetTopSongsForUser(userID, limit)
	require.NoError(t, err)
	require.NotNil(t, songs)
	require.LessOrEqual(t, len(songs), limit)

	for _, s := range songs {
		t.Logf("Song ID: %d, Name: %s", s.ID, s.Name)
	}
}

func TestGetMostPopularSongs(t *testing.T) {
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
	repo := NewSongRepo(testutils.TestDB)

	limit := 5

	songs, err := repo.GetMostPopularSongs(limit)
	require.NoError(t, err)
	require.NotNil(t, songs)
	require.LessOrEqual(t, len(songs), limit)

	for _, s := range songs {
		t.Logf("Song ID: %d, Name: %s", s.ID, s.Name)
	}
}
