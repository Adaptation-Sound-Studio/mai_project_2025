package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/genre"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
	"github.com/gorilla/mux"
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

func (h *GenreHandler) GetTopGenresForUser(w http.ResponseWriter, r *http.Request) {
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

	genres, err := h.Service.GetTopGenresForUser(userID, limit)
	if err != nil {
		http.Error(w, "Failed to get top genres", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}

func (h *GenreHandler) GetMostPopularGenres(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	genres, err := h.Service.GetMostPopularGenres(limit)
	if err != nil {
		http.Error(w, "Failed to get popular genres", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}

func (h *GenreHandler) GetTopGenresAtNight(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	genres, err := h.Service.GetTopGenresAtNight(limit)
	if err != nil {
		http.Error(w, "Failed to get top genres at night", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}

func (h *GenreHandler) GetTopGenresInMorning(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	genres, err := h.Service.GetTopGenresInMorning(limit)
	if err != nil {
		http.Error(w, "Failed to get top genres in morning", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}

func (h *GenreHandler) GetTopGenresInDay(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	genres, err := h.Service.GetTopGenresInDay(limit)
	if err != nil {
		http.Error(w, "Failed to get top genres in day", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}

func (h *GenreHandler) GetTopGenresInEvening(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	genres, err := h.Service.GetTopGenresInEvening(limit)
	if err != nil {
		http.Error(w, "Failed to get top genres in evening", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}
