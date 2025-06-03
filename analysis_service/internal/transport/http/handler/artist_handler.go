package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/artist"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
	"github.com/gorilla/mux"
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
func (h *ArtistHandler) GetTopArtistsForUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["user_id"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	artists, err := h.Service.GetTopArtistsForUser(userID, limit)
	if err != nil {
		http.Error(w, "Failed to get top artists", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artists)
}

func (h *ArtistHandler) GetMostPopularArtists(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	artists, err := h.Service.GetMostPopularArtists(limit)
	if err != nil {
		http.Error(w, "Failed to get popular artists", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artists)
}
