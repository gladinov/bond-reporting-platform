package app

import (
	handlers "bonds-report-service/internal/adapters/inbound/gateway"
	kafkaConsumer "bonds-report-service/internal/adapters/inbound/kafka"
	cbrtransport "bonds-report-service/internal/adapters/outbound/cbr/transport"
	moextransport "bonds-report-service/internal/adapters/outbound/moex/transport"
	tinkofftransport "bonds-report-service/internal/adapters/outbound/tinkoffApi/transport"
	"bonds-report-service/internal/application/ports"
	positionProcessor "bonds-report-service/internal/application/services/positionprocessor"
	tinkoffHelper "bonds-report-service/internal/application/services/tinkoff"
	"bonds-report-service/internal/application/usecases"
	config "bonds-report-service/internal/configs"
	"context"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kgo"
)

type diContainer struct {
	logger                   *slog.Logger
	config                   config.Config
	storage                  ports.Storage
	tinkoffTransport         tinkofftransport.TransportClient
	tinkoffAnalyticsClient   ports.TinkoffAnalyticsClient
	tinkoffInstrumentsClient ports.TinkoffInstrumentsClient
	tinkoffPortfolioClient   ports.TinkoffPortfolioClient
	tinkoffHelper            *tinkoffHelper.TinkoffHelper
	moexTransport            moextransport.TransportClient
	moexClient               ports.MoexClient
	cbrTransport             cbrtransport.TransportClient
	cbrClient                ports.CbrClient
	sberClient               ports.SberClient
	externalApis             *usecases.ExternalApis
	helpers                  *usecases.Helpers
	service                  *usecases.Service
	handler                  appHandler
	kafkaProducerClient      *kgo.Client
	kafkaProducer            usecases.Producer
	kafkaHandler             kafkaConsumer.Handler
	kafkaConsumerClient      *kgo.Client
	kafkaConsumer            KafkaConsumer
	cbrCurrencyGetter        usecases.CbrCurrencyGetter
	moexSpecificationGetter  usecases.MoexSpecificationGetter
	uidProvider              positionProcessor.UidProvider
	operationsUpdater        usecases.OperationsUpdater
	positionProcessor        usecases.PositionProcessor
	reportLineBuilder        usecases.ReportLineBuilder
	dividerByAssetType       usecases.DividerByAssetType
}

type KafkaConsumer interface {
	Run(ctx context.Context) error
}

func newDIContainer(logger *slog.Logger, config config.Config) *diContainer {
	return &diContainer{
		logger: logger,
		config: config,
	}
}

func (d *diContainer) Service() *usecases.Service {
	if d.service == nil {
		d.logger.Info("initialize Service client")
		d.service = usecases.NewService(
			d.logger,
			d.config.WorkersNubmer,
			d.ExternalApis(),
			d.Helpers(),
			d.Storage(),
			d.KafkaProducer(),
		)
	}

	return d.service
}

func (d *diContainer) Handler() appHandler {
	if d.handler == nil {
		d.logger.Info("initialize Handlers")
		d.handler = handlers.NewHandlers(d.logger, d.Service())
	}

	return d.handler
}
