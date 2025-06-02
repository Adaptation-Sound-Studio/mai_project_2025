package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"upload-service/internal/domain/album"
	"upload-service/internal/domain/model"
	"upload-service/internal/domain/response"

	"github.com/elastic/go-elasticsearch/v8"
)

var ErrAlbumNotFound = errors.New("альбом не найден")

type AlbumService struct {
	db            *sql.DB
	repo          album.Repository
	elasticClient *elasticsearch.Client
}

func NewAlbumService(db *sql.DB, r album.Repository, elasticClient *elasticsearch.Client) *AlbumService {
	return &AlbumService{
		db:            db,
		repo:          r,
		elasticClient: elasticClient,
	}
}

func (s *AlbumService) GetAllAlbums(ctx context.Context) ([]model.Album, error) {
	albums, err := s.repo.GetAllAlbums(ctx)
	if err != nil {
		log.Printf("Ошибка при получении всех альбомов: %v", err)
		return nil, err
	}
	return albums, nil
}

func (s *AlbumService) GetAlbumByID(ctx context.Context, id int64) (*response.AlbumResponse, error) {
	if id <= 0 {
		return nil, errors.New("некорректный ID альбома")
	}

	album, err := s.repo.GetAlbumByID(ctx, id)
	if err != nil {
		log.Printf("Ошибка при получении альбома с ID %d: %v", id, err)
		return nil, err
	}
	if album == nil {
		return nil, ErrAlbumNotFound
	}

	songs, err := s.repo.GetSongsByAlbumID(ctx, album.AlbumID)
	if err != nil {
		log.Printf("Ошибка при получении песен альбома с ID %d: %v", album.AlbumID, err)
		return nil, err
	}

	res := response.MapAlbumToResponse(album, songs)
	return &res, nil
}

func (s *AlbumService) CreateAlbum(ctx context.Context, album *model.Album, songIDs []int64) (int64, error) {
	if album.Name == "" {
		return 0, errors.New("название альбома не может быть пустым")
	}
	if album.ArtistID <= 0 {
		return 0, errors.New("не указан артист для альбома")
	}
	if album.GenreID <= 0 {
		return 0, errors.New("не указан жанр альбома")
	}
	if len(songIDs) > 50 {
		return 0, errors.New("нельзя добавить более 50 песен в альбом")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	album.Auditions = 0
	albumID, err := s.repo.CreateAlbum(ctx, tx, album)
	if err != nil {
		return 0, err
	}

	existingSongIDs, err := s.repo.CheckSongsExist(ctx, tx, songIDs)
	if err != nil {
		return 0, err
	}

	ownedSongIDs, err := s.repo.CheckSongsBelongToArtist(ctx, tx, album.ArtistID, songIDs)
	if err != nil {
		return 0, err
	}

	notOwned := findMissingAlbumSongIDs(songIDs, ownedSongIDs)
	if len(notOwned) > 0 {
		return 0, fmt.Errorf("вы не можете добавить чужие песни: %v", notOwned)
	}

	missing := findMissingAlbumSongIDs(songIDs, existingSongIDs)
	if len(missing) > 0 {
		return 0, fmt.Errorf("песни с ID %v не существуют", missing)
	}

	if len(songIDs) > 0 {
		if err := s.repo.BatchInsertSongsToAlbum(ctx, tx, albumID, songIDs); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	genreObj, err := s.repo.GetGenreByAlbumID(ctx, albumID)
	if err != nil {
		log.Printf("Ошибка получения жанра по альбому: %v", err)
	}
	genre := ""
	if genreObj != nil {
		genre = genreObj.Name
	}

	artistObj, err := s.repo.GetArtistByAlbumID(ctx, albumID)
	if err != nil {
		log.Printf("Ошибка получения артиста по альбому: %v", err)
	}
	artist := ""
	if artistObj != nil {
		artist = artistObj.Name
	}

	albumElastic := model.AlbumElastic{
		AlbumID:   albumID,
		Name:      album.Name,
		Auditions: album.Auditions,
		Genre:     genre,
		Artist:    artist,
		Date:      album.Date,
	}

	if err := s.indexAlbum(ctx, &albumElastic); err != nil {
		log.Printf("Ошибка индексации альбома в Elasticsearch: %v", err)
	}

	log.Printf("Альбом успешно создан с ID %d", albumID)
	return albumID, nil
}

func (s *AlbumService) indexAlbum(ctx context.Context, album *model.AlbumElastic) error {
	doc, err := json.Marshal(album)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %w", err)
	}

	res, err := s.elasticClient.Index(
		"albums",
		strings.NewReader(string(doc)),
		s.elasticClient.Index.WithDocumentID(fmt.Sprint(album.AlbumID)),
		s.elasticClient.Index.WithContext(ctx),
		s.elasticClient.Index.WithRefresh("true"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ошибка индексирования альбома: %s", res.String())
	}
	return nil
}

func (s *AlbumService) UpdateAlbum(ctx context.Context, album *model.Album) error {
	if album.AlbumID <= 0 {
		return errors.New("некорректный ID альбома для обновления")
	}

	existingAlbum, err := s.repo.GetAlbumByID(ctx, album.AlbumID)
	if err != nil {
		return err
	}
	if existingAlbum == nil {
		return ErrAlbumNotFound
	}

	if album.Name == "" {
		return errors.New("название альбома не может быть пустым")
	}
	if album.GenreID <= 0 {
		return errors.New("не указан жанр альбома")
	}

	if err := s.repo.UpdateAlbum(ctx, album); err != nil {
		log.Printf("Ошибка при обновлении альбома с ID %d: %v", album.AlbumID, err)
		return err
	}

	log.Printf("Альбом с ID %d успешно обновлён", album.AlbumID)
	return nil
}

func findMissingAlbumSongIDs(input, existing []int64) []int64 {
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

func (s *AlbumService) SearchAlbums(ctx context.Context, filters map[string]string) ([]model.AlbumElastic, error) {
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
		s.elasticClient.Search.WithIndex("albums"),
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
				Source model.AlbumElastic `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	album := make([]model.AlbumElastic, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		album = append(album, hit.Source)
	}

	return album, nil
}
