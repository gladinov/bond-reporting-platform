package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gladinov/valuefromcontext"
	storagemodels "main.go/internal/repository/models"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect to database: %w by path:%s", err, path)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Save(ctx context.Context, user_name string, token string) error {
	const op = "sqlite.Save"
	chatID, err := valuefromcontext.GetChatIDFromCtxInt(ctx)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	q := `INSERT INTO users(user_name, chatID, token) VALUES (?,?,?)`

	if token == "" {
		return errors.New("token are enpty")
	}
	if _, err := s.db.ExecContext(ctx, q, user_name, chatID, token); err != nil {
		return fmt.Errorf("can't save page: %w", err)
	}
	return nil
}

func (s *Storage) PickToken(ctx context.Context) (string, error) {
	const op = "sqlite.PickToken"
	chatID, err := valuefromcontext.GetChatIDFromCtxInt(ctx)
	if err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}
	q := `SELECT token FROM users WHERE chatID = ? LIMIT 1`

	var token string

	err = s.db.QueryRowContext(ctx, q, chatID).Scan(&token)
	if err == sql.ErrNoRows {
		return "", storagemodels.ErrNoSaveTokens
	}
	if err != nil {
		return "", fmt.Errorf("can't pick token: %w", err)
	}

	return token, nil
}

func (s *Storage) IsExistsToken(ctx context.Context) (bool, error) {
	const op = "sqlite.IsExistsToken"
	chatID, err := valuefromcontext.GetChatIDFromCtxInt(ctx)
	if err != nil {
		return false, fmt.Errorf("%s:%w", op, err)
	}

	q := `SELECT token FROM users WHERE chatID = ?`

	var token sql.NullString

	err = s.db.QueryRowContext(ctx, q, chatID).Scan(&token)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("can't check user in storage: %w", err)
	}

	return token.Valid, nil
}

func (s *Storage) Init(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Storage) CloseDB() {
	_ = s.db.Close()
}
