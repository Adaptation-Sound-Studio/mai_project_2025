package elastic

import (
	"log"
	"upload-service/internal/config"

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

	// Проверка подключения
	res, err := es.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	log.Println("Elasticsearch подключен:", res)

	return es, nil
}
