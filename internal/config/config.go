package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DB    DBConfig
	Admin AdminConfig
	Kafka KafkaConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type AdminConfig struct {
	Secret string
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
}

func splitEnv(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}

func LoadConfig() *Config {

	if err := godotenv.Load(); err != nil {
		log.Println(".env файл не найден, используются переменные среды")
	}

	return &Config{
		DB: DBConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
		},
		Admin: AdminConfig{
			Secret: os.Getenv("ADMIN_SECRET"),
		},
		Kafka: KafkaConfig{
			Brokers: splitEnv("KAFKA_BROKERS"),
			Topic:   os.Getenv("KAFKA_TOPIC"),
		},
	}
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name,
	)
}
