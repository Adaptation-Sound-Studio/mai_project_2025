package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"upload-service/internal/domain/model"
	"upload-service/internal/service"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/lib/pq"
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

func (m *MockAlbumRepo) CheckSongsExist(ctx context.Context, tx *sql.Tx, songIDs []int64) ([]int64, error) {
	args := m.Called(ctx, tx, songIDs)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockAlbumRepo) BatchInsertSongsToAlbum(ctx context.Context, tx *sql.Tx, albumID int64, songIDs []int64) error {
	args := m.Called(ctx, tx, albumID, songIDs)
	return args.Error(0)
}

func (m *MockAlbumRepo) UpdateAlbum(ctx context.Context, album *model.Album) error {
	args := m.Called(ctx, album)
	return args.Error(0)
}

type mockDB struct{}

func (m *mockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, nil
}

func TestGetAllAlbums_Success(t *testing.T) {
	repo := new(MockAlbumRepo)
	service := service.NewAlbumService(nil, repo)

	expected := []model.Album{{AlbumID: 1, Name: "A"}}
	repo.On("GetAllAlbums", mock.Anything).Return(expected, nil)

	got, err := service.GetAllAlbums(context.Background())
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestGetAlbumByID_Success(t *testing.T) {
	repo := new(MockAlbumRepo)
	svc := service.NewAlbumService(nil, repo)

	album := &model.Album{
		AlbumID: 1,
		Name:    "Test Album",
	}
	songs := []model.Song{
		{SongID: 1, Name: "Track 1"},
	}

	repo.On("GetAlbumByID", mock.Anything, int64(1)).Return(album, nil)
	repo.On("GetSongsByAlbumID", mock.Anything, int64(1)).Return(songs, nil)

	res, err := svc.GetAlbumByID(context.Background(), 1)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, album.Name, res.Name)
	assert.Len(t, res.Songs, 1)

	repo.AssertExpectations(t)
}

func TestGetAlbumByID_NotFound(t *testing.T) {
	repo := new(MockAlbumRepo)

	svc := service.NewAlbumService(nil, repo)

	repo.On("GetAlbumByID", mock.Anything, int64(123)).Return((*model.Album)(nil), nil)

	got, err := svc.GetAlbumByID(context.Background(), 123)

	assert.Nil(t, got)
	assert.Equal(t, service.ErrAlbumNotFound, err)

}
func TestCreateAlbum_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := new(MockAlbumRepo)
	svc := service.NewAlbumService(db, repo)

	album := &model.Album{
		Name:     "Test Album",
		ArtistID: 1,
		GenreID:  2,
	}
	songIDs := []int64{1, 2}

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repo.On("CreateAlbum", mock.Anything, mock.Anything, album).
		Return(int64(101), nil)
	repo.On("CheckSongsExist", mock.Anything, mock.Anything, songIDs).
		Return(songIDs, nil)
	repo.On("BatchInsertSongsToAlbum", mock.Anything, mock.Anything, int64(101), songIDs).
		Return(nil)

	id, err := svc.CreateAlbum(context.Background(), album, songIDs)
	require.NoError(t, err)
	assert.Equal(t, int64(101), id)

	repo.AssertExpectations(t)
	require.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateAlbum_ValidationError(t *testing.T) {
	s := service.NewAlbumService(nil, nil)

	err := s.UpdateAlbum(context.Background(), &model.Album{AlbumID: 0})
	assert.EqualError(t, err, "некорректный ID альбома для обновления")
}

func TestUpdateAlbum_Success(t *testing.T) {
	repo := new(MockAlbumRepo)
	svc := service.NewAlbumService(nil, repo)

	album := &model.Album{
		AlbumID: 1,
		Name:    "Updated",
		GenreID: 2,
	}

	repo.On("GetAlbumByID", mock.Anything, int64(1)).
		Return(&model.Album{AlbumID: 1, Name: "Old", GenreID: 1}, nil)
	repo.On("UpdateAlbum", mock.Anything, album).
		Return(nil)

	err := svc.UpdateAlbum(context.Background(), album)
	require.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestCreateAlbum_EmptyGenre(t *testing.T) {
	s := service.NewAlbumService(nil, nil)

	album := &model.Album{Name: "Test", ArtistID: 1, GenreID: 0}
	id, err := s.CreateAlbum(context.Background(), album, nil)
	assert.Equal(t, int64(0), id)
	assert.EqualError(t, err, "не указан жанр альбома")
}

func TestUpdateAlbum_InvalidID(t *testing.T) {
	repo := new(MockAlbumRepo)
	svc := service.NewAlbumService(nil, repo)

	err := svc.UpdateAlbum(context.Background(), &model.Album{
		AlbumID: 0,
		Name:    "Test",
		GenreID: 1,
	})

	assert.EqualError(t, err, "некорректный ID альбома для обновления")
}

func TestUpdateAlbum_EmptyName(t *testing.T) {
	repo := new(MockAlbumRepo)
	svc := service.NewAlbumService(nil, repo)

	repo.On("GetAlbumByID", mock.Anything, int64(1)).Return(&model.Album{AlbumID: 1}, nil)

	err := svc.UpdateAlbum(context.Background(), &model.Album{
		AlbumID: 1,
		Name:    "",
		GenreID: 1,
	})

	assert.EqualError(t, err, "название альбома не может быть пустым")
}

func TestUpdateAlbum_InvalidGenre(t *testing.T) {
	repo := new(MockAlbumRepo)
	svc := service.NewAlbumService(nil, repo)

	repo.On("GetAlbumByID", mock.Anything, int64(1)).Return(&model.Album{AlbumID: 1}, nil)

	err := svc.UpdateAlbum(context.Background(), &model.Album{
		AlbumID: 1,
		Name:    "Album",
		GenreID: 0,
	})

	assert.EqualError(t, err, "не указан жанр альбома")
}

func TestUpdateAlbum_UpdateError(t *testing.T) {
	repo := new(MockAlbumRepo)
	svc := service.NewAlbumService(nil, repo)

	album := &model.Album{
		AlbumID: 1,
		Name:    "Test",
		GenreID: 1,
	}

	repo.On("GetAlbumByID", mock.Anything, int64(1)).Return(album, nil)
	repo.On("UpdateAlbum", mock.Anything, album).Return(errors.New("ошибка БД"))

	err := svc.UpdateAlbum(context.Background(), album)

	assert.EqualError(t, err, "ошибка БД")
}
