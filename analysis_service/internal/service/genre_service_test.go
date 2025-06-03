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
func (m *MockGenreRepo) GetTopGenresForUser(userID int, limit int) ([]genre.Genre, error) {
	args := m.Called(userID, limit)
	if obj := args.Get(0); obj != nil {
		return obj.([]genre.Genre), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGenreRepo) GetMostPopularGenres(limit int) ([]genre.Genre, error) {
	args := m.Called(limit)
	if obj := args.Get(0); obj != nil {
		return obj.([]genre.Genre), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGenreRepo) GetTopGenresAtNight(limit int) ([]genre.Genre, error) {
	args := m.Called(limit)
	if obj := args.Get(0); obj != nil {
		return obj.([]genre.Genre), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGenreRepo) GetTopGenresInMorning(limit int) ([]genre.Genre, error) {
	args := m.Called(limit)
	if obj := args.Get(0); obj != nil {
		return obj.([]genre.Genre), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGenreRepo) GetTopGenresInDay(limit int) ([]genre.Genre, error) {
	args := m.Called(limit)
	if obj := args.Get(0); obj != nil {
		return obj.([]genre.Genre), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGenreRepo) GetTopGenresInEvening(limit int) ([]genre.Genre, error) {
	args := m.Called(limit)
	if obj := args.Get(0); obj != nil {
		return obj.([]genre.Genre), args.Error(1)
	}
	return nil, args.Error(1)
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

func TestGenreService(t *testing.T) {
	mockRepo := new(MockGenreRepo)
	service := NewGenreService(mockRepo)

	limit := 3
	userID := 102
	expected := []genre.Genre{
		{ID: 1, Name: "Rock"},
		{ID: 2, Name: "Jazz"},
		{ID: 3, Name: "Pop"},
	}

	mockRepo.On("GetTopGenresForUser", userID, limit).Return(expected, nil)
	mockRepo.On("GetMostPopularGenres", limit).Return(expected, nil)
	mockRepo.On("GetTopGenresAtNight", limit).Return(expected, nil)
	mockRepo.On("GetTopGenresInMorning", limit).Return(expected, nil)
	mockRepo.On("GetTopGenresInDay", limit).Return(expected, nil)
	mockRepo.On("GetTopGenresInEvening", limit).Return(expected, nil)

	res1, err1 := service.GetTopGenresForUser(userID, limit)
	assert.NoError(t, err1)
	assert.Equal(t, expected, res1)

	res2, err2 := service.GetMostPopularGenres(limit)
	assert.NoError(t, err2)
	assert.Equal(t, expected, res2)

	res3, err3 := service.GetTopGenresAtNight(limit)
	assert.NoError(t, err3)
	assert.Equal(t, expected, res3)

	res4, err4 := service.GetTopGenresInMorning(limit)
	assert.NoError(t, err4)
	assert.Equal(t, expected, res4)

	res5, err5 := service.GetTopGenresInDay(limit)
	assert.NoError(t, err5)
	assert.Equal(t, expected, res5)

	res6, err6 := service.GetTopGenresInEvening(limit)
	assert.NoError(t, err6)
	assert.Equal(t, expected, res6)

	mockRepo.AssertExpectations(t)
}
