package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/artist"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockArtistRepo struct {
	mock.Mock
}

func (m *MockArtistRepo) Create(a *artist.Artist) error {
	args := m.Called(a)
	return args.Error(0)
}

func TestCreateArtist_Success(t *testing.T) {
	mockRepo := &MockArtistRepo{}
	mockRepo.On("Create", mock.AnythingOfType("*artist.Artist")).Return(nil)

	artistService := service.NewArtistService(mockRepo)
	handler := NewArtistHandler(artistService, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/artists", handler.CreateArtist).Methods("POST")

	input := artist.Artist{Name: "Vlad"}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/artists", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", "admin-secret")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Contains(t, rec.Body.String(), "Vlad")

	mockRepo.AssertExpectations(t)
}

func TestCreateArtist_Forbidden(t *testing.T) {
	mockRepo := &MockArtistRepo{}
	artistService := service.NewArtistService(mockRepo)
	handler := NewArtistHandler(artistService, os.Getenv("ADMIN_SECRET"))

	router := mux.NewRouter()
	router.HandleFunc("/artists", handler.CreateArtist).Methods("POST")

	input := artist.Artist{Name: "Unknown"}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/artists", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", "wrong-secret")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Contains(t, rec.Body.String(), "Forbidden")

	mockRepo.AssertExpectations(t)
}
