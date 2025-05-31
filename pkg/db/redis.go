package db

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

func CacheCostData(client *redis.Client, key string, data []byte, expiration time.Duration) error {
	ctx := context.TODO()
	return client.Set(ctx, key, data, expiration).Err()
}

func GetCachedCostData(client *redis.Client, key string) ([]byte, error) {
	ctx := context.TODO()
	val, err := client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}