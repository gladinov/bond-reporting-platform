package app

import (
	handlers "bonds-report-service/internal/adapters/inbound/gateway"
	kafkaConsumer "bonds-report-service/internal/adapters/inbound/kafka"
	cbr "bonds-report-service/internal/adapters/outbound/cbr/client"
	cbrtransport "bonds-report-service/internal/adapters/outbound/cbr/transport"
	"bonds-report-service/internal/adapters/outbound/kafka"
	moex "bonds-report-service/internal/adapters/outbound/moex/client"
	moextransport "bonds-report-service/internal/adapters/outbound/moex/transport"
	"bonds-report-service/internal/adapters/outbound/repository/postgreSQL"
	"bonds-report-service/internal/adapters/outbound/sber"
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/client/analyticsclient"
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/client/instrumentsclient"
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/client/portfolioclient"
	tinkofftransport "bonds-report-service/internal/adapters/outbound/tinkoffApi/transport"
	bondreport "bonds-report-service/internal/application/helpers/bondReport"
	cbrHelper "bonds-report-service/internal/application/helpers/cbr"
	dividerbyassettype "bonds-report-service/internal/application/helpers/dividerByAssetType"
	generalbondreport "bonds-report-service/internal/application/helpers/generalBondReport"
	moexHelper "bonds-report-service/internal/application/helpers/moex"
	positionProcessor "bonds-report-service/internal/application/helpers/positionsToPositionsWithAssetUidProcessor"
	"bonds-report-service/internal/application/helpers/report"
	reportlinebuiler "bonds-report-service/internal/application/helpers/reportLineBuiler"
	tinkoffHelper "bonds-report-service/internal/application/helpers/tinkoff"
	"bonds-report-service/internal/application/helpers/uidprovider"
	updateoperations "bonds-report-service/internal/application/helpers/updateOperations"
	"bonds-report-service/internal/application/ports"
	"bonds-report-service/internal/application/usecases"
	"bonds-report-service/internal/closer"
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
	bondReportProcessor      ports.BondReportProcessor
	cbrCurrencyGetter        ports.CbrCurrencyGetter
	generalBondReporter      ports.GeneralBondReportProcessor
	moexSpecificationGetter  ports.MoexSpecificationGetter
	reportProcessor          ports.ReportProcessor
	uidProvider              ports.UidProvider
	operationsUpdater        ports.OperationsUpdater
	positionProcessor        ports.PositionProcessor
	reportLineBuilder        ports.ReportLineBuilder
	dividerByAssetType       ports.DividerByAssetType
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
		d.logger.Info("initialize Tinkoff helper")
		d.tinkoffHelper = tinkoffHelper.NewTinkoffHelper(
			d.logger,
			d.TinkoffInstrumentsClient(),
			d.TinkoffPortfolioClient(),
			d.TinkoffAnalyticsClient(),
		)
	}

	return d.tinkoffHelper
}

func (d *diContainer) TinkoffTransport() tinkofftransport.TransportClient {
	if d.tinkoffTransport == nil {
		host := d.config.Clients.TinkoffClient.GetTinkoffApiAddress()
		d.logger.Info("initialize Tinkoff transport", slog.String("address", host))
		if host == "" {
			panic("tinkoff host is empty")
		}

		d.tinkoffTransport = tinkofftransport.NewTransport(d.logger, host, d.config.Timeouts.RequestTimeout)
	}

	return d.tinkoffTransport
}

func (d *diContainer) TinkoffAnalyticsClient() ports.TinkoffAnalyticsClient {
	if d.tinkoffAnalyticsClient == nil {
		d.logger.Info("initialize Tinkoff analytics client")
		d.tinkoffAnalyticsClient = analyticsclient.NewAnalyticsTinkoffClient(d.logger, d.TinkoffTransport())
	}

	return d.tinkoffAnalyticsClient
}

func (d *diContainer) TinkoffInstrumentsClient() ports.TinkoffInstrumentsClient {
	if d.tinkoffInstrumentsClient == nil {
		d.logger.Info("initialize Tinkoff instruments client")
		d.tinkoffInstrumentsClient = instrumentsclient.NewInstrumentsTinkoffClient(d.logger, d.TinkoffTransport())
	}

	return d.tinkoffInstrumentsClient
}

func (d *diContainer) TinkoffPortfolioClient() ports.TinkoffPortfolioClient {
	if d.tinkoffPortfolioClient == nil {
		d.logger.Info("initialize Tinkoff portfolio client")
		d.tinkoffPortfolioClient = portfolioclient.NewPortfolioTinkoffClient(d.logger, d.TinkoffTransport())
	}

	return d.tinkoffPortfolioClient
}

func (d *diContainer) MoexClient() ports.MoexClient {
	if d.moexClient == nil {
		d.logger.Info("initialize Moex client")
		d.moexClient = moex.NewMoexClient(d.logger, d.MoexTransport())
	}

	return d.moexClient
}

func (d *diContainer) MoexTransport() moextransport.TransportClient {
	if d.moexTransport == nil {
		host := d.config.Clients.MoexClient.GetMoexAppAddress()
		d.logger.Info("initialize Moex transport", slog.String("address", host))
		if host == "" {
			panic("moex host is empty")
		}

		d.moexTransport = moextransport.NewTransport(d.logger, host, d.config.Timeouts.RequestTimeout)
	}

	return d.moexTransport
}

func (d *diContainer) CBRClient() ports.CbrClient {
	if d.cbrClient == nil {
		d.logger.Info("initialize CBR client")
		d.cbrClient = cbr.NewCbrClient(d.logger, d.CBRTransport())
	}

	return d.cbrClient
}

func (d *diContainer) CBRTransport() cbrtransport.TransportClient {
	if d.cbrTransport == nil {
		host := d.config.Clients.CBRClient.GetCBRAppAddress()
		d.logger.Info("initialize CBR transport", slog.String("address", host))
		if host == "" {
			panic("cbr host is empty")
		}

		d.cbrTransport = cbrtransport.NewTransport(d.logger, host, d.config.Timeouts.RequestTimeout)
	}

	return d.cbrTransport
}

func (d *diContainer) SberClient() ports.SberClient {
	if d.sberClient == nil {
		d.logger.Info("initialize Sber client", slog.String("address", d.config.SberConfigPath))

		sberClient, err := sber.NewClient(d.config.RootPath, d.config.SberConfigPath)
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
		d.logger.Info("initialize bond report processor")
		d.bondReportProcessor = bondreport.NewBondReporter(d.logger)
	}

	return d.bondReportProcessor
}

func (d *diContainer) CBRCurrencyGetter() ports.CbrCurrencyGetter {
	if d.cbrCurrencyGetter == nil {
		d.logger.Info("initialize cbr currency getter")
		d.cbrCurrencyGetter = cbrHelper.NewCbrHelper(d.logger, d.CBRClient(), d.Storage())
	}

	return d.cbrCurrencyGetter
}

func (d *diContainer) GeneralBondReporter() ports.GeneralBondReportProcessor {
	if d.generalBondReporter == nil {
		d.logger.Info("initialize general bond report processor")
		d.generalBondReporter = generalbondreport.NewGeneralBondReporter(d.logger)
	}

	return d.generalBondReporter
}

func (d *diContainer) MoexSpecificationGetter() ports.MoexSpecificationGetter {
	if d.moexSpecificationGetter == nil {
		d.logger.Info("initialize moex specification getter")
		d.moexSpecificationGetter = moexHelper.NewMoexHelper(d.logger, d.MoexClient())
	}

	return d.moexSpecificationGetter
}

func (d *diContainer) ReportProcessor() ports.ReportProcessor {
	if d.reportProcessor == nil {
		d.logger.Info("initialize report processor")
		d.reportProcessor = report.NewReportProcessor(d.logger)
	}

	return d.reportProcessor
}

func (d *diContainer) UidProvider() ports.UidProvider {
	if d.uidProvider == nil {
		d.logger.Info("initialize uid provider")
		d.uidProvider = uidprovider.NewUidProvider(d.Storage(), d.TinkoffAnalyticsClient())
	}

	return d.uidProvider
}

func (d *diContainer) OperationsUpdater() ports.OperationsUpdater {
	if d.operationsUpdater == nil {
		d.logger.Info("initialize operations updater")
		d.operationsUpdater = updateoperations.NewUpdater(d.logger, d.Storage(), d.TinkoffHelper())
	}

	return d.operationsUpdater
}

func (d *diContainer) PositionProcessor() ports.PositionProcessor {
	if d.positionProcessor == nil {
		d.logger.Info("initialize position processor")
		d.positionProcessor = positionProcessor.NewProcessor(d.logger, d.UidProvider())
	}

	return d.positionProcessor
}

func (d *diContainer) ReportLineBuilder() ports.ReportLineBuilder {
	if d.reportLineBuilder == nil {
		d.logger.Info("initialize report line builder")
		d.reportLineBuilder = reportlinebuiler.NewReportLineBuilder(d.logger, d.TinkoffHelper(), d.CBRCurrencyGetter())
	}

	return d.reportLineBuilder
}

func (d *diContainer) DividerByAssetType() ports.DividerByAssetType {
	if d.dividerByAssetType == nil {
		d.logger.Info("initialize divider by asset type")
		d.dividerByAssetType = dividerbyassettype.NewDividerByAssetType(
			d.logger,
			d.TinkoffHelper(),
			d.CBRCurrencyGetter(),
			d.config.WorkersNubmer,
		)
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

func (d *diContainer) Handler() appHandler {
	if d.handler == nil {
		d.logger.Info("initialize Handlers")
		d.handler = handlers.NewHandlers(d.logger, d.Service())
	}

	return d.handler
}
