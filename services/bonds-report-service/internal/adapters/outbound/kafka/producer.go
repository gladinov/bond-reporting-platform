package kafka

import (
	"bonds-report-service/internal/adapters/inbound/kafka"
	"bonds-report-service/internal/application/dto"
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gladinov/e"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	logg  *slog.Logger
	kafka *kgo.Client
}

func NewProducer(logg *slog.Logger, kafka *kgo.Client) *Producer {
	return &Producer{
		logg:  logg,
		kafka: kafka,
	}
}

func (p *Producer) PublishFailedBondReportWithPng(
	ctx context.Context,
	reportKind string,
	chatID string,
	traceID string,
	errCode string,
	errMesage string,
) error {
	// TODO: Доставать traceID и chatID из контекста
	resp := NewRepsponceReportFailed(reportKind, chatID, traceID, errCode, errMesage)

	body, err := json.Marshal(resp)
	if err != nil {
		return e.WrapIfErr("failed to marshall data", err)
	}

	record := kgo.Record{
		Topic: kafka.ReportFailed,
		Value: body,
	}
	// TODO: дождаться результата Produce и вернуть ошибку, если Kafka не подтвердила отправку
	p.kafka.Produce(ctx, &record, nil)
	return nil
}

func (p *Producer) PublishBondReportWithPng(
	ctx context.Context,
	reportKind string,
	chatID string,
	traceID string,
	bondReportsResponce dto.BondReportsResponce,
) error {
	bondResponce := mapBondReportsResponseToBondResponceBody(bondReportsResponce)

	// TODO: Доставать traceID и chatID из контекста
	resp := NewRequestReportGenerated(reportKind, chatID, traceID, bondResponce)

	body, err := json.Marshal(resp)
	if err != nil {
		return e.WrapIfErr("failed to marshall data", err)
	}

	record := kgo.Record{
		Topic: kafka.ReportGenerated,
		Value: body,
	}
	// TODO: дождаться результата Produce и вернуть ошибку, если Kafka не подтвердила отправку
	p.kafka.Produce(ctx, &record, nil)
	return nil
}
