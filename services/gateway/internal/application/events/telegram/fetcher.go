package telegram

import (
	"context"
	"log/slog"

	"github.com/gladinov/e"
	"main.go/internal/adapters/outbound/telegram"
	"main.go/internal/application/events"
)

type Fetcher struct {
	logger *slog.Logger
	tg     UpdatesFetcher
	offset int
}

type UpdatesFetcher interface {
	Updates(ctx context.Context, offset int, limit int) (updates []telegram.Update, err error)
}

func NewFetcher(logger *slog.Logger, client *telegram.Client) *Fetcher {
	return &Fetcher{
		logger: logger,
		tg:     client,
	}
}

func (f *Fetcher) Fetch(ctx context.Context, limit int) ([]events.Event, error) {
	updates, err := f.tg.Updates(ctx, f.offset, limit)
	if err != nil {
		return nil, e.Wrap("can't get events", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}
	res := make([]events.Event, 0, len(updates))

	for _, u := range updates {
		res = append(res, event(u))
	}

	f.offset = updates[len(updates)-1].ID + 1
	return res, nil
}

func event(upd telegram.Update) events.Event {
	updType := fetchType(upd)

	res := events.Event{
		Type: fetchType(upd),
		Text: fetchText(upd),
	}
	if updType == events.Message {
		res.Meta = Meta{
			ChatID:   upd.Message.Chat.ID,
			Username: upd.Message.From.Username,
		}
	}
	return res
}

func fetchText(upd telegram.Update) string {
	if upd.Message == nil {
		return ""
	}
	return upd.Message.Text
}

func fetchType(upd telegram.Update) events.Type {
	if upd.Message == nil {
		return events.Unknow
	}
	return events.Message
}
