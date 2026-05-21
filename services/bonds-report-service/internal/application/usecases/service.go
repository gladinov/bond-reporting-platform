package usecases

import (
	"bonds-report-service/internal/application/ports"
	"log/slog"
	"time"
)

const (
	layoutTime = "2006-01-02_15-04-05"
)

const (
	bond     = "bond"
	share    = "share"
	futures  = "futures"
	etf      = "etf"
	currency = "currency"
)

const (
	rub       = "rub"
	cny       = "cny"
	usd       = "usd"
	eur       = "eur"
	hkd       = "hkd"
	futuresPt = "pt."
)

const (
	commodityType = "TYPE_COMMODITY"
	currencyType  = "TYPE_CURRENCY"
	securityType  = "TYPE_SECURITY"
	indexType     = "TYPE_INDEX"
)

type ExternalApis struct {
	Moex ports.MoexClient
	Cbr  ports.CbrClient
	Sber ports.SberClient
}

func NewExternalApis(
	moex ports.MoexClient,
	cbr ports.CbrClient,
	sber ports.SberClient,
) *ExternalApis {
	return &ExternalApis{
		Moex: moex,
		Cbr:  cbr,
		Sber: sber,
	}
}

type Helpers struct {
	CbrGetter               CbrCurrencyGetter
	MoexSpecificationGetter MoexSpecificationGetter
	TinkoffProvider         TinkoffProvider
	OperationsUpdater       OperationsUpdater
	PositionProcessor       PositionProcessor
	ReportLineBuilder       ReportLineBuilder
	DividerByAssetType      DividerByAssetType
}

func NewHelpers(
	cbrGetter CbrCurrencyGetter,
	moexSpecificationGetter MoexSpecificationGetter,
	tinkoffProvider TinkoffProvider,
	operationsUpdater OperationsUpdater,
	positionProcessor PositionProcessor,
	reportLineBuilder ReportLineBuilder,
	dividerByAssetType DividerByAssetType,
) *Helpers {
	return &Helpers{
		CbrGetter:               cbrGetter,
		MoexSpecificationGetter: moexSpecificationGetter,
		TinkoffProvider:         tinkoffProvider,
		OperationsUpdater:       operationsUpdater,
		PositionProcessor:       positionProcessor,
		ReportLineBuilder:       reportLineBuilder,
		DividerByAssetType:      dividerByAssetType,
	}
}

type Service struct {
	logger        *slog.Logger
	WorkersNumber int
	External      *ExternalApis
	Helpers       *Helpers
	Storage       ports.Storage
	Producer      Producer
	now           func() time.Time
}

func NewService(
	logger *slog.Logger,
	workersNumber int,
	externalApis *ExternalApis,
	helpers *Helpers,
	storage ports.Storage,
	producer Producer,
) *Service {
	return &Service{
		logger:        logger,
		WorkersNumber: workersNumber,
		External:      externalApis,
		Helpers:       helpers,
		Storage:       storage,
		Producer:      producer,
		now:           time.Now,
	}
}
