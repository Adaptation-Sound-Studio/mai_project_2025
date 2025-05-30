package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/album"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockAlbumRepo struct {
	mock.Mock
}

func (m *MockAlbumRepo) Create(a *album.Album) error {
	args := m.Called(a)
	return args.Error(0)
}

func TestCreateAlbum_Success(t *testing.T) {

	mockRepo := &MockAlbumRepo{}
	mockRepo.On("Create", mock.AnythingOfType("*album.Album")).Return(nil)

	albumService := service.NewAlbumService(mockRepo)
	handler := NewAlbumHandler(albumService, os.Getenv("ADMIN_SECRET"))

	router := mux.NewRouter()
	router.HandleFunc("/albums", handler.CreateAlbum).Methods("POST")

	input := album.Album{Name: "Test Album"}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/albums", bytes.NewReader(body))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", os.Getenv("ADMIN_SECRET"))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Contains(t, rec.Body.String(), "Test Album")

	mockRepo.AssertExpectations(t)
}

func TestCreateAlbum_ServiceError(t *testing.T) {

	mockRepo := &MockAlbumRepo{}
	mockRepo.
		On("Create", mock.AnythingOfType("*album.Album")).
		Return(errors.New("DB failure"))

	albumService := service.NewAlbumService(mockRepo)
	handler := NewAlbumHandler(albumService, os.Getenv("ADMIN_SECRET"))

	router := mux.NewRouter()
	router.HandleFunc("/albums", handler.CreateAlbum).Methods("POST")

	input := album.Album{Name: "Bad Album"}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/albums", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", os.Getenv("ADMIN_SECRET"))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "Failed to create album")

	mockRepo.AssertExpectations(t)
}

func TestCreateAlbum_BadJSON(t *testing.T) {
	mockRepo := &MockAlbumRepo{}
	service := service.NewAlbumService(mockRepo)
	handler := NewAlbumHandler(service, os.Getenv("ADMIN_SECRET"))

	router := mux.NewRouter()
	router.HandleFunc("/albums", handler.CreateAlbum).Methods("POST")

	req := httptest.NewRequest(http.MethodPost, "/albums", bytes.NewReader([]byte(`not-json`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", os.Getenv("ADMIN_SECRET"))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "Invalid JSON")
}

func TestCreateAlbum_Forbidden_NoHeader(t *testing.T) {
	mockRepo := &MockAlbumRepo{}
	service := service.NewAlbumService(mockRepo)
	handler := NewAlbumHandler(service, "admin-secret")

	router := mux.NewRouter()
	router.HandleFunc("/albums", handler.CreateAlbum).Methods("POST")

	input := album.Album{Name: "No Admin Album"}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/albums", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Не устанавливаем ключ админа

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Contains(t, rec.Body.String(), "Forbidden")
	mockRepo.AssertNotCalled(t, "Create", mock.Anything)
}
