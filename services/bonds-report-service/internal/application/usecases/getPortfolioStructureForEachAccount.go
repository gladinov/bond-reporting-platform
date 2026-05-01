package usecases

import (
	"bonds-report-service/internal/application/dto"
	"bonds-report-service/internal/application/presenter"
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/utils/logging"
	"context"
	"sync"

	"github.com/gladinov/e"
)

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

func (s *Service) GetPortfolioStructureForEachAccount(ctx context.Context) (_ domain.PortfolioStructureForEachAccountResponce, err error) {
	const op = "service.GetPortfolioStructureForEachAccount"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	accounts, err := s.Helpers.TinkoffProvider.TinkoffGetAccounts(ctx)
	response := domain.PortfolioStructureForEachAccountResponce{}
	if err != nil {
		return domain.PortfolioStructureForEachAccountResponce{}, e.WrapIfErr("cant' get accounts from tinkoff", err)
	}

	ctxWorkers, cancel := context.WithCancel(ctx)
	defer cancel()
	errCh := make(chan error, 1)
	workers := s.WorkersNumber
	var wg sync.WaitGroup
	pipeline := NewPipeline(ctxWorkers, cancel, errCh)
	portfolioStructsCh := make(chan string, workers*2)

	produceAccountCh := s.produceAccounts(ctxWorkers, accounts)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.portfolioStructureWorker(pipeline, produceAccountCh, portfolioStructsCh)
		}()
	}

	go func() {
		wg.Wait()
		close(portfolioStructsCh)
	}()

loop:
	for {
		select {
		case <-ctxWorkers.Done():
			return domain.PortfolioStructureForEachAccountResponce{}, ctxWorkers.Err()
		case er := <-errCh:
			return domain.PortfolioStructureForEachAccountResponce{}, er
		case report, ok := <-portfolioStructsCh:
			if !ok {
				break loop
			}
			response.PortfolioStructures = append(response.PortfolioStructures, report)
		}
	}

	return response, nil
}

func (s *Service) portfolioStructureWorker(p *pipeline, accountCh <-chan domain.Account, reportCh chan<- string) {
	for account := range accountCh {
		report, err := s.getPortfolioStructure(p.ctx, account)
		if err != nil {
			p.sendErr(e.WrapIfErr("cant' get portfolio structure", err))
			return
		}
		select {
		case <-p.ctx.Done():
			return
		case reportCh <- report:
		}

	}
}

func (s *Service) getPortfolioStructure(ctx context.Context, account domain.Account) (_ string, err error) {
	const op = "service.getPortfolioStructure"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		portfolio, err := s.Helpers.TinkoffProvider.TinkoffGetPortfolio(ctx, account)
		if err != nil {
			return "", e.WrapIfErr("cant' get portfolio from Tinkoff", err)
		}
		positions := portfolio.Positions

		potfolioStructure, err := s.Helpers.DividerByAssetType.DivideByType(ctx, positions)
		if err != nil {
			return "", e.WrapIfErr("couldnot divide by type", err)
		}
		response := presenter.ResponsePortfolioStructure(ctx, s.logger, potfolioStructure, dto.EachPortf, account.Name)

		return response, nil
	}
}
