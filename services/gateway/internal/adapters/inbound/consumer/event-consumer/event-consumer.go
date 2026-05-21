package event_consumer

import (
	"context"
	"log/slog"
	"time"

	"main.go/internal/application/events"
)

type Consumer struct {
	logger    *slog.Logger
	fetcher   events.Fetcher
	handler   EventsHandler
	batchSize int
}

type EventsHandler interface {
	HandleEvents(ctx context.Context, events []events.Event) error
}

func New(logger *slog.Logger, fetcher events.Fetcher, handler EventsHandler, batchSize int) Consumer {
	return Consumer{
		logger:    logger,
		fetcher:   fetcher,
		handler:   handler,
		batchSize: batchSize,
	}
}

func (c Consumer) Start(ctx context.Context) error {
	const op = "event_consumer.Start"
	logg := c.logger.With(
		slog.String("op", op),
		slog.Any("batchSize", c.batchSize),
	)
	logg.Info("consumer started")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			gotEvents, err := c.fetcher.Fetch(ctx, c.batchSize)
			if err != nil {
				logg.Warn("fetch events failed",
					slog.Any("error", err),
				)
				continue
			}

			if len(gotEvents) == 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(1 * time.Second):
				}
				continue
			}

			if err := c.handler.HandleEvents(ctx, gotEvents); err != nil {
				logg.Error("handle events failed",
					slog.Any("error", err),
					slog.Any("events count", len(gotEvents)))

				continue
			}
		}
	}
}
