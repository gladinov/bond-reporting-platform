package kafka

import (
	"context"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Consumer struct {
	logger  *slog.Logger
	client  *kgo.Client
	handler Handler
}

type Handler interface {
	HandleRequest(ctx context.Context, record *kgo.Record) error
}

func NewConsumer(logg *slog.Logger, kafka *kgo.Client, handler Handler) *Consumer {
	return &Consumer{
		logger:  logg,
		client:  kafka,
		handler: handler,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			c.logger.InfoContext(ctx, "consumer stopping")
			return nil
		default:
		}
		fetches := c.client.PollFetches(ctx)
		// TODO: распаралелить, т.к. kafka присылает batch рекордов
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, err := range errs {
				c.logger.ErrorContext(ctx, "kafka fetch error",
					slog.String("topic", err.Topic),
					slog.Int64("partition", int64(err.Partition)),
					slog.Any("err", err.Err),
				)
			}

			if ctx.Err() != nil {
				return nil
			}
			continue
		}
		var runErr error

		fetches.EachRecord(func(record *kgo.Record) {
			if runErr != nil {
				return
			}
			c.logger.InfoContext(ctx, "kafka record received",
				slog.String("topic", record.Topic),
				slog.Int64("partition", int64(record.Partition)),
				slog.Int64("offset", record.Offset),
				slog.String("value", string(record.Value)),
			)
			err := c.handler.HandleRequest(ctx, record)
			if err != nil {
				runErr = err
				c.logger.ErrorContext(ctx, "failed to handle request", slog.Any("error", err))
				return
			}
			c.client.MarkCommitRecords(record)
		})
		if runErr != nil {
			// TODO: обрботать ошибку и не падать при первой ошибке(Сейчас consumer падает)
			return runErr
		}
		err := c.client.CommitMarkedOffsets(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
	}
}
