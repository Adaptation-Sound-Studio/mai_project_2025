package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
	"upload-service/internal/domain/event"
	"upload-service/internal/domain/model"
	"upload-service/internal/domain/response"
	"upload-service/internal/domain/song"
	"upload-service/internal/kafka"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/minio/minio-go/v7"
)

var ErrSongNotFound = errors.New("песня не найдена")

type SongService struct {
	repo          song.Repository
	minioClient   *minio.Client
	bucketName    string
	producer      *kafka.Producer
	elasticClient *elasticsearch.Client
}

func NewSongService(r song.Repository, minioClient *minio.Client, bucketName string, producer *kafka.Producer, elasticClient *elasticsearch.Client) *SongService {
	return &SongService{
		repo:          r,
		minioClient:   minioClient,
		bucketName:    bucketName,
		producer:      producer,
		elasticClient: elasticClient,
	}
}

func (s *SongService) GetAllSongs(ctx context.Context) ([]response.SongResponse, error) {
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

func (s *SongService) GetSongByID(ctx context.Context, id int64) (*response.SongResponse, error) {
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

func (s *SongService) CreateSong(ctx context.Context, song *model.Song, artistIDs []int64, file multipart.File, filename string) error {
	if song.Name == "" {
		return errors.New("название песни не может быть пустым")
	}
	if song.GenreID <= 0 {
		return errors.New("указан некорректный жанр")
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

	objectName := generateUniqueFilename(filename)
	_, err = s.minioClient.PutObject(ctx, s.bucketName, objectName, file, -1, minio.PutObjectOptions{
		ContentType: "audio/mpeg",
	})
	if err != nil {
		log.Printf("Ошибка при загрузке файла в MinIO: %v", err)
		return err
	}

	song.NameOfMinio = objectName
	song.Auditions = 0

	log.Printf("Файл успешно загружен в MinIO: %s", objectName)

	songID, err := s.repo.CreateSongWithArtists(ctx, song, artistIDs)
	if err != nil {
		log.Printf("Ошибка при создании песни: %v", err)
		return err
	}

	song.SongID = songID
	songEvent := event.SongCreatedEvent{
		ID:   songID,
		Name: song.Name,
	}

	err = s.producer.SendWrappedEvent("song_created", songEvent)
	if err != nil {
		log.Printf("Ошибка при отправке события song_created: %v", err)
	}

	artistObj, err := s.repo.GetOneArtistBySongID(ctx, songID)
	if err != nil {
		log.Printf("Ошибка получения артиста по песне: %v", err)
	}
	artist := ""
	if artistObj != nil {
		artist = artistObj.Name
	}

	genreObj, err := s.repo.GetGenreBySongID(ctx, songID)
	if err != nil {
		log.Printf("Ошибка получения жанра по песне: %v", err)
	}
	genre := ""
	if genreObj != nil {
		genre = genreObj.Name
	}

	songElastic := model.SongElastic{
		SongID:      song.SongID,
		Name:        song.Name,
		NameOfMinio: song.NameOfMinio,
		Auditions:   song.Auditions,
		Genre:       genre,
		Artist:      artist,
	}

	if err := s.indexSong(ctx, &songElastic); err != nil {
		log.Printf("Ошибка индексации песни в Elasticsearch: %v", err)
	}

	log.Printf("Песня успешно создана с ID %d", songID)
	return nil
}

func (s *SongService) indexSong(ctx context.Context, song *model.SongElastic) error {
	doc, err := json.Marshal(song)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %w", err)
	}

	res, err := s.elasticClient.Index(
		"songs",
		strings.NewReader(string(doc)),
		s.elasticClient.Index.WithDocumentID(fmt.Sprint(song.SongID)),
		s.elasticClient.Index.WithContext(ctx),
		s.elasticClient.Index.WithRefresh("true"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ошибка индексирования песни: %s", res.String())
	}
	return nil
}

func generateUniqueFilename(orig string) string {
	ext := filepath.Ext(orig)
	base := strings.TrimSuffix(orig, ext)
	uniquePart := fmt.Sprintf("%d", time.Now().UnixNano())
	return base + "_" + uniquePart + ext
}

func (s *SongService) UpdateSong(ctx context.Context, song *model.Song) error {
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

func (s *SongService) GetSongArtists(ctx context.Context, songID int64) ([]model.Artist, error) {
	return s.repo.GetArtistsBySongID(ctx, songID)
}

func (s *SongService) GetArtistsBySongID(ctx context.Context, songID int64) ([]model.Artist, error) {
	return s.repo.GetArtistsBySongID(ctx, songID)
}

func (s *SongService) SearchSongs(ctx context.Context, filters map[string]string) ([]model.SongElastic, error) {
	mustClauses := []string{}

	for field, value := range filters {
		if value != "" {
			mustClauses = append(mustClauses, fmt.Sprintf(`{
				"match": {
					"%s": {
						"query": "%s",
						"fuzziness": "AUTO"
					}
				}
			}`, field, value))
		}
	}

	body := fmt.Sprintf(`{
		"query": {
			"bool": {
				"must": [%s]
			}
		}
	}`, strings.Join(mustClauses, ","))

	res, err := s.elasticClient.Search(
		s.elasticClient.Search.WithContext(ctx),
		s.elasticClient.Search.WithIndex("songs"),
		s.elasticClient.Search.WithBody(strings.NewReader(body)),
		s.elasticClient.Search.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result struct {
		Hits struct {
			Hits []struct {
				Source model.SongElastic `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	songs := make([]model.SongElastic, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		songs = append(songs, hit.Source)
	}

	return songs, nil
}
