package session

import (
	"auth_service/internal/config"
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

func NewClient(cfg *config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: cfg.Password,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Не удалось подключиться к Redis: %v", err)
	}
	log.Println("Подключение к Redis установлено")
	return client
}
