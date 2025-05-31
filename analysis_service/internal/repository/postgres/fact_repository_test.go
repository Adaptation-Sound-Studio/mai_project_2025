package postgres

import (
	"testing"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/fact"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/testutils"
)

func TestFactRepo_Insert_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)

	testutils.TestDB.Exec(`INSERT INTO users (user_id, name) VALUES (1, 'User')`)
	testutils.TestDB.Exec(`INSERT INTO songs (song_id, name) VALUES (1, 'Track')`)
	testutils.TestDB.Exec(`INSERT INTO artists (artist_id, name) VALUES (1, 'Artist')`)
	testutils.TestDB.Exec(`INSERT INTO genres (genre_id, name) VALUES (1, 'Genre')`)

	repo := NewFactRepo(testutils.TestDB)

	fact := &fact.ListenFact{
		UserID:     1,
		SongID:     1,
		ArtistID:   1,
		GenreID:    1,
		ListenedAt: nil,
	}

	err := repo.Insert(fact)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	err = testutils.TestDB.QueryRow("SELECT COUNT(*) FROM fact_listens").Scan(&count)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}

func TestFactRepo_Insert_InvalidFK(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)

	testutils.TestDB.Exec(`INSERT INTO users (user_id, name) VALUES (1, 'User')`)
	testutils.TestDB.Exec(`INSERT INTO songs (song_id, name) VALUES (1, 'Track')`)
	testutils.TestDB.Exec(`INSERT INTO genres (genre_id, name) VALUES (1, 'Genre')`)

	repo := NewFactRepo(testutils.TestDB)

	fact := &fact.ListenFact{
		UserID:     1,
		SongID:     1,
		ArtistID:   999,
		GenreID:    1,
		ListenedAt: nil,
	}

	err := repo.Insert(fact)
	if err == nil {
		t.Fatal("expected foreign key constraint error, got nil")
	}

	t.Logf("got expected FK error: %v", err)
}
