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
	Addr string
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
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
		},
		Redis: &RedisConfig{
			Addr: os.Getenv("REDIS_ADDR"),
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
