package main

import (
	"auth_service/internal/app"
	"auth_service/internal/config"
	"auth_service/internal/infrastructure/db"
	"auth_service/internal/infrastructure/session"
	"log"
	nethttp "net/http"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	cfg := config.LoadConfig()
	log.Printf("INFO: Загружена конфигурация: server port=%s, redis addr=%s", cfg.Server.Port, cfg.Redis.Host+":"+cfg.Redis.Port)

	redisClient := session.NewClient(cfg.Redis)
	log.Println("INFO: Redis клиент инициализирован")

	dbConn, err := db.NewPostgresConnection(cfg.DB)
	if err != nil {
		log.Fatalf("FATAL: Ошибка подключения к базе данных: %v", err)
	}

	defer func() {
		if err := dbConn.Close(); err != nil {
			log.Printf("ERROR: Ошибка закрытия подключения к БД: %v", err)
		} else {
			log.Println("INFO: Подключение к базе данных закрыто")
		}
	}()
	log.Println("INFO: Подключение к базе данных успешно")

	router, err := app.BuildRouter(dbConn, redisClient)
	if err != nil {
		log.Fatalf("FATAL: Ошибка инициализации роутера: %v", err)
	}

	log.Println("INFO: Роутер успешно построен")
	if err := nethttp.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
		log.Printf("ERROR: Ошибка запуска сервера: %v", err)
		os.Exit(1)
	}
}
