package redis

import (
	"context"
	"fmt"
	"upload-service/internal/config"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func NewRedisClient(cfg *config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
	})
}

func GetUserID(rdb *redis.Client, sessionID string) (string, error) {
	return rdb.Get(ctx, "session:"+sessionID).Result()
}

func SetArtistID(rdb *redis.Client, userID string, artistID int64) error {
	return rdb.Set(ctx, "artist:"+userID, artistID, 0).Err()
}

func GetArtistID(rdb *redis.Client, userID string) (string, error) {
	return rdb.Get(ctx, "artist:"+userID).Result()
}

func GetUserRole(rdb *redis.Client, userID string) (string, error) {
	roleKey := "role:" + userID
	role, err := rdb.Get(ctx, roleKey).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("роль не найдена для пользователя %s", userID)
	} else if err != nil {
		return "", fmt.Errorf("ошибка при получении роли пользователя %s: %v", userID, err)
	}
	return role, nil
}
