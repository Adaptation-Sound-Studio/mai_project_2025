package handler

import (
	"encoding/json"
	"net/http"
	"upload-service/internal/service"
)

type SearchHandler struct {
	searchSvc *service.SearchService
}

func NewSearchHandler(svc *service.SearchService) *SearchHandler {
	return &SearchHandler{searchSvc: svc}
}

func (h *SearchHandler) SearchSongs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "`q` parameter is required", http.StatusBadRequest)
		return
	}

	songs, err := h.searchSvc.SearchSongs(r.Context(), q)
	if err != nil {
		http.Error(w, "search error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songs)
}

func (h *SearchHandler) SearchSongsByArtist(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "`q` parameter is required", http.StatusBadRequest)
		return
	}

	songs, err := h.searchSvc.SearchSongsByArtist(r.Context(), q)
	if err != nil {
		http.Error(w, "search by artist error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songs)
}
