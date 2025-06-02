package service

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
)

func getMockedESClient(responseBody string, statusCode int) *elasticsearch.Client {
	es, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{
				StatusCode: statusCode,
				Body:       ioutil.NopCloser(bytes.NewBufferString(responseBody)),
				Header:     make(http.Header),
			}
			resp.Header.Set("X-Elastic-Product", "Elasticsearch")
			return resp, nil
		}),
	})
	return es
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSearchSongs_Success(t *testing.T) {
	mockJSON := `
	{
		"hits": {
			"hits": [
				{ "_source": { "song_id": 1, "name": "Test Song" } }
			]
		}
	}`
	es := getMockedESClient(mockJSON, http.StatusOK)
	svc := NewSearchService(es)

	results, err := svc.SearchSongs(context.Background(), "Test")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, int64(1), results[0].SongID)
	assert.Equal(t, "Test Song", results[0].Name)
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
	svc := NewSearchService(es)

	results, err := svc.SearchSongsByArtist(context.Background(), "John")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, int64(2), results[0].SongID)
	assert.Equal(t, "By Artist", results[0].Name)
}

func TestSearchSongs_ESReturnsError(t *testing.T) {
	es := getMockedESClient("something went wrong", http.StatusInternalServerError)
	svc := NewSearchService(es)

	results, err := svc.SearchSongs(context.Background(), "x")

	assert.Nil(t, results)
	assert.ErrorContains(t, err, "ES error")
}

func TestSearchSongs_InvalidJSON(t *testing.T) {
	es := getMockedESClient("not-json", http.StatusOK)
	svc := NewSearchService(es)

	results, err := svc.SearchSongs(context.Background(), "x")

	assert.Nil(t, results)
	assert.Error(t, err)
}
