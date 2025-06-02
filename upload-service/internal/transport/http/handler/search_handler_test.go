package handler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"upload-service/internal/domain/model"
	"upload-service/internal/service"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSearchService struct {
	mock.Mock
}

func (m *MockSearchService) SearchSongs(ctx context.Context, query string) ([]*model.Song, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]*model.Song), args.Error(1)
}

func (m *MockSearchService) SearchSongsByArtist(ctx context.Context, query string) ([]*model.Song, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]*model.Song), args.Error(1)
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func getMockedESClient(responseBody string, statusCode int) *elasticsearch.Client {
	es, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: statusCode,
				Body:       ioutil.NopCloser(strings.NewReader(responseBody)),
				Header: http.Header{
					"X-Elastic-Product": []string{"Elasticsearch"},
				},
			}, nil
		}),
	})
	return es
}

func TestSearchSongs_Success(t *testing.T) {
	mockJSON := `
	{
		"hits": {
			"hits": [
				{ "_source": { "song_id": 1, "name": "Hello" } }
			]
		}
	}`

	es := getMockedESClient(mockJSON, http.StatusOK)
	svc := service.NewSearchService(es)
	handler := NewSearchHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/search?q=Hello", nil)
	rec := httptest.NewRecorder()

	handler.SearchSongs(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var songs []model.Song
	err := json.NewDecoder(rec.Body).Decode(&songs)
	require.NoError(t, err)
	require.Len(t, songs, 1)
	assert.Equal(t, int64(1), songs[0].SongID)
	assert.Equal(t, "Hello", songs[0].Name)
}

func TestSearchSongsByArtist_Success(t *testing.T) {
	mockJSON := `
	{
		"hits": {
			"hits": [
				{ "_source": { "song_id": 2, "name": "By Artist" } }
			]
		}
	}`

	es := getMockedESClient(mockJSON, http.StatusOK)
	svc := service.NewSearchService(es)
	handler := NewSearchHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/search/by-artist?q=John", nil)
	rec := httptest.NewRecorder()

	handler.SearchSongsByArtist(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var songs []model.Song
	err := json.NewDecoder(rec.Body).Decode(&songs)
	require.NoError(t, err)
	require.Len(t, songs, 1)
	assert.Equal(t, int64(2), songs[0].SongID)
	assert.Equal(t, "By Artist", songs[0].Name)
}
