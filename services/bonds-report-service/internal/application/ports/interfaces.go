package ports

import (
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/domain/generalbondreport"
	report "bonds-report-service/internal/domain/report"
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
	Close() error
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
