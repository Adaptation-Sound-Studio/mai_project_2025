package router

import (
	handler "auth_service/internal/transport/http/handler"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(r *mux.Router, h *handler.UserHandler) {
	r.HandleFunc("/users", h.CreateUser).Methods("POST")
}
