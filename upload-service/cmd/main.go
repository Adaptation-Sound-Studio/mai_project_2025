package main

import (
	"context"
	"fmt"
	"log"
	"upload-service/internal/infrastructure/db"
	"upload-service/internal/repository"
)

func main() {
	database, err := db.NewPostgresDB()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	songRepo := repository.NewSongRepository(database)
	songs, err := songRepo.GetAllSongs(ctx)
	if err != nil {
		log.Fatalf("Ошибка при получении песен: %v", err)
	}
	fmt.Println("Все песни:", songs)

	albumRepo := repository.NewAlbumRepository(database)
	albums, err := albumRepo.GetAllAlbums(ctx)
	if err != nil {
		log.Fatalf("Ошибка при получении альбомов: %v", err)
	}
	fmt.Println("Все альбомы:", albums)

	artistRepo := repository.NewArtistRepository(database)
	artists, err := artistRepo.GetAllArtists(ctx)
	if err != nil {
		log.Fatalf("Ошибка при получении артистов: %v", err)
	}
	fmt.Println("Все артисты:", artists)

	genreRepo := repository.NewGenreRepository(database)
	genres, err := genreRepo.GetAllGenres(ctx)
	if err != nil {
		log.Fatalf("Ошибка при получении жанров: %v", err)
	}
	fmt.Println("Все жанры:", genres)
}
