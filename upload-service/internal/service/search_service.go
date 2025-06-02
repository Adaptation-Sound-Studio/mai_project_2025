package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"upload-service/internal/domain/model"

	"github.com/elastic/go-elasticsearch/v8"
)

type SearchService struct {
	esClient *elasticsearch.Client
	esIndex  string
}

func NewSearchService(esClient *elasticsearch.Client) *SearchService {
	return &SearchService{
		esClient: esClient,
		esIndex:  "songs",
	}
}

func (s *SearchService) SearchSongs(ctx context.Context, query string) ([]*model.Song, error) {
	body := fmt.Sprintf(`
    {
      "query": {
        "multi_match": {
          "query": "%s",
          "fields": ["name"]
        }
      }
    }`, query)

	res, err := s.esClient.Search(
		s.esClient.Search.WithContext(ctx),
		s.esClient.Search.WithIndex(s.esIndex),
		s.esClient.Search.WithBody(strings.NewReader(body)),
		s.esClient.Search.WithSize(20),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		data, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("ES error: %s", string(data))
	}

	var parsed struct {
		Hits struct {
			Hits []struct {
				Source model.Song `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	var results []*model.Song
	for _, hit := range parsed.Hits.Hits {
		song := hit.Source
		results = append(results, &song)
	}
	return results, nil
}

func (s *SearchService) SearchSongsByArtist(ctx context.Context, query string) ([]*model.Song, error) {
	body := fmt.Sprintf(`
    {
      "query": {
        "nested": {
          "path": "artists",
          "query": {
            "match": {
              "artists.name": "%s"
            }
          }
        }
      }
    }`, query)

	res, err := s.esClient.Search(
		s.esClient.Search.WithContext(ctx),
		s.esClient.Search.WithIndex(s.esIndex),
		s.esClient.Search.WithBody(strings.NewReader(body)),
		s.esClient.Search.WithSize(20),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		data, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("ES error: %s", string(data))
	}

	var parsed struct {
		Hits struct {
			Hits []struct {
				Source model.Song `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	var results []*model.Song
	for _, hit := range parsed.Hits.Hits {
		song := hit.Source
		results = append(results, &song)
	}
	return results, nil
}
