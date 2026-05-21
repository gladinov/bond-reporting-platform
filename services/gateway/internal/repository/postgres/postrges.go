package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gladinov/valuefromcontext"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"main.go/internal/config"
)

const (
// postgresHost = "host=localhost user=user password=parol dbname=storage port=5433 sslmode=disable"
)

type Storage struct {
	db *pgxpool.Pool
}

func NewStorage(ctx context.Context, postgresConfig config.Config) (*Storage, error) {
	postgresHost, err := postgresConfig.PostgresHost.GetStringHost()
	if err != nil {
		return nil, err
	}
	db, err := pgxpool.New(ctx, postgresHost)
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Save(ctx context.Context, user_name string, token string) error {
	const op = "postgres.Save"
	chatId, err := valuefromcontext.GetChatIDFromCtxInt(ctx)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	q := `INSERT INTO users (
                   user_name,
                   chatID,
                   token) VALUES ($1,$2,$3)`

	if token == "" {
		return errors.New("token are enpty")
	}
	_, err = s.db.Exec(ctx, q, user_name, int64(chatId), token)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	return nil
}

func (s *Storage) PickToken(ctx context.Context) (string, error) {
	const op = "postgres.PickToken"
	chatId, err := valuefromcontext.GetChatIDFromCtxInt(ctx)
	if err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}
	q := `SELECT token FROM users WHERE chatID = $1`

	var token string
	err = s.db.QueryRow(ctx, q, int64(chatId)).Scan(&token)
	if err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}
	return token, nil
}

func (s *Storage) IsExistsToken(ctx context.Context) (bool, error) {
	const op = "postgres.IsExistsToken"
	chatId, err := valuefromcontext.GetChatIDFromCtxInt(ctx)
	if err != nil {
		return false, fmt.Errorf("%s:%w", op, err)
	}
	q := `SELECT token FROM users WHERE chatID = $1`
	var token sql.NullString
	err = s.db.QueryRow(ctx, q, int64(chatId)).Scan(&token)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("%s:%w", op, err)
	}

	return token.Valid, nil
}

func (s *Storage) Init(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *Storage) CloseDB() {
	s.db.Close()
}
