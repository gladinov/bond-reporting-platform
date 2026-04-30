package app

import (
	handlers "bonds-report-service/internal/adapters/inbound/gateway"
	kafkaConsumer "bonds-report-service/internal/adapters/inbound/kafka"
	"bonds-report-service/internal/adapters/outbound/kafka"
	"bonds-report-service/internal/adapters/outbound/repository/postgreSQL"
	tinkoffHelper "bonds-report-service/internal/application/helpers/tinkoff"
	"bonds-report-service/internal/application/ports"
	"bonds-report-service/internal/application/usecases"
	"bonds-report-service/internal/closer"
	config "bonds-report-service/internal/configs"
	"context"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kgo"
)

type diContainer struct {
	logger                  *slog.Logger
	config                  config.Config
	storage                 ports.Storage
	tinkoffHelper           *tinkoffHelper.TinkoffHelper
	moexClient              ports.MoexClient
	cbrClient               ports.CbrClient
	sberClient              ports.SberClient
	externalApis            *usecases.ExternalApis
	helpers                 *usecases.Helpers
	service                 *usecases.Service
	handler                 appHandler
	kafkaProducer           usecases.Producer
	kafkaHandler            kafkaConsumer.Handler
	kafkaConsumer           KafkaConsumer
	bondReportProcessor     ports.BondReportProcessor
	cbrCurrencyGetter       ports.CbrCurrencyGetter
	generalBondReporter     ports.GeneralBondReportProcessor
	moexSpecificationGetter ports.MoexSpecificationGetter
	reportProcessor         ports.ReportProcessor
	uidProvider             ports.UidProvider
	operationsUpdater       ports.OperationsUpdater
	positionProcessor       ports.PositionProcessor
	reportLineBuilder       ports.ReportLineBuilder
	dividerByAssetType      ports.DividerByAssetType
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

func (d *diContainer) Storage() ports.Storage {
	if d.storage == nil {
		ctx := context.Background()

		serviceStorage, err := postgreSQL.NewStorage(ctx, d.logger, d.config)
		if err != nil {
			d.logger.Error("failed to create PostgreSQL storage", "err", err)
			panic(err)
		}

		if err := serviceStorage.InitDB(ctx); err != nil {
			d.logger.Error("failed to init PostgreSQL database", "err", err)
			panic(err)
		}

		d.logger.Info("PostgreSQL storage initialized successfully")

		d.storage = serviceStorage
		closer.Add("storage db", func(context.Context) error {
			d.storage.CloseDB()
			return nil
		})
	}

	return d.storage
}

func (d *diContainer) TinkoffHelper() *tinkoffHelper.TinkoffHelper {
	if d.tinkoffHelper == nil {
		d.tinkoffHelper = InitTinkoffApiHelper(d.logger, d.config.Clients.TinkoffClient.GetTinkoffApiAddress())
	}

	return d.tinkoffHelper
}

func (d *diContainer) MoexClient() ports.MoexClient {
	if d.moexClient == nil {
		d.moexClient = InitTiMoexClient(d.logger, d.config.Clients.MoexClient.GetMoexAppAddress())
	}

	return d.moexClient
}

func (d *diContainer) CBRClient() ports.CbrClient {
	if d.cbrClient == nil {
		d.cbrClient = InitCBRClient(d.logger, d.config.Clients.CBRClient.GetCBRAppAddress())
	}

	return d.cbrClient
}

func (d *diContainer) SberClient() ports.SberClient {
	if d.sberClient == nil {
		sberClient, err := InitSberClient(d.logger, &d.config)
		if err != nil {
			d.logger.Error("could not create sber client", slog.String("error", err.Error()))
			panic(err)
		}

		d.sberClient = sberClient
	}

	return d.sberClient
}

func (d *diContainer) BondReportProcessor() ports.BondReportProcessor {
	if d.bondReportProcessor == nil {
		d.bondReportProcessor = InitBondReportProcessor(d.logger)
	}

	return d.bondReportProcessor
}

func (d *diContainer) CBRCurrencyGetter() ports.CbrCurrencyGetter {
	if d.cbrCurrencyGetter == nil {
		d.cbrCurrencyGetter = InitCBRCurrencyGetter(d.logger, d.CBRClient(), d.Storage())
	}

	return d.cbrCurrencyGetter
}

func (d *diContainer) GeneralBondReporter() ports.GeneralBondReportProcessor {
	if d.generalBondReporter == nil {
		d.generalBondReporter = InitGeneralReportProcessor(d.logger)
	}

	return d.generalBondReporter
}

func (d *diContainer) MoexSpecificationGetter() ports.MoexSpecificationGetter {
	if d.moexSpecificationGetter == nil {
		d.moexSpecificationGetter = InitMoexSpecificationGetter(d.logger, d.MoexClient())
	}

	return d.moexSpecificationGetter
}

func (d *diContainer) ReportProcessor() ports.ReportProcessor {
	if d.reportProcessor == nil {
		d.reportProcessor = InitReportProcessor(d.logger)
	}

	return d.reportProcessor
}

func (d *diContainer) UidProvider() ports.UidProvider {
	if d.uidProvider == nil {
		d.uidProvider = InitUidProvider(d.logger, d.Storage(), d.TinkoffHelper().Analytics)
	}

	return d.uidProvider
}

func (d *diContainer) OperationsUpdater() ports.OperationsUpdater {
	if d.operationsUpdater == nil {
		d.operationsUpdater = InitOperationsUpdater(d.logger, d.TinkoffHelper(), d.Storage())
	}

	return d.operationsUpdater
}

func (d *diContainer) PositionProcessor() ports.PositionProcessor {
	if d.positionProcessor == nil {
		d.positionProcessor = InitPositionProcessor(d.logger, d.UidProvider())
	}

	return d.positionProcessor
}

func (d *diContainer) ReportLineBuilder() ports.ReportLineBuilder {
	if d.reportLineBuilder == nil {
		d.reportLineBuilder = InitReportLineBuilder(d.logger, d.TinkoffHelper(), d.CBRCurrencyGetter())
	}

	return d.reportLineBuilder
}

func (d *diContainer) DividerByAssetType() ports.DividerByAssetType {
	if d.dividerByAssetType == nil {
		d.dividerByAssetType = InitDividerByAssetType(d.logger, d.TinkoffHelper(), d.CBRCurrencyGetter(), d.config.WorkersNubmer)
	}

	return d.dividerByAssetType
}

func (d *diContainer) ExternalApis() *usecases.ExternalApis {
	if d.externalApis == nil {
		d.externalApis = usecases.NewExternalApis(d.MoexClient(), d.CBRClient(), d.SberClient())
	}

	return d.externalApis
}

func (d *diContainer) Helpers() *usecases.Helpers {
	if d.helpers == nil {
		d.helpers = usecases.NewHelpers(
			d.BondReportProcessor(),
			d.CBRCurrencyGetter(),
			d.GeneralBondReporter(),
			d.MoexSpecificationGetter(),
			d.ReportProcessor(),
			d.TinkoffHelper(),
			d.OperationsUpdater(),
			d.PositionProcessor(),
			d.ReportLineBuilder(),
			d.DividerByAssetType(),
		)
	}

	return d.helpers
}

func (d *diContainer) KafkaProducer() usecases.Producer {
	if d.kafkaProducer == nil {
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

		closer.Add("kafka producer", func(context.Context) error {
			producerClient.Close()
			return nil
		})

		d.kafkaProducer = kafka.NewProducer(d.logger, producerClient)
	}

	return d.kafkaProducer
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

func (d *diContainer) KafkaHandler() kafkaConsumer.Handler {
	if d.kafkaHandler == nil {
		d.kafkaHandler = kafkaConsumer.NewHandler(d.logger, d.Service())
	}

	return d.kafkaHandler
}

func (d *diContainer) Consumer() KafkaConsumer {
	if d.kafkaConsumer == nil {
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

		d.kafkaConsumer = kafkaConsumer.NewConsumer(d.logger, consumerClient, d.KafkaHandler())
	}

	return d.kafkaConsumer
}

func (d *diContainer) Handler() appHandler {
	if d.handler == nil {
		d.logger.Info("initialize Handlers")
		d.handler = handlers.NewHandlers(d.logger, d.Service())
	}

	return d.handler
}
