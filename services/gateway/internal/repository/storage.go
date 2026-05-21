package storage

import (
	"context"
	"errors"

	"main.go/internal/config"
	"main.go/internal/repository/postgres"
)

const (
	postreSQL = "postgreSQL"
	SQLite    = "SQLite"
)

type Storage interface {
	Save(ctx context.Context, user_name string, token string) error
	PickToken(ctx context.Context) (string, error)
	IsExistsToken(ctx context.Context) (bool, error)
	CloseDB()
}

func NewStorage(ctx context.Context, config config.Config) (Storage, error) {
	switch config.DbType {
	case postreSQL:
		storage, err := postgres.NewStorage(ctx, config)
		if err != nil {
			return nil, err
		}
		err = storage.Init(ctx)
		if err != nil {
			return nil, err
		}

		return storage, nil
	default:
		return nil, errors.New("possible init only PostgreSQL databases")
	}
}
