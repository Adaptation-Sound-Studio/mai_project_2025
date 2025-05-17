package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/fact"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockFactRepo struct {
	mock.Mock
}

func (m *MockFactRepo) Insert(f *fact.ListenFact) error {
	args := m.Called(f)
	return args.Error(0)
}

func TestCreateFact_Success(t *testing.T) {
	mockRepo := &MockFactRepo{}
	mockRepo.On("Insert", mock.AnythingOfType("*fact.ListenFact")).Return(nil)

	factService := service.NewFactService(mockRepo)
	handler := NewFactHandler(factService, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/facts", handler.CreateFact).Methods("POST")

	payload := fact.ListenFact{
		UserID:     1,
		SongID:     2,
		ArtistID:   4,
		AlbumID:    3,
		GenreID:    5,
		ListenedAt: time.Now(),
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/facts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", "admin-secret")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Contains(t, rec.Body.String(), `"user_id":1`)

	mockRepo.AssertExpectations(t)
}

func TestCreateFact_Failure(t *testing.T) {
	mockRepo := &MockFactRepo{}
	mockRepo.On("Insert", mock.AnythingOfType("*fact.ListenFact")).Return(errors.New("db error"))

	factService := service.NewFactService(mockRepo)
	handler := NewFactHandler(factService, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/facts", handler.CreateFact).Methods("POST")

	req := httptest.NewRequest(http.MethodPost, "/facts", bytes.NewReader([]byte(`{
		"user_id":1, "song_id":2, "artist_id":4, "album_id":3, "genre_id":5, "listened_at":"2024-05-16T00:00:00Z"
	}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", "admin-secret")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "Failed to store listen fact")
}
