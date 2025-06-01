package main

import (
	"log"
	"net/http"
	"strings"

	"upload-service/internal/config"
	dbinfra "upload-service/internal/infrastructure/db"
	minioinfra "upload-service/internal/infrastructure/minio"
	redisinfra "upload-service/internal/infrastructure/redis"
	kafkapkg "upload-service/internal/kafka"
	"upload-service/internal/repository"
	"upload-service/internal/service"
	router "upload-service/internal/transport/http"
	"upload-service/internal/transport/http/handler"
)

func main() {
	cfg := config.LoadConfig()

	db, err := dbinfra.NewPostgresConnection(cfg.DB)
	if err != nil {
		log.Fatalf("Ошибка подключения к Postgres: %v", err)
	}
	defer db.Close()

	redisClient := redisinfra.NewRedisClient(cfg.Redis)

	minioClient, minioBucket, err := minioinfra.NewMinioClientFromEnv()
	if err != nil {
		log.Fatalf("Ошибка подключения к MinIO: %v", err)
	}

	kafkaProducer, err := kafkapkg.NewProducer(strings.Split(cfg.Kafka.Brokers, ","), cfg.Kafka.Topic)
	if err != nil {
		log.Fatalf("Ошибка подключения к Kafka: %v", err)
	}
	defer kafkaProducer.Close()

	genreRepo := repository.NewGenreRepository(db)
	artistRepo := repository.NewArtistRepository(db)
	songRepo := repository.NewSongRepository(db)
	albumRepo := repository.NewAlbumRepository(db)

	genreService := service.NewGenreService(genreRepo)
	artistService := service.NewArtistService(artistRepo, cfg.AuthService.URL, cfg.AuthService.APIKey)
	songService := service.NewSongService(songRepo, minioClient, minioBucket)
	albumService := service.NewAlbumService(db, albumRepo)

	genreHandler := handler.NewGenreHandler(genreService)
	artistHandler := handler.NewArtistHandler(artistService, redisClient)
	songHandler := handler.NewSongHandler(songService, minioClient, minioBucket, redisClient, kafkaProducer)

	albumHandler := handler.NewAlbumHandler(albumService, redisClient)

	router := router.NewRouter(
		genreHandler,
		artistHandler,
		songHandler,
		albumHandler,
		redisClient,
		cfg,
	)

	port := "8082"
	log.Printf("Сервер запущен на порту %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Ошибка запуска HTTP-сервера: %v", err)
	}
}
