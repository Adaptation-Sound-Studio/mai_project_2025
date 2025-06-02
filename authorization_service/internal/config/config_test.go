package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {

	os.Setenv("AUTH_DB_HOST", "localhost")
	os.Setenv("AUTH_DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("AUTH_OF_DB_NAME", "auth_db")

	os.Setenv("REDIS_HOST", "redis.local")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_PASSWORD", "redispass")

	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("SERVICE_API_KEY", "api-secret-key")
	os.Setenv("UPL_SERVICE_URL", "https://upl.service")

	cfg := LoadConfig()

	require.Equal(t, "localhost", cfg.DB.Host)
	require.Equal(t, "5432", cfg.DB.Port)
	require.Equal(t, "postgres", cfg.DB.User)
	require.Equal(t, "secret", cfg.DB.Password)
	require.Equal(t, "auth_db", cfg.DB.Name)

	require.Equal(t, "redis.local", cfg.Redis.Host)
	require.Equal(t, "6379", cfg.Redis.Port)
	require.Equal(t, "redispass", cfg.Redis.Password)

	require.Equal(t, "8080", cfg.Server.Port)

	require.Equal(t, "api-secret-key", cfg.ServiceApiKey)

	require.Equal(t, "https://upl.service", cfg.UplService.URL)
	require.Equal(t, "api-secret-key", cfg.UplService.APIKey)
}

func TestDBConfig_DSN(t *testing.T) {
	db := &DBConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "secret",
		Name:     "auth_db",
	}

	expected := "host=localhost port=5432 user=postgres password=secret dbname=auth_db sslmode=disable"
	require.Equal(t, expected, db.DSN())
}
