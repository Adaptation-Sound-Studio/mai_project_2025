package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"upload-service/internal/domain/model"
	"upload-service/internal/domain/request"
	"upload-service/internal/service"
)

type AlbumHandler struct {
	Service *service.AlbumService
}

func NewAlbumHandler(service *service.AlbumService) *AlbumHandler {
	return &AlbumHandler{Service: service}
}

func (h *AlbumHandler) GetAllAlbums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	albums, err := h.Service.GetAllAlbums(r.Context())
	if err != nil {
		log.Printf("Ошибка при получении альбомов: %v", err)
		http.Error(w, "Ошибка при получении альбомов", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(albums)
}

func (h *AlbumHandler) GetAlbumByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		log.Printf("Некорректный ID альбома: %v", err)
		http.Error(w, "Некорректный ID альбома", http.StatusBadRequest)
		return
	}

	album, err := h.Service.GetAlbumByID(r.Context(), id)
	if err != nil {
		log.Printf("Ошибка при получении альбома с ID %d: %v", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if album == nil {
		log.Printf("Альбом с ID %d не найден", id)
		http.Error(w, "Альбом не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(album)
}

func (h *AlbumHandler) CreateAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	var req request.CreateAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Неверный формат запроса: %v", err)
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	album := &model.Album{
		Name:     req.Name,
		ArtistID: req.ArtistID,
		GenreID:  req.GenreID,
	}

	albumID, err := h.Service.CreateAlbum(r.Context(), album, req.SongIDs)
	if err != nil {
		log.Printf("Ошибка при создании альбома: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Альбом успешно создан",
		"album_id": strconv.FormatInt(albumID, 10),
	})
}

func (h *AlbumHandler) UpdateAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		log.Printf("Некорректный ID альбома для обновления: %v", err)
		http.Error(w, "Некорректный ID альбома", http.StatusBadRequest)
		return
	}

	var req request.UpdateAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Неверный формат запроса при обновлении альбома: %v", err)
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	album := &model.Album{
		AlbumID:  id,
		Name:     req.Name,
		GenreID:  req.GenreID,
		ArtistID: req.ArtistID,
	}

	if err := h.Service.UpdateAlbum(r.Context(), album); err != nil {
		log.Printf("Ошибка при обновлении альбома с ID %d: %v", id, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Альбом успешно обновлён"})
}
