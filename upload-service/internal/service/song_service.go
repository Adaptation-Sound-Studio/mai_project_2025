package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"upload-service/internal/domain"
	"upload-service/internal/domain/response"
	"upload-service/internal/repository"
)

var ErrSongNotFound = errors.New("песня не найдена")

type SongService interface {
	GetAllSongs(ctx context.Context) ([]response.SongResponse, error)
	GetSongByID(ctx context.Context, id int64) (*response.SongResponse, error)
	CreateSong(ctx context.Context, song *domain.Song, artistIDs []int64) error
	UpdateSong(ctx context.Context, song *domain.Song) error
}

type songService struct {
	repo repository.SongRepository
}

func NewSongService(r repository.SongRepository) SongService {
	return &songService{repo: r}
}

func (s *songService) GetAllSongs(ctx context.Context) ([]response.SongResponse, error) {
	songs, err := s.repo.GetAllSongs(ctx)
	if err != nil {
		log.Printf("Ошибка при получении песен: %v", err)
		return nil, err
	}

	var responses []response.SongResponse
	for _, song := range songs {
		artists, err := s.repo.GetArtistsBySongID(ctx, song.SongID)
		if err != nil {
			log.Printf("Ошибка при получении артистов для песни с ID %d: %v", song.SongID, err)
			return nil, err
		}

		album, err := s.repo.GetAlbumBySongID(ctx, song.SongID)
		if err != nil {
			log.Printf("Ошибка при получении альбома для песни с ID %d: %v", song.SongID, err)
			return nil, err
		}

		responses = append(responses, response.MapSongToResponse(&song, artists, album))
	}

	return responses, nil
}

func (s *songService) GetSongByID(ctx context.Context, id int64) (*response.SongResponse, error) {
	if id <= 0 {
		return nil, errors.New("ID песни должен быть положительным числом")
	}

	song, err := s.repo.GetSongByID(ctx, id)
	if err != nil {
		log.Printf("Ошибка при получении песни по ID %d: %v", id, err)
		return nil, err
	}
	if song == nil {
		return nil, ErrSongNotFound
	}

	artists, err := s.repo.GetArtistsBySongID(ctx, song.SongID)
	if err != nil {
		log.Printf("Ошибка при получении артистов для песни с ID %d: %v", song.SongID, err)
		return nil, err
	}

	album, err := s.repo.GetAlbumBySongID(ctx, song.SongID)
	if err != nil {
		log.Printf("Ошибка при получении альбома для песни с ID %d: %v", song.SongID, err)
		return nil, err
	}

	res := response.MapSongToResponse(song, artists, album)
	return &res, nil
}

func (s *songService) CreateSong(ctx context.Context, song *domain.Song, artistIDs []int64) error {
	if song.Name == "" {
		return errors.New("название песни не может быть пустым")
	}
	if song.GenreID <= 0 {
		return errors.New("указан некорректный жанр")
	}
	if song.Link == "" {
		return errors.New("ссылка на песню не может быть пустой")
	}
	if len(artistIDs) == 0 {
		return errors.New("нужно указать хотя бы одного артиста")
	}
	if len(artistIDs) > 10 {
		return errors.New("можно указать не более 10 артистов")
	}

	existingIDs, err := s.repo.CheckArtistsExist(ctx, artistIDs)
	if err != nil {
		log.Printf("Ошибка при проверке артистов: %v", err)
		return err
	}
	if len(existingIDs) != len(artistIDs) {
		missing := findMissingIDs(artistIDs, existingIDs)
		return fmt.Errorf("артисты с ID %v не существуют", missing)
	}

	song.Auditions = 0
	songID, err := s.repo.CreateSongWithArtists(ctx, song, artistIDs)
	if err != nil {
		log.Printf("Ошибка при создании песни: %v", err)
		return err
	}

	log.Printf("Песня успешно создана с ID %d", songID)
	return nil
}

func (s *songService) UpdateSong(ctx context.Context, song *domain.Song) error {
	if song.SongID <= 0 {
		return errors.New("некорректный ID песни для обновления")
	}

	existingSong, err := s.repo.GetSongByID(ctx, song.SongID)
	if err != nil {
		log.Printf("Ошибка при проверке существования песни с ID %d: %v", song.SongID, err)
		return err
	}
	if existingSong == nil {
		return ErrSongNotFound
	}

	if song.Name == "" {
		return errors.New("название песни не может быть пустым")
	}
	if song.GenreID <= 0 {
		return errors.New("указан некорректный жанр")
	}
	if song.Link == "" {
		return errors.New("ссылка на песню не может быть пустой")
	}

	if err := s.repo.UpdateSong(ctx, song); err != nil {
		log.Printf("Ошибка при обновлении песни с ID %d: %v", song.SongID, err)
		return err
	}
	return nil
}

func findMissingIDs(input, existing []int64) []int64 {
	existingMap := make(map[int64]bool)
	for _, id := range existing {
		existingMap[id] = true
	}
	var missing []int64
	for _, id := range input {
		if !existingMap[id] {
			missing = append(missing, id)
		}
	}
	return missing
}
