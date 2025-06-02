package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"upload-service/internal/domain/model"
	"upload-service/internal/kafka"
	"upload-service/internal/service"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockArtistRepo struct {
	mock.Mock
}

func (m *MockArtistRepo) GetAllArtists(ctx context.Context) ([]model.Artist, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Artist), args.Error(1)
}

func (m *MockArtistRepo) GetArtistByID(ctx context.Context, id int64) (*model.Artist, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Artist), args.Error(1)
}

func (m *MockArtistRepo) GetSongsByArtistID(ctx context.Context, id int64) ([]model.Song, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]model.Song), args.Error(1)
}

func (m *MockArtistRepo) GetAlbumsByArtistID(ctx context.Context, id int64) ([]model.Album, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]model.Album), args.Error(1)
}

func (m *MockArtistRepo) GetArtistByUserID(ctx context.Context, id int64) (*model.Artist, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Artist), args.Error(1)
}

func (m *MockArtistRepo) CreateArtist(ctx context.Context, artist *model.Artist, userID int64) (int64, error) {
	args := m.Called(ctx, artist, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockArtistRepo) UpdateArtist(ctx context.Context, artist *model.Artist) error {
	args := m.Called(ctx, artist)
	return args.Error(0)
}

func TestGetAllArtists_Success(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	dummyProducer := new(kafka.Producer)
	svc := service.NewArtistService(mockRepo, "test-bucket", "folder", dummyProducer)
	handler := NewArtistHandler(svc, nil)

	mockRepo.On("GetAllArtists", mock.Anything).Return([]model.Artist{
		{ArtistID: 1, Name: "Test Artist"},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/artists", nil)
	rr := httptest.NewRecorder()

	handler.GetAllArtists(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestGetArtistByID_Success(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	dummyProducer := new(kafka.Producer)
	svc := service.NewArtistService(mockRepo, "test-bucket", "folder", dummyProducer)
	handler := NewArtistHandler(svc, nil)

	artistID := int64(1)

	mockRepo.On("GetArtistByID", mock.Anything, artistID).
		Return(&model.Artist{ArtistID: artistID, Name: "Artist"}, nil)

	mockRepo.On("GetSongsByArtistID", mock.Anything, artistID).
		Return([]model.Song{{SongID: 1, Name: "Song"}}, nil)

	mockRepo.On("GetAlbumsByArtistID", mock.Anything, artistID).
		Return([]model.Album{{AlbumID: 1, Name: "Album"}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/artists/1", nil)
	req = mux.SetURLVars(req, map[string]string{"artist_id": "1"})
	rr := httptest.NewRecorder()

	handler.GetArtistByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}
func TestRegisterArtist_Success(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	dummyProducer := new(kafka.Producer)
	svc := service.NewArtistService(mockRepo, "test-bucket", "folder", dummyProducer)

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "Rbkkth3920",
		DB:       0,
	})

	_ = rdb.Set(context.Background(), "session:token", "10", 0).Err()
	_ = rdb.Del(context.Background(), "artist:10").Err()

	handler := NewArtistHandler(svc, rdb)

	reqBody := `{"name": "New Artist"}`
	req := httptest.NewRequest(http.MethodPost, "/artists/register/me", bytes.NewBufferString(reqBody))
	req.Header.Set("Authorization", "token")
	rr := httptest.NewRecorder()

	expectedArtist := &model.Artist{Name: "New Artist", UserID: 10}
	mockRepo.On("GetArtistByUserID", mock.Anything, int64(10)).Return((*model.Artist)(nil), nil)

	mockRepo.On("CreateArtist", mock.Anything, expectedArtist, int64(10)).Return(int64(123), nil)

	handler.RegisterArtist(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}
func TestUpdateArtist_Success(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	dummyProducer := new(kafka.Producer)
	svc := service.NewArtistService(mockRepo, "test-bucket", "folder", dummyProducer)

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "Rbkkth3920",
		DB:       0,
	})

	_ = rdb.Set(context.Background(), "session:token", "10", 0).Err()
	_ = rdb.Set(context.Background(), "artist:10", "20", 0).Err()

	handler := NewArtistHandler(svc, rdb)

	reqBody := `{"name": "Updated Artist"}`
	req := httptest.NewRequest(http.MethodPut, "/artists/20", bytes.NewBufferString(reqBody))
	req = mux.SetURLVars(req, map[string]string{"artist_id": "20"})
	req.Header.Set("Authorization", "token")
	rr := httptest.NewRecorder()

	expectedArtist := &model.Artist{ArtistID: 20, Name: "Updated Artist"}

	mockRepo.On("GetArtistByID", mock.Anything, int64(20)).
		Return(&model.Artist{ArtistID: 20, Name: "Old Artist"}, nil)

	mockRepo.On("UpdateArtist", mock.Anything, expectedArtist).Return(nil)

	handler.UpdateArtist(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUpdateArtist_MethodNotAllowed(t *testing.T) {
	handler := NewArtistHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/artists/20", nil)
	req = mux.SetURLVars(req, map[string]string{"artist_id": "20"})
	rr := httptest.NewRecorder()

	handler.UpdateArtist(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "Метод не разрешён")
}

func TestUpdateArtist_InvalidID(t *testing.T) {
	handler := NewArtistHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPut, "/artists/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"artist_id": "abc"})
	rr := httptest.NewRecorder()

	handler.UpdateArtist(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Некорректный ID артиста")
}

func TestUpdateArtist_ForbiddenAnotherArtist(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "Rbkkth3920",
		DB:       0,
	})
	_ = rdb.Set(context.Background(), "session:token", "10", 0).Err()
	_ = rdb.Set(context.Background(), "artist:10", "99", 0).Err()

	handler := NewArtistHandler(nil, rdb)

	req := httptest.NewRequest(http.MethodPut, "/artists/20", bytes.NewBufferString(`{"name": "X"}`))
	req = mux.SetURLVars(req, map[string]string{"artist_id": "20"})
	req.Header.Set("Authorization", "token")
	rr := httptest.NewRecorder()

	handler.UpdateArtist(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "не можете редактировать")
}

func TestUpdateArtist_InvalidJSON(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "Rbkkth3920",
		DB:       0,
	})
	_ = rdb.Set(context.Background(), "session:token", "10", 0).Err()
	_ = rdb.Set(context.Background(), "artist:10", "20", 0).Err()

	handler := NewArtistHandler(nil, rdb)

	req := httptest.NewRequest(http.MethodPut, "/artists/20", bytes.NewBufferString(`{invalid}`))
	req = mux.SetURLVars(req, map[string]string{"artist_id": "20"})
	req.Header.Set("Authorization", "token")
	rr := httptest.NewRecorder()

	handler.UpdateArtist(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Неверный формат запроса")
}
func TestUpdateArtist_ServiceError(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	dummyProducer := new(kafka.Producer)
	svc := service.NewArtistService(mockRepo, "test-bucket", "folder", dummyProducer)

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "Rbkkth3920",
		DB:       0,
	})
	_ = rdb.Set(context.Background(), "session:token", "10", 0).Err()
	_ = rdb.Set(context.Background(), "artist:10", "20", 0).Err()

	handler := NewArtistHandler(svc, rdb)

	expectedArtist := &model.Artist{ArtistID: 20, Name: "Updated Artist"}

	mockRepo.On("GetArtistByID", mock.Anything, int64(20)).
		Return(&model.Artist{ArtistID: 20, Name: "Old Artist"}, nil)

	mockRepo.On("UpdateArtist", mock.Anything, expectedArtist).
		Return(assert.AnError)

	req := httptest.NewRequest(http.MethodPut, "/artists/20", bytes.NewBufferString(`{"name":"Updated Artist"}`))
	req = mux.SetURLVars(req, map[string]string{"artist_id": "20"})
	req.Header.Set("Authorization", "token")
	rr := httptest.NewRecorder()

	handler.UpdateArtist(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), assert.AnError.Error())

	mockRepo.AssertExpectations(t)
}

func TestGetArtistByID_MethodNotAllowed(t *testing.T) {
	handler := NewArtistHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/artists/1", nil)
	req = mux.SetURLVars(req, map[string]string{"artist_id": "1"})
	rr := httptest.NewRecorder()

	handler.GetArtistByID(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}
func TestGetArtistByID_ServiceError(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	dummyProducer := new(kafka.Producer)
	svc := service.NewArtistService(mockRepo, "test-bucket", "folder", dummyProducer)
	handler := NewArtistHandler(svc, nil)

	mockRepo.On("GetArtistByID", mock.Anything, int64(42)).
		Return((*model.Artist)(nil), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/artists/42", nil)
	req = mux.SetURLVars(req, map[string]string{"artist_id": "42"})
	rr := httptest.NewRecorder()

	handler.GetArtistByID(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), assert.AnError.Error())
	mockRepo.AssertExpectations(t)
}

func TestGetAllArtists_MethodNotAllowed(t *testing.T) {
	handler := NewArtistHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/artists", nil)
	rr := httptest.NewRecorder()

	handler.GetAllArtists(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "Метод не разрешён")
}

func TestGetAllArtists_ServiceError(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	dummyProducer := new(kafka.Producer)
	svc := service.NewArtistService(mockRepo, "test-bucket", "folder", dummyProducer)
	handler := NewArtistHandler(svc, nil)

	mockRepo.On("GetAllArtists", mock.Anything).Return([]model.Artist{}, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/artists", nil)
	rr := httptest.NewRecorder()

	handler.GetAllArtists(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Ошибка при получении артистов")

	mockRepo.AssertExpectations(t)
}

func TestRegisterArtist_MethodNotAllowed(t *testing.T) {
	handler := NewArtistHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/artists/register/me", nil)
	rr := httptest.NewRecorder()

	handler.RegisterArtist(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "Метод не разрешён")
}

func TestRegisterArtist_InvalidJSON(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "Rbkkth3920",
		DB:       0,
	})
	_ = rdb.Set(context.Background(), "session:token", "10", 0).Err()

	handler := NewArtistHandler(nil, rdb)

	req := httptest.NewRequest(http.MethodPost, "/artists/register/me", bytes.NewBufferString(`{invalid_json}`))
	req.Header.Set("Authorization", "token")
	rr := httptest.NewRecorder()

	handler.RegisterArtist(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Неверный формат запроса")
}

func TestRegisterArtist_ServiceError(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	dummyProducer := new(kafka.Producer)
	svc := service.NewArtistService(mockRepo, "test-bucket", "folder", dummyProducer)

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "Rbkkth3920",
		DB:       0,
	})
	_ = rdb.Set(context.Background(), "session:token", "10", 0).Err()
	_ = rdb.Del(context.Background(), "artist:10").Err()

	handler := NewArtistHandler(svc, rdb)

	reqBody := `{"name": "Test Artist"}`
	req := httptest.NewRequest(http.MethodPost, "/artists/register/me", bytes.NewBufferString(reqBody))
	req.Header.Set("Authorization", "token")
	rr := httptest.NewRecorder()

	mockRepo.On("GetArtistByUserID", mock.Anything, int64(10)).Return((*model.Artist)(nil), nil)
	mockRepo.On("CreateArtist", mock.Anything, mock.AnythingOfType("*model.Artist"), int64(10)).Return(int64(0), assert.AnError)

	handler.RegisterArtist(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), assert.AnError.Error())

	mockRepo.AssertExpectations(t)
}
