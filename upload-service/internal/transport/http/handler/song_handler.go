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

type SongHandler struct {
	Service *service.SongService
}

func NewSongHandler(service *service.SongService) *SongHandler {
	return &SongHandler{Service: service}
}

func (h *SongHandler) GetAllSongs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	songs, err := h.Service.GetAllSongs(r.Context())
	if err != nil {
		log.Printf("Ошибка при получении песен: %v", err)
		http.Error(w, `{"error": "Ошибка при получении песен"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(songs)
}

func (h *SongHandler) GetSongByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		log.Printf("Некорректный ID песни: %v", err)
		http.Error(w, `{"error": "Некорректный ID песни"}`, http.StatusBadRequest)
		return
	}

	song, err := h.Service.GetSongByID(r.Context(), id)
	if err != nil {
		log.Printf("Ошибка при получении песни с ID %d: %v", id, err)
		http.Error(w, `{"error": "Ошибка при получении песни"}`, http.StatusInternalServerError)
		return
	}
	if song == nil {
		log.Printf("Песня с ID %d не найдена", id)
		http.Error(w, `{"error": "Песня не найдена"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(song)
}

func (h *SongHandler) CreateSong(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req request.CreateSongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Неверный формат запроса: %v", err)
		http.Error(w, `{"error": "Неверный формат запроса"}`, http.StatusBadRequest)
		return
	}

	song := &model.Song{
		Name:    req.Name,
		GenreID: req.GenreID,
		Link:    req.Link,
	}

	if err := h.Service.CreateSong(r.Context(), song, req.ArtistIDs); err != nil {
		log.Printf("Ошибка при создании песни: %v", err)
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Песня успешно создана"})
}

func (h *SongHandler) UpdateSong(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		log.Printf("Некорректный ID песни для обновления: %v", err)
		http.Error(w, `{"error": "Некорректный ID песни"}`, http.StatusBadRequest)
		return
	}

	var req request.UpdateSongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Неверный формат запроса при обновлении песни: %v", err)
		http.Error(w, `{"error": "Неверный формат запроса"}`, http.StatusBadRequest)
		return
	}

	song := &model.Song{
		SongID:  id,
		Name:    req.Name,
		GenreID: req.GenreID,
		Link:    req.Link,
	}

	if err := h.Service.UpdateSong(r.Context(), song); err != nil {
		log.Printf("Ошибка при обновлении песни с ID %d: %v", id, err)
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Песня успешно обновлена"})
}
