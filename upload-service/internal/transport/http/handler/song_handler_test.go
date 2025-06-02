package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"upload-service/internal/domain/model"
	"upload-service/internal/domain/request"
	"upload-service/internal/kafka"
	"upload-service/internal/service"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	minio "github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestCreateSong_Success(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	rdb.Set(context.Background(), "session:token", "10", 0)
	rdb.Set(context.Background(), "artist:10", "10", 0)

	mockRepo := new(MockSongRepo)
	var mockMinioClient *minio.Client = nil
	var mockKafkaProducer *kafka.Producer = nil
	svc := service.NewSongService(mockRepo, mockMinioClient, "test-bucket", mockKafkaProducer)
	handler := NewSongHandler(svc, mockMinioClient, "test-bucket", rdb, mockKafkaProducer)

	songReq := request.CreateSongRequest{
		Name:      "My Song",
		GenreID:   1,
		ArtistIDs: []int64{11},
	}

	mockRepo.On("CheckArtistsExist", mock.Anything, mock.MatchedBy(func(ids []int64) bool {
		set := map[int64]bool{}
		for _, id := range ids {
			set[id] = true
		}
		return set[10] && set[11] && len(set) == 2
	})).Return([]int64{10, 11}, nil)

	mockRepo.On("CreateSongWithArtists", mock.Anything, mock.AnythingOfType("*model.Song"), mock.MatchedBy(func(ids []int64) bool {
		set := map[int64]bool{}
		for _, id := range ids {
			set[id] = true
		}
		return set[10] && set[11] && len(set) == 2
	})).Return(int64(1), nil)

	body, _ := json.Marshal(songReq)
	req := httptest.NewRequest(http.MethodPost, "/songs", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "token")
	rr := httptest.NewRecorder()

	handler.CreateSong(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUpdateSong_Success(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	rdb.Set(context.Background(), "session:token", "10", 0)
	rdb.Set(context.Background(), "artist:10", "10", 0)

	mockRepo := new(MockSongRepo)
	var mockMinioClient *minio.Client = nil
	var mockKafkaProducer *kafka.Producer = nil

	svc := service.NewSongService(mockRepo, mockMinioClient, "test-bucket", mockKafkaProducer)
	handler := NewSongHandler(svc, mockMinioClient, "test-bucket", rdb, mockKafkaProducer)

	songID := int64(1)
	reqBody := request.UpdateSongRequest{
		Name:    "Updated",
		GenreID: 2,
	}
	updatedSong := &model.Song{
		SongID:  songID,
		Name:    "Updated",
		GenreID: 2,
	}

	mockRepo.On("GetSongByID", mock.Anything, songID).Return(&model.Song{SongID: songID}, nil)
	mockRepo.On("GetArtistsBySongID", mock.Anything, songID).Return([]model.Artist{{ArtistID: 10}}, nil)
	mockRepo.On("UpdateSong", mock.Anything, updatedSong).Return(nil)

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/songs/1", bytes.NewBuffer(jsonBody))
	req = mux.SetURLVars(req, map[string]string{"song_id": "1"})
	req.Header.Set("Authorization", "token")
	rr := httptest.NewRecorder()

	handler.UpdateSong(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestGetAllSongs_Success(t *testing.T) {
	mockRepo := new(MockSongRepo)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	var mockMinioClient *minio.Client = nil
	var mockKafkaProducer *kafka.Producer = nil

	svc := service.NewSongService(mockRepo, mockMinioClient, "test-bucket", mockKafkaProducer)
	handler := NewSongHandler(svc, mockMinioClient, "test-bucket", rdb, mockKafkaProducer)

	expectedSongs := []model.Song{
		{SongID: 1, Name: "Song One", GenreID: 1, NameOfMinio: "link1"},
		{SongID: 2, Name: "Song Two", GenreID: 2, NameOfMinio: "link2"},
	}

	mockRepo.On("GetAllSongs", mock.Anything).Return(expectedSongs, nil)

	mockRepo.On("GetArtistsBySongID", mock.Anything, int64(1)).Return([]model.Artist{
		{ArtistID: 10, Name: "Artist A"},
	}, nil)
	mockRepo.On("GetArtistsBySongID", mock.Anything, int64(2)).Return([]model.Artist{
		{ArtistID: 20, Name: "Artist B"},
	}, nil)

	mockRepo.On("GetAlbumBySongID", mock.Anything, int64(1)).Return(&model.Album{
		AlbumID: 1, Name: "Album A",
	}, nil)
	mockRepo.On("GetAlbumBySongID", mock.Anything, int64(2)).Return(&model.Album{
		AlbumID: 2, Name: "Album B",
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/songs", nil)
	rr := httptest.NewRecorder()

	handler.GetAllSongs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result []model.Song
	err := json.NewDecoder(rr.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, expectedSongs, result)

	mockRepo.AssertExpectations(t)
}

func TestGetSongByID_Success(t *testing.T) {
	mockRepo := new(MockSongRepo)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	var mockMinioClient *minio.Client = nil
	var mockKafkaProducer *kafka.Producer = nil

	svc := service.NewSongService(mockRepo, mockMinioClient, "test-bucket", mockKafkaProducer)
	handler := NewSongHandler(svc, mockMinioClient, "test-bucket", rdb, mockKafkaProducer)

	expectedSong := &model.Song{SongID: 1, Name: "Test Song", GenreID: 1, NameOfMinio: "testlink"}
	expectedArtists := []model.Artist{{ArtistID: 10, Name: "Artist"}}
	expectedAlbum := &model.Album{AlbumID: 5, Name: "Album"}

	mockRepo.On("GetSongByID", mock.Anything, int64(1)).Return(expectedSong, nil)
	mockRepo.On("GetArtistsBySongID", mock.Anything, int64(1)).Return(expectedArtists, nil)
	mockRepo.On("GetAlbumBySongID", mock.Anything, int64(1)).Return(expectedAlbum, nil)

	req := httptest.NewRequest(http.MethodGet, "/songs/1", nil)
	req = mux.SetURLVars(req, map[string]string{"song_id": "1"})
	rr := httptest.NewRecorder()

	handler.GetSongByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result model.Song
	err := json.NewDecoder(rr.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, *expectedSong, result)

	mockRepo.AssertExpectations(t)
}

func TestUpdateSong_Unauthorized(t *testing.T) {
	handler := NewSongHandler(nil, redis.NewClient(&redis.Options{Addr: "localhost:6379"}))

	req := httptest.NewRequest(http.MethodPut, "/songs/10", nil)
	req = mux.SetURLVars(req, map[string]string{"song_id": "10"})
	req.Header.Set("Authorization", "")
	rr := httptest.NewRecorder()

	handler.UpdateSong(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestGetSongByID_InvalidID(t *testing.T) {
	handler := NewSongHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/songs/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"song_id": "invalid"})
	rr := httptest.NewRecorder()

	handler.GetSongByID(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Некорректный ID песни")
}

func TestGetSongByID_ServiceError(t *testing.T) {
	mockRepo := new(MockSongRepo)
	svc := service.NewSongService(mockRepo)
	handler := NewSongHandler(svc, nil)

	mockRepo.On("GetSongByID", mock.Anything, int64(42)).
		Return((*model.Song)(nil), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/songs/42", nil)
	req = mux.SetURLVars(req, map[string]string{"song_id": "42"})
	rr := httptest.NewRecorder()

	handler.GetSongByID(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Ошибка при получении песни")
	mockRepo.AssertExpectations(t)
}
