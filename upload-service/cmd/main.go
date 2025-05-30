package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"upload-service/internal/config"
	dbinfra "upload-service/internal/infrastructure/db"
	redisinfra "upload-service/internal/infrastructure/redis"
	"upload-service/internal/repository"
	"upload-service/internal/service"
	router "upload-service/internal/transport/http"
	"upload-service/internal/transport/http/handler"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден — используем переменные окружения")
	}

	cfg := config.LoadConfig()

	db, err := dbinfra.NewPostgresConnection(cfg.DB)
	if err != nil {
		log.Fatalf("Ошибка подключения к Postgres: %v", err)
	}
	defer db.Close()

	redisClient := redisinfra.NewRedisClient(cfg.Redis)
	defer redisClient.Close()

	genreRepo := repository.NewGenreRepository(db)
	artistRepo := repository.NewArtistRepository(db)
	songRepo := repository.NewSongRepository(db)
	albumRepo := repository.NewAlbumRepository(db)

	genreService := service.NewGenreService(genreRepo)
	artistService := service.NewArtistService(artistRepo)
	songService := service.NewSongService(songRepo)
	albumService := service.NewAlbumService(db, albumRepo)

	genreHandler := handler.NewGenreHandler(genreService)
	artistHandler := handler.NewArtistHandler(artistService, redisClient)
	songHandler := handler.NewSongHandler(songService, redisClient)
	albumHandler := handler.NewAlbumHandler(albumService, redisClient)

	router := router.NewRouter(
		genreHandler,
		artistHandler,
		songHandler,
		albumHandler,
		redisClient,
	)

	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("Сервер запущен на порту %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Ошибка запуска HTTP-сервера: %v", err)
	}
}
