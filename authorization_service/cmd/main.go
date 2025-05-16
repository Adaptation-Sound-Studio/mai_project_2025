package main

import (
	"auth_service/internal/app"
	"auth_service/internal/config"
	"auth_service/internal/infrastructure/db"
	"log"
	nethttp "net/http"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден")
	}

	cfg := config.LoadConfig()

	dbConn, err := db.NewPostgresConnection(cfg.DB)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer dbConn.Close()

	router, err := app.BuildRouter(dbConn)
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}

	log.Println("Сервер запущен")
	if err := nethttp.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
