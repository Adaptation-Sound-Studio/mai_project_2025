package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/genre"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockGenreRepo struct {
	mock.Mock
}

func (m *MockGenreRepo) Create(g *genre.Genre) error {
	args := m.Called(g)
	return args.Error(0)
}

func (m *MockGenreRepo) GetMostPopularGenres(limit int) ([]genre.Genre, error) {
	args := m.Called(limit)
	if obj := args.Get(0); obj != nil {
		return obj.([]genre.Genre), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGenreRepo) GetTopGenresForUser(userID int, limit int) ([]genre.Genre, error) {
	args := m.Called(userID, limit)
	if obj := args.Get(0); obj != nil {
		return obj.([]genre.Genre), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGenreRepo) GetTopGenresAtNight(limit int) ([]genre.Genre, error) {
	args := m.Called(limit)
	if obj := args.Get(0); obj != nil {
		return obj.([]genre.Genre), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGenreRepo) GetTopGenresInMorning(limit int) ([]genre.Genre, error) {
	args := m.Called(limit)
	if obj := args.Get(0); obj != nil {
		return obj.([]genre.Genre), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGenreRepo) GetTopGenresInDay(limit int) ([]genre.Genre, error) {
	args := m.Called(limit)
	if obj := args.Get(0); obj != nil {
		return obj.([]genre.Genre), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGenreRepo) GetTopGenresInEvening(limit int) ([]genre.Genre, error) {
	args := m.Called(limit)
	if obj := args.Get(0); obj != nil {
		return obj.([]genre.Genre), args.Error(1)
	}
	return nil, args.Error(1)
}
func TestCreateGenre_Success(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	mockRepo.On("Create", mock.AnythingOfType("*genre.Genre")).Return(nil)

	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/genres", handler.CreateGenre).Methods("POST")

	input := genre.Genre{Name: "Jazz"}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/genres", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", "admin-secret")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Contains(t, rec.Body.String(), "Jazz")

	mockRepo.AssertExpectations(t)
}

func TestCreateGenre_BadJSON(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/genres", handler.CreateGenre).Methods("POST")

	req := httptest.NewRequest(http.MethodPost, "/genres", bytes.NewReader([]byte(`not-json`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", "admin-secret")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "Invalid JSON")
}

func TestCreateGenre_Forbidden(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/genres", handler.CreateGenre).Methods("POST")

	req := httptest.NewRequest(http.MethodPost, "/genres", bytes.NewReader([]byte(`{"name":"Rock"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", "wrong-key")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Contains(t, rec.Body.String(), "Forbidden")
}

func TestCreateGenre_InternalError(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	mockRepo.
		On("Create", mock.AnythingOfType("*genre.Genre")).
		Return(errors.New("database failure"))

	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/genres", handler.CreateGenre).Methods("POST")

	input := genre.Genre{Name: "Blues"}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/genres", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", "admin-secret")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "Failed to create genre")

	mockRepo.AssertExpectations(t)
}

func TestGetTopGenresForUser_Success(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	expectedGenres := []genre.Genre{{ID: 1, Name: "Rock"}, {ID: 2, Name: "Jazz"}}
	mockRepo.On("GetTopGenresForUser", 42, 10).Return(expectedGenres, nil)

	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/users/{user_id}/top-genres", handler.GetTopGenresForUser).Methods("GET")

	req := httptest.NewRequest(http.MethodGet, "/users/42/top-genres?limit=10", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var genres []genre.Genre
	err := json.NewDecoder(rec.Body).Decode(&genres)
	require.NoError(t, err)
	require.Equal(t, expectedGenres, genres)

	mockRepo.AssertExpectations(t)
}

func TestGetTopGenresForUser_Error(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	mockRepo.On("GetTopGenresForUser", 42, 10).Return(nil, errors.New("some error"))

	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "111")

	router := mux.NewRouter()
	router.HandleFunc("/users/{user_id}/top-genres", handler.GetTopGenresForUser).Methods("GET")

	req := httptest.NewRequest(http.MethodGet, "/users/42/top-genres?limit=10", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetMostPopularGenres_Success(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	expectedGenres := []genre.Genre{{ID: 1, Name: "Pop"}, {ID: 2, Name: "Classical"}}
	mockRepo.On("GetMostPopularGenres", 10).Return(expectedGenres, nil)

	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "")

	req := httptest.NewRequest(http.MethodGet, "/genres/popular?limit=10", nil)
	rec := httptest.NewRecorder()

	handler.GetMostPopularGenres(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var genres []genre.Genre
	err := json.NewDecoder(rec.Body).Decode(&genres)
	require.NoError(t, err)
	require.Equal(t, expectedGenres, genres)

	mockRepo.AssertExpectations(t)
}

func TestGetMostPopularGenres_Error(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	mockRepo.On("GetMostPopularGenres", 10).Return(nil, errors.New("some error"))

	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "admin-secret")

	req := httptest.NewRequest(http.MethodGet, "/genres/popular?limit=10", nil)
	rec := httptest.NewRecorder()

	handler.GetMostPopularGenres(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetTopGenresAtNight_Success(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	expectedGenres := []genre.Genre{{ID: 3, Name: "Blues"}}
	mockRepo.On("GetTopGenresAtNight", 10).Return(expectedGenres, nil)

	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "admin-secret")

	req := httptest.NewRequest(http.MethodGet, "/genres/top/night?limit=10", nil)
	rec := httptest.NewRecorder()

	handler.GetTopGenresAtNight(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var genres []genre.Genre
	err := json.NewDecoder(rec.Body).Decode(&genres)
	require.NoError(t, err)
	require.Equal(t, expectedGenres, genres)

	mockRepo.AssertExpectations(t)
}

func TestGetTopGenresAtNight_Error(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	mockRepo.On("GetTopGenresAtNight", 10).Return(nil, errors.New("some error"))

	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "admin-secret")

	req := httptest.NewRequest(http.MethodGet, "/genres/top/night?limit=10", nil)
	rec := httptest.NewRecorder()

	handler.GetTopGenresAtNight(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetTopGenresInMorning_Success(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	expectedGenres := []genre.Genre{{ID: 4, Name: "Country"}}
	mockRepo.On("GetTopGenresInMorning", 10).Return(expectedGenres, nil)

	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "admin-secret")

	req := httptest.NewRequest(http.MethodGet, "/genres/top/morning?limit=10", nil)
	rec := httptest.NewRecorder()

	handler.GetTopGenresInMorning(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var genres []genre.Genre
	err := json.NewDecoder(rec.Body).Decode(&genres)
	require.NoError(t, err)
	require.Equal(t, expectedGenres, genres)

	mockRepo.AssertExpectations(t)
}

func TestGetTopGenresInMorning_Error(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	mockRepo.On("GetTopGenresInMorning", 10).Return(nil, errors.New("some error"))

	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "admin-secret")

	req := httptest.NewRequest(http.MethodGet, "/genres/top/morning?limit=10", nil)
	rec := httptest.NewRecorder()

	handler.GetTopGenresInMorning(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetTopGenresInDay_Success(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	expectedGenres := []genre.Genre{{ID: 5, Name: "Disco"}}
	mockRepo.On("GetTopGenresInDay", 10).Return(expectedGenres, nil)

	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "admin-secret")

	req := httptest.NewRequest(http.MethodGet, "/genres/top/day?limit=10", nil)
	rec := httptest.NewRecorder()

	handler.GetTopGenresInDay(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var genres []genre.Genre
	err := json.NewDecoder(rec.Body).Decode(&genres)
	require.NoError(t, err)
	require.Equal(t, expectedGenres, genres)

	mockRepo.AssertExpectations(t)
}

func TestGetTopGenresInDay_Error(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	mockRepo.On("GetTopGenresInDay", 10).Return(nil, errors.New("some error"))

	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "admin-secret")

	req := httptest.NewRequest(http.MethodGet, "/genres/top/day?limit=10", nil)
	rec := httptest.NewRecorder()

	handler.GetTopGenresInDay(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetTopGenresInEvening_Success(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	expectedGenres := []genre.Genre{{ID: 6, Name: "Soul"}}
	mockRepo.On("GetTopGenresInEvening", 10).Return(expectedGenres, nil)

	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "admin-secret")

	req := httptest.NewRequest(http.MethodGet, "/genres/top/evening?limit=10", nil)
	rec := httptest.NewRecorder()

	handler.GetTopGenresInEvening(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var genres []genre.Genre
	err := json.NewDecoder(rec.Body).Decode(&genres)
	require.NoError(t, err)
	require.Equal(t, expectedGenres, genres)

	mockRepo.AssertExpectations(t)
}

func TestGetTopGenresInEvening_Error(t *testing.T) {
	mockRepo := &MockGenreRepo{}
	mockRepo.On("GetTopGenresInEvening", 10).Return(nil, errors.New("some error"))

	service := service.NewGenreService(mockRepo)
	handler := NewGenreHandler(service, "admin-secret")

	req := httptest.NewRequest(http.MethodGet, "/genres/top/evening?limit=10", nil)
	rec := httptest.NewRecorder()

	handler.GetTopGenresInEvening(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
