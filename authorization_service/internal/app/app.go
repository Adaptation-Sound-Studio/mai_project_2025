package app

import (
	"auth_service/internal/infrastructure/session"
	"auth_service/internal/repository/postgres"
	"auth_service/internal/service"
	"auth_service/internal/transport/http/handler"
	"auth_service/internal/transport/http/router"
	"database/sql"
	"time"

	"log"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func BuildRouter(dbConn *sql.DB, redisClient *redis.Client) (*mux.Router, error) {

	sessionManager := session.NewRedisSessionManager(redisClient)

	// Инициализация зависимостей
	userRepo := postgres.NewUserRepository(dbConn)

	authService := service.NewAuthService(userRepo, sessionManager, time.Hour*24)
	userService := service.NewUserService(userRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService, sessionManager)

	// Создание роутера
	mux_router := router.NewRouter()

	// Регистрация маршрутов по сущностям
	router.RegisterAuthRoutes(mux_router, authHandler)
	router.RegisterUserRoutes(mux_router, userHandler)

	log.Println("Маршруты успешно зарегистрированы")
	return mux_router, nil
}
