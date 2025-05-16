package app

import (
	"auth_service/internal/repository/postgres"
	"auth_service/internal/service"
	"auth_service/internal/transport/http/handler"
	"auth_service/internal/transport/http/router"
	"database/sql"

	"log"

	"github.com/gorilla/mux"
)

func BuildRouter(dbConn *sql.DB) (*mux.Router, error) {

	// Инициализация зависимостей
	userRepo := postgres.NewUserRepository(dbConn)
	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	// Создание роутера
	mux_router := router.NewRouter()

	// Регистрация маршрутов по сущностям
	router.RegisterAuthRoutes(mux_router, authHandler)
	router.RegisterUserRoutes(mux_router, userHandler)

	log.Println("Маршруты успешно зарегистрированы")
	return mux_router, nil
}
