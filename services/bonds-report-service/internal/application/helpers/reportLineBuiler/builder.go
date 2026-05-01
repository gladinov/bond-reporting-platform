package reportlinebuiler

import (
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/utils/logging"
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gladinov/e"
)

type ReportLineBuilder struct {
	logger            *slog.Logger
	tinkoffProvider   TinkoffProvider
	CbrCurrencyGetter CbrCurrencyGetter
	now               func() time.Time
}

func NewReportLineBuilder(logger *slog.Logger, tinkoffProvider TinkoffProvider, cbrCurrencyGetter CbrCurrencyGetter) *ReportLineBuilder {
	return &ReportLineBuilder{
		logger:            logger,
		tinkoffProvider:   tinkoffProvider,
		CbrCurrencyGetter: cbrCurrencyGetter,
	}
}

// TODO: нужно добавить решение через worker pool + limiter для ограничения запросов в АПИ
func (b *ReportLineBuilder) CreateNewReportLines(
	ctx context.Context,
	position domain.PortfolioPositionsWithAssetUid,
	operationsDb []domain.OperationWithoutCustomTypes,
) (_ *domain.ReportLine, err error) {
	const op = "service.CreateNewReportLines"
	defer logging.LogOperation_Debug(ctx, b.logger, op, &err)()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	var lastPrice domain.LastPrice
	var bondActions domain.BondIdentIdentifiers
	var vunitRate domain.Rate

	wg.Add(1)
	go func() {
		defer wg.Done()
		price, e := b.tinkoffProvider.TinkoffGetLastPriceInPersentageToNominal(ctx, position.InstrumentUid)
		if e != nil {
			select {
			case errCh <- e:
				cancel()
			default:
			}
			return
		}

		lastPrice = price
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		actions, e := b.tinkoffProvider.TinkoffGetBondActions(ctx, position.InstrumentUid)
		if e != nil {
			select {
			case errCh <- e:
				cancel()
			default:
			}
			return
		}

		bondActions = actions
	}()

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	vunitRate, err = b.buildVunitRate(ctx, bondActions)
	if err != nil {
		return nil, err
	}

	reportLine := domain.NewReportLine(
		operationsDb,
		bondActions,
		lastPrice,
		vunitRate,
	)

	return &reportLine, nil
}

func (b *ReportLineBuilder) buildVunitRate(ctx context.Context, bondActions domain.BondIdentIdentifiers) (domain.Rate, error) {
	// TODO: Не очевидная логика! Потом есть проверка,
	if !bondActions.Replaced {
		return domain.Rate{}, nil
	}

	isoCurrName := bondActions.NominalCurrency

	rate, err := b.CbrCurrencyGetter.GetCurrencyFromCB(ctx, isoCurrName, b.now())
	if err != nil {
		return domain.Rate{}, e.WrapIfErr("failed to get currency rate", err)
	}

	return domain.Rate{
		IsoCurrencyName: isoCurrName,
		Vunit_Rate:      domain.NewNullFloat64(rate, true, false),
	}, nil
}
