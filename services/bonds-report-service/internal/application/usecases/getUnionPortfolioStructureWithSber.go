package usecases

import (
	"bonds-report-service/internal/application/dto"
	"bonds-report-service/internal/application/presenter"
	portfolio "bonds-report-service/internal/application/services/portfolio"
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/utils/logging"
	"context"
	"errors"
	"sync"

	"github.com/gladinov/e"
)

var (
	ErrEmptyReport                = errors.New("no elements in report")
	ErrEmptyPosition              = errors.New("positions are empty")
	ErrpositionsClassCodeVariants = errors.New("positions class code variants are empty")
)

func (s *Service) GetUnionPortfolioStructureWithSber(ctx context.Context) (_ dto.UnionPortfolioStructureWithSberResponce, err error) {
	const op = "service.GetUnionPortfolioStructureWithSber"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	responce := dto.UnionPortfolioStructureWithSberResponce{}
	accounts, err := s.Helpers.TinkoffProvider.TinkoffGetAccounts(ctx)
	if err != nil {
		return dto.UnionPortfolioStructureWithSberResponce{}, e.WrapIfErr("cant' get accounts from tinkoff", err)
	}

	ctxWorkers, cancel := context.WithCancel(ctx)
	defer cancel()
	errCh := make(chan error, 1)
	pipeline := NewPipeline(ctxWorkers, cancel, errCh)
	workers := s.WorkersNumber

	accountsCh := s.produceAccounts(pipeline.ctx, accounts)
	portfolioCh := make(chan *domain.PortfolioByTypeAndCurrency, workers*2)

	var wgStage1 sync.WaitGroup
	for i := 0; i < workers; i++ {
		wgStage1.Add(1)
		go func() {
			defer wgStage1.Done()
			s.worker(pipeline, accountsCh, portfolioCh)
		}()
	}

	wgStage1.Add(1)
	go func() {
		defer wgStage1.Done()
		sberPortfolio, sberErr := s.divideByTypeFromSber(ctxWorkers, s.External.Sber.GetPortfolio())
		if sberErr != nil {
			pipeline.sendErr(e.WrapIfErr("couldnot divide by type from sber", sberErr))
			return
		}
		select {
		case portfolioCh <- sberPortfolio:
		case <-ctxWorkers.Done():
		}
	}()

	go func() {
		wgStage1.Wait()
		close(portfolioCh)
	}()

	positionsList := make([]*domain.PortfolioByTypeAndCurrency, 0, len(accounts)+1)
loop:
	for {
		select {
		case er := <-errCh:
			return dto.UnionPortfolioStructureWithSberResponce{}, er
		default:
			select {
			case <-ctxWorkers.Done():
				return dto.UnionPortfolioStructureWithSberResponce{}, ctxWorkers.Err()
			case portfolio, ok := <-portfolioCh:
				if !ok {
					break loop
				}
				if portfolio != nil {
					positionsList = append(positionsList, portfolio)
				}
			case er := <-errCh:
				return dto.UnionPortfolioStructureWithSberResponce{}, er

			}
		}
	}

	unionPositions := portfolio.UnionPortf(positionsList)

	vizualizeUnionPositions := presenter.ResponsePortfolioStructure(ctx, s.logger, unionPositions, dto.UnionPortfWithSber, "")

	responce.Report = vizualizeUnionPositions
	return responce, nil
}

func (s *Service) worker(p *pipeline, in <-chan domain.Account, out chan<- *domain.PortfolioByTypeAndCurrency) {
	for account := range in {
		portfolio, err := s.Helpers.TinkoffProvider.TinkoffGetPortfolio(p.ctx, account)
		if err != nil {
			p.sendErr(e.WrapIfErr("cant' get portfolio from Tinkoff", err))
			return
		}
		positions := portfolio.Positions

		portfolioStructure, err := s.Helpers.DividerByAssetType.DivideByType(p.ctx, positions)
		if err != nil {
			p.sendErr(e.WrapIfErr("couldnot divide by type", err))
			return
		}
		select {
		case <-p.ctx.Done():
			return
		case out <- portfolioStructure:
		}
	}
}

func (s *Service) divideByTypeFromSber(ctx context.Context, positions map[string]float64) (_ *domain.PortfolioByTypeAndCurrency, err error) {
	const op = "service.DivideByTypeFromSber"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	portfolio := domain.NewPortfolioByTypeAndCurrency()

	if len(positions) == 0 {
		return portfolio, ErrEmptyPosition
	}
	for ticker, quantity := range positions {
		positionsClassCodeVariants, err := s.Helpers.TinkoffProvider.TinkoffFindBy(ctx, ticker)
		if err != nil {
			return nil, e.WrapIfErr("can't find by ticker from tinkoff", err)
		}
		if len(positionsClassCodeVariants) == 0 {
			return nil, ErrpositionsClassCodeVariants
		}

		switch positionsClassCodeVariants[0].InstrumentType {
		case bond:
			bondUid := positionsClassCodeVariants[0].Uid
			bond, err := s.Helpers.TinkoffProvider.TinkoffGetBondByUid(ctx, bondUid)
			if err != nil {
				return nil, e.WrapIfErr("can't get bond by uid from tinkoff", err)
			}
			currentNkd := bond.AciValue.ToFloat()
			currency := bond.Currency
			resp, err := s.Helpers.TinkoffProvider.TinkoffGetLastPriceInPersentageToNominal(ctx, bondUid)
			if err != nil {
				return nil, e.WrapIfErr("can't get last price in persentage to nominal from tinkoff", err)
			}
			currentPriceInPers := resp.LastPrice.ToFloat()
			currentPrice := currentPriceInPers / 100 * bond.Nominal.ToFloat()
			currentNkdOfPosition := currentNkd * quantity
			positionPrice := currentPrice*quantity + currentNkdOfPosition

			portfolio.AllAssets += positionPrice
			portfolio.BondsAssets.SumOfAssets += positionPrice

			if existing, exist := portfolio.BondsAssets.AssetsByCurrency[currency]; !exist {
				portfolio.BondsAssets.AssetsByCurrency[currency] = &domain.AssetByParam{
					SumOfAssets: positionPrice,
				}
			} else {
				existing.SumOfAssets += positionPrice
			}

		case share:
		case futures:
		case etf:
		case currency:
		default:
		}

	}
	return portfolio, nil
}
