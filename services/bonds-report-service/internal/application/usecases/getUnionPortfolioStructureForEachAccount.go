package usecases

import (
	"bonds-report-service/internal/application/dto"
	"bonds-report-service/internal/application/presenter"
	portfolio "bonds-report-service/internal/application/services/portfolio"
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/utils/logging"
	"context"
	"sync"

	"github.com/gladinov/e"
)

func (s *Service) GetUnionPortfolioStructureForEachAccount(ctx context.Context) (_ domain.UnionPortfolioStructureResponce, err error) {
	const op = "service.GetUnionPortfolioStructureForEachAccount"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	accounts, err := s.Helpers.TinkoffProvider.TinkoffGetAccounts(ctx)
	response := domain.UnionPortfolioStructureResponce{}
	if err != nil {
		return domain.UnionPortfolioStructureResponce{}, e.WrapIfErr("cant' get accounts from tinkoff", err)
	}
	unionPortfolioStructure, err := s.getUnionPortfolioStructure(ctx, accounts)
	if err != nil {
		return domain.UnionPortfolioStructureResponce{}, e.WrapIfErr("cant' get union portfolio structure", err)
	}
	response.Report = unionPortfolioStructure

	return response, nil
}

func (s *Service) getUnionPortfolioStructure(ctx context.Context, accounts map[string]domain.Account) (_ string, err error) {
	const op = "service.getUnionPortfolioStructure"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	positionsList := make([]*domain.PortfolioByTypeAndCurrency, 0)

	ctxWorkers, cancel := context.WithCancel(ctx)
	defer cancel()
	workers := s.WorkersNumber

	portfolioCh := make(chan domain.Portfolio, workers*2) // TODO: what size i must do in chan?
	portfolioStructCh := make(chan *domain.PortfolioByTypeAndCurrency, workers*2)
	accountsCh := s.produceAccounts(ctxWorkers, accounts)

	errCh := make(chan error, 1)
	pipeline := NewPipeline(ctxWorkers, cancel, errCh)
	var wgStage1 sync.WaitGroup
	for i := 0; i < workers; i++ {
		wgStage1.Add(1)
		go func() {
			defer wgStage1.Done()
			s.portfolioWorkers(pipeline, accountsCh, portfolioCh)
		}()
	}

	go func() {
		wgStage1.Wait()
		close(portfolioCh)
	}()

	var wgStage2 sync.WaitGroup
	for i := 0; i < workers; i++ {
		wgStage2.Add(1)
		go func() {
			defer wgStage2.Done()
			s.portfolioStructureWorkers(pipeline, portfolioCh, portfolioStructCh)
		}()
	}

	go func() {
		wgStage2.Wait()
		close(portfolioStructCh)
	}()

loop:
	for {
		select {
		case <-ctxWorkers.Done():
			return "", ctxWorkers.Err()
		case er := <-errCh:
			cancel()
			return "", er
		case portfolioStructure, ok := <-portfolioStructCh:
			if !ok {
				break loop
			}
			positionsList = append(positionsList, portfolioStructure)
		}
	}
	unionPositions := portfolio.UnionPortf(positionsList)

	vizualizeUnionPositions := presenter.ResponsePortfolioStructure(ctx, s.logger, unionPositions, dto.UnionPortf, "")

	return vizualizeUnionPositions, nil
}

func (s *Service) portfolioWorkers(p *pipeline, in <-chan domain.Account, out chan<- domain.Portfolio) {
	for account := range in {
		portfolio, err := s.Helpers.TinkoffProvider.TinkoffGetPortfolio(p.ctx, account)
		if err != nil {
			p.sendErr(e.WrapIfErr("can't get portfolio from Tinkoff", err))
			return
		}

		select {
		case <-p.ctx.Done():
			return
		case out <- portfolio:
		}
	}
}

func (s *Service) portfolioStructureWorkers(p *pipeline, in <-chan domain.Portfolio, out chan<- *domain.PortfolioByTypeAndCurrency) {
	for portfolio := range in {
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
