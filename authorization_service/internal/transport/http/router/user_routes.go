package router

import (
	handler "auth_service/internal/transport/http/handler"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(r *mux.Router, h *handler.UserHandler, sessionMiddleware func(http.Handler) http.Handler, roleMiddleware func(allowedRoles ...string) func(http.Handler) http.Handler) {
	r.Handle("/users/me", sessionMiddleware(http.HandlerFunc(h.GetCurrentUser))).Methods("GET")
	r.Handle("/users/me", sessionMiddleware(http.HandlerFunc(h.UpdateCurrentUser))).Methods("POST")
	r.Handle("/users/{user_id}", sessionMiddleware(roleMiddleware("admin")(http.HandlerFunc(h.GetUserByID)))).Methods("GET")
	r.Handle("/users/{user_id}", sessionMiddleware(roleMiddleware("admin")(http.HandlerFunc(h.UpdateUserByID)))).Methods("POST")
	r.Handle("/users/{user_id}", sessionMiddleware(roleMiddleware("admin")(http.HandlerFunc(h.SoftDeleteUserByID)))).Methods("DELETE")

	r.HandleFunc("/users", h.CreateUser).Methods("POST")
}
