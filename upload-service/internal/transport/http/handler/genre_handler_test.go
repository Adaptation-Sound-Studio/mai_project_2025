package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"upload-service/internal/domain/model"
	"upload-service/internal/service"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGenreRepo struct{ mock.Mock }

func (m *MockGenreRepo) GetAllGenres(ctx context.Context) ([]model.Genre, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Genre), args.Error(1)
}
func (m *MockGenreRepo) CreateGenre(ctx context.Context, g *model.Genre) (int64, error) {
	args := m.Called(ctx, g)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockGenreRepo) UpdateGenre(ctx context.Context, g *model.Genre) error {
	args := m.Called(ctx, g)
	return args.Error(0)
}

type MockProducer struct{ mock.Mock }

func (m *MockProducer) SendWrappedEvent(topic string, payload interface{}) error {
	args := m.Called(topic, payload)
	return args.Error(0)
}

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type MockGenreService struct {
	mock.Mock
}

func (m *MockGenreService) UpdateGenre(ctx context.Context, genre *model.Genre) error {
	args := m.Called(ctx, genre)
	return args.Error(0)
}

func esMock() *elasticsearch.Client {
	es, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: rtFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 201,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
				Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
			}, nil
		}),
	})
	return es
}

func newTestHandler(repo *MockGenreRepo) *GenreHandler {
	svc := service.NewGenreService(repo, nil, esMock())
	return NewGenreHandler(svc)
}

type genreServiceStub struct {
	*service.GenreService
	getAllGenresFunc func(ctx context.Context) ([]*model.Genre, error)
}

func (s *genreServiceStub) GetAllGenres(ctx context.Context) ([]*model.Genre, error) {
	if s.getAllGenresFunc != nil {
		return s.getAllGenresFunc(ctx)
	}
	return nil, nil
}

func TestGetAllGenres_Success(t *testing.T) {
	repo := new(MockGenreRepo)
	h := newTestHandler(repo)

	want := []model.Genre{{GenreID: 1, Name: "Rock"}}
	repo.On("GetAllGenres", mock.Anything).Return(want, nil)

	req := httptest.NewRequest(http.MethodGet, "/genres", nil)
	rec := httptest.NewRecorder()

	h.GetAllGenres(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var got []model.Genre
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, want, got)
	repo.AssertExpectations(t)
}

func TestCreateGenre_Success(t *testing.T) {
	repo := new(MockGenreRepo)
	h := newTestHandler(repo)

	body := `{"name":"Jazz"}`
	repo.
		On("CreateGenre", mock.Anything, mock.MatchedBy(
			func(g *model.Genre) bool { return g.Name == "Jazz" }),
		).Return(int64(5), nil)

	req := httptest.NewRequest(http.MethodPost, "/genres", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.CreateGenre(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]string
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "5", resp["genre_id"])
	repo.AssertExpectations(t)
}

func TestCreateGenre_BadJSON(t *testing.T) {
	repo := new(MockGenreRepo)
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/genres", bytes.NewBufferString(`{`))
	rec := httptest.NewRecorder()

	h.CreateGenre(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	repo.AssertNotCalled(t, "CreateGenre", mock.Anything, mock.Anything)
}

func TestCreateGenre_ServiceError(t *testing.T) {
	repo := new(MockGenreRepo)
	h := newTestHandler(repo)

	repo.
		On("CreateGenre", mock.Anything, mock.Anything).
		Return(int64(0), errors.New("duplicate"))

	req := httptest.NewRequest(http.MethodPost, "/genres",
		bytes.NewBufferString(`{"name":"Rock"}`))
	rec := httptest.NewRecorder()

	h.CreateGenre(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	repo.AssertExpectations(t)
}
