package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"upload-service/internal/domain/model"
	"upload-service/internal/service"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockAlbumRepo struct {
	mock.Mock
}

func (m *MockAlbumRepo) GetAllAlbums(ctx context.Context) ([]model.Album, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Album), args.Error(1)
}

func (m *MockAlbumRepo) GetAlbumByID(ctx context.Context, id int64) (*model.Album, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Album), args.Error(1)
}

func (m *MockAlbumRepo) GetSongsByAlbumID(ctx context.Context, albumID int64) ([]model.Song, error) {
	args := m.Called(ctx, albumID)
	return args.Get(0).([]model.Song), args.Error(1)
}

func (m *MockAlbumRepo) CreateAlbum(ctx context.Context, tx *sql.Tx, album *model.Album) (int64, error) {
	args := m.Called(ctx, tx, album)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAlbumRepo) UpdateAlbum(ctx context.Context, album *model.Album) error {
	args := m.Called(ctx, album)
	return args.Error(0)
}

func (m *MockAlbumRepo) CheckSongsExist(ctx context.Context, tx *sql.Tx, songIDs []int64) ([]int64, error) {
	args := m.Called(ctx, tx, songIDs)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockAlbumRepo) BatchInsertSongsToAlbum(ctx context.Context, tx *sql.Tx, albumID int64, songIDs []int64) error {
	args := m.Called(ctx, tx, albumID, songIDs)
	return args.Error(0)
}
func TestGetAllAlbums_Success(t *testing.T) {
	mockRepo := new(MockAlbumRepo)
	db, _, _ := sqlmock.New()
	service := service.NewAlbumService(db, mockRepo)
	handler := NewAlbumHandler(service, nil)

	expectedAlbums := []model.Album{
		{AlbumID: 1, Name: "Test", ArtistID: 42, GenreID: 1},
	}

	mockRepo.On("GetAllAlbums", mock.Anything).Return(expectedAlbums, nil)

	req := httptest.NewRequest(http.MethodGet, "/albums", nil)
	rr := httptest.NewRecorder()

	handler.GetAllAlbums(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var actual []model.Album
	err := json.Unmarshal(rr.Body.Bytes(), &actual)
	require.NoError(t, err)
	assert.Equal(t, expectedAlbums, actual)

	mockRepo.AssertExpectations(t)
}

func TestGetAlbumByID_Success(t *testing.T) {
	mockRepo := new(MockAlbumRepo)
	db, _, _ := sqlmock.New()
	service := service.NewAlbumService(db, mockRepo)
	handler := NewAlbumHandler(service, nil)

	album := &model.Album{AlbumID: 1, Name: "Test", GenreID: 1, ArtistID: 42}
	mockRepo.On("GetAlbumByID", mock.Anything, int64(1)).Return(album, nil)
	mockRepo.On("GetSongsByAlbumID", mock.Anything, int64(1)).Return([]model.Song{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/albums/1", nil)
	req = mux.SetURLVars(req, map[string]string{"album_id": "1"})
	rr := httptest.NewRecorder()

	handler.GetAlbumByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}
func TestCreateAlbum_Success(t *testing.T) {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	_ = rdb.Set(ctx, "session:test-token", "42", 0).Err()
	_ = rdb.Set(ctx, "artist:42", "42", 0).Err()

	db, sqlMock, _ := sqlmock.New()
	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	mockRepo := new(MockAlbumRepo)
	service := service.NewAlbumService(db, mockRepo)
	handler := NewAlbumHandler(service, rdb)

	songIDs := []int64{1, 2}
	album := &model.Album{Name: "New Album", GenreID: 1, ArtistID: 42}

	mockRepo.On("CreateAlbum", mock.Anything, mock.Anything, album).Return(int64(100), nil)
	mockRepo.On("CheckSongsExist", mock.Anything, mock.Anything, songIDs).Return(songIDs, nil)
	mockRepo.On("BatchInsertSongsToAlbum", mock.Anything, mock.Anything, int64(100), songIDs).Return(nil)

	body := `{"name": "New Album", "genre_id": 1, "song_ids": [1, 2]}`
	req := httptest.NewRequest(http.MethodPost, "/albums", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "test-token")
	rr := httptest.NewRecorder()

	handler.CreateAlbum(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Альбом успешно создан", response["message"])
	assert.Equal(t, "100", response["album_id"])

	mockRepo.AssertExpectations(t)
}

func TestGetAlbumByID_InvalidID(t *testing.T) {
	handler := NewAlbumHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/albums/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"album_id": "invalid"})
	rr := httptest.NewRecorder()

	handler.GetAlbumByID(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Некорректный ID альбома")

}

func TestGetAlbumByID_ServiceError(t *testing.T) {
	mockRepo := new(MockAlbumRepo)
	svc := service.NewAlbumService(nil, mockRepo)
	handler := NewAlbumHandler(svc, nil)

	mockRepo.On("GetAlbumByID", mock.Anything, int64(42)).
		Return((*model.Album)(nil), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/albums/42", nil)
	req = mux.SetURLVars(req, map[string]string{"album_id": "42"})
	rr := httptest.NewRecorder()

	handler.GetAlbumByID(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), assert.AnError.Error())
	mockRepo.AssertExpectations(t)
}

func TestCreateAlbum_InvalidJSON(t *testing.T) {
	h := NewAlbumHandler(nil, redis.NewClient(&redis.Options{Addr: "localhost:6379"}))

	r := mux.NewRouter()
	r.HandleFunc("/albums", h.CreateAlbum).Methods("POST")

	req := httptest.NewRequest("POST", "/albums", bytes.NewBufferString(`{invalid json}`))
	req.Header.Set("Authorization", "bad-token")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
func TestUpdateAlbum_Success(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()

	_ = rdb.Set(ctx, "session:test-token", "42", 0).Err()
	_ = rdb.Set(ctx, "artist:42", "42", 0).Err()

	mockRepo := new(MockAlbumRepo)
	db, _, _ := sqlmock.New()
	svc := service.NewAlbumService(db, mockRepo)
	handler := NewAlbumHandler(svc, rdb)

	albumID := int64(1)
	currentArtistID := int64(42)

	mockRepo.On("GetAlbumByID", mock.Anything, albumID).
		Return(&model.Album{
			AlbumID:  albumID,
			Name:     "Old",
			GenreID:  1,
			ArtistID: currentArtistID,
		}, nil)

	mockRepo.On("GetSongsByAlbumID", mock.Anything, albumID).
		Return([]model.Song{}, nil)

	mockRepo.On("UpdateAlbum", mock.Anything, &model.Album{
		AlbumID:  albumID,
		Name:     "Updated Album",
		GenreID:  2,
		ArtistID: currentArtistID,
	}).Return(nil)

	body := `{"name": "Updated Album", "genre_id": 2}`
	req := httptest.NewRequest(http.MethodPut, "/albums/1", bytes.NewBufferString(body))
	req = mux.SetURLVars(req, map[string]string{"album_id": "1"})
	req.Header.Set("Authorization", "test-token")

	rr := httptest.NewRecorder()
	handler.UpdateAlbum(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var responseBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "Альбом успешно обновлён", responseBody["message"])

	mockRepo.AssertExpectations(t)
}

func TestUpdateAlbum_InvalidID(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	_ = rdb.Set(context.Background(), "session:token", "10", 0).Err()
	_ = rdb.Set(context.Background(), "artist:10", "20", 0).Err()

	handler := NewAlbumHandler(nil, rdb)

	req := httptest.NewRequest(http.MethodPut, "/albums/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"album_id": "invalid"})
	req.Header.Set("Authorization", "token")
	rr := httptest.NewRecorder()

	handler.UpdateAlbum(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdateAlbum_Forbidden(t *testing.T) {
	mockRepo := new(MockAlbumRepo)
	svc := service.NewAlbumService(nil, mockRepo)

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	_ = rdb.Set(context.Background(), "session:token", "10", 0).Err()
	_ = rdb.Set(context.Background(), "artist:10", "20", 0).Err()

	handler := NewAlbumHandler(svc, rdb)

	mockRepo.On("GetAlbumByID", mock.Anything, int64(30)).
		Return(&model.Album{AlbumID: 30, ArtistID: 99}, nil)

	mockRepo.On("GetSongsByAlbumID", mock.Anything, int64(30)).
		Return([]model.Song{}, nil)

	req := httptest.NewRequest(http.MethodPut, "/albums/30", bytes.NewBufferString(`{"name": "Updated", "genre_id": 2}`))
	req = mux.SetURLVars(req, map[string]string{"album_id": "30"})
	req.Header.Set("Authorization", "token")
	rr := httptest.NewRecorder()

	handler.UpdateAlbum(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestUpdateAlbum_Unauthorized(t *testing.T) {
	repo := new(MockAlbumRepo)
	svc := service.NewAlbumService(nil, repo)
	handler := NewAlbumHandler(svc, nil)

	req := httptest.NewRequest(http.MethodPut, "/albums/1", nil)
	req = mux.SetURLVars(req, map[string]string{"album_id": "1"})
	req.Header.Set("Authorization", "")
	rr := httptest.NewRecorder()

	handler.UpdateAlbum(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Неавторизованный доступ")
}
func TestGetAllAlbums_ServiceError(t *testing.T) {
	repo := new(MockAlbumRepo)
	svc := service.NewAlbumService(nil, repo)
	handler := NewAlbumHandler(svc, nil)

	repo.On("GetAllAlbums", mock.Anything).Return([]model.Album(nil), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/albums", nil)
	rr := httptest.NewRecorder()

	handler.GetAllAlbums(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Ошибка при получении альбомов")

	repo.AssertExpectations(t)
}

func TestGetAllAlbums_MethodNotAllowed(t *testing.T) {
	handler := NewAlbumHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/albums", nil)
	rr := httptest.NewRecorder()

	handler.GetAllAlbums(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "Метод не разрешён")
}
