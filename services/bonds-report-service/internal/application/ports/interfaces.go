package ports

import (
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/domain/generalbondreport"
	report "bonds-report-service/internal/domain/report"
	report_position "bonds-report-service/internal/domain/report_position"
	"context"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=Storage
type Storage interface {
	OperationStorage
	BondReportStorage
	GeneralBondReportStorage
	CurrencyStorage
	UidsStorage
	CloseStorage
}

type SberClient interface {
	GetPortfolio() map[string]float64
}

type OperationStorage interface {
	LastOperationTime(ctx context.Context, chatID int, accountId string) (time.Time, error)
	SaveOperations(ctx context.Context, chatID int, accountId string, operations []domain.OperationWithoutCustomTypes) error
	GetOperations(ctx context.Context, chatId int, assetUid string, accountId string) ([]domain.OperationWithoutCustomTypes, error)
	GetAllOperations(ctx context.Context, chatId int, accountId string) ([]domain.OperationWithoutCustomTypes, error)
}

type BondReportStorage interface {
	DeleteBondReport(ctx context.Context, chatID int, accountId string) (err error)
	SaveBondReport(ctx context.Context, chatID int, accountId string, bondReport []report.BondReport) error
}

type GeneralBondReportStorage interface {
	DeleteGeneralBondReport(ctx context.Context, chatID int, accountId string) (err error)
	SaveGeneralBondReport(ctx context.Context, chatID int, accountId string, bondReport []generalbondreport.GeneralBondReportPosition) error
}

type CurrencyStorage interface {
	SaveCurrency(ctx context.Context, currencies domain.CurrenciesCBR, date time.Time) error
	GetCurrency(ctx context.Context, currency string, date time.Time) (float64, error)
}

type UidsStorage interface {
	SaveUids(ctx context.Context, uids map[string]string) error
	IsUpdatedUids(ctx context.Context) (time.Time, error)
	GetUid(ctx context.Context, instrumentUid string) (string, error)
}

type CloseStorage interface {
	CloseDB()
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=GeneralBondReportProcessor
type GeneralBondReportProcessor interface {
	GetGeneralBondReportPosition(
		ctx context.Context,
		currentPositions []report_position.PositionByFIFO,
		totalAmount float64,
		moexBuyDateData domain.ValuesMoex,
		moexNowData domain.ValuesMoex,
		firstBondsBuyDate time.Time,
	) (_ generalbondreport.GeneralBondReportPosition, err error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=BondReportProcessor
type BondReportProcessor interface {
	CreateBondReport(ctx context.Context, currentPositions []report_position.PositionByFIFO, moexBuyDateData domain.ValuesMoex, moexNowData domain.ValuesMoex) (_ report.Report, err error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=ReportProcessor
type ReportProcessor interface {
	ProcessOperations(ctx context.Context, reportLine *domain.ReportLine) (_ *report_position.ReportPositions, err error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=UidProvider
type UidProvider interface {
	GetUid(ctx context.Context, instrumentUid string) (string, error)
	UpdateAndGetUid(ctx context.Context, instrumentUid string) (string, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=CbrClient
type CbrClient interface {
	GetAllCurrencies(ctx context.Context, date time.Time) (res domain.CurrenciesCBR, err error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=MoexClient
type MoexClient interface {
	GetSpecifications(ctx context.Context, ticker string, date time.Time) (data domain.ValuesMoex, err error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=TinkoffInstrumentsClient
type TinkoffInstrumentsClient interface {
	FindBy(ctx context.Context, query string) (domain.InstrumentShortList, error)
	GetBondByUid(ctx context.Context, uid string) (domain.Bond, error)
	GetCurrencyBy(ctx context.Context, figi string) (domain.Currency, error)
	GetFutureBy(ctx context.Context, figi string) (domain.Future, error)
	GetShareCurrencyBy(ctx context.Context, figi string) (domain.ShareCurrency, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=TinkoffPortfolioClient
type TinkoffPortfolioClient interface {
	GetAccounts(ctx context.Context) (_ map[string]domain.Account, err error)
	GetPortfolio(ctx context.Context, accountID string, accountStatus int64) (domain.Portfolio, error)
	GetOperations(ctx context.Context, accountId string, date time.Time) (_ []domain.Operation, err error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=TinkoffAnalyticsClient
type TinkoffAnalyticsClient interface {
	GetLastPriceInPersentageToNominal(ctx context.Context, instrumentUid string) (domain.LastPrice, error)
	GetAllAssetUids(ctx context.Context) (map[string]string, error)
	GetBondsActions(ctx context.Context, instrumentUid string) (domain.BondIdentIdentifiers, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=CbrCurrencyGetter
type CbrCurrencyGetter interface {
	GetCurrencyFromCB(ctx context.Context, charCode string, date time.Time) (vunit_rate float64, err error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=MoexSpecificationGetter
type MoexSpecificationGetter interface {
	GetSpecificationsFromMoex(ctx context.Context, ticker string, date time.Time) (data domain.ValuesMoex, err error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=OperationsUpdater
type OperationsUpdater interface {
	UpdateOperations(ctx context.Context, chatID int, accountID string, openDate time.Time) (err error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=PositionProcessor
type PositionProcessor interface {
	ProcessPositionsToPositionsWithAssetUid(ctx context.Context, portffolioPositions []domain.PortfolioPosition) (_ []domain.PortfolioPositionsWithAssetUid, err error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=ReportLineBuilder
type ReportLineBuilder interface {
	CreateNewReportLines(ctx context.Context, position domain.PortfolioPositionsWithAssetUid, operationsDb []domain.OperationWithoutCustomTypes) (_ *domain.ReportLine, err error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=DividerByAssetType
type DividerByAssetType interface {
	DivideByType(ctx context.Context, positions []domain.PortfolioPosition) (_ *domain.PortfolioByTypeAndCurrency, err error)
}
