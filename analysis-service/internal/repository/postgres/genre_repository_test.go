package postgres

import (
	"testing"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/genre"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/testutils"
)

func TestGenretRepo_Create_Success(t *testing.T) {

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
