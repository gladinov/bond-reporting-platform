package usecases

import (
	"bonds-report-service/internal/application/dto"
	tinkoffHelper "bonds-report-service/internal/application/helpers/tinkoff"
	"bonds-report-service/internal/application/ports"
	"context"
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
	BondReportProcessor        ports.BondReportProcessor
	CbrGetter                  ports.CbrCurrencyGetter
	GeneralBondReportProcessor ports.GeneralBondReportProcessor
	MoexSpecificationGetter    ports.MoexSpecificationGetter
	ReportProcessor            ports.ReportProcessor
	TinkoffHelper              *tinkoffHelper.TinkoffHelper
	OperationsUpdater          ports.OperationsUpdater
	PositionProcessor          ports.PositionProcessor
	ReportLineBuilder          ports.ReportLineBuilder
	DividerByAssetType         ports.DividerByAssetType
}

func NewHelpers(
	bondReportProcessor ports.BondReportProcessor,
	cbrGetter ports.CbrCurrencyGetter,
	generalBondReportProcessor ports.GeneralBondReportProcessor,
	moexSpecificationGetter ports.MoexSpecificationGetter,
	reportProcessor ports.ReportProcessor,
	tinkoffHelper *tinkoffHelper.TinkoffHelper,
	operationsUpdater ports.OperationsUpdater,
	positionProcessor ports.PositionProcessor,
	reportLineBuilder ports.ReportLineBuilder,
	dividerByAssetType ports.DividerByAssetType,
) *Helpers {
	return &Helpers{
		BondReportProcessor:        bondReportProcessor,
		CbrGetter:                  cbrGetter,
		GeneralBondReportProcessor: generalBondReportProcessor,
		MoexSpecificationGetter:    moexSpecificationGetter,
		ReportProcessor:            reportProcessor,
		TinkoffHelper:              tinkoffHelper,
		OperationsUpdater:          operationsUpdater,
		PositionProcessor:          positionProcessor,
		ReportLineBuilder:          reportLineBuilder,
		DividerByAssetType:         dividerByAssetType,
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

type Producer interface {
	PublishFailedBondReportWithPng(ctx context.Context, reportKind, chatID, traceID, errCode, errMesage string) error
	PublishBondReportWithPng(ctx context.Context, reportKind, chatID, traceID string, bondReportsResponce dto.BondReportsResponce) error
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
