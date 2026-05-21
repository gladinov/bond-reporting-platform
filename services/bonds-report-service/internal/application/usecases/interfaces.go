package usecases

import (
	"bonds-report-service/internal/application/dto"
	"bonds-report-service/internal/domain"
	"context"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=CbrCurrencyGetter
type CbrCurrencyGetter interface {
	GetCurrencyFromCB(ctx context.Context, charCode string, date time.Time) (float64, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=MoexSpecificationGetter
type MoexSpecificationGetter interface {
	GetSpecificationsFromMoex(ctx context.Context, ticker string, date time.Time) (domain.ValuesMoex, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=TinkoffProvider
type TinkoffProvider interface {
	TinkoffGetAccounts(ctx context.Context) (map[string]domain.Account, error)
	TinkoffGetPortfolio(ctx context.Context, account domain.Account) (domain.Portfolio, error)
	TinkoffFindBy(ctx context.Context, query string) (domain.InstrumentShortList, error)
	TinkoffGetBondByUid(ctx context.Context, uid string) (domain.Bond, error)
	TinkoffGetLastPriceInPersentageToNominal(ctx context.Context, instrumentUid string) (domain.LastPrice, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=OperationsUpdater
type OperationsUpdater interface {
	UpdateOperations(ctx context.Context, chatID int, accountID string, openDate time.Time) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=PositionProcessor
type PositionProcessor interface {
	ProcessPositionsToPositionsWithAssetUid(ctx context.Context, portfolioPositions []domain.PortfolioPosition) ([]domain.PortfolioPositionsWithAssetUid, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=ReportLineBuilder
type ReportLineBuilder interface {
	CreateNewReportLines(ctx context.Context, position domain.PortfolioPositionsWithAssetUid, operationsDb []domain.OperationWithoutCustomTypes) (*domain.ReportLine, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=DividerByAssetType
type DividerByAssetType interface {
	DivideByType(ctx context.Context, positions []domain.PortfolioPosition) (*domain.PortfolioByTypeAndCurrency, error)
}

type Producer interface {
	PublishFailedBondReportWithPng(ctx context.Context, reportKind, chatID, traceID, errCode, errMesage string) error
	PublishBondReportWithPng(ctx context.Context, reportKind, chatID, traceID string, bondReportsResponce dto.BondReportsResponce) error
}
