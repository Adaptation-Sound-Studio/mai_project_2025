package elastic

import (
	"os"
	"testing"
	"upload-service/internal/config"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/require"
)

func TestCreateAllIndices(t *testing.T) {
	cfg := &config.ElasticConfig{
		URL:      os.Getenv("ELASTIC_URL"),
		Username: os.Getenv("ELASTIC_USER"),
		Password: os.Getenv("ELASTIC_PASS"),
	}

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.URL},
		Username:  cfg.Username,
		Password:  cfg.Password,
	})
	require.NoError(t, err)

	err = CreateAllIndices(client)
	require.NoError(t, err)
}

func TestCreateIndex(t *testing.T) {
	esURL := os.Getenv("ELASTIC_URL")
	esUser := os.Getenv("ELASTIC_USER")
	esPass := os.Getenv("ELASTIC_PASS")

	if esURL == "" {
		t.Skip("ELASTIC_URL не установлен — пропускаем интеграционный тест")
	}

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{esURL},
		Username:  esUser,
		Password:  esPass,
	})
	require.NoError(t, err)

	indexName := "test_songs_index"
	mapping := `{
		"mappings": {
			"properties": {
				"song_id":     { "type": "long" },
				"name":        { "type": "text" },
				"genre":       { "type": "keyword" },
				"artist_name": { "type": "text" }
			}
		}
	}`

	err = CreateIndex(client, indexName, mapping)
	require.NoError(t, err)
}

func TestCreateIndex_AlreadyExists(t *testing.T) {
	esURL := os.Getenv("ELASTIC_URL")
	esUser := os.Getenv("ELASTIC_USER")
	esPass := os.Getenv("ELASTIC_PASS")

	if esURL == "" {
		t.Skip("ELASTIC_URL не установлен — пропускаем интеграционный тест")
	}

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{esURL},
		Username:  esUser,
		Password:  esPass,
	})
	require.NoError(t, err)

	indexName := "test_index_duplicate"
	mapping := `{
		"mappings": {
			"properties": {
				"name": { "type": "text" }
			}
		}
	}`

	err = CreateIndex(client, indexName, mapping)
	require.NoError(t, err)

	err = CreateIndex(client, indexName, mapping)
	require.Error(t, err)
	require.Contains(t, err.Error(), "resource_already_exists_exception")

	t.Cleanup(func() {
		client.Indices.Delete([]string{indexName})
	})
}
