package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	appredis "upload-service/internal/infrastructure/redis"

	redislib "github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"

	"upload-service/internal/domain/model"
	"upload-service/internal/domain/request"
	"upload-service/internal/service"
)

type ArtistHandler struct {
	Service     *service.ArtistService
	RedisClient *redislib.Client
}

func NewArtistHandler(service *service.ArtistService, redisClient *redislib.Client) *ArtistHandler {
	return &ArtistHandler{
		Service:     service,
		RedisClient: redisClient,
	}
}

func (h *ArtistHandler) GetAllArtists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	artists, err := h.Service.GetAllArtists(r.Context())
	if err != nil {
		log.Printf("Ошибка при получении артистов: %v", err)
		http.Error(w, "Ошибка при получении артистов", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artists)
}

func (h *ArtistHandler) GetArtistByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		log.Printf("Некорректный ID артиста: %v", err)
		http.Error(w, "Некорректный ID артиста", http.StatusBadRequest)
		return
	}

	artist, err := h.Service.GetArtistByID(r.Context(), id)
	if err != nil {
		log.Printf("Ошибка при получении артиста с ID %d: %v", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if artist == nil {
		log.Printf("Артист с ID %d не найден", id)
		http.Error(w, "Артист не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artist)
}

func (h *ArtistHandler) RegisterArtist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.Header.Get("Authorization")
	if sessionID == "" {
		log.Printf("Отсутствует токен сессии")
		http.Error(w, "Отсутствует токен сессии", http.StatusUnauthorized)
		return
	}

	userIDStr, err := appredis.GetUserID(h.RedisClient, sessionID)
	if err != nil {
		log.Printf("Ошибка при получении user_id из Redis для сессии %s: %v", sessionID, err)
		http.Error(w, "Сессия недействительна или просрочена", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Printf("Ошибка при конвертации user_id %s в int64: %v", userIDStr, err)
		http.Error(w, "Некорректный user_id в Redis", http.StatusInternalServerError)
		return
	}

	var req request.RegisterArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Неверный формат запроса: %v", err)
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	artist := &model.Artist{
		Name:   req.Name,
		UserID: userID,
	}

	artistID, err := h.Service.RegisterArtist(r.Context(), artist, userID)
	if err != nil {
		log.Printf("Ошибка при регистрации артиста для user_id %d: %v", userID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := appredis.SetArtistID(h.RedisClient, userIDStr, artistID); err != nil {
		log.Printf("Ошибка при сохранении artist_id %d для user_id %s в Redis: %v", artistID, userIDStr, err)
		http.Error(w, "Ошибка при сохранении artist_id в Redis", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Артист успешно зарегистрирован"})
}

func (h *ArtistHandler) UpdateArtist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		log.Printf("Некорректный ID артиста для обновления: %v", err)
		http.Error(w, "Некорректный ID артиста", http.StatusBadRequest)
		return
	}

	var req request.UpdateArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Неверный формат запроса при обновлении артиста: %v", err)
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	artist := &model.Artist{
		ArtistID: id,
		Name:     req.Name,
	}

	if err := h.Service.UpdateArtist(r.Context(), artist); err != nil {
		log.Printf("Ошибка при обновлении артиста с ID %d: %v", id, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Артист успешно обновлён"})
}
