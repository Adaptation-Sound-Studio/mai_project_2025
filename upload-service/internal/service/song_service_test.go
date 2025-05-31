package service

import (
	"context"
	"errors"
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

func TestFindMissingIDs(t *testing.T) {
	input := []int64{1, 2, 3, 4, 5}
	existing := []int64{2, 4, 5}

	expected := []int64{1, 3}
	result := findMissingIDs(input, existing)

	assert.ElementsMatch(t, expected, result)
}

func TestUpdateSong_InvalidID(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	err := svc.UpdateSong(context.Background(), &model.Song{SongID: 0})
	assert.EqualError(t, err, "некорректный ID песни для обновления")
}

func TestUpdateSong_EmptyName(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	repo.On("GetSongByID", mock.Anything, int64(1)).
		Return(&model.Song{SongID: 1}, nil)

	err := svc.UpdateSong(context.Background(), &model.Song{
		SongID: 1, Name: "", GenreID: 1, Link: "link",
	})
	assert.EqualError(t, err, "название песни не может быть пустым")
}

func TestUpdateSong_InvalidGenre(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	repo.On("GetSongByID", mock.Anything, int64(2)).
		Return(&model.Song{SongID: 2}, nil)

	err := svc.UpdateSong(context.Background(), &model.Song{
		SongID: 2, Name: "test", GenreID: 0, Link: "link",
	})
	assert.EqualError(t, err, "указан некорректный жанр")
}

func TestUpdateSong_EmptyLink(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	repo.On("GetSongByID", mock.Anything, int64(3)).
		Return(&model.Song{SongID: 3}, nil)

	err := svc.UpdateSong(context.Background(), &model.Song{
		SongID: 3, Name: "test", GenreID: 1, Link: "",
	})
	assert.EqualError(t, err, "ссылка на песню не может быть пустой")
}

func TestUpdateSong_RepoUpdateError(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	song := &model.Song{
		SongID: 4, Name: "test", GenreID: 1, Link: "link",
	}
	repo.On("GetSongByID", mock.Anything, int64(4)).
		Return(song, nil)
	repo.On("UpdateSong", mock.Anything, song).
		Return(errors.New("update failed"))

	err := svc.UpdateSong(context.Background(), song)
	assert.EqualError(t, err, "update failed")
}

func TestGetSongByID_InvalidID(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	res, err := svc.GetSongByID(context.Background(), 0)

	assert.Nil(t, res)
	assert.EqualError(t, err, "ID песни должен быть положительным числом")
}

func TestGetSongByID_DBError(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	repo.On("GetSongByID", mock.Anything, int64(1)).
		Return((*model.Song)(nil), errors.New("db error"))

	res, err := svc.GetSongByID(context.Background(), 1)

	assert.Nil(t, res)
	assert.EqualError(t, err, "db error")
	repo.AssertExpectations(t)
}

func TestGetSongByID_NotFound(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	repo.On("GetSongByID", mock.Anything, int64(1)).
		Return((*model.Song)(nil), nil)

	res, err := svc.GetSongByID(context.Background(), 1)

	assert.Nil(t, res)
	assert.Equal(t, ErrSongNotFound, err)
	repo.AssertExpectations(t)
}

func TestGetSongByID_AlbumError(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

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
	svc := NewSongService(repo)

	err := svc.CreateSong(context.Background(), &model.Song{
		Name:    "",
		GenreID: 1,
		Link:    "link.mp3",
	}, []int64{1})

	assert.EqualError(t, err, "название песни не может быть пустым")
}

func TestCreateSong_InvalidGenre(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	err := svc.CreateSong(context.Background(), &model.Song{
		Name:    "Test Song",
		GenreID: 0,
		Link:    "link.mp3",
	}, []int64{1})

	assert.EqualError(t, err, "указан некорректный жанр")
}

func TestCreateSong_EmptyLink(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	err := svc.CreateSong(context.Background(), &model.Song{
		Name:    "Test Song",
		GenreID: 1,
		Link:    "",
	}, []int64{1})

	assert.EqualError(t, err, "ссылка на песню не может быть пустой")
}

func TestCreateSong_NoArtists(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	err := svc.CreateSong(context.Background(), &model.Song{
		Name:    "Test Song",
		GenreID: 1,
		Link:    "link.mp3",
	}, []int64{})

	assert.EqualError(t, err, "нужно указать хотя бы одного артиста")
}

func TestCreateSong_TooManyArtists(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	tooMany := make([]int64, 11)
	for i := 0; i < 11; i++ {
		tooMany[i] = int64(i + 1)
	}

	err := svc.CreateSong(context.Background(), &model.Song{
		Name:    "Test Song",
		GenreID: 1,
		Link:    "link.mp3",
	}, tooMany)

	assert.EqualError(t, err, "можно указать не более 10 артистов")
}

func TestCreateSong_ArtistNotExist(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	input := []int64{1, 2, 3}
	existing := []int64{1, 3}
	song := &model.Song{Name: "Song", GenreID: 1, Link: "link.mp3"}

	repo.On("CheckArtistsExist", mock.Anything, input).
		Return(existing, nil)

	err := svc.CreateSong(context.Background(), song, input)

	assert.EqualError(t, err, "артисты с ID [2] не существуют")
	repo.AssertExpectations(t)
}

func TestCreateSong_CreateError(t *testing.T) {
	repo := new(MockSongRepo)
	svc := NewSongService(repo)

	artistIDs := []int64{1, 2}
	song := &model.Song{Name: "Song", GenreID: 1, Link: "link.mp3"}

	repo.On("CheckArtistsExist", mock.Anything, artistIDs).
		Return(artistIDs, nil)

	repo.On("CreateSongWithArtists", mock.Anything, song, artistIDs).
		Return(int64(0), errors.New("insert error"))

	err := svc.CreateSong(context.Background(), song, artistIDs)

	assert.EqualError(t, err, "insert error")
	repo.AssertExpectations(t)
}
