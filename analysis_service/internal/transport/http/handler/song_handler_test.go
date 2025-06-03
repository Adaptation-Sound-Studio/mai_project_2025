package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/song"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSongRepo struct {
	mock.Mock
}

func (m *MockSongRepo) Create(s *song.Song) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *MockSongRepo) GetPopularSongs(limit int) ([]song.PopularSong, error) {
	args := m.Called(limit)
	return args.Get(0).([]song.PopularSong), args.Error(1)
}

func (m *MockSongRepo) GetTopSongsForUser(userID int, limit int) ([]song.Song, error) {
	args := m.Called(userID, limit)
	return args.Get(0).([]song.Song), args.Error(1)
}

func (m *MockSongRepo) GetMostPopularSongs(limit int) ([]song.Song, error) {
	args := m.Called(limit)
	return args.Get(0).([]song.Song), args.Error(1)
}
func TestCreateSong_Success(t *testing.T) {
	mockRepo := &MockSongRepo{}
	mockRepo.On("Create", mock.AnythingOfType("*song.Song")).Return(nil)

	service := service.NewSongService(mockRepo)
	handler := NewSongHandler(service, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/songs", handler.CreateSong).Methods("POST")

	input := song.Song{Name: "My Song"}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/songs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", "admin-secret")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Contains(t, rec.Body.String(), "My Song")

	mockRepo.AssertExpectations(t)
}

func TestCreateSong_Failure(t *testing.T) {
	mockRepo := &MockSongRepo{}
	service := service.NewSongService(mockRepo)
	handler := NewSongHandler(service, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/songs", handler.CreateSong).Methods("POST")

	req := httptest.NewRequest(http.MethodPost, "/songs", bytes.NewReader([]byte(`not-json`)))
	req.Header.Set("X-Admin-Key", "admin-secret")
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "Invalid JSON")
}

func TestGetSongsAdmin_Success(t *testing.T) {
	mockRepo := &MockSongRepo{}
	mockRepo.On("GetPopularSongs", 5).Return([]song.PopularSong{
		{ID: 1, Name: "Top Hit", Artist: "I", Genre: "Pop", Listens: 123},
	}, nil)

	service := service.NewSongService(mockRepo)
	handler := NewSongHandler(service, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/songs/popular", handler.GetPopularSongs).Methods("GET")

	req := httptest.NewRequest(http.MethodGet, "/songs/popular?limit=5", nil)
	req.Header.Set("X-Admin-Key", "admin-secret")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Top Hit")

	mockRepo.AssertExpectations(t)
}

func TestGetSongsAdmin_Failure(t *testing.T) {
	mockRepo := &MockSongRepo{}
	mockRepo.On("GetPopularSongs", 5).Return([]song.PopularSong(nil), errors.New("db error"))

	service := service.NewSongService(mockRepo)
	handler := NewSongHandler(service, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/songs/popular", handler.GetPopularSongs).Methods("GET")

	req := httptest.NewRequest(http.MethodGet, "/songs/popular?limit=5", nil)
	req.Header.Set("X-Admin-Key", "admin-secret")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "Failed to fetch popular songs")
}

func TestGetTopSongsForUser_Success(t *testing.T) {
	mockRepo := &MockSongRepo{}
	userID := 42
	limit := 3
	mockRepo.On("GetTopSongsForUser", userID, limit).Return([]song.Song{
		{ID: 1, Name: "Song One"},
		{ID: 2, Name: "Song Two"},
	}, nil)

	service := service.NewSongService(mockRepo)
	handler := NewSongHandler(service, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/users/{user_id}/top-songs", handler.GetTopSongsForUser).Methods("GET")

	req := httptest.NewRequest(http.MethodGet, "/users/42/top-songs?limit=3", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Song One")

	mockRepo.AssertExpectations(t)
}

func TestGetTopSongsForUser_Failure(t *testing.T) {
	mockRepo := &MockSongRepo{}
	userID := 42
	limit := 3
	mockRepo.On("GetTopSongsForUser", userID, limit).Return([]song.Song(nil), errors.New("db error"))

	service := service.NewSongService(mockRepo)
	handler := NewSongHandler(service, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/users/{user_id}/top-songs", handler.GetTopSongsForUser).Methods("GET")

	req := httptest.NewRequest(http.MethodGet, "/users/42/top-songs?limit=3", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "Failed to get top songs")

	mockRepo.AssertExpectations(t)
}

func TestGetMostPopularSongs_Success(t *testing.T) {
	mockRepo := &MockSongRepo{}
	limit := 10
	mockRepo.On("GetMostPopularSongs", limit).Return([]song.Song{
		{ID: 1, Name: "Popular Song"},
		{ID: 2, Name: "Another Hit"},
	}, nil)

	service := service.NewSongService(mockRepo)
	handler := NewSongHandler(service, "admin-secret")

	req := httptest.NewRequest(http.MethodGet, "/songs/popular?limit=10", nil)
	rec := httptest.NewRecorder()

	handler.GetMostPopularSongs(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Popular Song")

	mockRepo.AssertExpectations(t)
}

func TestGetMostPopularSongs_Failure(t *testing.T) {
	mockRepo := &MockSongRepo{}
	limit := 10
	mockRepo.On("GetMostPopularSongs", limit).Return([]song.Song(nil), errors.New("db error"))

	service := service.NewSongService(mockRepo)
	handler := NewSongHandler(service, "admin-secret")

	req := httptest.NewRequest(http.MethodGet, "/songs/popular?limit=10", nil)
	rec := httptest.NewRecorder()

	handler.GetMostPopularSongs(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "Failed to get popular songs")

	mockRepo.AssertExpectations(t)
}
