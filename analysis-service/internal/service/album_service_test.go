package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/album"
)

type MockAlbumRepo struct {
	mock.Mock
}

func (m *MockAlbumRepo) Create(a *album.Album) error {
	args := m.Called(a)
	return args.Error(0)
}

func TestCreateAlbum_Success(t *testing.T) {

	// Успешное создание
	mockRepo := new(MockAlbumRepo)
	s := NewAlbumService(mockRepo)

	a := &album.Album{Name: "Test Album"}

	mockRepo.On("Create", a).Return(nil)

	err := s.CreateAlbum(a)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateAlbum_Failure(t *testing.T) {

	// Правильное пробрасывание ошибки
	mockRepo := new(MockAlbumRepo)
	s := NewAlbumService(mockRepo)

	a := &album.Album{Name: "Unknown Album"}

	mockRepo.On("Create", a).Return(errors.New("db error"))

	err := s.CreateAlbum(a)

	assert.EqualError(t, err, "db error")
	mockRepo.AssertExpectations(t)
}
