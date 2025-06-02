package service

import (
	"context"
	"errors"
	"testing"
	"upload-service/internal/domain/event"
	"upload-service/internal/domain/model"

	"github.com/elastic/go-elasticsearch/v8"
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

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) SendWrappedEvent(topic string, payload interface{}) error {
	args := m.Called(topic, payload)
	return args.Error(0)
}

func TestCreateGenre_Success(t *testing.T) {
	repo := new(MockGenreRepo)
	producer := new(MockProducer)
	es := &elasticsearch.Client{}

	svc := NewGenreService(repo, nil, es)

	genre := &model.Genre{Name: "Rock"}

	repo.On("CreateGenre", mock.Anything, genre).Return(int64(1), nil)

	producer.On("SendWrappedEvent", "genre_created", event.Genre{
		ID:   1,
		Name: "Rock",
	}).Return(nil)

	id, err := svc.CreateGenre(context.Background(), genre)

	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	repo.AssertExpectations(t)
	producer.AssertExpectations(t)
}

func TestCreateGenre_EmptyName(t *testing.T) {
	repo := new(MockGenreRepo)
	svc := NewGenreService(repo, nil, nil)

	genre := &model.Genre{Name: ""}

	id, err := svc.CreateGenre(context.Background(), genre)

	assert.EqualError(t, err, "название жанра не может быть пустым")
	assert.Equal(t, int64(0), id)
}

func TestCreateGenre_RepoError(t *testing.T) {
	repo := new(MockGenreRepo)
	svc := NewGenreService(repo, nil, nil)

	genre := &model.Genre{Name: "Jazz"}

	repo.On("CreateGenre", mock.Anything, genre).Return(int64(0), errors.New("db failure"))

	id, err := svc.CreateGenre(context.Background(), genre)

	assert.EqualError(t, err, "db failure")
	assert.Equal(t, int64(0), id)
	repo.AssertExpectations(t)
}

func TestUpdateGenre_Success(t *testing.T) {
	repo := new(MockGenreRepo)
	svc := NewGenreService(repo, nil, nil)

	genre := &model.Genre{GenreID: 1, Name: "Jazz"}
	repo.On("UpdateGenre", mock.Anything, genre).Return(nil)

	err := svc.UpdateGenre(context.Background(), genre)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdateGenre_InvalidID(t *testing.T) {
	repo := new(MockGenreRepo)
	svc := NewGenreService(repo, nil, nil)

	err := svc.UpdateGenre(context.Background(), &model.Genre{
		GenreID: 0,
		Name:    "Rock",
	})

	assert.EqualError(t, err, "некорректный ID жанра для обновления")
}

func TestUpdateGenre_EmptyName(t *testing.T) {
	repo := new(MockGenreRepo)
	svc := NewGenreService(repo, nil, nil)

	err := svc.UpdateGenre(context.Background(), &model.Genre{
		GenreID: 1,
		Name:    "",
	})

	assert.EqualError(t, err, "название жанра не может быть пустым")
}

func TestUpdateGenre_UpdateFailed(t *testing.T) {
	repo := new(MockGenreRepo)
	svc := NewGenreService(repo, nil, nil)

	genre := &model.Genre{
		GenreID: 2,
		Name:    "Jazz",
	}

	repo.On("UpdateGenre", mock.Anything, genre).
		Return(errors.New("ошибка обновления в БД"))

	err := svc.UpdateGenre(context.Background(), genre)

	assert.EqualError(t, err, "ошибка обновления в БД")
	repo.AssertExpectations(t)
}

func TestGetAllGenres_Error(t *testing.T) {
	repo := new(MockGenreRepo)
	svc := NewGenreService(repo, nil, nil)

	repo.On("GetAllGenres", mock.Anything).
		Return([]model.Genre{}, errors.New("db error"))

	genres, err := svc.GetAllGenres(context.Background())

	assert.Nil(t, genres)
	assert.EqualError(t, err, "db error")
	repo.AssertExpectations(t)
}

func TestCreateGenre_DBError(t *testing.T) {
	repo := new(MockGenreRepo)
	svc := NewGenreService(repo, nil, nil)

	genre := &model.Genre{Name: "Jazz"}
	repo.On("CreateGenre", mock.Anything, genre).
		Return(int64(0), errors.New("insert error"))

	id, err := svc.CreateGenre(context.Background(), genre)

	assert.Equal(t, int64(0), id)
	assert.EqualError(t, err, "insert error")
	repo.AssertExpectations(t)
}
