package postgreSQL

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"github.com/gladinov/e"

	config "bonds-report-service/internal/configs"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	logger *slog.Logger
	db     *pgxpool.Pool
}

func NewStorage(logger *slog.Logger, db *pgxpool.Pool) *Storage {
	return &Storage{db: db, logger: logger}
}

func NewPool(ctx context.Context, postgresConfig config.Config) (*pgxpool.Pool, error) {
	postgresHost, err := postgresConfig.PostgresHost.GetStringHost()
	if err != nil {
		return nil, e.WrapIfErr("get postgres host", err)
	}

	poolConfig, err := newPoolConfig(postgresHost)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, e.WrapIfErr("create postgres pool", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, e.WrapIfErr("ping postgres", err)
	}

	return pool, nil
}

func newPoolConfig(postgresHost string) (*pgxpool.Config, error) {
	poolConfig, err := pgxpool.ParseConfig(postgresHost)
	if err != nil {
		return nil, e.WrapIfErr("parse postgres pool config", err)
	}

	poolConfig.MaxConns = int32(runtime.NumCPU() * 2)
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.MinConns = 2

	return poolConfig, nil
}
