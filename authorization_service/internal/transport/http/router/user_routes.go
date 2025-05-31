package router

import (
	handler "auth_service/internal/transport/http/handler"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(r *mux.Router, h *handler.UserHandler, sessionMiddleware func(http.Handler) http.Handler) {
	r.HandleFunc("/users/me", h.GetCurrentUser).Methods("GET")
	r.HandleFunc("/users/me", h.UpdateCurrentUser).Methods("POST")
	r.Handle("/users/{user_id}", sessionMiddleware(http.HandlerFunc(h.GetUserByID))).Methods("GET")
	r.HandleFunc("/users/{user_id}", h.UpdateUserByID).Methods("POST")
	r.HandleFunc("/users", h.CreateUser).Methods("POST")
}
