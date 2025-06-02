package repository

import (
	"context"
	"testing"
	"upload-service/internal/domain/model"
	"upload-service/internal/testutils"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGenre(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewGenreRepository(testDB)

	genre := &model.Genre{Name: "Electronic"}
	id, err := repo.CreateGenre(ctx, genre)
	require.NoError(t, err)
	assert.NotZero(t, id)

	var name string
	err = testDB.QueryRowContext(ctx, "SELECT name FROM genres WHERE genre_id = $1", id).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "Electronic", name)
}

func TestCreateGenre_DuplicateName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewGenreRepository(testDB)

	genre := &model.Genre{Name: "UniqueGenre"}

	_, err := repo.CreateGenre(ctx, genre)
	require.NoError(t, err)

	_, err = repo.CreateGenre(ctx, genre)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key")
}

func TestGetAllGenres(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewGenreRepository(testDB)

	_, err := repo.CreateGenre(ctx, &model.Genre{Name: "Jazz"})
	require.NoError(t, err)
	_, err = repo.CreateGenre(ctx, &model.Genre{Name: "Blues"})
	require.NoError(t, err)

	genres, err := repo.GetAllGenres(ctx)
	require.NoError(t, err)
	require.Len(t, genres, 2)

	names := []string{genres[0].Name, genres[1].Name}
	assert.ElementsMatch(t, []string{"Jazz", "Blues"}, names)
}

func TestGetAllGenres_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewGenreRepository(testDB)

	genres, err := repo.GetAllGenres(ctx)
	require.NoError(t, err)
	assert.Empty(t, genres)
}

func TestUpdateGenre(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewGenreRepository(testDB)

	genre := &model.Genre{Name: "HipHop"}
	id, err := repo.CreateGenre(ctx, genre)
	require.NoError(t, err)

	updated := &model.Genre{GenreID: id, Name: "Hip-Hop"}
	err = repo.UpdateGenre(ctx, updated)
	require.NoError(t, err)

	var name string
	err = testDB.QueryRowContext(ctx, "SELECT name FROM genres WHERE genre_id = $1", id).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "Hip-Hop", name)
}

func TestUpdateGenre_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewGenreRepository(testDB)

	genre := &model.Genre{GenreID: 999, Name: "Non-existent"}

	err := repo.UpdateGenre(ctx, genre)
	require.Error(t, err)
	assert.EqualError(t, err, "жанр с ID 999 не найден")
}

func TestUpdateGenre_InvalidID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	testutils.CleanTables(t, testDB)
	ctx := context.Background()
	repo := NewGenreRepository(testDB)

	genre := &model.Genre{GenreID: -1, Name: "WrongID"}
	err := repo.UpdateGenre(ctx, genre)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "жанр с ID -1 не найден")
}
