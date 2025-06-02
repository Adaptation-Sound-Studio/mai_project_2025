package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"upload-service/internal/domain/artist"
	"upload-service/internal/domain/event"
	"upload-service/internal/domain/model"
	"upload-service/internal/domain/response"
	"upload-service/internal/kafka"

	"github.com/elastic/go-elasticsearch/v8"
)

var ErrArtistNotFound = errors.New("артист не найден")

type ArtistService struct {
	repo          artist.Repository
	authBaseURL   string
	apiKey        string
	producer      kafka.ProducerIface
	elasticClient *elasticsearch.Client
}

type roleUpdateRequest struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}

func NewArtistService(r artist.Repository, authBaseURL, apiKey string, producer *kafka.Producer, elasticClient *elasticsearch.Client) *ArtistService {
	return &ArtistService{
		repo:          r,
		authBaseURL:   authBaseURL,
		apiKey:        apiKey,
		producer:      producer,
		elasticClient: elasticClient,
	}
}

func (s *ArtistService) GetAllArtists(ctx context.Context) ([]model.Artist, error) {
	artists, err := s.repo.GetAllArtists(ctx)
	if err != nil {
		log.Printf("Ошибка при получении всех артистов: %v", err)
		return nil, err
	}
	return artists, nil
}

func (s *ArtistService) GetArtistByID(ctx context.Context, id int64) (*response.ArtistResponse, error) {
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

func (s *ArtistService) RegisterArtist(ctx context.Context, artist *model.Artist, userID int64) (int64, error) {
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

	reqBody := roleUpdateRequest{
		UserID: userID,
		Role:   "artist",
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Ошибка сериализации запроса на обновление роли: %v", err)
		return artistID, nil
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/users/role", s.authBaseURL), bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Ошибка создания HTTP-запроса в auth-сервис: %v", err)
		return artistID, nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", s.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Ошибка при выполнении запроса в auth-сервис: %v", err)
		return artistID, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Auth-сервис вернул статус %d при обновлении роли", resp.StatusCode)
	}

	artist.ArtistID = artistID
	artistEvent := event.ArtistCreatedEvent{
		ID:   artist.ArtistID,
		Name: artist.Name,
	}

	err = s.producer.SendWrappedEvent("artist_created", artistEvent)
	if err != nil {
		log.Printf("Ошибка отправки события artist_created: %v", err)
	}

	if err := s.indexArtist(ctx, artist); err != nil {
		log.Printf("Ошибка индексации артиста в Elasticsearch: %v", err)
	}

	return artistID, nil
}

func (s *ArtistService) indexArtist(ctx context.Context, artist *model.Artist) error {
	if s.elasticClient == nil {
		return nil
	}
	body := fmt.Sprintf(`{
		"artist_id": %d,
		"name": "%s",
		"user_id": %d
	}`, artist.ArtistID, artist.Name, artist.UserID)

	res, err := s.elasticClient.Index(
		"artists",
		strings.NewReader(body),
		s.elasticClient.Index.WithDocumentID(fmt.Sprint(artist.ArtistID)),
		s.elasticClient.Index.WithContext(ctx),
		s.elasticClient.Index.WithRefresh("true"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ошибка индексирования артиста: %s", res.String())
	}
	return nil
}

func (s *ArtistService) UpdateArtist(ctx context.Context, artist *model.Artist) error {
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

func (s *ArtistService) GetArtistIDByUserID(ctx context.Context, userID int64) (string, error) {
	artist, err := s.repo.GetArtistByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if artist == nil {
		return "", nil
	}
	return strconv.FormatInt(artist.ArtistID, 10), nil
}

func (s *ArtistService) SearchArtists(ctx context.Context, query string) ([]model.Artist, error) {
	log.Println("SearchArtists hit with query:", query)

	res, err := s.elasticClient.Search(
		s.elasticClient.Search.WithContext(ctx),
		s.elasticClient.Search.WithIndex("artists"),
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
				Source model.Artist `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	artists := make([]model.Artist, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		artists = append(artists, hit.Source)
	}

	return artists, nil
}
