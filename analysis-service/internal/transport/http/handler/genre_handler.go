package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/genre"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
)

type GenreHandler struct {
	Service     *service.GenreService
	AdminSecret string
}

func NewGenreHandler(service *service.GenreService, secret string) *GenreHandler {
	return &GenreHandler{Service: service, AdminSecret: secret}
}

func (h *GenreHandler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Admin-Key") != h.AdminSecret {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var genre genre.Genre
	if err := json.NewDecoder(r.Body).Decode(&genre); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.Service.CreateGenre(&genre); err != nil {
		http.Error(w, "Failed to create genre", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(genre)
}
