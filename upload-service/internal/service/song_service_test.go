package service

import (
	"context"
	"testing"
	"upload-service/internal/domain/model"

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

func TestGetAllSongs_Success(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	songs := []model.Song{{SongID: 1, Name: "Track 1"}}
	artists := []model.Artist{{ArtistID: 1, Name: "Artist"}}
	album := &model.Album{AlbumID: 1, Name: "Album"}

	repo.On("GetAllSongs", mock.Anything).Return(songs, nil)
	repo.On("GetArtistsBySongID", mock.Anything, int64(1)).Return(artists, nil)
	repo.On("GetAlbumBySongID", mock.Anything, int64(1)).Return(album, nil)

	result, err := svc.GetAllSongs(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Track 1", result[0].Name)
	repo.AssertExpectations(t)
}

func TestGetSongByID_Success(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

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
	repo.AssertExpectations(t)
}

func TestCreateSong_Success(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	song := &model.Song{
		Name:    "New Song",
		GenreID: 2,
		Link:    "https://song.link",
	}
	artistIDs := []int64{1, 2}

	repo.On("CheckArtistsExist", mock.Anything, artistIDs).Return(artistIDs, nil)
	repo.On("CreateSongWithArtists", mock.Anything, song, artistIDs).Return(int64(10), nil)

	err := svc.CreateSong(context.Background(), song, artistIDs)

	require.NoError(t, err)
	assert.Equal(t, int64(0), song.Auditions)
	repo.AssertExpectations(t)
}

func TestUpdateSong_Success(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	song := &model.Song{SongID: 1, Name: "Update", GenreID: 1, Link: "link"}
	repo.On("GetSongByID", mock.Anything, int64(1)).Return(song, nil)
	repo.On("UpdateSong", mock.Anything, song).Return(nil)

	err := svc.UpdateSong(context.Background(), song)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetSongArtists_Success(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	expected := []model.Artist{{ArtistID: 1, Name: "Artist"}}
	repo.On("GetArtistsBySongID", mock.Anything, int64(1)).Return(expected, nil)

	result, err := svc.GetSongArtists(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}
