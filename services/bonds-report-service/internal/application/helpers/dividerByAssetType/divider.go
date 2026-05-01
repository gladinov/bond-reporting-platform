package dividerbyassettype

import (
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/utils/logging"
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/gladinov/e"
)

var ErrEmptyPosition = errors.New("positions are empty")

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

type DividerByAssetType struct {
	logger          *slog.Logger
	CbrGetter       CbrCurrencyGetter
	tinkoffProvider TinkoffProvider
	WorkersNumber   int
	now             func() time.Time
}

func NewDividerByAssetType(logger *slog.Logger, tinkoffProvider TinkoffProvider, cbrGetter CbrCurrencyGetter, workerNumber int) *DividerByAssetType {
	return &DividerByAssetType{
		logger:          logger,
		tinkoffProvider: tinkoffProvider,
		CbrGetter:       cbrGetter,
		WorkersNumber:   workerNumber,
		now:             time.Now,
	}
}

func (d *DividerByAssetType) DivideByType(ctx context.Context, positions []domain.PortfolioPosition) (_ *domain.PortfolioByTypeAndCurrency, err error) {
	const op = "service.DivideByType"

	defer logging.LogOperation_Debug(ctx, d.logger, op, &err)()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if len(positions) == 0 {
			return nil, ErrEmptyPosition
		}

		ctxWorkers, cancel := context.WithCancel(ctx)
		defer cancel()
		errCh := make(chan error, 1)
		pipeline := NewPipeline(ctxWorkers, cancel, errCh)
		workers := d.WorkersNumber

		positionCh := d.producePosition(ctxWorkers, positions)

		portfiliosCh := make(chan *domain.PortfolioByTypeAndCurrency, workers*2)

		var wgStage1 sync.WaitGroup
		for i := 0; i < workers; i++ {
			wgStage1.Add(1)
			go func() {
				defer wgStage1.Done()
				d.worker(pipeline, positionCh, portfiliosCh)
			}()
		}
		go func() {
			wgStage1.Wait()
			close(portfiliosCh)
		}()

		portfolio := domain.NewPortfolioByTypeAndCurrency()
	loop:
		for {
			select {
			case <-ctxWorkers.Done():
				return nil, ctxWorkers.Err()
			case er := <-errCh:
				cancel()
				return nil, er
			case portfolioWithOnePos, ok := <-portfiliosCh:
				if !ok {
					break loop
				}
				portfolio.SumWithPortfolio(portfolioWithOnePos)
			}
		}

		return portfolio, nil

	}
}

func (d *DividerByAssetType) worker(
	p *pipeline,
	in <-chan domain.PortfolioPosition,
	out chan<- *domain.PortfolioByTypeAndCurrency,
) {
	for pos := range in {
		vunit, err := d.getCurrencyVunitRate(p.ctx, pos)
		if err != nil {
			p.sendErr(err)
			return
		}

		portfolio, err := d.processAsset(p.ctx, pos, vunit)
		if err != nil {
			p.sendErr(err)
			return
		}

		select {
		case <-p.ctx.Done():
			return
		case out <- portfolio:
		}
	}
}

func (d *DividerByAssetType) producePosition(ctx context.Context, positions []domain.PortfolioPosition) <-chan domain.PortfolioPosition {
	out := make(chan domain.PortfolioPosition, d.WorkersNumber*2)
	go func() {
		defer close(out)
		for _, pos := range positions {
			select {
			case <-ctx.Done():
				return
			case out <- pos:
			}
		}
	}()
	return out
}

func (d *DividerByAssetType) getCurrencyVunitRate(ctx context.Context, position domain.PortfolioPosition) (vunit_rate float64, err error) {
	currencyOfPos := position.CurrentPrice.Currency
	if currencyOfPos != futuresPt && currencyOfPos != rub {
		vunit_rate, err = d.CbrGetter.GetCurrencyFromCB(ctx, currencyOfPos, d.now())
		if err != nil {
			return 0, e.WrapIfErr("can't get currency from CB", err)
		}
	} else {
		vunit_rate = 1.0
	}
	return vunit_rate, nil
}

func (d *DividerByAssetType) processAsset(ctx context.Context, pos domain.PortfolioPosition, vunit_rate float64) (_ *domain.PortfolioByTypeAndCurrency, err error) {
	var portfolioWithOnePos *domain.PortfolioByTypeAndCurrency
	switch pos.InstrumentType {
	case bond:
		portfolioWithOnePos = d.processBond(pos, vunit_rate)
	case share:
		portfolioWithOnePos = d.processShare(pos, vunit_rate)
	case futures:
		portfolioWithOnePos, err = d.processFutures(ctx, pos, vunit_rate)
		if err != nil {
			return nil, e.WrapIfErr("failed to process futures", err)
		}
	case etf:
		portfolioWithOnePos = d.processEtf(pos, vunit_rate)
	case currency:
		portfolioWithOnePos, err = d.processCurrency(ctx, pos, vunit_rate)
		if err != nil {
			return nil, e.WrapIfErr("failed to process currency", err)
		}
	default:
		d.logger.WarnContext(ctx, "unexpected instrument type", slog.Any("instrument type", pos.InstrumentType))
	}
	return portfolioWithOnePos, nil
}

func (d *DividerByAssetType) processBond(pos domain.PortfolioPosition, vunit_rate float64) *domain.PortfolioByTypeAndCurrency {
	portfolioWithOnePos := domain.NewPortfolioByTypeAndCurrency()

	positionPrice := pos.Quantity.ToFloat() * pos.CurrentPrice.ToFloat() * vunit_rate
	positionPrice += pos.CurrentNkd.ToFloat() * pos.Quantity.ToFloat() * vunit_rate
	portfolioWithOnePos.BondsAssets.SumOfAssets += positionPrice
	currencyOfPos := pos.CurrentPrice.Currency

	domain.AddToMap(portfolioWithOnePos.BondsAssets.AssetsByCurrency, currencyOfPos, positionPrice)
	portfolioWithOnePos.AllAssets += positionPrice
	return portfolioWithOnePos
}

func (d *DividerByAssetType) processShare(pos domain.PortfolioPosition, vunit_rate float64) *domain.PortfolioByTypeAndCurrency {
	portfolioWithOnePos := domain.NewPortfolioByTypeAndCurrency()
	currencyOfPos := pos.CurrentPrice.Currency
	positionPrice := pos.Quantity.ToFloat() * pos.CurrentPrice.ToFloat() * vunit_rate
	portfolioWithOnePos.SharesAssets.SumOfAssets += positionPrice

	domain.AddToMap(portfolioWithOnePos.SharesAssets.AssetsByCurrency, currencyOfPos, positionPrice)
	portfolioWithOnePos.AllAssets += positionPrice
	return portfolioWithOnePos
}

func (d *DividerByAssetType) processFutures(ctx context.Context, pos domain.PortfolioPosition, vunit_rate float64) (*domain.PortfolioByTypeAndCurrency, error) {
	portfolioWithOnePos := domain.NewPortfolioByTypeAndCurrency()
	positionPrice := pos.Quantity.ToFloat() * pos.CurrentPrice.ToFloat() * vunit_rate
	futures, err := d.tinkoffProvider.TinkoffGetFutureBy(ctx, pos.Figi)
	if err != nil {
		return nil, e.WrapIfErr("can't get future data", err)
	}

	positionPrice = positionPrice / futures.MinPriceIncrement.ToFloat() * futures.MinPriceIncrementAmount.ToFloat()
	portfolioWithOnePos.FuturesAssets.SumOfAssets += positionPrice

	futureType := futures.AssetType
	switch futureType {
	case commodityType:
		domain.AddToMap(portfolioWithOnePos.FuturesAssets.AssetsByType.Commodity.AssetsByCurrency, futures.Name, positionPrice)

		portfolioWithOnePos.FuturesAssets.AssetsByType.Commodity.SumOfAssets += positionPrice

	case currencyType:
		domain.AddToMap(portfolioWithOnePos.FuturesAssets.AssetsByType.Currency.AssetsByCurrency, futures.Name, positionPrice)
		portfolioWithOnePos.FuturesAssets.AssetsByType.Currency.SumOfAssets += positionPrice

	case securityType:
		resp, err := d.tinkoffProvider.TinkoffGetBaseShareFutureValute(ctx, futures.BasicAssetPositionUid)
		if err != nil {
			return nil, e.WrapIfErr("can't dget base share future valute from tinkoff", err)
		}
		valute := resp.Currency
		domain.AddToMap(portfolioWithOnePos.FuturesAssets.AssetsByType.Security.AssetsByCurrency, valute, positionPrice)

		portfolioWithOnePos.FuturesAssets.AssetsByType.Security.SumOfAssets += positionPrice

	case indexType:
		domain.AddToMap(portfolioWithOnePos.FuturesAssets.AssetsByType.Index.AssetsByCurrency, futures.Name, positionPrice)

		portfolioWithOnePos.FuturesAssets.AssetsByType.Index.SumOfAssets += positionPrice

	}
	// Чтобы сумма фьюча не сумировалась с суммой всех активов, так как фактически я за тело фьючерса не заплатил
	// positionPrice = 0
	// portfolioWithOnePos.AllAssets += positionPrice
	return portfolioWithOnePos, nil
}

func (d *DividerByAssetType) processEtf(pos domain.PortfolioPosition, vunit_rate float64) *domain.PortfolioByTypeAndCurrency {
	portfolioWithOnePos := domain.NewPortfolioByTypeAndCurrency()

	currencyOfPos := pos.CurrentPrice.Currency
	positionPrice := pos.Quantity.ToFloat() * pos.CurrentPrice.ToFloat() * vunit_rate
	portfolioWithOnePos.EtfsAssets.SumOfAssets += positionPrice

	domain.AddToMap(portfolioWithOnePos.EtfsAssets.AssetsByCurrency, currencyOfPos, positionPrice)

	portfolioWithOnePos.AllAssets += positionPrice
	return portfolioWithOnePos
}

func (d *DividerByAssetType) processCurrency(ctx context.Context, pos domain.PortfolioPosition, vunit_rate float64) (*domain.PortfolioByTypeAndCurrency, error) {
	portfolioWithOnePos := domain.NewPortfolioByTypeAndCurrency()
	positionPrice := pos.Quantity.ToFloat() * pos.CurrentPrice.ToFloat() * vunit_rate
	curr, err := d.tinkoffProvider.TinkoffGetCurrencyBy(ctx, pos.Figi)
	if err != nil {
		return nil, e.WrapIfErr("can't get currency by figi from tinkoff", err)
	}
	currName := curr.Isin
	portfolioWithOnePos.CurrenciesAssets.SumOfAssets += positionPrice
	domain.AddToMap(portfolioWithOnePos.CurrenciesAssets.AssetsByCurrency, currName, positionPrice)

	portfolioWithOnePos.AllAssets += positionPrice
	return portfolioWithOnePos, nil
}

type pipeline struct {
	ctx    context.Context
	cancel context.CancelFunc
	errCh  chan error
}

func NewPipeline(ctx context.Context, cancel context.CancelFunc, errCh chan error) *pipeline {
	return &pipeline{
		ctx:    ctx,
		cancel: cancel,
		errCh:  errCh,
	}
}

func (p *pipeline) sendErr(err error) {
	select {
	case p.errCh <- err:
	default:
	}
	p.cancel()
}
