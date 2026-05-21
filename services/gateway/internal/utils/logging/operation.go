package logging

import (
	"context"
	"log/slog"
	"time"
)

func LogOperation_Debug(
	ctx context.Context,
	logger *slog.Logger,
	op string,
	err *error,
) func() {
	start := time.Now()
	logg := logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")

	return func() {
		logg.DebugContext(ctx, "finished",
			slog.Duration("duration", time.Since(start)),
			slog.Any("error", err),
		)
	}
}
