package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"upload-service/internal/config"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func NewRedisClient(cfg *config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: cfg.Password,
	})
}

func GetUserID(rdb *redis.Client, ctx context.Context, sessionID string) (int64, error) {
	val, err := rdb.Get(ctx, sessionID).Result()
	if err != nil {
		return 0, err
	}

	var sessionData map[string]interface{}
	if err := json.Unmarshal([]byte(val), &sessionData); err != nil {
		return 0, err
	}

	rawID, ok := sessionData["user_id"]
	if !ok {
		return 0, fmt.Errorf("user_id не найден в сессии")
	}

	switch id := rawID.(type) {
	case float64:
		return int64(id), nil
	case int64:
		return id, nil
	default:
		return 0, fmt.Errorf("user_id имеет неподдерживаемый тип: %T", id)
	}
}

func SetArtistID(rdb *redis.Client, userID string, artistID int64) error {
	return rdb.Set(ctx, "artist:"+userID, artistID, 0).Err()
}

func GetArtistID(rdb *redis.Client, userID string) (string, error) {
	return rdb.Get(ctx, "artist:"+userID).Result()
}

func GetUserRole(rdb *redis.Client, sessionID string) (string, error) {
	val, err := rdb.Get(ctx, sessionID).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("сессия %s не найдена", sessionID)
	} else if err != nil {
		return "", fmt.Errorf("ошибка при получении сессии: %v", err)
	}

	var sessionData struct {
		Role string `json:"role"`
	}
	if err := json.Unmarshal([]byte(val), &sessionData); err != nil {
		return "", fmt.Errorf("не удалось распарсить сессию: %v", err)
	}
	return sessionData.Role, nil
}
