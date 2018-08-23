package db

import (
	"time"

	"github.com/go-redis/redis"
)

// GetClient get redis client
func GetClient() *redis.Client {
	return redis.NewClient(
		&redis.Options{
			Network:      "tcp",
			Addr:         "localhost:6379",
			DB:           0,
			DialTimeout:  10 * time.Second,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			PoolSize:     10,
			PoolTimeout:  11 * time.Second,
		},
	)
}
