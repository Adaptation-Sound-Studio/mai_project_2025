package main

import (
	"context"
	"log"
	nethttp "net/http"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/config"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/infrastructure/db"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/kafka"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/repository/postgres"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/transport/http"
)

func main() {

	cfg := config.LoadConfig()

	dbConn, err := db.NewPostgresConnection(&cfg.DB)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer dbConn.Close()

	FactRepo := postgres.NewFactRepo(dbConn)
	FactService := service.NewFactService(FactRepo)

	ArtistRepo := postgres.NewArtistRepo(dbConn)
	ArtistService := service.NewArtistService(ArtistRepo)

	SongRepo := postgres.NewSongRepo(dbConn)
	SongService := service.NewSongService(SongRepo)

	GenreRepo := postgres.NewGenreRepo(dbConn)
	GenreService := service.NewGenreService(GenreRepo)

	consumer := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, FactService, ArtistService, SongService, GenreService)
	go consumer.Start(context.Background())

	router := http.NewRouter(dbConn, cfg.Admin.Secret)

	log.Println("Сервер запущен на порту 8081")
	if err := nethttp.ListenAndServe(":8081", router); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
