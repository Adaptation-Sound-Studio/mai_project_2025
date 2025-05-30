package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/artist"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
)

type ArtistHandler struct {
	Service     *service.ArtistService
	AdminSecret string
}

func NewArtistHandler(s *service.ArtistService, secret string) *ArtistHandler {
	return &ArtistHandler{Service: s, AdminSecret: secret}
}

func (h *ArtistHandler) CreateArtist(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Admin-Key") != h.AdminSecret {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var artist artist.Artist
	if err := json.NewDecoder(r.Body).Decode(&artist); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := h.Service.CreateArtist(&artist); err != nil {
		http.Error(w, "Failed to create artist", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(artist)
}
