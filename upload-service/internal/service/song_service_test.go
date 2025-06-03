package service

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"upload-service/internal/config"
	"upload-service/internal/domain/event"
	"upload-service/internal/domain/model"
	"upload-service/internal/infrastructure/elastic"
	"upload-service/internal/kafka"
	"upload-service/internal/repository"
	"upload-service/internal/testutils"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSongRepo struct {
	mock.Mock
}

func (m *MockSongRepo) GetAllSongs(ctx context.Context) ([]model.Song, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Song), args.Error(1)
}

func (m *MockSongRepo) GetSongByID(ctx context.Context, id int64) (*model.Song, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Song), args.Error(1)
}

func (m *MockSongRepo) GetArtistsBySongID(ctx context.Context, songID int64) ([]model.Artist, error) {
	args := m.Called(ctx, songID)
	return args.Get(0).([]model.Artist), args.Error(1)
}

func (m *MockSongRepo) GetAlbumBySongID(ctx context.Context, songID int64) (*model.Album, error) {
	args := m.Called(ctx, songID)
	return args.Get(0).(*model.Album), args.Error(1)
}

func (m *MockSongRepo) CheckArtistsExist(ctx context.Context, ids []int64) ([]int64, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockSongRepo) CreateSongWithArtists(ctx context.Context, song *model.Song, artistIDs []int64) (int64, error) {
	args := m.Called(ctx, song, artistIDs)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSongRepo) UpdateSong(ctx context.Context, song *model.Song) error {
	args := m.Called(ctx, song)
	return args.Error(0)
}
func (m *MockSongRepo) GetGenreBySongID(ctx context.Context, songID int64) (*model.Genre, error) {
	args := m.Called(ctx, songID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Genre), args.Error(1)
}

func (m *MockSongRepo) GetOneArtistBySongID(ctx context.Context, songID int64) (*model.Artist, error) {
	args := m.Called(ctx, songID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Artist), args.Error(1)
}

func (m *MockSongRepo) IncrementAuditions(ctx context.Context, songID int64) error {
	args := m.Called(ctx, songID)
	return args.Error(0)
}

type MockMinioClient struct {
	mock.Mock
}

func (m *MockMinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	args := m.Called(ctx, bucketName, objectName, reader, objectSize, opts)
	return args.Get(0).(minio.UploadInfo), args.Error(1)
}

type fakeMultipartFile struct {
	*bytes.Reader
}

func (f *fakeMultipartFile) Close() error {
	return nil
}

func (f *fakeMultipartFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.Reader.ReadAt(p, off)
}

func TestGetAllSongs_Success(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	songs := []model.Song{{SongID: 1, Name: "Track 1"}}
	artists := []model.Artist{{ArtistID: 1, Name: "Artist"}}
	album := &model.Album{AlbumID: 1, Name: "Album"}

	repo.On("GetAllSongs", mock.Anything).Return(songs, nil)
	repo.On("GetArtistsBySongID", mock.Anything, int64(1)).Return(artists, nil)
	repo.On("GetAlbumBySongID", mock.Anything, int64(1)).Return(album, nil)

	result, err := svc.GetAllSongs(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Track 1", result[0].Name)
	assert.Equal(t, "Artist", result[0].Artists[0].Name)
	assert.Equal(t, "Album", result[0].Album.Name)

	repo.AssertExpectations(t)
}
func TestGetSongByID_Success(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	song := &model.Song{SongID: 1, Name: "Track"}
	artists := []model.Artist{{ArtistID: 1, Name: "Artist"}}
	album := &model.Album{AlbumID: 1, Name: "Album"}

	repo.On("GetSongByID", mock.Anything, int64(1)).Return(song, nil)
	repo.On("GetArtistsBySongID", mock.Anything, int64(1)).Return(artists, nil)
	repo.On("GetAlbumBySongID", mock.Anything, int64(1)).Return(album, nil)

	result, err := svc.GetSongByID(context.Background(), 1)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Track", result.Name)
	assert.Equal(t, "Artist", result.Artists[0].Name)
	assert.Equal(t, "Album", result.Album.Name)

	repo.AssertExpectations(t)
}

func TestCreateSong_WithRealMinio_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	minioEndpoint := "localhost:9000"
	minioAccessKey := "admin"
	minioSecretKey := "secret123"
	bucketName := "music"

	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: false,
	})
	require.NoError(t, err)

	ctx := context.Background()
	found, err := minioClient.BucketExists(ctx, bucketName)
	require.NoError(t, err)
	if !found {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		require.NoError(t, err)
	}

	db, err := sql.Open("postgres", "host=localhost port=5434 user=postgres password=Rbkkth3920 dbname=upl_db sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	testutils.CleanTables(t, db)

	repo := repository.NewSongRepository(db)

	producer := new(MockProducer)

	svc := NewSongService(repo, minioClient, bucketName, producer, nil)

	song := &model.Song{
		Name:    "Integration Song",
		GenreID: 1,
	}
	artistIDs := []int64{1}

	_, err = db.Exec(`INSERT INTO artists (artist_id, name, user_id) VALUES (1, 'Test Artist', 10) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO genres (genre_id, name) VALUES (1, 'Test Genre') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	content := []byte("fake audio content")
	fakeFile := &fakeMultipartFile{bytes.NewReader(content)}
	filename := "integration_song.mp3"

	producer.On("SendWrappedEvent", "song_created", event.SongCreatedEvent{
		ID:   1,
		Name: "Integration Song",
	}).Return(nil)

	err = svc.CreateSong(ctx, song, artistIDs, fakeFile, filename)

	require.NoError(t, err)

	require.NotEmpty(t, song.NameOfMinio)

	_, err = minioClient.StatObject(ctx, bucketName, song.NameOfMinio, minio.StatObjectOptions{})
	require.NoError(t, err)

	err = minioClient.RemoveObject(ctx, bucketName, song.NameOfMinio, minio.RemoveObjectOptions{})
	require.NoError(t, err)
}

func TestUpdateSong_Success(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	song := &model.Song{SongID: 1, Name: "Update", GenreID: 1}

	repo.On("GetSongByID", mock.Anything, int64(1)).Return(song, nil)
	repo.On("UpdateSong", mock.Anything, song).Return(nil)

	err := svc.UpdateSong(context.Background(), song)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetSongArtists_Success(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	expected := []model.Artist{{ArtistID: 1, Name: "Artist"}}
	repo.On("GetArtistsBySongID", mock.Anything, int64(1)).Return(expected, nil)

	result, err := svc.GetSongArtists(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestFindMissingIDs(t *testing.T) {
	input := []int64{1, 2, 3, 4, 5}
	existing := []int64{2, 4, 5}

	expected := []int64{1, 3}
	result := findMissingIDs(input, existing)

	assert.ElementsMatch(t, expected, result)
}

func TestUpdateSong_InvalidID(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	err := svc.UpdateSong(context.Background(), &model.Song{SongID: 0})

	assert.EqualError(t, err, "некорректный ID песни для обновления")
}

func TestUpdateSong_EmptyName(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	repo.On("GetSongByID", mock.Anything, int64(1)).
		Return(&model.Song{SongID: 1}, nil)

	err := svc.UpdateSong(context.Background(), &model.Song{
		SongID:  1,
		Name:    "",
		GenreID: 1,
	})

	assert.EqualError(t, err, "название песни не может быть пустым")
	repo.AssertExpectations(t)
}

func TestUpdateSong_InvalidGenre(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	repo.On("GetSongByID", mock.Anything, int64(2)).
		Return(&model.Song{SongID: 2}, nil)

	err := svc.UpdateSong(context.Background(), &model.Song{
		SongID:  2,
		Name:    "test",
		GenreID: 0,
	})

	assert.EqualError(t, err, "указан некорректный жанр")
	repo.AssertExpectations(t)
}

func TestUpdateSong_RepoUpdateError(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	song := &model.Song{
		SongID:  4,
		Name:    "test",
		GenreID: 1,
	}

	repo.On("GetSongByID", mock.Anything, int64(4)).Return(song, nil)
	repo.On("UpdateSong", mock.Anything, song).Return(errors.New("update failed"))

	err := svc.UpdateSong(context.Background(), song)

	assert.EqualError(t, err, "update failed")
	repo.AssertExpectations(t)
}

func TestGetSongByID_InvalidID(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	res, err := svc.GetSongByID(context.Background(), 0)

	assert.Nil(t, res)
	assert.EqualError(t, err, "ID песни должен быть положительным числом")
}

func TestGetSongByID_DBError(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	repo.On("GetSongByID", mock.Anything, int64(1)).
		Return((*model.Song)(nil), errors.New("db error"))

	res, err := svc.GetSongByID(context.Background(), 1)

	assert.Nil(t, res)
	assert.EqualError(t, err, "db error")
	repo.AssertExpectations(t)
}

func TestGetSongByID_NotFound(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	repo.On("GetSongByID", mock.Anything, int64(1)).
		Return((*model.Song)(nil), nil)

	res, err := svc.GetSongByID(context.Background(), 1)

	assert.Nil(t, res)
	assert.Equal(t, ErrSongNotFound, err)
	repo.AssertExpectations(t)
}

func TestGetSongByID_AlbumError(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	song := &model.Song{SongID: 1, Name: "Test Song"}
	artists := []model.Artist{{ArtistID: 1, Name: "Artist"}}

	repo.On("GetSongByID", mock.Anything, int64(1)).
		Return(song, nil)
	repo.On("GetArtistsBySongID", mock.Anything, int64(1)).
		Return(artists, nil)
	repo.On("GetAlbumBySongID", mock.Anything, int64(1)).
		Return((*model.Album)(nil), errors.New("album error"))

	res, err := svc.GetSongByID(context.Background(), 1)

	assert.Nil(t, res)
	assert.EqualError(t, err, "album error")
	repo.AssertExpectations(t)
}

func TestCreateSong_EmptyName(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	err := svc.CreateSong(context.Background(), &model.Song{
		Name:    "",
		GenreID: 1,
	}, []int64{1}, nil, "")

	assert.EqualError(t, err, "название песни не может быть пустым")
}

func TestCreateSong_InvalidGenre(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	err := svc.CreateSong(context.Background(), &model.Song{
		Name:    "Test Song",
		GenreID: 0,
	}, []int64{1}, nil, "")

	assert.EqualError(t, err, "указан некорректный жанр")
}

func TestCreateSong_NoArtists(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	err := svc.CreateSong(context.Background(), &model.Song{
		Name:    "Test Song",
		GenreID: 1,
	}, []int64{}, nil, "file.mp3")

	assert.EqualError(t, err, "нужно указать хотя бы одного артиста")
}
func TestCreateSong_TooManyArtists(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	tooMany := make([]int64, 11)
	for i := 0; i < 11; i++ {
		tooMany[i] = int64(i + 1)
	}

	err := svc.CreateSong(context.Background(), &model.Song{
		Name:    "Test Song",
		GenreID: 1,
	}, tooMany, nil, "file.mp3")

	assert.EqualError(t, err, "можно указать не более 10 артистов")
}

func TestCreateSong_ArtistNotExist(t *testing.T) {
	repo := new(MockSongRepo)

	var fakeMinio *minio.Client = nil
	var fakeProducer *kafka.Producer = nil
	var fakeES *elasticsearch.Client = nil
	bucket := "test-bucket"

	svc := NewSongService(repo, fakeMinio, bucket, fakeProducer, fakeES)

	input := []int64{1, 2, 3}
	existing := []int64{1, 3}
	song := &model.Song{Name: "Song", GenreID: 1}

	repo.On("CheckArtistsExist", mock.Anything, input).
		Return(existing, nil)

	err := svc.CreateSong(context.Background(), song, input, nil, "file.mp3")

	assert.EqualError(t, err, "артисты с ID [2] не существуют")
	repo.AssertExpectations(t)
}

func TestSearchSongs_Integration(t *testing.T) {

	cfg := &config.ElasticConfig{
		URL:      "http://localhost:9200",
		Username: "elastic",
		Password: "your_password",
	}

	esClient, err := elastic.NewElasticClient(cfg)
	require.NoError(t, err)

	service := &SongService{
		elasticClient: esClient,
	}

	ctx := context.Background()

	songsToIndex := []model.SongElastic{
		{SongID: 1, Name: "Electro Beat", Genre: "Electronic", Artist: "DJ Alpha"},
		{SongID: 2, Name: "Rock Anthem", Genre: "Rock", Artist: "The Rockers"},
	}

	for _, song := range songsToIndex {
		docBody := fmt.Sprintf(
			`{"song_id": %d, "name": "%s", "genre": "%s", "artist_name": "%s"}`,
			song.SongID, song.Name, song.Genre, song.Artist,
		)

		res, err := esClient.Index("songs", strings.NewReader(docBody),
			esClient.Index.WithDocumentID(fmt.Sprint(song.SongID)),
			esClient.Index.WithRefresh("true"),
		)
		require.NoError(t, err)
		defer res.Body.Close()
		require.False(t, res.IsError(), "Ошибка индексации: %s", res.String())
	}

	filters := map[string]string{
		"name": "Electro",
	}

	resultSongs, err := service.SearchSongs(ctx, filters)
	require.NoError(t, err)
	require.NotEmpty(t, resultSongs)

	found := false
	for _, s := range resultSongs {
		if s.SongID == 1 && s.Name == "Electro Beat" {
			found = true
			break
		}
	}
	assert.True(t, found, "Ожидаемая песня не найдена в результатах поиска")

	for _, song := range songsToIndex {
		_, _ = esClient.Delete("songs", fmt.Sprint(song.SongID))
	}
}

func TestGenerateUniqueFilename(t *testing.T) {
	original := "cover.png"

	result1 := generateUniqueFilename(original)

	time.Sleep(5 * time.Second)

	result2 := generateUniqueFilename(original)

	if ext := filepath.Ext(result1); ext != ".png" {
		t.Errorf("ожидали расширение '.png', но получили: %s", ext)
	}

	if !strings.HasPrefix(result1, "cover_") {
		t.Errorf("ожидали префикс 'cover_', но получили: %s", result1)
	}

	if result1 == result2 {
		t.Errorf("ожидали уникальные имена, но получили одинаковые: %s", result1)
	}
}

func TestIndexSong_Integration(t *testing.T) {
	cfg := &config.ElasticConfig{
		URL:      "http://localhost:9200",
		Username: "elastic",
		Password: "your_password",
	}

	esClient, err := elastic.NewElasticClient(cfg)
	require.NoError(t, err)

	service := &SongService{
		elasticClient: esClient,
	}

	ctx := context.Background()

	song := &model.SongElastic{
		SongID: 12345,
		Name:   "Test Song",
		Genre:  "Rock",
		Artist: "Test Artist",
	}

	err = service.indexSong(ctx, song)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	docID := fmt.Sprint(song.SongID)
	getResp, err := esClient.Get("songs", docID)
	require.NoError(t, err)
	defer getResp.Body.Close()

	assert.False(t, getResp.IsError(), "Документ не найден в индексе")

	bodyBytes, err := io.ReadAll(getResp.Body)
	require.NoError(t, err)

	assert.Contains(t, string(bodyBytes), song.Name)

	_, _ = esClient.Delete("songs", docID)
}
