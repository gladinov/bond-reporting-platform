package service

import (
	"context"
	"log/slog"
	"sync"
)

type Service struct {
	logg    *slog.Logger
	mu      sync.Mutex
	updates map[int]Update
	ID      int
}

func NewService(logger *slog.Logger) *Service {
	return &Service{
		logg:    logger,
		updates: make(map[int]Update),
	}
}

func (s *Service) Updates(ctx context.Context, offset, limit int) []Update {
	s.mu.Lock()
	defer s.mu.Unlock()
	res := []Update{}
	if len(s.updates) == 0 {
		return res
	}
	for i := offset; i < offset+limit; i++ {
		update, exist := s.updates[i]
		if !exist {
			break
		}

		res = append(res, update)
		delete(s.updates, i)
	}

	return res
}

func (s *Service) SaveMessage(ctx context.Context, message IncomingMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	update := Update{ID: s.ID, Message: &message}
	s.updates[s.ID] = update
	s.ID++
	s.logg.DebugContext(ctx, "save message updates", slog.Any("updates", s.updates))
}
