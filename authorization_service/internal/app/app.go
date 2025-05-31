package app

import (
	"auth_service/internal/repository/postgres"
	redisrepo "auth_service/internal/repository/redis"
	"auth_service/internal/service"
	"auth_service/internal/transport/http/handler"
	"auth_service/internal/transport/http/middleware"
	"auth_service/internal/transport/http/router"
	"database/sql"
	"time"

	"log"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func BuildRouter(dbConn *sql.DB, redisClient *redis.Client) (*mux.Router, error) {

	sessionRepo := redisrepo.NewSessionRepository(redisClient)
	sessionService := service.NewSessionService(sessionRepo)

	userRepo := postgres.NewUserRepository(dbConn)

	authService := service.NewAuthService(userRepo, sessionService, time.Hour*24)
	userService := service.NewUserService(userRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService, sessionService)
	log.Println("INFO: HTTP обработчики (AuthHandler, UserHandler) инициализированы")

	mux_router := router.NewRouter()

	router.RegisterAuthRoutes(mux_router, authHandler)
	authMiddleware := middleware.AuthMiddleware(sessionService)
	router.RegisterUserRoutes(mux_router, userHandler, authMiddleware)
	log.Println("INFO: Маршруты успешно зарегистрированы")

	return mux_router, nil
}
