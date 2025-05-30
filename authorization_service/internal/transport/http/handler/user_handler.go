package handler

import (
	"auth_service/internal/domain/user"
	"auth_service/internal/infrastructure/session"
	"auth_service/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type UserHandler struct {
	Service        *service.UserService
	SessionManager *session.RedisSessionManager
}

func NewUserHandler(service *service.UserService, sessionManager *session.RedisSessionManager) *UserHandler {
	return &UserHandler{
		Service:        service,
		SessionManager: sessionManager,
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user user.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.Service.CreateUser(&user); err != nil {
		http.Error(w, "Ошибка при создании пользователя", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "Неверный user_id", http.StatusBadRequest)
		return
	}

	user, err := h.Service.GetUserByID(id)
	if err != nil {
		http.Error(w, "Ошибка при получении пользователя", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "Неверный user_id", http.StatusBadRequest)
		return
	}

	var user user.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	user.ID = id

	if err := h.Service.UpdateUser(&user); err != nil {
		http.Error(w, "Ошибка при обновлении пользователя", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Session ID not found in cookies", http.StatusUnauthorized)
		return
	}
	sessionID := cookie.Value

	userID, err := h.SessionManager.GetUserIDFromSession(ctx, sessionID)

	if err != nil {
		http.Error(w, "Сессия не найдена или истекла", http.StatusUnauthorized)
		return
	}

	user, err := h.Service.GetUserByID(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении пользователя", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Session ID not found in cookies", http.StatusUnauthorized)
		return
	}
	sessionID := cookie.Value

	userID, err := h.SessionManager.GetUserIDFromSession(ctx, sessionID)

	if err != nil {
		http.Error(w, "Сессия не найдена или истекла", http.StatusUnauthorized)
		return
	}

	var u user.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Неверный формат", http.StatusBadRequest)
		return
	}

	u.ID = userID

	if err := h.Service.UpdateUser(&u); err != nil {
		http.Error(w, "Ошибка при обновлении", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(u)
}
