package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"upload-service/internal/domain/model"
	"upload-service/internal/service"

	"github.com/DATA-DOG/go-sqlmock"
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
func (m *MockAlbumRepo) CheckSongsBelongToArtist(ctx context.Context, tx *sql.Tx, artistID int64, songIDs []int64) ([]int64, error) {
	args := m.Called(ctx, tx, artistID, songIDs)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockAlbumRepo) GetArtistByAlbumID(ctx context.Context, albumID int64) (*model.Artist, error) {
	args := m.Called(ctx, albumID)
	return args.Get(0).(*model.Artist), args.Error(1)
}

func (m *MockAlbumRepo) GetGenreByAlbumID(ctx context.Context, albumID int64) (*model.Genre, error) {
	args := m.Called(ctx, albumID)
	return args.Get(0).(*model.Genre), args.Error(1)
}

func TestGetAllAlbums_Success(t *testing.T) {
	mockRepo := new(MockAlbumRepo)
	db, _, _ := sqlmock.New()
	service := service.NewAlbumService(db, mockRepo, nil)
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
	service := service.NewAlbumService(db, mockRepo, nil)
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
	svc := service.NewAlbumService(nil, mockRepo, nil)
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

func TestGetAllAlbums_ServiceError(t *testing.T) {
	repo := new(MockAlbumRepo)
	svc := service.NewAlbumService(nil, repo, nil)
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
