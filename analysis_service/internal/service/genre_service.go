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

func (s *GenreService) GetTopGenresForUser(userID int, limit int) ([]genre.Genre, error) {
	return s.Repo.GetTopGenresForUser(userID, limit)
}

func (s *GenreService) GetMostPopularGenres(limit int) ([]genre.Genre, error) {
	return s.Repo.GetMostPopularGenres(limit)
}

func (s *GenreService) GetTopGenresAtNight(limit int) ([]genre.Genre, error) {
	return s.Repo.GetTopGenresAtNight(limit)
}

func (s *GenreService) GetTopGenresInMorning(limit int) ([]genre.Genre, error) {
	return s.Repo.GetTopGenresInMorning(limit)
}

func (s *GenreService) GetTopGenresInDay(limit int) ([]genre.Genre, error) {
	return s.Repo.GetTopGenresInDay(limit)
}

func (s *GenreService) GetTopGenresInEvening(limit int) ([]genre.Genre, error) {
	return s.Repo.GetTopGenresInEvening(limit)
}
