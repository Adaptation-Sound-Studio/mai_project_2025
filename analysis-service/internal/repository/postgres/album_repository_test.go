package postgres

import (
	"testing"

	_ "github.com/lib/pq"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/album"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/testutils"
)

func TestAlbumRepo_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	testutils.CleanTables(t, testutils.TestDB)
	repo := NewAlbumRepo(testutils.TestDB)

	album := &album.Album{Name: "Test Album"}
	err := repo.Create(album)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	err = testutils.TestDB.QueryRow("SELECT COUNT(*) FROM albums WHERE name = $1", album.Name).Scan(&count)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}

func TestAlbumRepo_Create_NullManually(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testutils.TestDB)
	_, err := testutils.TestDB.Exec("INSERT INTO albums (name) VALUES ($1)", nil)

	if err == nil {
		t.Fatal("expected NOT NULL constraint violation, got nil")
	}
	t.Logf("got expected error: %v", err)
}
