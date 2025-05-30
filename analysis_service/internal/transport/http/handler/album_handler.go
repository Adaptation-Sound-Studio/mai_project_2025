package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/album"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
)

type AlbumHandler struct {
	Service     *service.AlbumService
	AdminSecret string
}

func NewAlbumHandler(service *service.AlbumService, adminSecret string) *AlbumHandler {
	return &AlbumHandler{Service: service, AdminSecret: adminSecret}
}

func (h *AlbumHandler) CreateAlbum(w http.ResponseWriter, r *http.Request) {
	// Получаем ключ админа через фронтенд-заголовок
	if r.Header.Get("X-Admin-Key") != h.AdminSecret {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var album album.Album
	if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.Service.CreateAlbum(&album); err != nil {
		http.Error(w, "Failed to create album", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(album)
}
