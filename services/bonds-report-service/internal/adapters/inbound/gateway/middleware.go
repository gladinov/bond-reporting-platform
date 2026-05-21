package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gladinov/traceidgenerator"

	"github.com/gin-gonic/gin"
	contextkeys "github.com/gladinov/contracts/context"
	httpheaders "github.com/gladinov/contracts/http"
	trace "github.com/gladinov/contracts/trace"
)

func (h *Handler) ContextHeaderTraceIdMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.ContextHeaderTraceIdMiddleWare"
		logg := h.logger.With(slog.String("op", op))
		traceID := c.GetHeader(httpheaders.HeaderTraceID)
		if traceID == "" {
			logg.Warn("traceID is empty")
			var err error
			traceID, err = traceidgenerator.New()
			if err != nil {
				logg.Error("could not generate traceID uuid", slog.Any("error", err))
			}
		}
		logg.Debug("trace_id", slog.String("trace_id", traceID))

		ctx := context.WithValue(c.Request.Context(), contextkeys.TraceIDKey, traceID)

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.AuthMiddleware"
		logg := h.logger.With(slog.String("op", op))
		chatID := c.GetHeader(httpheaders.HeaderChatID)
		if chatID == "" {
			logg.Warn("missing X-Chat-ID header")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing X-Chat-ID header"})
			return
		}
		ctx := context.WithValue(c.Request.Context(), contextkeys.ChatIDKey, chatID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (h *Handler) LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logg := h.logger.With(
			slog.String("component", "middleware/logger"),
		)

		traceID, _ := trace.TraceIDFromContext(c.Request.Context())

		req := c.Request
		resp := c.Writer
		entry := logg.With(
			slog.String("method", req.Method),
			slog.String("path", req.URL.Path),
			slog.String("remote_addr", req.RemoteAddr),
			slog.String("user_agent", req.UserAgent()),
			slog.String("trace_id", traceID),
		)
		t1 := time.Now()
		c.Next()
		defer func() {
			duration := time.Since(t1)
			entry.Info("request completed",
				slog.Int("status", resp.Status()),
				slog.Int64("bytes", int64(resp.Size())),
				slog.Duration("duration", duration),
			)
		}()
	}
}
