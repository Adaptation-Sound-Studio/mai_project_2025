package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"upload-service/internal/domain/model"
	"upload-service/internal/domain/request"
	"upload-service/internal/kafka"
	"upload-service/internal/service"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	mockRepo := new(MockGenreRepo)
	dummyProducer := new(kafka.ProducerIface)
	var elasticClient *elasticsearch.Client = nil
	svc := service.NewGenreService(mockRepo, *dummyProducer, elasticClient)

	handler := NewGenreHandler(svc)

	mockGenres := []model.Genre{{GenreID: 1, Name: "Rock"}}
	mockRepo.On("GetAllGenres", mock.Anything).Return(mockGenres, nil)

	req := httptest.NewRequest(http.MethodGet, "/genres", nil)
	rr := httptest.NewRecorder()

	handler.GetAllGenres(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUpdateGenre_Success(t *testing.T) {
	mockRepo := new(MockGenreRepo)
	var dummyProducer kafka.ProducerIface = nil
	var elasticClient *elasticsearch.Client = nil

	svc := service.NewGenreService(mockRepo, dummyProducer, elasticClient)
	handler := NewGenreHandler(svc)

	genre := &model.Genre{GenreID: 1, Name: "Blues"}
	mockRepo.On("UpdateGenre", mock.Anything, genre).Return(nil)

	body := `{"name": "Blues"}`
	req := httptest.NewRequest(http.MethodPut, "/genres/1", bytes.NewBufferString(body))
	req = mux.SetURLVars(req, map[string]string{"genre_id": "1"})
	rr := httptest.NewRecorder()

	handler.UpdateGenre(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}
func TestCreateGenre_BadRequest(t *testing.T) {
	mockRepo := new(MockGenreRepo)
	var dummyProducer kafka.ProducerIface = nil
	var elasticClient *elasticsearch.Client = nil

	svc := service.NewGenreService(mockRepo, dummyProducer, elasticClient)
	handler := NewGenreHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/genres", bytes.NewBufferString("invalid-json"))
	rr := httptest.NewRecorder()

	handler.CreateGenre(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Неверный формат запроса")
}

func TestCreateGenre_MethodNotAllowed(t *testing.T) {
	mockRepo := new(MockGenreRepo)
	var dummyProducer kafka.ProducerIface = nil
	var elasticClient *elasticsearch.Client = nil

	svc := service.NewGenreService(mockRepo, dummyProducer, elasticClient)
	handler := NewGenreHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/genres", nil)
	rr := httptest.NewRecorder()

	handler.CreateGenre(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "Метод не разрешён")
}

func TestCreateGenre_ServiceError(t *testing.T) {
	mockRepo := new(MockGenreRepo)
	var dummyProducer kafka.ProducerIface = nil
	var elasticClient *elasticsearch.Client = nil

	svc := service.NewGenreService(mockRepo, dummyProducer, elasticClient)
	handler := NewGenreHandler(svc)

	genreReq := request.CreateGenreRequest{Name: "Rock"}
	mockRepo.On("CreateGenre", mock.Anything, &model.Genre{Name: "Rock"}).
		Return(int64(0), errors.New("жанр уже существует"))

	body, _ := json.Marshal(genreReq)
	req := httptest.NewRequest(http.MethodPost, "/genres", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.CreateGenre(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "жанр уже существует")
	mockRepo.AssertExpectations(t)
}

func TestUpdateGenre_MethodNotAllowed(t *testing.T) {
	handler := NewGenreHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/genres/1", nil)
	req = mux.SetURLVars(req, map[string]string{"genre_id": "1"})
	rr := httptest.NewRecorder()

	handler.UpdateGenre(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "Метод не разрешён")
}

func TestUpdateGenre_InvalidID(t *testing.T) {
	handler := NewGenreHandler(nil)

	req := httptest.NewRequest(http.MethodPut, "/genres/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"genre_id": "invalid"})
	rr := httptest.NewRecorder()

	handler.UpdateGenre(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Некорректный ID жанра")
}

func TestUpdateGenre_BadJSON(t *testing.T) {
	handler := NewGenreHandler(nil)

	req := httptest.NewRequest(http.MethodPut, "/genres/1", bytes.NewBufferString("{invalid_json}"))
	req = mux.SetURLVars(req, map[string]string{"genre_id": "1"})
	rr := httptest.NewRecorder()

	handler.UpdateGenre(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Неверный формат запроса")
}

func TestUpdateGenre_ServiceError(t *testing.T) {
	mockRepo := new(MockGenreRepo)
	var dummyProducer kafka.ProducerIface = nil
	var elasticClient *elasticsearch.Client = nil

	svc := service.NewGenreService(mockRepo, dummyProducer, elasticClient)
	handler := NewGenreHandler(svc)

	genre := &model.Genre{GenreID: 1, Name: "Rock"}
	mockRepo.On("UpdateGenre", mock.Anything, genre).Return(errors.New("жанр не найден"))

	body, _ := json.Marshal(map[string]string{"name": "Rock"})
	req := httptest.NewRequest(http.MethodPut, "/genres/1", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"genre_id": "1"})
	rr := httptest.NewRecorder()

	handler.UpdateGenre(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "жанр не найден")
	mockRepo.AssertExpectations(t)
}

func TestGetAllGenres_ServiceError(t *testing.T) {
	mockRepo := new(MockGenreRepo)
	var dummyProducer kafka.ProducerIface = nil
	var elasticClient *elasticsearch.Client = nil

	svc := service.NewGenreService(mockRepo, dummyProducer, elasticClient)
	handler := NewGenreHandler(svc)

	mockRepo.On("GetAllGenres", mock.Anything).Return([]model.Genre(nil), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/genres", nil)
	rr := httptest.NewRecorder()

	handler.GetAllGenres(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Ошибка при получении жанров")

	mockRepo.AssertExpectations(t)
}

func TestGetAllGenres_MethodNotAllowed(t *testing.T) {
	handler := NewGenreHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/genres", nil)
	rr := httptest.NewRecorder()

	handler.GetAllGenres(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "Метод не разрешён")
}
