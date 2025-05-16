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

var ErrArtistNotFound = errors.New("артист не найден")

type ArtistService interface {
	GetAllArtists(ctx context.Context) ([]domain.Artist, error)
	GetArtistByID(ctx context.Context, id int64) (*response.ArtistResponse, error)
	RegisterArtist(ctx context.Context, artist *domain.Artist, userID int64) (int64, error)
	UpdateArtist(ctx context.Context, artist *domain.Artist) error
}

type artistService struct {
	repo repository.ArtistRepository
}

func NewArtistService(r repository.ArtistRepository) ArtistService {
	return &artistService{repo: r}
}

func (s *artistService) GetAllArtists(ctx context.Context) ([]domain.Artist, error) {
	artists, err := s.repo.GetAllArtists(ctx)
	if err != nil {
		log.Printf("Ошибка при получении всех артистов: %v", err)
		return nil, err
	}
	return artists, nil
}

func (s *artistService) GetArtistByID(ctx context.Context, id int64) (*response.ArtistResponse, error) {
	if id <= 0 {
		return nil, errors.New("некорректный ID артиста")
	}

	artist, err := s.repo.GetArtistByID(ctx, id)
	if err != nil {
		log.Printf("Ошибка при получении артиста с ID %d: %v", id, err)
		return nil, err
	}
	if artist == nil {
		return nil, ErrArtistNotFound
	}

	songs, err := s.repo.GetSongsByArtistID(ctx, artist.ArtistID)
	if err != nil {
		log.Printf("Ошибка при получении песен артиста с ID %d: %v", artist.ArtistID, err)
		return nil, err
	}

	albums, err := s.repo.GetAlbumsByArtistID(ctx, artist.ArtistID)
	if err != nil {
		log.Printf("Ошибка при получении альбомов артиста с ID %d: %v", artist.ArtistID, err)
		return nil, err
	}

	res := response.MapArtistToResponse(artist, songs, albums)
	return &res, nil
}

func (s *artistService) RegisterArtist(ctx context.Context, artist *domain.Artist, userID int64) (int64, error) {
	if artist.Name == "" {
		return 0, errors.New("имя артиста не может быть пустым")
	}

	existingArtist, err := s.repo.GetArtistByUserID(ctx, userID)
	if err != nil {
		log.Printf("Ошибка при проверке существования артиста по user_id %d: %v", userID, err)
		return 0, err
	}
	if existingArtist != nil {
		return 0, fmt.Errorf("пользователь с user_id %d уже зарегистрирован как артист", userID)
	}

	artistID, err := s.repo.CreateArtist(ctx, artist, userID)
	if err != nil {
		log.Printf("Ошибка при создании артиста: %v", err)
		return 0, err
	}

	log.Printf("Артист успешно зарегистрирован с ID %d", artistID)
	return artistID, nil
}

func (s *artistService) UpdateArtist(ctx context.Context, artist *domain.Artist) error {
	if artist.ArtistID <= 0 {
		return errors.New("некорректный ID артиста для обновления")
	}

	existingArtist, err := s.repo.GetArtistByID(ctx, artist.ArtistID)
	if err != nil {
		log.Printf("Ошибка при проверке существования артиста с ID %d: %v", artist.ArtistID, err)
		return err
	}
	if existingArtist == nil {
		return ErrArtistNotFound
	}

	if artist.Name == "" {
		return errors.New("имя артиста не может быть пустым")
	}

	if err := s.repo.UpdateArtist(ctx, artist); err != nil {
		log.Printf("Ошибка при обновлении артиста с ID %d: %v", artist.ArtistID, err)
		return err
	}

	log.Printf("Артист с ID %d успешно обновлён", artist.ArtistID)
	return nil
}
