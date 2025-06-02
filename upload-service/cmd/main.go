package main

import (
	"log"
	"net/http"
	"strings"

	"upload-service/internal/config"
	dbinfra "upload-service/internal/infrastructure/db"
	"upload-service/internal/infrastructure/elastic"
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

	elasticClient, err := elastic.NewElasticClient(cfg.Elastic)
	if err != nil {
		log.Fatalf("Ошибка подключения к Elasticsearch: %v", err)
	}

	if err := elastic.CreateAllIndices(elasticClient); err != nil {
		log.Printf("Ошибка при создании индексов: %v", err)
	}

	genreRepo := repository.NewGenreRepository(db)
	artistRepo := repository.NewArtistRepository(db)
	songRepo := repository.NewSongRepository(db)
	albumRepo := repository.NewAlbumRepository(db)

	genreService := service.NewGenreService(genreRepo, kafkaProducer, elasticClient)
	artistService := service.NewArtistService(artistRepo, cfg.AuthService.URL, cfg.AuthService.APIKey, kafkaProducer, elasticClient)
	songService := service.NewSongService(songRepo, minioClient, minioBucket, kafkaProducer, elasticClient)
	albumService := service.NewAlbumService(db, albumRepo, elasticClient)

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
