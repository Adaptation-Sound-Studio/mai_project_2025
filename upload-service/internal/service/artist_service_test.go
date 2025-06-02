package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
	"upload-service/internal/config"
	"upload-service/internal/domain/model"
	"upload-service/internal/infrastructure/elastic"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func (m *MockArtistRepo) GetSongsByArtistID(ctx context.Context, artistID int64) ([]model.Song, error) {
	args := m.Called(ctx, artistID)
	return args.Get(0).([]model.Song), args.Error(1)
}

func (m *MockArtistRepo) GetAlbumsByArtistID(ctx context.Context, artistID int64) ([]model.Album, error) {
	args := m.Called(ctx, artistID)
	return args.Get(0).([]model.Album), args.Error(1)
}

func (m *MockArtistRepo) GetArtistByUserID(ctx context.Context, userID int64) (*model.Artist, error) {
	args := m.Called(ctx, userID)
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

type mockProducer struct{}

func TestGetAllArtists_Success(t *testing.T) {
	repo := new(MockArtistRepo)
	es := &elasticsearch.Client{}
	svc := NewArtistService(repo, "", "", nil, es)

	expected := []model.Artist{{ArtistID: 1, Name: "Artist 1"}}
	repo.On("GetAllArtists", mock.Anything).Return(expected, nil)

	result, err := svc.GetAllArtists(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestArtistService_GetAllArtists_Error(t *testing.T) {
	repo := new(MockArtistRepo)
	svc := NewArtistService(repo, "", "", nil, nil)

	ctx := context.Background()

	expectedErr := errors.New("db error")

	repo.On("GetAllArtists", ctx).Return([]model.Artist{}, expectedErr)

	artists, err := svc.GetAllArtists(ctx)

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Empty(t, artists)

	repo.AssertExpectations(t)
}

func TestGetArtistByID_Success(t *testing.T) {
	repo := new(MockArtistRepo)
	svc := NewArtistService(repo, "", "", nil, nil)

	artist := &model.Artist{ArtistID: 1, Name: "Artist"}
	songs := []model.Song{{SongID: 1, Name: "Song"}}
	albums := []model.Album{{AlbumID: 1, Name: "Album"}}

	repo.On("GetArtistByID", mock.Anything, int64(1)).Return(artist, nil)
	repo.On("GetSongsByArtistID", mock.Anything, int64(1)).Return(songs, nil)
	repo.On("GetAlbumsByArtistID", mock.Anything, int64(1)).Return(albums, nil)

	res, err := svc.GetArtistByID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, artist.Name, res.Name)
	assert.Len(t, res.Songs, 1)
	assert.Len(t, res.Albums, 1)
	repo.AssertExpectations(t)
}

func TestRegisterArtist_Success(t *testing.T) {
	repo := new(MockArtistRepo)
	svc := NewArtistService(repo, "", "", nil, nil)

	userID := int64(42)
	artist := &model.Artist{Name: "New Artist"}

	repo.On("GetArtistByUserID", mock.Anything, userID).Return((*model.Artist)(nil), nil)
	repo.On("CreateArtist", mock.Anything, artist, userID).Return(int64(10), nil)

	id, err := svc.RegisterArtist(context.Background(), artist, userID)

	require.NoError(t, err)
	assert.Equal(t, int64(10), id)
	repo.AssertExpectations(t)
}

func TestUpdateArtist_Success(t *testing.T) {
	repo := new(MockArtistRepo)
	svc := NewArtistService(repo, "", "", nil, nil)

	artist := &model.Artist{ArtistID: 1, Name: "Updated"}

	repo.On("GetArtistByID", mock.Anything, int64(1)).Return(artist, nil)
	repo.On("UpdateArtist", mock.Anything, artist).Return(nil)

	err := svc.UpdateArtist(context.Background(), artist)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdateArtist_InvalidID(t *testing.T) {
	repo := new(MockArtistRepo)
	svc := NewArtistService(repo, "", "", nil, nil)

	err := svc.UpdateArtist(context.Background(), &model.Artist{
		ArtistID: 0,
		Name:     "X",
	})

	assert.EqualError(t, err, "некорректный ID артиста для обновления")
}

func TestUpdateArtist_ArtistNotFound(t *testing.T) {
	repo := new(MockArtistRepo)
	svc := NewArtistService(repo, "", "", nil, nil)

	artist := &model.Artist{
		ArtistID: 7,
		Name:     "New Name",
	}

	repo.On("GetArtistByID", mock.Anything, int64(7)).
		Return((*model.Artist)(nil), nil)

	err := svc.UpdateArtist(context.Background(), artist)

	require.Error(t, err)
	assert.Equal(t, ErrArtistNotFound, err)
	repo.AssertExpectations(t)
}

func TestUpdateArtist_EmptyName(t *testing.T) {
	repo := new(MockArtistRepo)
	svc := NewArtistService(repo, "", "", nil, nil)

	repo.On("GetArtistByID", mock.Anything, int64(8)).
		Return(&model.Artist{ArtistID: 8, Name: "Y"}, nil)

	err := svc.UpdateArtist(context.Background(), &model.Artist{ArtistID: 8, Name: ""})

	assert.EqualError(t, err, "имя артиста не может быть пустым")
	repo.AssertExpectations(t)
}

func TestUpdateArtist_RepoUpdateError(t *testing.T) {
	repo := new(MockArtistRepo)
	svc := NewArtistService(repo, "", "", nil, nil)

	repo.On("GetArtistByID", mock.Anything, int64(9)).
		Return(&model.Artist{ArtistID: 9, Name: "Old"}, nil)

	repo.On("UpdateArtist", mock.Anything, &model.Artist{ArtistID: 9, Name: "New"}).
		Return(errors.New("update failed"))

	err := svc.UpdateArtist(context.Background(), &model.Artist{ArtistID: 9, Name: "New"})

	assert.EqualError(t, err, "update failed")
	repo.AssertExpectations(t)
}

func TestGetArtistByID_InvalidID(t *testing.T) {
	repo := new(MockArtistRepo)
	svc := NewArtistService(repo, "", "", nil, nil)

	_, err := svc.GetArtistByID(context.Background(), 0)

	assert.EqualError(t, err, "некорректный ID артиста")
}

func TestGetArtistByID_GetArtistError(t *testing.T) {
	repo := new(MockArtistRepo)
	svc := NewArtistService(repo, "", "", nil, nil)

	repo.On("GetArtistByID", mock.Anything, int64(1)).
		Return((*model.Artist)(nil), errors.New("db error"))

	_, err := svc.GetArtistByID(context.Background(), 1)

	assert.EqualError(t, err, "db error")
	repo.AssertExpectations(t)
}

func TestGetArtistByID_ArtistNotFound(t *testing.T) {
	repo := new(MockArtistRepo)
	svc := NewArtistService(repo, "", "", nil, nil)

	repo.On("GetArtistByID", mock.Anything, int64(1)).
		Return((*model.Artist)(nil), nil)

	_, err := svc.GetArtistByID(context.Background(), 1)

	assert.Equal(t, ErrArtistNotFound, err)
	repo.AssertExpectations(t)
}

func TestSearchArtists_Integration(t *testing.T) {
	cfg := &config.ElasticConfig{
		URL:      "http://localhost:9200",
		Username: "elastic",
		Password: "your_password",
	}

	esClient, err := elastic.NewElasticClient(cfg)
	require.NoError(t, err)

	service := &ArtistService{
		elasticClient: esClient,
	}

	ctx := context.Background()

	artist := model.Artist{
		ArtistID: 9999,
		Name:     "Electroman",
		UserID:   1234,
	}

	body := fmt.Sprintf(`{"artist_id": %d, "name": "%s", "user_id": %d}`, artist.ArtistID, artist.Name, artist.UserID)
	_, err = esClient.Index("artists", strings.NewReader(body),
		esClient.Index.WithDocumentID(fmt.Sprint(artist.ArtistID)),
		esClient.Index.WithRefresh("true"),
	)
	require.NoError(t, err)

	results, err := service.SearchArtists(ctx, "Electroman")
	require.NoError(t, err)

	require.NotEmpty(t, results)

	var found bool
	for _, a := range results {
		if a.ArtistID == artist.ArtistID && a.Name == artist.Name {
			found = true
			break
		}
	}
	assert.True(t, found, "Artist not found in search results")
}

func TestGetArtistIDByUserID(t *testing.T) {
	ctx := context.Background()

	topic := "artist-topic"
	bucket := "test-bucket"

	t.Run("artist found", func(t *testing.T) {
		mockRepo := new(MockArtistRepo)
		svc := NewArtistService(mockRepo, topic, bucket, nil, nil)

		artist := &model.Artist{
			ArtistID: 1234,
			Name:     "Test Artist",
			UserID:   1,
		}

		mockRepo.On("GetArtistByUserID", ctx, int64(1)).Return(artist, nil)

		gotID, err := svc.GetArtistIDByUserID(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, strconv.FormatInt(artist.ArtistID, 10), gotID)
		mockRepo.AssertCalled(t, "GetArtistByUserID", ctx, int64(1))
	})

	t.Run("artist not found (nil result)", func(t *testing.T) {
		mockRepo := new(MockArtistRepo)
		svc := NewArtistService(mockRepo, topic, bucket, nil, nil)

		mockRepo.On("GetArtistByUserID", ctx, int64(2)).Return((*model.Artist)(nil), nil)

		gotID, err := svc.GetArtistIDByUserID(ctx, 2)

		assert.NoError(t, err)
		assert.Equal(t, "", gotID)
		mockRepo.AssertCalled(t, "GetArtistByUserID", ctx, int64(2))
	})

	t.Run("repository returns error", func(t *testing.T) {
		mockRepo := new(MockArtistRepo)
		svc := NewArtistService(mockRepo, topic, bucket, nil, nil)

		mockRepo.On("GetArtistByUserID", ctx, int64(3)).Return((*model.Artist)(nil), errors.New("db error"))

		gotID, err := svc.GetArtistIDByUserID(ctx, 3)

		assert.Error(t, err)
		assert.Equal(t, "", gotID)
		mockRepo.AssertCalled(t, "GetArtistByUserID", ctx, int64(3))
	})
}
func TestIndexArtist_Integration(t *testing.T) {
	cfg := &config.ElasticConfig{
		URL:      "http://localhost:9200",
		Username: "elastic",
		Password: "your_password",
	}

	esClient, err := elastic.NewElasticClient(cfg)
	require.NoError(t, err, "ошибка подключения к Elasticsearch")

	svc := &ArtistService{
		elasticClient: esClient,
	}

	ctx := context.Background()

	artist := &model.Artist{
		ArtistID: 123456789,
		Name:     "Integration Test Artist",
		UserID:   777,
	}

	err = svc.indexArtist(ctx, artist)
	require.NoError(t, err, "ошибка при индексировании")

	time.Sleep(1 * time.Second)

	docID := fmt.Sprint(artist.ArtistID)
	getResp, err := esClient.Get("artists", docID)
	require.NoError(t, err)
	defer getResp.Body.Close()

	assert.False(t, getResp.IsError(), "документ не найден в индексе")

	body, err := io.ReadAll(getResp.Body)
	require.NoError(t, err)

	assert.Contains(t, string(body), artist.Name)

	_, _ = esClient.Delete("artists", docID)
}

func TestRegisterArtist_ErrorMarshalling(t *testing.T) {
	repo := new(MockArtistRepo)
	producer := new(MockProducer)

	svc := &ArtistService{
		repo:        repo,
		producer:    nil,
		authBaseURL: "http://fake-auth",
		apiKey:      "test-key",
	}

	artist := &model.Artist{Name: "Test Artist"}

	repo.On("GetArtistByUserID", mock.Anything, int64(42)).Return((*model.Artist)(nil), nil)
	repo.On("CreateArtist", mock.Anything, artist, int64(42)).Return(int64(101), nil)

	id, err := svc.RegisterArtist(context.Background(), artist, 42)

	assert.Equal(t, int64(101), id)
	assert.NoError(t, err)

	producer.AssertNotCalled(t, "SendWrappedEvent")
}
func TestRegisterArtist_AuthServiceFailure(t *testing.T) {
	repo := new(MockArtistRepo)
	producer := new(MockProducer)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	svc := &ArtistService{
		repo:        repo,
		producer:    producer,
		authBaseURL: testServer.URL,
		apiKey:      "fake",
	}

	artist := &model.Artist{Name: "AuthFail"}

	repo.On("GetArtistByUserID", mock.Anything, int64(1)).
		Return((*model.Artist)(nil), nil)

	repo.On("CreateArtist", mock.Anything, artist, int64(1)).
		Return(int64(10), nil)

	producer.On("SendWrappedEvent", "artist_created", mock.Anything).
		Return(nil)

	id, err := svc.RegisterArtist(context.Background(), artist, 1)

	assert.NoError(t, err)
	assert.Equal(t, int64(10), id)
	producer.AssertCalled(t, "SendWrappedEvent", "artist_created", mock.Anything)
}
