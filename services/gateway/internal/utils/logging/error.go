package logging

import (
	"context"
	"log/slog"
	"net/http"
	"os"
)

func LoggerError(ctx context.Context, logger *slog.Logger, msg string, op string, err error) {
	logg := logger.With(slog.String("op", op),
		slog.Any("error", err))
	logg.ErrorContext(ctx, msg)
}

func LoggHTTPError(ctx context.Context, logger *slog.Logger, req *http.Request, msg, op string, err error) {
	errLogg := logger.With(
		slog.String("op", op),
		slog.Bool("is_timeout", os.IsTimeout(err)),
		slog.Any("error", err),
	)

	if req != nil {
		errLogg = errLogg.With(
			slog.String("endpoint", req.URL.String()),
			slog.String("method", req.Method),
		)
	}

	errLogg.ErrorContext(ctx, msg)
}
