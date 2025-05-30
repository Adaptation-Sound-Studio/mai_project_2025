package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/genre"
)

type MockGenreRepo struct {
	mock.Mock
}

func (m *MockGenreRepo) Create(g *genre.Genre) error {
	args := m.Called(g)
	return args.Error(0)
}

func TestCreateGenre_Success(t *testing.T) {
	mockRepo := new(MockGenreRepo)
	s := NewGenreService(mockRepo)

	g := &genre.Genre{Name: "Pop"}

	mockRepo.On("Create", g).Return(nil)

	err := s.CreateGenre(g)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateGenre_Failure(t *testing.T) {
	mockRepo := new(MockGenreRepo)
	s := NewGenreService(mockRepo)

	g := &genre.Genre{Name: "Unknown genre"}

	mockRepo.On("Create", g).Return(errors.New("db error"))

	err := s.CreateGenre(g)

	assert.EqualError(t, err, "db error")
	mockRepo.AssertExpectations(t)
}
