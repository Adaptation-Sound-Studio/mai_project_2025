package service

import "github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/genre"

type GenreService struct {
	Repo genre.GenreRepository
}

func NewGenreService(repo genre.GenreRepository) *GenreService {
	return &GenreService{Repo: repo}
}

func (s *GenreService) CreateGenre(genre *genre.Genre) error {
	return s.Repo.Create(genre)
}
