package service

import "github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/song"

type SongService struct {
	Repo song.SongRepository
}

func NewSongService(repo song.SongRepository) *SongService {
	return &SongService{Repo: repo}
}

func (s *SongService) CreateSong(song *song.Song) error {
	return s.Repo.Create(song)
}

func (s *SongService) GetPopularSongs(limit int) ([]song.PopularSong, error) {
	return s.Repo.GetPopularSongs(limit)
}
func (s *SongService) GetTopSongsForUser(userID int, limit int) ([]song.Song, error) {
	return s.Repo.GetTopSongsForUser(userID, limit)
}
func (s *SongService) GetMostPopularSongs(limit int) ([]song.Song, error) {
	return s.Repo.GetMostPopularSongs(limit)
}
