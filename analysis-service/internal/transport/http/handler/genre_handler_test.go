package handler

import (
	"bytes"
	"encoding/json"
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
	mockRepo := &MockGenreRepo{} // не нужен вызов On, потому что не дойдёт до него
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
