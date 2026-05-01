package reportlinebuiler

import (
	"bonds-report-service/internal/domain"
	"context"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=TinkoffProvider
type TinkoffProvider interface {
	TinkoffGetBondActions(ctx context.Context, instrumentUid string) (domain.BondIdentIdentifiers, error)
	TinkoffGetLastPriceInPersentageToNominal(ctx context.Context, instrumentUid string) (domain.LastPrice, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=CbrCurrencyGetter
type CbrCurrencyGetter interface {
	GetCurrencyFromCB(ctx context.Context, charCode string, date time.Time) (float64, error)
}
