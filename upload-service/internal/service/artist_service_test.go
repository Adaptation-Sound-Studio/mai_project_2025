package service

import (
	"context"
	"testing"
	"upload-service/internal/domain/model"

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

func TestGetAllArtists_Success(t *testing.T) {
	repo := new(MockArtistRepo)
	svc := NewArtistService(repo)

	expected := []model.Artist{{ArtistID: 1, Name: "Artist 1"}}
	repo.On("GetAllArtists", mock.Anything).Return(expected, nil)

	result, err := svc.GetAllArtists(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestGetArtistByID_Success(t *testing.T) {
	repo := new(MockArtistRepo)
	svc := NewArtistService(repo)

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
	svc := NewArtistService(repo)

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
	svc := NewArtistService(repo)

	artist := &model.Artist{ArtistID: 1, Name: "Updated"}

	repo.On("GetArtistByID", mock.Anything, int64(1)).Return(artist, nil)
	repo.On("UpdateArtist", mock.Anything, artist).Return(nil)

	err := svc.UpdateArtist(context.Background(), artist)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}
