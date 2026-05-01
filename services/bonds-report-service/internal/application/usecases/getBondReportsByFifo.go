package usecases

import (
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/domain/report"
	"bonds-report-service/internal/utils/logging"
	"context"
	"sync"

	"github.com/gladinov/e"
)

func (s *Service) GetBondReportsByFifo(ctx context.Context, chatID int) (err error) {
	const op = "service.GetBondReportsByFifo"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	accounts, err := s.Helpers.TinkoffProvider.TinkoffGetAccounts(ctx)
	if err != nil {
		return e.WrapIfErr("get accounts error", err)
	}

	ctxWorkers, cancel := context.WithCancel(ctx)
	defer cancel()
	workers := s.WorkersNumber
	errCh := make(chan error, 1)

	var wg sync.WaitGroup

	accountsCh := s.produceAccounts(ctxWorkers, accounts)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.accountWorkerForBondReportByFifo(ctxWorkers, accountsCh, errCh, chatID)
		}()
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	select {
	case <-ctxWorkers.Done():
		return ctxWorkers.Err()
	case er, ok := <-errCh:
		if !ok {
			return nil
		}
		cancel()
		return er
	}
}

func (s *Service) accountWorkerForBondReportByFifo(ctx context.Context, accountsCh <-chan domain.Account, errCh chan<- error, chatID int) {
	for {
		select {
		case <-ctx.Done():
			return
		case account, ok := <-accountsCh:
			if !ok {
				return
			}
			err := s.processAccountForBondReportByFifo(ctx, chatID, account)
			if err != nil {
				select {
				case errCh <- e.WrapIfErr("failed to process account", err):
				case <-ctx.Done():
				}
				return
			}
		}
	}
}

func (s *Service) processAccountForBondReportByFifo(ctx context.Context, chatID int, account domain.Account) (err error) {
	const op = "service.processAccountForBondReportByFifo"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		err := s.Helpers.OperationsUpdater.UpdateOperations(ctx, chatID, account.ID, account.OpenedDate)
		if err != nil {
			return e.WrapIfErr("update operation error", err)
		}

		portfolio, err := s.Helpers.TinkoffProvider.TinkoffGetPortfolio(ctx, account)
		if err != nil {
			return e.WrapIfErr("tinkoffGetPortfolio err", err)
		}

		portfolioPositions, err := s.Helpers.PositionProcessor.ProcessPositionsToPositionsWithAssetUid(ctx, portfolio.Positions)
		if err != nil {
			return e.WrapIfErr("transformPositions err", err)
		}
		err = s.Storage.DeleteBondReport(ctx, chatID, account.ID)
		if err != nil {
			return e.WrapIfErr("deleteBondReport err", err)
		}

		operationsDb, err := s.Storage.GetAllOperations(ctx, chatID, account.ID)
		if err != nil {
			return e.WrapIfErr("failed to get all operations ", err)
		}
		operationsByAssetUid := mapOperationsWithoutCustomTypesToMapByAssetUid(operationsDb)

		var wg sync.WaitGroup
		ctxWorkers, cancel := context.WithCancel(ctx)
		defer cancel()
		bondsInRub := make([]report.BondReport, 0)
		workers := s.WorkersNumber
		positionCh := make(chan domain.PortfolioPositionsWithAssetUid, workers*2)
		bondReportCh := make(chan report.Report, workers*2)
		errCh := make(chan error, 1)

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				s.bondReportWorker(ctxWorkers, positionCh, bondReportCh, errCh, operationsByAssetUid)
			}()
		}

		go func() {
			s.producePositionsForBondReportByFifo(ctxWorkers, positionCh, portfolioPositions)
		}()

		go func() {
			wg.Wait()
			close(bondReportCh)
		}()

	loop:
		for {
			select {
			case <-ctxWorkers.Done():
				return ctxWorkers.Err()
			case errAgg := <-errCh:
				cancel()
				return e.WrapIfErr("failed to process positions for bond reports by FIFO", errAgg)
			case bondReport, ok := <-bondReportCh:
				if !ok {
					break loop
				}
				bondsInRub = append(bondsInRub, bondReport.BondsInRUB...)

			}
		}
		err = s.Storage.SaveBondReport(ctx, chatID, account.ID, bondsInRub)
		if err != nil {
			return e.WrapIfErr("Storage.SaveBondReport error", err)
		}
		return nil

	}
}

func (s *Service) bondReportWorker(ctx context.Context, positionCh <-chan domain.PortfolioPositionsWithAssetUid, bondReportCh chan<- report.Report, errCh chan<- error, operationsByAssetUid map[string][]domain.OperationWithoutCustomTypes) {
	for {
		select {
		case <-ctx.Done():
			return
		case position, ok := <-positionCh:
			if !ok {
				return
			}
			operationsOfPosition := operationsByAssetUid[position.AssetUid]
			bondReport, errWorkers := s.processPositionsForBondReportByFifo(ctx, position, operationsOfPosition)
			if errWorkers != nil {
				select {
				case <-ctx.Done():
				case errCh <- errWorkers:
				}
				return
			}
			select {
			case <-ctx.Done():
				return
			case bondReportCh <- bondReport:
			}
		}
	}
}

func (s *Service) producePositionsForBondReportByFifo(ctx context.Context, positionCh chan<- domain.PortfolioPositionsWithAssetUid, portfolioPositions []domain.PortfolioPositionsWithAssetUid) {
	defer close(positionCh)
	for _, position := range portfolioPositions {
		if !isBondType(position) {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case positionCh <- position:
		}
	}
}

func (s *Service) processPositionsForBondReportByFifo(ctx context.Context, position domain.PortfolioPositionsWithAssetUid, operationDbByAssetUid []domain.OperationWithoutCustomTypes) (report.Report, error) {
	select {
	case <-ctx.Done():
		return report.Report{}, ctx.Err()
	default:
		reporLines, err := s.Helpers.ReportLineBuilder.CreateNewReportLines(ctx, position, operationDbByAssetUid)
		if err != nil {
			return report.Report{}, e.WrapIfErr("failed to create new report lines", err)
		}

		resultBondPosition, err := s.Helpers.ReportProcessor.ProcessOperations(ctx, reporLines)
		if err != nil {
			return report.Report{}, e.WrapIfErr("failed to process operation", err)
		}

		firstBuyDate := resultBondPosition.CurrentPositions[0].BuyDate

		moexBuyDateData, err := s.Helpers.MoexSpecificationGetter.GetSpecificationsFromMoex(ctx, reporLines.Bond.Ticker, firstBuyDate)
		if err != nil {
			return report.Report{}, e.WrapIfErr("failed to get specifications from moex to buy date", err)
		}
		moexNowData, err := s.Helpers.MoexSpecificationGetter.GetSpecificationsFromMoex(ctx, reporLines.Bond.Ticker, firstBuyDate)
		if err != nil {
			return report.Report{}, e.WrapIfErr("failed to get specifications from moex to buy now", err)
		}

		bondReport, err := s.Helpers.BondReportProcessor.CreateBondReport(
			ctx,
			resultBondPosition.CurrentPositions,
			moexBuyDateData,
			moexNowData,
		)
		if err != nil {
			return report.Report{}, e.WrapIfErr("CreateBondReport error", err)
		}
		return bondReport, nil
	}
}

func mapOperationsWithoutCustomTypesToMapByAssetUid(operationsDb []domain.OperationWithoutCustomTypes) map[string][]domain.OperationWithoutCustomTypes {
	operationsByAssetUid := make(map[string][]domain.OperationWithoutCustomTypes)

	for _, operations := range operationsDb {
		operationsByAssetUid[operations.AssetUid] = append(operationsByAssetUid[operations.AssetUid], operations)
	}
	return operationsByAssetUid
}
