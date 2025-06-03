package app

import (
	"auth_service/internal/config"
	"auth_service/internal/repository/postgres"
	redisrepo "auth_service/internal/repository/redis"
	"auth_service/internal/service"
	"auth_service/internal/transport/http/handler"
	"auth_service/internal/transport/http/middleware"
	"auth_service/internal/transport/http/router"
	"database/sql"
	"net/http"
	"time"

	"log"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func BuildRouter(dbConn *sql.DB, redisClient *redis.Client, cfg *config.Config) (*mux.Router, error) {

	sessionRepo := redisrepo.NewSessionRepository(redisClient)
	sessionService := service.NewSessionService(sessionRepo)

	userRepo := postgres.NewUserRepository(dbConn)

	authService := service.NewAuthService(userRepo, sessionRepo, time.Hour*24, cfg.UplService.URL, cfg.UplService.APIKey)
	userService := service.NewUserService(userRepo, sessionService)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService, sessionService)
	log.Println("INFO: HTTP обработчики (AuthHandler, UserHandler) инициализированы")

	muxRouter := router.NewRouter()

	router.RegisterAuthRoutes(muxRouter, authHandler)

	authMiddleware := middleware.AuthMiddleware(sessionService)
	roleMiddleware := func(allowedRoles ...string) func(http.Handler) http.Handler {
		return middleware.RoleMiddleware(sessionService, allowedRoles...)
	}

	apiKey := cfg.ServiceApiKey
	apiKeyMiddleware := middleware.ApiKeyMiddleware(apiKey)

	router.RegisterUserRoutes(muxRouter, userHandler, authMiddleware, roleMiddleware, apiKeyMiddleware)
	log.Println("INFO: Маршруты успешно зарегистрированы")

	return muxRouter, nil
}
