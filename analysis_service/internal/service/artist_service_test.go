package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/artist"
)

type MockArtistRepo struct {
	mock.Mock
}

func (m *MockArtistRepo) Create(a *artist.Artist) error {
	args := m.Called(a)
	return args.Error(0)
}

func TestCreateArtist_Success(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	s := NewArtistService(mockRepo)

	a := &artist.Artist{Name: "Vlad"}

	mockRepo.On("Create", a).Return(nil)

	err := s.CreateArtist(a)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateArtist_Failure(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	s := NewArtistService(mockRepo)

	a := &artist.Artist{Name: "Unknown"}

	mockRepo.On("Create", a).Return(errors.New("db error"))

	err := s.CreateArtist(a)

	assert.EqualError(t, err, "db error")
	mockRepo.AssertExpectations(t)
}
