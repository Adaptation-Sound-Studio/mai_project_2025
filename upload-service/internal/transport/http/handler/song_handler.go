package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"upload-service/internal/domain/model"
	"upload-service/internal/domain/request"
	"upload-service/internal/service"
	auth "upload-service/internal/transport/http/helper"

	redislib "github.com/go-redis/redis/v8"
)

type SongHandler struct {
	Service     *service.SongService
	RedisClient *redislib.Client
}

func NewSongHandler(service *service.SongService, redisClient *redislib.Client) *SongHandler {
	return &SongHandler{Service: service, RedisClient: redisClient}
}

func (h *SongHandler) GetAllSongs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}
	songs, err := h.Service.GetAllSongs(r.Context())
	if err != nil {
		http.Error(w, "Ошибка при получении песен", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(songs)
}

func (h *SongHandler) GetSongByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["song_id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Некорректный ID песни", http.StatusBadRequest)
		return
	}

	song, err := h.Service.GetSongByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Ошибка при получении песни", http.StatusInternalServerError)
		return
	}
	if song == nil {
		http.Error(w, "Песня не найдена", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(song)
}

func (h *SongHandler) CreateSong(w http.ResponseWriter, r *http.Request) {
	currentArtistID, err := auth.GetCurrentArtistID(r, h.RedisClient)
	if err != nil {
		http.Error(w, err.Error(), err.(*auth.HttpError).Code)
		return
	}

	var req request.CreateSongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	artistSet := map[int64]struct{}{currentArtistID: {}}
	for _, id := range req.ArtistIDs {
		artistSet[id] = struct{}{}
	}
	var allArtistIDs []int64
	for id := range artistSet {
		allArtistIDs = append(allArtistIDs, id)
	}

	song := &model.Song{
		Name:    req.Name,
		GenreID: req.GenreID,
	}

	if err := h.Service.CreateSong(r.Context(), song, allArtistIDs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Песня успешно создана"})
}

func (h *SongHandler) UpdateSong(w http.ResponseWriter, r *http.Request) {
	currentArtistID, err := auth.GetCurrentArtistID(r, h.RedisClient)
	if err != nil {
		http.Error(w, err.Error(), err.(*auth.HttpError).Code)
		return
	}

	idStr := mux.Vars(r)["song_id"]
	songID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || songID <= 0 {
		http.Error(w, "Некорректный ID песни", http.StatusBadRequest)
		return
	}

	artists, err := h.Service.GetSongArtists(r.Context(), songID)
	if err != nil {
		http.Error(w, "Ошибка при проверке прав", http.StatusInternalServerError)
		return
	}

	allowed := false
	for _, a := range artists {
		if a.ArtistID == currentArtistID {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "Вы не можете редактировать эту песню", http.StatusForbidden)
		return
	}

	var req request.UpdateSongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	song := &model.Song{
		SongID:  songID,
		Name:    req.Name,
		GenreID: req.GenreID,
	}

	if err := h.Service.UpdateSong(r.Context(), song); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Песня успешно обновлена"})
}
