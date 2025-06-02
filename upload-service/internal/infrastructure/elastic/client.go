package elastic

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"upload-service/internal/config"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v8"
)

func NewElasticClient(cfg *config.ElasticConfig) (*elasticsearch.Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.URL},
		Username:  cfg.Username,
		Password:  cfg.Password,
	})
	if err != nil {
		return nil, err
	}

	res, err := es.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	log.Println("Elasticsearch подключен:", res)

	return es, nil
}

func CreateIndex(client *elasticsearch.Client, indexName string, mapping string) error {
	req := esapi.IndicesCreateRequest{
		Index: indexName,
		Body:  strings.NewReader(mapping),
	}
	res, err := req.Do(context.Background(), client)
	if err != nil {
		return fmt.Errorf("ошибка создания индекса %s: %w", indexName, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("ошибка индекса %s: %s", indexName, body)
	}
	return nil
}

func CreateAllIndices(client *elasticsearch.Client) error {
	songMapping := `{
  "mappings": {
   "properties": {
    "song_id": { "type": "long" },
    "name": { "type": "text" },
    "genre": { "type": "keyword" },
    "date": { "type": "date" }
   }
  }
 }`

	albumMapping := `{
  "mappings": {
   "properties": {
    "album_id": { "type": "long" },
    "name": { "type": "text" },
    "artist_id": { "type": "long" },
    "genre": { "type": "keyword" },
    "date": { "type": "date" }
   }
  }
 }`

	artistMapping := `{
  "mappings": {
   "properties": {
    "artist_id": { "type": "long" },
    "name": { "type": "text" },
    "user_id": { "type": "long" }
   }
  }
 }`

	genreMapping := `{
  "mappings": {
   "properties": {
    "genre_id": { "type": "long" },
    "name": { "type": "text" }
   }
  }
 }`

	if err := CreateIndex(client, "songs", songMapping); err != nil {
		fmt.Println("Индекс songs: ", err)
	}
	if err := CreateIndex(client, "albums", albumMapping); err != nil {
		fmt.Println("Индекс albums: ", err)
	}
	if err := CreateIndex(client, "artists", artistMapping); err != nil {
		fmt.Println("Индекс artists: ", err)
	}
	if err := CreateIndex(client, "genres", genreMapping); err != nil {
		fmt.Println("Индекс genres: ", err)
	}

	return nil
}
