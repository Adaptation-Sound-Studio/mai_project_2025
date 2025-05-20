package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"upload-service/internal/domain/model"
	"upload-service/internal/domain/request"
	"upload-service/internal/service"

	"github.com/gorilla/mux"
)

type GenreHandler struct {
	Service *service.GenreService
}

func NewGenreHandler(service *service.GenreService) *GenreHandler {
	return &GenreHandler{Service: service}
}

func (h *GenreHandler) GetAllGenres(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	genres, err := h.Service.GetAllGenres(r.Context())
	if err != nil {
		log.Printf("Ошибка при получении жанров: %v", err)
		http.Error(w, "Ошибка при получении жанров", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}

func (h *GenreHandler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	var req request.CreateGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Неверный формат запроса: %v", err)
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	genre := &model.Genre{
		Name: req.Name,
	}

	genreID, err := h.Service.CreateGenre(r.Context(), genre)
	if err != nil {
		log.Printf("Ошибка при создании жанра: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Жанр успешно создан",
		"genre_id": strconv.FormatInt(genreID, 10),
	})
}

func (h *GenreHandler) UpdateGenre(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	idStr := mux.Vars(r)["genre_id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		log.Printf("Некорректный ID жанра для обновления: %v", err)
		http.Error(w, "Некорректный ID жанра", http.StatusBadRequest)
		return
	}

	var req request.UpdateGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Неверный формат запроса при обновлении жанра: %v", err)
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	genre := &model.Genre{
		GenreID: id,
		Name:    req.Name,
	}

	if err := h.Service.UpdateGenre(r.Context(), genre); err != nil {
		log.Printf("Ошибка при обновлении жанра с ID %d: %v", id, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Жанр успешно обновлён"})
}
