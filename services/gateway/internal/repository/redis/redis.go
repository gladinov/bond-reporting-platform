package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"main.go/internal/config"
)

func NewClient(ctx context.Context, cfg config.Config) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHTTPServer.GetAddress(),
		Password: cfg.RedisHTTPServer.Password,
		DB:       cfg.RedisHTTPServer.DB,
		// Username:     cfg.RedisHTTPServer.User,
		MaxRetries:   cfg.RedisHTTPServer.MaxRetries,
		DialTimeout:  cfg.RedisHTTPServer.DialTimeout,
		ReadTimeout:  cfg.RedisHTTPServer.Timeout,
		WriteTimeout: cfg.RedisHTTPServer.Timeout,
	})

	if err := db.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return db, nil
}
