package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {

	os.Setenv("UPL_DB_HOST", "localhost")
	os.Setenv("UPL_DB_PORT", "5432")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("UPL_OF_DB_NAME", "testdb")

	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_PASSWORD", "redispass")

	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("SERVICE_API_KEY", "testapikey")

	os.Setenv("AUTH_SERVICE_URL", "http://auth:8000")

	os.Setenv("KAFKA_BROKERS", "localhost:9092")
	os.Setenv("KAFKA_TOPIC", "test-topic")

	os.Setenv("ELASTIC_URL", "http://elastic:9200")
	os.Setenv("ELASTIC_USER", "elasticuser")
	os.Setenv("ELASTIC_PASS", "elasticpass")

	cfg := LoadConfig()

	require.Equal(t, "localhost", cfg.DB.Host)
	require.Equal(t, "5432", cfg.DB.Port)
	require.Equal(t, "testuser", cfg.DB.User)
	require.Equal(t, "testpass", cfg.DB.Password)
	require.Equal(t, "testdb", cfg.DB.Name)

	require.Equal(t, "localhost", cfg.Redis.Host)
	require.Equal(t, "6379", cfg.Redis.Port)
	require.Equal(t, "redispass", cfg.Redis.Password)

	require.Equal(t, "8080", cfg.Server.Port)
	require.Equal(t, "testapikey", cfg.APIKey)

	require.Equal(t, "http://auth:8000", cfg.AuthService.URL)
	require.Equal(t, "testapikey", cfg.AuthService.APIKey)

	require.Equal(t, "localhost:9092", cfg.Kafka.Brokers)
	require.Equal(t, "test-topic", cfg.Kafka.Topic)

	require.Equal(t, "http://elastic:9200", cfg.Elastic.URL)
	require.Equal(t, "elasticuser", cfg.Elastic.Username)
	require.Equal(t, "elasticpass", cfg.Elastic.Password)

	expectedDSN := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	require.Equal(t, expectedDSN, cfg.DB.DSN())
}
