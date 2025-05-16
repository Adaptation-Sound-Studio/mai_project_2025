package postgres

import (
	"strings"
	"testing"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/artist"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/testutils"
	_ "github.com/lib/pq"
)

func TestArtistRepo_Create_Success(t *testing.T) {
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
	testutils.CleanTables(t, testutils.TestDB)
	repo := NewArtistRepo(testutils.TestDB)

	// Создаем слишком длинное имя (>100 символов)
	longName := "A" + strings.Repeat("verylongartistname", 10) // получится 161 символ

	a := &artist.Artist{Name: longName}

	err := repo.Create(a)

	if err == nil {
		t.Fatal("expected error due to name exceeding length constraint, got nil")
	}

	t.Logf("got expected error: %v", err)
}
