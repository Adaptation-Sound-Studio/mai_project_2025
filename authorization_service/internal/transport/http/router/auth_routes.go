package router

import (
	handler "auth_service/internal/transport/http/handler"

	"github.com/gorilla/mux"
)

func RegisterAuthRoutes(r *mux.Router, h *handler.AuthHandler) {
	r.HandleFunc("/register", h.RegisterUser).Methods("POST")
	r.HandleFunc("/login", h.Login).Methods("POST")
	r.HandleFunc("/logout", h.Logout).Methods("POST")
}
