package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"upload-service/internal/domain/event"
	"upload-service/internal/domain/genre"
	"upload-service/internal/domain/model"
	"upload-service/internal/kafka"

	"github.com/elastic/go-elasticsearch/v8"
)

type GenreService struct {
	repo          genre.Repository
	producer      *kafka.Producer
	elasticClient *elasticsearch.Client
}

func NewGenreService(r genre.Repository, producer *kafka.Producer, elasticClient *elasticsearch.Client) *GenreService {
	return &GenreService{
		repo:          r,
		producer:      producer,
		elasticClient: elasticClient,
	}
}

func (s *GenreService) GetAllGenres(ctx context.Context) ([]model.Genre, error) {
	genres, err := s.repo.GetAllGenres(ctx)
	if err != nil {
		log.Printf("Ошибка при получении жанров: %v", err)
		return nil, err
	}
	return genres, nil
}

func (s *GenreService) CreateGenre(ctx context.Context, genre *model.Genre) (int64, error) {
	if genre.Name == "" {
		return 0, errors.New("название жанра не может быть пустым")
	}

	genreID, err := s.repo.CreateGenre(ctx, genre)
	if err != nil {
		log.Printf("Ошибка при создании жанра: %v", err)
		return 0, err
	}

	event := event.Genre{
		ID:   genreID,
		Name: genre.Name,
	}

	if err := s.producer.SendWrappedEvent("genre_created", event); err != nil {
		log.Printf("Ошибка отправки события genre_created: %v", err)
	}

	log.Printf("Жанр успешно создан с ID %d", genreID)
	return genreID, nil
}

func (s *GenreService) UpdateGenre(ctx context.Context, genre *model.Genre) error {
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

func (s *GenreService) SearchGenres(ctx context.Context, query string) ([]model.Genre, error) {
	res, err := s.elasticClient.Search(
		s.elasticClient.Search.WithContext(ctx),
		s.elasticClient.Search.WithIndex("genres"),
		s.elasticClient.Search.WithBody(strings.NewReader(fmt.Sprintf(`{
			"query": {
				"match": {
					"name": {
						"query": "%s",
						"fuzziness": "AUTO"
					}
				}
			}
		}`, query))),
		s.elasticClient.Search.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result struct {
		Hits struct {
			Hits []struct {
				Source model.Genre `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	genres := make([]model.Genre, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		genres = append(genres, hit.Source)
	}

	return genres, nil
}
