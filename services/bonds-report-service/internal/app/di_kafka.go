package app

import (
	kafkaConsumer "bonds-report-service/internal/adapters/inbound/kafka"
	"bonds-report-service/internal/adapters/outbound/kafka"
	"bonds-report-service/internal/application/usecases"
	"bonds-report-service/internal/closer"
	"context"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kgo"
)

func (d *diContainer) KafkaProducer() usecases.Producer {
	if d.kafkaProducer == nil {
		d.logger.Info("initialize kafka producer")
		d.kafkaProducer = kafka.NewProducer(d.logger, d.KafkaProducerClient())
	}

	return d.kafkaProducer
}

func (d *diContainer) KafkaProducerClient() *kgo.Client {
	if d.kafkaProducerClient == nil {
		d.logger.Info("initialize producer kafka client")

		producerClient, err := kgo.NewClient(kgo.SeedBrokers(d.config.Kafka.GetKafkaAddress()))
		if err != nil {
			d.logger.Error("failed to create producer client", slog.String("err", err.Error()))
			panic(err)
		}

		if err := producerClient.Ping(context.Background()); err != nil {
			d.logger.Error("producer kafka not available", slog.Any("error", err))
			panic(err)
		}

		d.kafkaProducerClient = producerClient
		closer.Add("kafka producer", func(context.Context) error {
			producerClient.Close()
			return nil
		})
	}

	return d.kafkaProducerClient
}

func (d *diContainer) KafkaHandler() kafkaConsumer.Handler {
	if d.kafkaHandler == nil {
		d.kafkaHandler = kafkaConsumer.NewHandler(d.logger, d.Service())
	}

	return d.kafkaHandler
}

func (d *diContainer) Consumer() KafkaConsumer {
	if d.kafkaConsumer == nil {
		d.logger.Info("initialize kafka consumer")
		d.kafkaConsumer = kafkaConsumer.NewConsumer(d.logger, d.KafkaConsumerClient(), d.KafkaHandler())
	}

	return d.kafkaConsumer
}

func (d *diContainer) KafkaConsumerClient() *kgo.Client {
	if d.kafkaConsumerClient == nil {
		d.logger.Info("initialize consumer kafka client")

		consumerClient, err := kgo.NewClient(
			kgo.SeedBrokers(d.config.Kafka.GetKafkaAddress()),
			kgo.ConsumerGroup("bond-report-service-group"),
			kgo.ConsumeTopics(kafka.ReportRequested),
		)
		if err != nil {
			d.logger.Error("failed to create consumer client", slog.String("err", err.Error()))
			panic(err)
		}

		if err := consumerClient.Ping(context.Background()); err != nil {
			d.logger.Error("consumer kafka not available", slog.Any("error", err))
			panic(err)
		}

		closer.Add("kafka consumer", func(context.Context) error {
			consumerClient.Close()
			return nil
		})

		d.kafkaConsumerClient = consumerClient
	}

	return d.kafkaConsumerClient
}
