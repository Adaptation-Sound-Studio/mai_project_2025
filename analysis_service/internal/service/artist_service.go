package service

import (
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/artist"
)

type ArtistService struct {
	Repo artist.ArtistRepository
}

func NewArtistService(repo artist.ArtistRepository) *ArtistService {
	return &ArtistService{Repo: repo}
}

func (s *ArtistService) CreateArtist(artist *artist.Artist) error {
	return s.Repo.Create(artist)
}

func (s *ArtistService) GetTopArtistsForUser(userID int, limit int) ([]artist.Artist, error) {
	return s.Repo.GetTopArtistsForUser(userID, limit)
}

func (s *ArtistService) GetMostPopularArtists(limit int) ([]artist.Artist, error) {
	return s.Repo.GetMostPopularArtists(limit)
}
