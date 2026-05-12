package cache

import (
	"context"
	"crypto/tls"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func NewCache() *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:      os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password:  os.Getenv("REDIS_PASSWORD"),
		TLSConfig: &tls.Config{},
	})
	return &Cache{client: client}
}

func (c *Cache) Set(key string, value string, duration time.Duration) error {
	return c.client.Set(context.Background(), key, value, duration).Err()
}

func (c *Cache) Get(key string) (string, error) {
	return c.client.Get(context.Background(), key).Result()
}

func (c *Cache) Delete(key string) error {
	return c.client.Del(context.Background(), key).Err()
}
