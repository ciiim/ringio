package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	RedisDB *redis.Client
)

const (
	REDIS_TIMEOUT = 5 * time.Second
)

func InitRedis(addr, password string, db int) error {
	RedisDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	_, err := RedisDB.Ping(context.Background()).Result()
	return err
}

func RedisGet(key string) *redis.StringCmd {
	if RedisDB == nil {
		return redis.NewStringResult("", fmt.Errorf("redis is not enabled"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), REDIS_TIMEOUT)
	defer cancel()
	return RedisDB.Get(ctx, key)
}

func RedisSet(key string, value interface{}, expireTime time.Duration) error {
	if RedisDB == nil {
		return fmt.Errorf("redis is not enabled")
	}
	ctx, cancel := context.WithTimeout(context.Background(), REDIS_TIMEOUT)
	defer cancel()
	err := RedisDB.Set(ctx, key, value, expireTime).Err()
	if err != nil {
		return err
	}
	return nil
}

func RedisDel(key string) error {
	if RedisDB == nil {
		return fmt.Errorf("redis is not enabled")
	}
	ctx, cancel := context.WithTimeout(context.Background(), REDIS_TIMEOUT)
	defer cancel()
	err := RedisDB.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}
