package config

import (
	"fmt"
	"os"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type ServerConfig struct {
	Port string
}

type AuthServiceConfig struct {
	URL    string
	APIKey string
}

type KafkaConfig struct {
	Brokers string
	Topic   string
}

type ElasticConfig struct {
	URL      string
	Username string
	Password string
}

type Config struct {
	DB          *DBConfig
	Redis       *RedisConfig
	Server      *ServerConfig
	AuthService *AuthServiceConfig
	APIKey      string
	Kafka       *KafkaConfig
	Elastic     *ElasticConfig
}

func LoadConfig() *Config {
	return &Config{
		DB: &DBConfig{
			Host:     os.Getenv("UPL_DB_HOST"),
			Port:     os.Getenv("UPL_DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("UPL_OF_DB_NAME"),
		},
		Redis: &RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     os.Getenv("REDIS_PORT"),
			Password: os.Getenv("REDIS_PASSWORD"),
		},
		Server: &ServerConfig{
			Port: os.Getenv("SERVER_PORT"),
		},
		APIKey: os.Getenv("SERVICE_API_KEY"),
		AuthService: &AuthServiceConfig{
			URL:    os.Getenv("AUTH_SERVICE_URL"),
			APIKey: os.Getenv("SERVICE_API_KEY"),
		},
		Kafka: &KafkaConfig{
			Brokers: os.Getenv("KAFKA_BROKERS"),
			Topic:   os.Getenv("KAFKA_TOPIC"),
		},
		Elastic: &ElasticConfig{
			URL:      os.Getenv("ELASTIC_URL"),
			Username: os.Getenv("ELASTIC_USER"),
			Password: os.Getenv("ELASTIC_PASS"),
		},
	}
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name,
	)
}
