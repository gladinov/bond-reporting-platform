package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	contextkeys "github.com/gladinov/contracts/context"
	"github.com/gladinov/e"
	"github.com/gladinov/notification-service/internal/application/usecases"
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
	SendReportFailed(ctx context.Context, body usecases.ReportFailed) error
	SendReportGenerated(ctx context.Context, body usecases.ReportGenerated) error
}

func (h *HandlerClient) HandleRequest(ctx context.Context, record *kgo.Record) error {
	switch record.Topic {
	case ReportGenerated:
		err := h.handleReportGenerated(ctx, record.Value)
		if err != nil {
			return err
		}
	case ReportFailed:
		err := h.handleReportFailed(ctx, record.Value)
		if err != nil {
			return err
		}
	default:
		h.logger.WarnContext(ctx, "unexpected topic", slog.String("topic", record.Topic))
	}
	return nil
}

func (h *HandlerClient) handleReportGenerated(ctx context.Context, value []byte) error {
	var body RequestReportGenerated

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

	usecaseBody := body.ResponceReportGeneratedToUsecaseDto()

	err = h.service.SendReportGenerated(ctx, usecaseBody)
	if err != nil {
		return e.WrapIfErr("failed to send report generated", err)
	}
	return nil
}

func (h *HandlerClient) handleReportFailed(ctx context.Context, value []byte) error {
	var body ResponceReportFailed
	err := json.Unmarshal(value, &body)
	if err != nil {
		return e.WrapIfErr("failed to unmarshall record value", err)
	}

	err = body.Validate()
	if err != nil {
		if errors.Is(err, ErrEmptyTraceID) {
			h.logger.WarnContext(
				ctx,
				"traceId is empty",
				slog.Any("chatId", body.ChatID),
				slog.Any("reportKind", body.ReportKind),
				slog.Any("error code from kafka", body.ErrorCode),
				slog.Any("error message from kafka", body.ErrorMessage))
		}
		if errors.Is(err, ErrEmptyErr) {
			h.logger.WarnContext(ctx, err.Error(), slog.Any("chatId", body.ChatID), slog.Any("reportKind", body.ReportKind), slog.Any("traceID", body.TraceID))
		}
		return err
	}

	ctx = setCtx(ctx, body.ChatID, body.TraceID)

	usecaseBody := body.ResponceReportFailedToUsecaseDto()

	err = h.service.SendReportFailed(ctx, usecaseBody)
	if err != nil {
		return e.WrapIfErr("failed to send report failed", err)
	}
	return nil
}

func setCtx(ctx context.Context, chatID, traceID string) context.Context {
	newCtx := context.WithValue(ctx, contextkeys.ChatIDKey, chatID)
	newCtx = context.WithValue(ctx, contextkeys.TraceIDKey, traceID)
	return newCtx
}
