package usecases

import (
	"context"
	"log/slog"
)

type Service struct {
	logger   *slog.Logger
	telegram Telegram
}

type Telegram interface {
	SendImageFromBuffer(ctx context.Context, chatID int, imageData []byte, caption string) error
	SendMediaGroupFromBuffer(ctx context.Context, chatID int, images []*ImageData) error
	SendMessage(ctx context.Context, chatID int, text string) error
}

func NewService(logger *slog.Logger, telegram Telegram) *Service {
	return &Service{
		logger:   logger,
		telegram: telegram,
	}
}
