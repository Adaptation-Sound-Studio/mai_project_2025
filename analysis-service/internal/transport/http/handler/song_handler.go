package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/song"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
)

type SongHandler struct {
	Service     *service.SongService
	AdminSecret string
}

func NewSongHandler(service *service.SongService, secret string) *SongHandler {
	return &SongHandler{Service: service, AdminSecret: secret}
}

func (h *SongHandler) CreateSong(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Admin-Key") != h.AdminSecret {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var song song.Song
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.Service.CreateSong(&song); err != nil {
		http.Error(w, "Failed to create song", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(song)
}

func (h *SongHandler) GetPopularSongs(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Admin-Key") != h.AdminSecret {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	songs, err := h.Service.GetPopularSongs(limit)
	if err != nil {
		http.Error(w, "Failed to fetch popular songs", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(songs)
}
