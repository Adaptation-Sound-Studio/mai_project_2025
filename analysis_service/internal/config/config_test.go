package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {

	os.Setenv("ANAL_DB_HOST", "localhost")
	os.Setenv("ANAL_DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("ANAL_OF_DB_NAME", "analytics")
	os.Setenv("ADMIN_SECRET", "supersecret")
	os.Setenv("KAFKA_BROKERS", "kafka1:9092,kafka2:9092")
	os.Setenv("KAFKA_TOPIC", "songs")

	cfg := LoadConfig()

	require.Equal(t, "localhost", cfg.DB.Host)
	require.Equal(t, "5432", cfg.DB.Port)
	require.Equal(t, "postgres", cfg.DB.User)
	require.Equal(t, "secret", cfg.DB.Password)
	require.Equal(t, "analytics", cfg.DB.Name)

	require.Equal(t, "supersecret", cfg.Admin.Secret)

	require.Equal(t, []string{"kafka1:9092", "kafka2:9092"}, cfg.Kafka.Brokers)
	require.Equal(t, "songs", cfg.Kafka.Topic)
}

func TestDBConfig_DSN(t *testing.T) {
	cfg := DBConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "secret",
		Name:     "analytics",
	}
	expected := "host=localhost port=5432 user=postgres password=secret dbname=analytics sslmode=disable"
	require.Equal(t, expected, cfg.DSN())
}
