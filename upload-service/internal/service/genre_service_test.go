package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"upload-service/internal/config"
	"upload-service/internal/domain/event"
	"upload-service/internal/domain/model"
	"upload-service/internal/infrastructure/elastic"

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

func (m *MockProducer) SendWrappedEvent(eventType string, payload interface{}) error {
	args := m.Called(eventType, payload)
	return args.Error(0)
}

func TestCreateGenre_Success(t *testing.T) {
	repo := new(MockGenreRepo)
	producer := new(MockProducer)

	svc := NewGenreService(repo, producer, nil)

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

func TestSearchGenres_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	cfg := &config.ElasticConfig{
		URL:      "http://localhost:9200",
		Username: "elastic",
		Password: "your_password",
	}

	esClient, err := elastic.NewElasticClient(cfg)
	require.NoError(t, err)

	service := &GenreService{
		elasticClient: esClient,
	}

	body := `{"genre_id": 100, "name": "Electro"}`
	_, err = esClient.Index(
		"genres",
		strings.NewReader(body),
		esClient.Index.WithDocumentID("100"),
		esClient.Index.WithRefresh("true"),
	)
	require.NoError(t, err)

	results, err := service.SearchGenres(context.Background(), "Electro")
	require.NoError(t, err)
	require.NotEmpty(t, results)

	found := false
	for _, g := range results {
		if g.Name == "Electro" && g.GenreID == 100 {
			found = true
			break
		}
	}
	require.True(t, found, "Жанр 'Electro' не найден в результатах поиска")
}

func TestSearchGenres_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	cfg := &config.ElasticConfig{
		URL:      "http://localhost:9200",
		Username: "elastic",
		Password: "your_password",
	}

	esClient, err := elastic.NewElasticClient(cfg)
	require.NoError(t, err)

	service := &GenreService{
		elasticClient: esClient,
	}

	_, _ = esClient.Indices.Delete([]string{"genres"})

	body := `{"genre_id": 200, "name": "Jazz Fusion"}`
	_, err = esClient.Index(
		"genres",
		strings.NewReader(body),
		esClient.Index.WithDocumentID("200"),
		esClient.Index.WithRefresh("true"),
	)
	require.NoError(t, err)

	results, err := service.SearchGenres(context.Background(), "Jazz")
	require.NoError(t, err)
	require.NotEmpty(t, results)

	found := false
	for _, g := range results {
		if g.GenreID == 200 && g.Name == "Jazz Fusion" {
			found = true
			break
		}
	}
	require.True(t, found, "Жанр 'Jazz Fusion' не найден в результатах поиска")
}

func TestIndexGenre_Integration(t *testing.T) {
	cfg := &config.ElasticConfig{
		URL:      "http://localhost:9200",
		Username: "elastic",
		Password: "your_password",
	}
	esClient, err := elastic.NewElasticClient(cfg)
	require.NoError(t, err)

	service := &GenreService{
		elasticClient: esClient,
	}

	ctx := context.Background()
	genre := &model.Genre{
		GenreID: 12345,
		Name:    "TestGenre123",
	}

	err = service.indexGenre(ctx, genre)

	require.NoError(t, err)

	res, err := esClient.Get("genres", fmt.Sprint(genre.GenreID))
	require.NoError(t, err)
	defer res.Body.Close()

	require.False(t, res.IsError(), "ожидалось, что документ будет найден")

	var doc struct {
		Source model.Genre `json:"_source"`
	}
	err = json.NewDecoder(res.Body).Decode(&doc)
	require.NoError(t, err)

	assert.Equal(t, genre.GenreID, doc.Source.GenreID)
	assert.Equal(t, genre.Name, doc.Source.Name)
}
