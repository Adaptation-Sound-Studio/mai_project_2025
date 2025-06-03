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

func (m *MockArtistRepo) Create(artist *artist.Artist) error {
	args := m.Called(artist)
	return args.Error(0)
}

func (m *MockArtistRepo) GetTopArtistsForUser(userID int, limit int) ([]artist.Artist, error) {
	args := m.Called(userID, limit)
	return args.Get(0).([]artist.Artist), args.Error(1)
}

func (m *MockArtistRepo) GetMostPopularArtists(limit int) ([]artist.Artist, error) {
	args := m.Called(limit)
	if obj := args.Get(0); obj != nil {
		return obj.([]artist.Artist), args.Error(1)
	}
	return nil, args.Error(1)
}
func TestCreateArtist_Success_Service(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	s := NewArtistService(mockRepo)

	a := &artist.Artist{Name: "Vlad"}

	mockRepo.On("Create", a).Return(nil)

	err := s.CreateArtist(a)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateArtist_Failure_Service(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	s := NewArtistService(mockRepo)

	a := &artist.Artist{Name: "Unknown"}

	mockRepo.On("Create", a).Return(errors.New("db error"))

	err := s.CreateArtist(a)

	assert.EqualError(t, err, "db error")
	mockRepo.AssertExpectations(t)
}

func TestArtistService_GetTopArtistsForUser(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	svc := NewArtistService(mockRepo)

	userID := 1
	limit := 3
	expected := []artist.Artist{
		{ID: 1, Name: "Artist_1"},
		{ID: 2, Name: "Artist_2"},
	}

	mockRepo.On("GetTopArtistsForUser", userID, limit).Return(expected, nil)

	result, err := svc.GetTopArtistsForUser(userID, limit)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	mockRepo.AssertExpectations(t)
}

func TestArtistService_GetMostPopularArtists(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	svc := NewArtistService(mockRepo)

	limit := 3
	expected := []artist.Artist{
		{ID: 1, Name: "Artist_1"},
		{ID: 2, Name: "Artist_2"},
	}

	mockRepo.On("GetMostPopularArtists", limit).Return(expected, nil)

	result, err := svc.GetMostPopularArtists(limit)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	mockRepo.AssertExpectations(t)
}

func TestArtistService_GetTopArtistsForUserError(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	svc := NewArtistService(mockRepo)

	userID := 3
	limit := 3

	expected := []artist.Artist{
		{ID: 1, Name: "Artist_1"},
		{ID: 2, Name: "Artist_2"},
	}
	mockRepo.On("GetTopArtistsForUser", userID, limit).Return(expected, nil).Once()

	result, err := svc.GetTopArtistsForUser(userID, limit)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	mockRepo.On("GetTopArtistsForUser", userID, limit).Return([]artist.Artist{}, assert.AnError).Once()

	result, err = svc.GetTopArtistsForUser(userID, limit)
	assert.Error(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestArtistService_GetMostPopularArtistsError(t *testing.T) {
	mockRepo := new(MockArtistRepo)
	svc := NewArtistService(mockRepo)

	limit := 3
	expected := []artist.Artist{
		{ID: 1, Name: "Artist_1"},
		{ID: 2, Name: "Artist_2"},
	}

	mockRepo.On("GetMostPopularArtists", limit).Return(expected, nil).Once()

	result, err := svc.GetMostPopularArtists(limit)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	mockRepo.On("GetMostPopularArtists", limit).Return(nil, assert.AnError).Once()

	result, err = svc.GetMostPopularArtists(limit)
	assert.Error(t, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}
