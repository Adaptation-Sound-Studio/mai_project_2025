package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/song"
)

type MockSongRepo struct {
	mock.Mock
}

func (m *MockSongRepo) Create(s *song.Song) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *MockSongRepo) GetPopularSongs(limit int) ([]song.PopularSong, error) {
	args := m.Called(limit)
	return args.Get(0).([]song.PopularSong), args.Error(1)
}

func TestCreateSong_Success(t *testing.T) {
	mockRepo := new(MockSongRepo)
	s := NewSongService(mockRepo)

	songInput := &song.Song{Name: "My Song"}

	mockRepo.On("Create", songInput).Return(nil)

	err := s.CreateSong(songInput)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateSong_Failure(t *testing.T) {
	mockRepo := new(MockSongRepo)
	s := NewSongService(mockRepo)

	songInput := &song.Song{Name: "Unknown Song"}

	mockRepo.On("Create", songInput).Return(errors.New("db error"))

	err := s.CreateSong(songInput)

	assert.EqualError(t, err, "db error")
	mockRepo.AssertExpectations(t)
}

func TestGetPopularSongs_Success(t *testing.T) {
	mockRepo := new(MockSongRepo)
	s := NewSongService(mockRepo)

	expected := []song.PopularSong{
		{ID: 1, Name: "Hit", Artist: "Artist", Album: "Album", Genre: "Rock", Listens: 1000},
	}

	mockRepo.On("GetPopularSongs", 1).Return(expected, nil)

	result, err := s.GetPopularSongs(1)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestGetPopularSongs_Failure(t *testing.T) {
	mockRepo := new(MockSongRepo)
	s := NewSongService(mockRepo)

	mockRepo.On("GetPopularSongs", 5).Return([]song.PopularSong(nil), errors.New("db error"))

	result, err := s.GetPopularSongs(5)

	assert.Nil(t, result)
	assert.EqualError(t, err, "db error")
	mockRepo.AssertExpectations(t)
}
