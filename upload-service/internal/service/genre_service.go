package service

import (
	"context"
	"errors"
	"log"
	"upload-service/internal/domain"
	"upload-service/internal/repository"
)

type GenreService interface {
	GetAllGenres(ctx context.Context) ([]domain.Genre, error)
	CreateGenre(ctx context.Context, genre *domain.Genre) (int64, error)
	UpdateGenre(ctx context.Context, genre *domain.Genre) error
}

type genreService struct {
	repo repository.GenreRepository
}

func NewGenreService(r repository.GenreRepository) GenreService {
	return &genreService{repo: r}
}

func (s *genreService) GetAllGenres(ctx context.Context) ([]domain.Genre, error) {
	genres, err := s.repo.GetAllGenres(ctx)
	if err != nil {
		log.Printf("Ошибка при получении жанров: %v", err)
		return nil, err
	}
	return genres, nil
}

func (s *genreService) CreateGenre(ctx context.Context, genre *domain.Genre) (int64, error) {
	if genre.Name == "" {
		return 0, errors.New("название жанра не может быть пустым")
	}

	genreID, err := s.repo.CreateGenre(ctx, genre)
	if err != nil {
		log.Printf("Ошибка при создании жанра: %v", err)
		return 0, err
	}

	log.Printf("Жанр успешно создан с ID %d", genreID)
	return genreID, nil
}

func (s *genreService) UpdateGenre(ctx context.Context, genre *domain.Genre) error {
	if genre.GenreID <= 0 {
		return errors.New("некорректный ID жанра для обновления")
	}
	if genre.Name == "" {
		return errors.New("название жанра не может быть пустым")
	}

	err := s.repo.UpdateGenre(ctx, genre)
	if err != nil {
		log.Printf("Ошибка при обновлении жанра с ID %d: %v", genre.GenreID, err)
		return err
	}

	log.Printf("Жанр с ID %d успешно обновлён", genre.GenreID)
	return nil
}
