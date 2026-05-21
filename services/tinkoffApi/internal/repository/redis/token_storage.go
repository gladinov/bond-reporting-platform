package redisClient

import (
	"context"
	"errors"
	"tinkoffApi/internal/configs"
	"tinkoffApi/internal/handlers"

	"github.com/redis/go-redis/v9"
)

type TokenStorage struct {
	client *redis.Client
}

func NewTokenStorage(ctx context.Context, cfg *configs.Config) (*TokenStorage, error) {
	client, err := NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &TokenStorage{
		client: client,
	}, nil
}

func (s *TokenStorage) GetToken(ctx context.Context, chatID string) (string, error) {
	token, err := s.client.Get(ctx, chatID).Result()
	if errors.Is(err, redis.Nil) {
		return "", handlers.ErrTokenNotFound
	}

	return token, err
}

func (s *TokenStorage) Close(_ context.Context) error {
	if s.client == nil {
		return nil
	}

	return s.client.Close()
}
