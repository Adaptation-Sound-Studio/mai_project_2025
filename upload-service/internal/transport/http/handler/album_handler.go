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
	auth "upload-service/internal/transport/http/helper"

	redislib "github.com/go-redis/redis/v8"
)

type AlbumHandler struct {
	Service     *service.AlbumService
	RedisClient *redislib.Client
}

func NewAlbumHandler(service *service.AlbumService, redisClient *redislib.Client) *AlbumHandler {
	return &AlbumHandler{
		Service:     service,
		RedisClient: redisClient,
	}
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

	idStr := mux.Vars(r)["album_id"]
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
	currentArtistID, err := auth.GetCurrentArtistID(r, h.RedisClient)
	if err != nil {
		http.Error(w, err.Error(), err.(*auth.HttpError).Code)
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
		GenreID:  req.GenreID,
		ArtistID: currentArtistID,
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
	currentArtistID, err := auth.GetCurrentArtistID(r, h.RedisClient)
	if err != nil {
		http.Error(w, err.Error(), err.(*auth.HttpError).Code)
		return
	}

	idStr := mux.Vars(r)["album_id"]
	albumID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || albumID <= 0 {
		log.Printf("Некорректный ID альбома: %v", err)
		http.Error(w, "Некорректный ID альбома", http.StatusBadRequest)
		return
	}

	albumResp, err := h.Service.GetAlbumByID(r.Context(), albumID)
	if err != nil {
		log.Printf("Ошибка при получении альбома: %v", err)
		http.Error(w, "Ошибка при получении альбома", http.StatusInternalServerError)
		return
	}
	if albumResp == nil || albumResp.ArtistID != currentArtistID {
		http.Error(w, "Вы не можете редактировать этот альбом", http.StatusForbidden)
		return
	}

	var req request.UpdateAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Неверный формат запроса: %v", err)
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	album := &model.Album{
		AlbumID:  albumID,
		Name:     req.Name,
		GenreID:  req.GenreID,
		ArtistID: currentArtistID,
	}

	if err := h.Service.UpdateAlbum(r.Context(), album); err != nil {
		log.Printf("Ошибка при обновлении альбома: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Альбом успешно обновлён"})
}

func (h *AlbumHandler) SearchAlbums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	filters := map[string]string{
		"name":   r.URL.Query().Get("name"),
		"genre":  r.URL.Query().Get("genre"),
		"artist": r.URL.Query().Get("artist"),
	}

	albums, err := h.Service.SearchAlbums(r.Context(), filters)
	if err != nil {
		log.Printf("Ошибка при поиске альбома: %v", err)
		http.Error(w, "Ошибка при поиске альбома", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(albums)
}
