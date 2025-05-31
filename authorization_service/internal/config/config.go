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

type Config struct {
	DB     *DBConfig
	Redis  *RedisConfig
	Server *ServerConfig
}

func LoadConfig() *Config {
	return &Config{
		DB: &DBConfig{
			Host:     os.Getenv("AUTH_DB_HOST"),
			Port:     os.Getenv("AUTH_DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("AUTH_OF_DB_NAME"),
		},
		Redis: &RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     os.Getenv("REDIS_PORT"),
			Password: os.Getenv("REDIS_PASSWORD"),
		},
		Server: &ServerConfig{
			Port: os.Getenv("SERVER_PORT"),
		},
	}
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name,
	)
}
