package dividerbyassettype

import (
	"bonds-report-service/internal/domain"
	"context"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=TinkoffProvider
type TinkoffProvider interface {
	TinkoffGetFutureBy(ctx context.Context, figi string) (domain.Future, error)
	TinkoffGetBaseShareFutureValute(ctx context.Context, positionUid string) (domain.ShareCurrency, error)
	TinkoffGetCurrencyBy(ctx context.Context, figi string) (domain.Currency, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=CbrCurrencyGetter
type CbrCurrencyGetter interface {
	GetCurrencyFromCB(ctx context.Context, charCode string, date time.Time) (float64, error)
}
