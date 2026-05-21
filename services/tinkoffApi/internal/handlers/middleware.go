package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gladinov/valuefromcontext"

	"github.com/gladinov/traceidgenerator"

	contextkeys "github.com/gladinov/contracts/context"
	httpheaders "github.com/gladinov/contracts/http"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) ContextHeaderTraceIdMiddleWare(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		const op = "handlers.ContextHeaderTraceIdMiddleWare"
		logg := h.logger.With(slog.String("op", op))
		logg.Debug("start")

		traceID := c.Request().Header.Get(httpheaders.HeaderTraceID)
		if traceID == "" {
			logg.Warn("traceID is empty")
			traceID, err = traceidgenerator.New()
			if err != nil {
				logg.Error("could not generate traceID uuid", slog.Any("error", err))
			}
		}

		ctx := context.WithValue(c.Request().Context(), contextkeys.TraceIDKey, traceID)

		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}

func (h *Handlers) LoggerMiddleWare(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		logg := h.logger.With(
			slog.String("component", "middleware/logger"),
		)

		traceID, _ := valuefromcontext.GetTraceID(c.Request().Context())

		req := c.Request()
		resp := c.Response()
		entry := logg.With(
			slog.String("method", req.Method),
			slog.String("path", req.URL.Path),
			slog.String("remote_addr", req.RemoteAddr),
			slog.String("user_agent", req.UserAgent()),
			slog.String("trace_id", traceID),
		)
		start := time.Now()

		defer func() {
			var status int

			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				} else {
					status = http.StatusInternalServerError
				}
			} else {
				status = resp.Status
			}

			attrs := []any{
				slog.Int("status", status),
				slog.Int64("bytes", resp.Size),
				slog.Duration("duration", time.Since(start)),
			}

			if err != nil {
				entry.Error("request failed",
					append(attrs, slog.Any("error", err))...,
				)
			} else {
				entry.Info("request completed", attrs...)
			}
		}()
		err = next(c)

		return err
	}
}

func (h *Handlers) CheckTokenFromRedisByChatIDMiddleWare(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		const op = "handlers.CheckTokenFromRedisByChatIDMiddleWare"

		if c.Path() == "/tinkoff/checktoken" {
			return next(c)
		}

		chatID := c.Request().Header.Get(httpheaders.HeaderChatID)
		if chatID == "" {
			err = errHeaderRequired
			return echo.NewHTTPError(http.StatusUnauthorized, errHeaderRequired)
		}

		ctx := c.Request().Context()
		tokenInBase64, err := h.tokenStorage.GetToken(ctx, chatID)
		switch err {
		case nil:
		case ErrTokenNotFound:
			return echo.NewHTTPError(http.StatusUnauthorized, errNoTokenInRedis)
		default:
			return echo.NewHTTPError(http.StatusServiceUnavailable, errRedisDoNotAnswer)
		}

		token, err := h.tokenDecrypter.DecryptTokenFromBase64(tokenInBase64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errInvalidAuthFormat)
		}

		ctx = context.WithValue(c.Request().Context(), contextkeys.EncryptedTokenKey, token)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}

func (h *Handlers) CheckTokenFromHeadersMiddleWare(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		const op = "handlers.CheckTokenFromHeadersMiddleWare"

		tokenInBase64 := c.Request().Header.Get(httpheaders.HeaderEncryptedToken)
		if tokenInBase64 == "" {
			err = errHeaderRequired
			return echo.NewHTTPError(http.StatusUnauthorized, err)
		}
		token, err := h.tokenDecrypter.DecryptTokenFromBase64(tokenInBase64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errInvalidAuthFormat)
		}
		ctx := context.WithValue(c.Request().Context(), contextkeys.EncryptedTokenKey, token)

		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
