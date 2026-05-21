package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	sl "github.com/gladinov/mylogger"
	"github.com/gladinov/notification-service/internal/adapters/inbound/kafka"
	"github.com/gladinov/notification-service/internal/adapters/outbound/telegram"
	tgClient "github.com/gladinov/notification-service/internal/adapters/outbound/telegram"
	"github.com/gladinov/notification-service/internal/application/usecases"
	"github.com/gladinov/notification-service/internal/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	conf := config.MustInitConfig()

	logg := sl.NewLogger(conf.Env)

	logg.Info("start app",
		slog.String("env", conf.Env),
		slog.String("notification-service_app_host", conf.Clients.NotificationService.Host),
		slog.String("notification-service_app_port", conf.Clients.NotificationService.Port))

	logg.Info("initialize consumer kafka client")
	consumerClient, err := kgo.NewClient(
		kgo.SeedBrokers(conf.Kafka.GetKafkaAddress()),
		kgo.ConsumerGroup("notification-service-group"),
		kgo.ConsumeTopics(kafka.ReportFailed, kafka.ReportGenerated),
	)
	if err != nil {
		logg.Error("failed to create consumer client", slog.String("err", err.Error()))
		return
	}

	if err := consumerClient.Ping(ctx); err != nil {
		logg.ErrorContext(ctx, "consumer kafka not available", slog.Any("error", err))
		return
	}

	logg.Info("initialize Telegram client", slog.String("addres", conf.Clients.Telegram.Host))
	telegramClient := tgClient.New(logg, conf.Clients.Telegram.Host, conf.Clients.Telegram.Token)

	service := usecases.NewService(logg, telegramClient)

	handler := kafka.NewHandler(logg, service)

	consumer := kafka.NewConsumer(logg, consumerClient, handler)

	errCh := make(chan error, 1)
	go func() {
		logg.InfoContext(ctx, "run kafka consumer")
		if err := consumer.Run(ctx); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logg.InfoContext(ctx, "Shutdown signal received")
	case er := <-errCh:
		logg.ErrorContext(ctx, "consumer error", slog.Any("error", er))
	}

	gracefulShutdown(logg, consumerClient, telegramClient)
}

func gracefulShutdown(logg *slog.Logger, kafkaClient *kgo.Client, telegramClient *telegram.Client) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logg.InfoContext(shutdownCtx, "close kafka")
	kafkaClient.LeaveGroupContext(shutdownCtx)
	kafkaClient.Close()
	logg.InfoContext(shutdownCtx, "close tg client")
	telegramClient.Close()

	<-shutdownCtx.Done()

	logg.InfoContext(shutdownCtx, "Server exited gracefully")
}
