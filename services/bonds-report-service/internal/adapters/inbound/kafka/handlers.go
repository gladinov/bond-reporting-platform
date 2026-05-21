package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	contextkeys "github.com/gladinov/contracts/context"
	"github.com/gladinov/e"
	"github.com/twmb/franz-go/pkg/kgo"
)

type HandlerClient struct {
	logger  *slog.Logger
	service Service
}

func NewHandler(logger *slog.Logger, service Service) *HandlerClient {
	return &HandlerClient{
		logger:  logger,
		service: service,
	}
}

type Service interface {
	ProduceBondReports(ctx context.Context, reportKind, traceID, chatIDStr string) (err error)
}

func (h *HandlerClient) HandleRequest(ctx context.Context, record *kgo.Record) error {
	switch record.Topic {
	case ReportRequested:
		err := h.handleReportRequested(ctx, record.Value)
		if err != nil {
			return err
		}
	default:
		h.logger.WarnContext(ctx, "unexpected topic", slog.String("topic", record.Topic))
	}
	return nil
}

func (h *HandlerClient) handleReportRequested(ctx context.Context, value []byte) error {
	var body kafkaRequest

	err := json.Unmarshal(value, &body)
	if err != nil {
		return e.WrapIfErr("failed to unmarshall record value", err)
	}

	err = body.Validate()
	if err != nil {
		if errors.Is(err, ErrEmptyTraceID) {
			h.logger.WarnContext(ctx, "traceId is empty", slog.Any("chatId", body.ChatID), slog.Any("reportKind", body.ReportKind))
		}
		return err
	}

	ctx = setCtx(ctx, body.ChatID, body.TraceID)

	err = h.service.ProduceBondReports(ctx, body.ReportKind, body.TraceID, body.ChatID)
	if err != nil {
		return e.WrapIfErr("failed to produce bond report", err)
	}
	return nil
}

func setCtx(ctx context.Context, chatID, traceID string) context.Context {
	newCtx := context.WithValue(ctx, contextkeys.ChatIDKey, chatID)
	newCtx2 := context.WithValue(newCtx, contextkeys.TraceIDKey, traceID)
	return newCtx2
}
