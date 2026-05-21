package event_consumer

import (
	"context"
	"log/slog"
	"time"

	"github.com/gladinov/contracts/trace"
	"github.com/gladinov/e"
	"github.com/gladinov/traceidgenerator"
	"main.go/internal/application/events"
	"main.go/internal/application/events/telegram"
)

type Handler struct {
	logger    *slog.Logger
	processor events.Processor
}

func NewHandler(logger *slog.Logger, processor events.Processor) *Handler {
	return &Handler{
		logger:    logger,
		processor: processor,
	}
}

func (h *Handler) HandleEvents(ctx context.Context, events []events.Event) error {
	const op = "event_consumer.handleEvents"

	start := time.Now()
	logg := h.logger.With(slog.String("op", op))
	logg.Debug("start")
	defer func() {
		logg.Info("finished",
			slog.Duration("duration", time.Since(start)),
		)
	}()

	for _, event := range events {
		traceID, err := traceidgenerator.New()
		if err != nil {
			return e.WrapIfErr(op, err)
		}
		ctx := trace.WithTraceID(ctx, traceID)

		eventLog := logg.With(
			slog.Any("event_type", event.Type),
		)

		// TODO: Подумать насколько это корректно? Разные логи для предусмотренных и не предусмотренных комманд
		if telegram.ContainsInConstantCommands(event.Text) {
			eventLog.DebugContext(ctx, "got new event", slog.String("event", event.Text))
		} else {
			eventLog.DebugContext(ctx, "got new other event")
		}

		if err := h.processor.Process(ctx, event); err != nil {
			if telegram.ContainsInConstantCommands(event.Text) {
				eventLog.Error("process event failed",
					slog.String("event", event.Text),
					slog.Any("error", err))
			} else {
				eventLog.Error("process other event failed",
					slog.Any("error", err))
			}
			continue
		}
	}

	return nil
}
