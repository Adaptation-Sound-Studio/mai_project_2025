package router

import (
	handler "auth_service/internal/transport/http/handler"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(r *mux.Router, h *handler.UserHandler) {
	r.HandleFunc("/users/me", h.GetCurrentUser).Methods("GET")
	r.HandleFunc("/users/me", h.UpdateCurrentUser).Methods("POST")
	r.HandleFunc("/users/{user_id}", h.GetUserByID).Methods("GET")
	r.HandleFunc("/users/{user_id}", h.UpdateUserByID).Methods("POST")
	r.HandleFunc("/users", h.CreateUser).Methods("POST")
}
