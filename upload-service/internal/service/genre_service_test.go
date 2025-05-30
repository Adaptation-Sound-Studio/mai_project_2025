package service

import (
	"context"
	"testing"
	"upload-service/internal/domain/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockGenreRepo struct {
	mock.Mock
}

func (m *MockGenreRepo) GetAllGenres(ctx context.Context) ([]model.Genre, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Genre), args.Error(1)
}

func (m *MockGenreRepo) CreateGenre(ctx context.Context, genre *model.Genre) (int64, error) {
	args := m.Called(ctx, genre)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockGenreRepo) UpdateGenre(ctx context.Context, genre *model.Genre) error {
	args := m.Called(ctx, genre)
	return args.Error(0)
}

func TestGetAllGenres_Success(t *testing.T) {
	repo := new(MockGenreRepo)
	svc := NewGenreService(repo)

	expected := []model.Genre{{GenreID: 1, Name: "Pop"}}
	repo.On("GetAllGenres", mock.Anything).Return(expected, nil)

	result, err := svc.GetAllGenres(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestCreateGenre_Success(t *testing.T) {
	repo := new(MockGenreRepo)
	svc := NewGenreService(repo)

	genre := &model.Genre{Name: "Rock"}
	repo.On("CreateGenre", mock.Anything, genre).Return(int64(10), nil)

	id, err := svc.CreateGenre(context.Background(), genre)

	require.NoError(t, err)
	assert.Equal(t, int64(10), id)
	repo.AssertExpectations(t)
}

func TestUpdateGenre_Success(t *testing.T) {
	repo := new(MockGenreRepo)
	svc := NewGenreService(repo)

	genre := &model.Genre{GenreID: 1, Name: "Jazz"}
	repo.On("UpdateGenre", mock.Anything, genre).Return(nil)

	err := svc.UpdateGenre(context.Background(), genre)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}
