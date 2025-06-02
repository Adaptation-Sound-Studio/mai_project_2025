package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	appredis "upload-service/internal/infrastructure/redis"
	auth "upload-service/internal/transport/http/helper"

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

	idStr := mux.Vars(r)["artist_id"]
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

	userID, userIDStr, err := auth.GetUserID(r, h.RedisClient)
	if err != nil {
		http.Error(w, err.Error(), err.(*auth.HttpError).Code)
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

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "missing session cookie", http.StatusUnauthorized)
		return
	}
	sessionID := cookie.Value

	artistIDStr := strconv.FormatInt(artistID, 10)
	if err := appredis.SetArtistID(h.RedisClient, sessionID, artistIDStr); err != nil {
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

	idStr := mux.Vars(r)["artist_id"]
	artistIDParam, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || artistIDParam <= 0 {
		log.Printf("Некорректный ID артиста для обновления: %v", err)
		http.Error(w, "Некорректный ID артиста", http.StatusBadRequest)
		return
	}

	currentArtistID, err := auth.GetCurrentArtistID(r, h.RedisClient)
	if err != nil {
		http.Error(w, err.Error(), err.(*auth.HttpError).Code)
		return
	}

	if artistIDParam != currentArtistID {
		http.Error(w, "Вы не можете редактировать другого артиста", http.StatusForbidden)
		return
	}

	var req request.UpdateArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Неверный формат запроса при обновлении артиста: %v", err)
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	artist := &model.Artist{
		ArtistID: artistIDParam,
		Name:     req.Name,
	}

	if err := h.Service.UpdateArtist(r.Context(), artist); err != nil {
		log.Printf("Ошибка при обновлении артиста с ID %d: %v", artistIDParam, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Артист успешно обновлён"})
}

func (h *ArtistHandler) GetArtistIDByUserID(w http.ResponseWriter, r *http.Request) {
	userIDStr := mux.Vars(r)["user_id"]

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user_id: "+err.Error(), http.StatusBadRequest)
		return
	}

	artistID, err := h.Service.GetArtistIDByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "error fetching artist ID: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if artistID == "" {
		http.Error(w, "artist not found for user_id "+userIDStr, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"artist_id": artistID})
}

func (h *ArtistHandler) SearchArtists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "Параметр q обязателен", http.StatusBadRequest)
		return
	}

	artists, err := h.Service.SearchArtists(r.Context(), q)
	if err != nil {
		log.Printf("Ошибка при поиске артистов: %v", err)
		http.Error(w, "Ошибка при поиске артистов", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artists)
}
