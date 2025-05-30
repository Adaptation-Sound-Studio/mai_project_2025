package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/fact"
)

type MockFactRepo struct {
	mock.Mock
}

func (m *MockFactRepo) Insert(f *fact.ListenFact) error {
	args := m.Called(f)
	return args.Error(0)
}

func TestStoreFact_Success(t *testing.T) {
	mockRepo := new(MockFactRepo)
	s := NewFactService(mockRepo)

	input := &fact.ListenFact{
		ID:         1,
		UserID:     101,
		SongID:     202,
		ArtistID:   303,
		AlbumID:    404,
		GenreID:    505,
		ListenedAt: time.Now(),
	}

	mockRepo.On("Insert", input).Return(nil)

	err := s.Store(input)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestStoreFact_Failure(t *testing.T) {
	mockRepo := new(MockFactRepo)
	s := NewFactService(mockRepo)

	input := &fact.ListenFact{
		ID:         2,
		UserID:     111,
		SongID:     222,
		ArtistID:   333,
		AlbumID:    444,
		GenreID:    555,
		ListenedAt: time.Now(),
	}

	mockRepo.On("Insert", input).Return(errors.New("db error"))

	err := s.Store(input)

	assert.EqualError(t, err, "db error")
	mockRepo.AssertExpectations(t)
}
