package main

import (
	"auth_service/internal/app"
	"auth_service/internal/config"
	"auth_service/internal/infrastructure/db"
	"log"
	nethttp "net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден")
	}

	cfg := config.LoadConfig()

	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	dbConn, err := db.NewPostgresConnection(cfg.DB)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer dbConn.Close()

	router, err := app.BuildRouter(dbConn, redisClient)
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}

	log.Println("Сервер запущен")
	if err := nethttp.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
