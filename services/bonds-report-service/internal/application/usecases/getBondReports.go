package usecases

import (
	"bonds-report-service/internal/application/dto"
	"bonds-report-service/internal/application/presenter"
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/domain/generalbondreport"
	"bonds-report-service/internal/utils/logging"
	"context"
	"errors"
	"log/slog"
	"strconv"
	"sync"

	"github.com/gladinov/e"
)

type reportJob struct {
	GeneralBondReport generalbondreport.GeneralBondReports
	AccountID         string
}

var ErrEmptyBondPositions = errors.New("len of result bond positions is empty")

func (s *Service) GetBondReports(ctx context.Context, chatID int) (_ dto.BondReportsResponce, err error) {
	const op = "service.GetBondReports"
	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	reportsInByteByAccounts := make([][]*dto.MediaGroup, 0)

	accounts, err := s.Helpers.TinkoffProvider.TinkoffGetAccounts(ctx)
	if err != nil {
		return dto.BondReportsResponce{}, e.WrapIfErr("failde to get accounts from Tinkoff", err)
	}
	ctxWorkers, cancel := context.WithCancel(ctx)
	defer cancel()
	bufSize := min(len(accounts), s.WorkersNumber)

	reportsCh := make(chan reportJob, s.WorkersNumber)
	mediaGroupsCh := make(chan []*dto.MediaGroup, s.WorkersNumber*2)
	errCh := make(chan error, 1)
	var wgStage1 sync.WaitGroup

	accountsCh := s.produceAccounts(ctxWorkers, accounts)

	for i := 0; i < bufSize; i++ {
		wgStage1.Add(1)
		go func() {
			defer wgStage1.Done()
			s.fetchReportsWorker(ctxWorkers, accountsCh, errCh, reportsCh, chatID)
		}()
	}

	var wgStage2 sync.WaitGroup
	for i := 0; i < bufSize; i++ {
		wgStage2.Add(1)
		go func() {
			defer wgStage2.Done()
			s.renderReportsWorker(ctxWorkers, reportsCh, errCh, mediaGroupsCh, chatID)
		}()
	}

	go func() {
		wgStage1.Wait()
		close(reportsCh)
	}()

	go func() {
		wgStage2.Wait()
		close(mediaGroupsCh)
	}()

loop:
	for {
		select {
		case er := <-errCh:
			cancel()
			return dto.BondReportsResponce{}, er
		case reportsInByte, ok := <-mediaGroupsCh:
			if !ok {
				break loop
			}
			reportsInByteByAccounts = append(reportsInByteByAccounts, reportsInByte)
		}
	}

	getBondReportsResponce := dto.BondReportsResponce{Media: reportsInByteByAccounts}
	return getBondReportsResponce, nil
}

func (s *Service) ProduceBondReports(ctx context.Context, reportKind, traceID, chatIDStr string) (err error) {
	const op = "service.ProduceBondReports"
	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		return s.handleProduceError(ctx, reportKind, chatIDStr, traceID, dto.ErrInternal, "internal err", "failed to convert chatID to int", err)
	}

	accounts, err := s.Helpers.TinkoffProvider.TinkoffGetAccounts(ctx)
	if err != nil {
		return s.handleProduceError(ctx, reportKind, chatIDStr, traceID, dto.ErrTinkoffAPI, "tinkoff API err", "failed to get accounts from Tinkoff", err)
	}

	reportsInBytes, err := s.buildBondReportsPipeline(ctx, accounts, chatID)
	if err != nil {
		return s.handleProduceError(ctx, reportKind, chatIDStr, traceID, dto.ErrInternal, "internal err", "failed to build bond reports", err)
	}

	response := dto.BondReportsResponce{Media: reportsInBytes}
	return s.Producer.PublishBondReportWithPng(ctx, reportKind, chatIDStr, traceID, response)
}

func (s *Service) buildBondReportsPipeline(
	ctx context.Context,
	accounts map[string]domain.Account,
	chatID int,
) ([][]*dto.MediaGroup, error) {
	if len(accounts) == 0 {
		return nil, nil
	}

	workersCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	workers := min(len(accounts), s.WorkersNumber)

	reportsCh := make(chan reportJob, workers)
	mediaGroupsCh := make(chan []*dto.MediaGroup, workers*2)
	errCh := make(chan error, 1)

	accountsCh := s.produceAccounts(workersCtx, accounts)

	var wgFetch sync.WaitGroup
	for i := 0; i < workers; i++ {
		wgFetch.Add(1)
		go func() {
			defer wgFetch.Done()
			s.fetchReportsWorker(workersCtx, accountsCh, errCh, reportsCh, chatID)
		}()
	}

	var wgRender sync.WaitGroup
	for i := 0; i < workers; i++ {
		wgRender.Add(1)
		go func() {
			defer wgRender.Done()
			s.renderReportsWorker(workersCtx, reportsCh, errCh, mediaGroupsCh, chatID)
		}()
	}

	go func() {
		wgFetch.Wait()
		close(reportsCh)
	}()

	go func() {
		wgRender.Wait()
		close(mediaGroupsCh)
	}()

	var result [][]*dto.MediaGroup

	for {
		select {
		case pipelineErr := <-errCh:
			if pipelineErr != nil {
				cancel()
				return nil, pipelineErr
			}
		case mediaGroups, ok := <-mediaGroupsCh:
			if !ok {
				return result, nil
			}
			result = append(result, mediaGroups)
		}
	}
}

func (s *Service) produceAccounts(ctx context.Context, accounts map[string]domain.Account) <-chan domain.Account {
	out := make(chan domain.Account, s.WorkersNumber*2)

	go func() {
		defer close(out)
		for _, account := range accounts {

			if !isActiveAccounts(account) {
				continue
			}

			select {
			case <-ctx.Done():
				return
			case out <- account:
			}
		}
	}()

	return out
}

func (s *Service) fetchReportsWorker(
	ctx context.Context,
	jobCh <-chan domain.Account,
	errCh chan<- error,
	generalBondReportsCh chan<- reportJob,
	chatID int,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case account, ok := <-jobCh:
			if !ok {
				return
			}

			generalBondReports, err := s.processAccount(ctx, chatID, account)
			if err != nil {
				select {
				case errCh <- e.WrapIfErr("failed to procces account", err):
				case <-ctx.Done():
				}
				return

			}

			select {
			case generalBondReportsCh <- reportJob{
				GeneralBondReport: generalBondReports,
				AccountID:         account.ID,
			}:
			case <-ctx.Done():
				return
			}

		}
	}
}

func (s *Service) renderReportsWorker(
	ctx context.Context,
	generalBondReportsCh <-chan reportJob,
	errCh chan<- error,
	reportCh chan<- []*dto.MediaGroup,
	chatID int,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case genralBondReportWithAccount, ok := <-generalBondReportsCh:
			if !ok {
				return
			}
			reportsInByte, err := presenter.GenerateTablePNG(ctx,
				s.logger,
				&genralBondReportWithAccount.GeneralBondReport,
				chatID,
				genralBondReportWithAccount.AccountID)
			if err != nil {
				select {
				case errCh <- e.WrapIfErr("failed to GenerateTablePNG", err):
				case <-ctx.Done():
				}
				return
			}
			select {
			case reportCh <- reportsInByte:
			case <-ctx.Done():
				return
			}
		}
	}
}

func isActiveAccounts(account domain.Account) bool {
	if account.Status == 2 { // TODO: Magic Number
		return true
	}
	return false
}

func (s *Service) processAccount(ctx context.Context, chatID int, account domain.Account) (_ generalbondreport.GeneralBondReports, err error) {
	const op = "service.processAccount"
	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	select {
	case <-ctx.Done():
		return generalbondreport.GeneralBondReports{}, ctx.Err()
	default:

		err = s.Helpers.OperationsUpdater.UpdateOperations(ctx, chatID, account.ID, account.OpenedDate)
		if err != nil {
			return generalbondreport.GeneralBondReports{}, e.WrapIfErr("failed to update operations", err)
		}
		portfolio, err := s.Helpers.TinkoffProvider.TinkoffGetPortfolio(ctx, account)
		if err != nil {
			return generalbondreport.GeneralBondReports{}, e.WrapIfErr("failed to get portfolio from tinkoff", err)
		}

		totalAmount := portfolio.TotalAmount.ToFloat()

		portfolioPositions, err := s.Helpers.PositionProcessor.ProcessPositionsToPositionsWithAssetUid(ctx, portfolio.Positions)
		if err != nil {
			return generalbondreport.GeneralBondReports{}, e.WrapIfErr("failed to process position to postition with asset uid", err)
		}

		allOperations, err := s.Storage.GetAllOperations(ctx, chatID, account.ID)
		if err != nil {
			return generalbondreport.GeneralBondReports{}, e.WrapIfErr("failed to get all operations from storage", err)
		}

		operationsByAssetUid := mapOperationsWithoutCustomTypesToMapByAssetUid(allOperations)

		generalBondReports := generalbondreport.NewGeneralBondReports()

		workers := s.WorkersNumber
		ctxWorkers, cancel := context.WithCancel(ctx)
		defer cancel()
		workList := make(chan domain.PortfolioPositionsWithAssetUid, workers*2)
		bondReportCh := make(chan generalbondreport.GeneralBondReportPosition, workers*2)
		errCh := make(chan error, 1)
		var wg sync.WaitGroup

		// Создаю N Воркеров
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				s.bondWorker(ctxWorkers, workList, bondReportCh, errCh, totalAmount, operationsByAssetUid)
			}()

		}
		// 1. Начинаем скармливать работу вокрерам
		go s.producePositions(ctxWorkers, portfolioPositions, workList)

		go func() {
			wg.Wait()
			close(bondReportCh)
		}()

	loop:
		for {
			select {
			case bondReport, ok := <-bondReportCh:
				if !ok {
					break loop
				}
				s.addBondReport(&generalBondReports, bondReport)
			case er := <-errCh:
				cancel()
				return generalbondreport.GeneralBondReports{}, e.WrapIfErr("failed to process BondPosition", er)

			}
		}

		return generalBondReports, nil
	}
}

func (s *Service) bondWorker(
	ctx context.Context,
	workList <-chan domain.PortfolioPositionsWithAssetUid,
	resultCh chan<- generalbondreport.GeneralBondReportPosition,
	errCh chan<- error,
	totalAmount float64,
	operationsByAssetUid map[string][]domain.OperationWithoutCustomTypes,
) {
	for position := range workList {
		bondReport, er := s.processBondPosition(ctx,
			position,
			totalAmount,
			operationsByAssetUid[position.AssetUid])
		if er != nil {
			if errors.Is(er, ErrEmptyBondPositions) {
				continue
			}
			select {
			case errCh <- er:
			case <-ctx.Done():
			}
			return
		}
		select {
		case resultCh <- bondReport:
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) producePositions(
	ctx context.Context,
	positions []domain.PortfolioPositionsWithAssetUid,
	out chan<- domain.PortfolioPositionsWithAssetUid,
) {
	defer close(out)

	for _, position := range positions {
		if !isBondType(position) {
			continue
		}
		select {
		case out <- position:
		case <-ctx.Done():
			return
		}
	}
}

func isBondType(position domain.PortfolioPositionsWithAssetUid) bool {
	if position.InstrumentType != "bond" {
		return false
	}
	return true
}

func (s *Service) processBondPosition(ctx context.Context,
	position domain.PortfolioPositionsWithAssetUid,
	totalAmount float64,
	operationsDb []domain.OperationWithoutCustomTypes,
) (_ generalbondreport.GeneralBondReportPosition, err error) {
	const op = "service.processBondPosition"
	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	select {
	case <-ctx.Done():
		return generalbondreport.GeneralBondReportPosition{}, ctx.Err()
	default:

		reporLines, err := s.Helpers.ReportLineBuilder.CreateNewReportLines(ctx, position, operationsDb)
		if err != nil {
			return generalbondreport.GeneralBondReportPosition{}, e.WrapIfErr("failed to create new report lines", err)
		}
		// Обрабатываем операции и получаем открытые позиции по данной бумаге
		resultBondPosition, err := s.Helpers.ReportProcessor.ProcessOperations(ctx, reporLines)
		if err != nil {
			return generalbondreport.GeneralBondReportPosition{}, e.WrapIfErr("failed to process operation", err)
		}
		// Общая стоимость портфеля

		if len(resultBondPosition.CurrentPositions) == 0 {
			s.logger.WarnContext(
				ctx,
				"len of result bond positions is empty. But in portfolio request to TinkoffApi position in portfolio",
				slog.Any("postion", position))
			return generalbondreport.GeneralBondReportPosition{}, ErrEmptyBondPositions
		}

		firstBuyDate := resultBondPosition.CurrentPositions[0].BuyDate

		// TODO : Кэшифрование запросов в MOEX
		moexBuyDateData, err := s.Helpers.MoexSpecificationGetter.GetSpecificationsFromMoex(ctx, reporLines.Bond.Ticker, firstBuyDate)
		if err != nil {
			return generalbondreport.GeneralBondReportPosition{}, e.WrapIfErr("failed to get specifications from moex to buy date", err)
		}
		moexNowData, err := s.Helpers.MoexSpecificationGetter.GetSpecificationsFromMoex(ctx, reporLines.Bond.Ticker, firstBuyDate)
		if err != nil {
			return generalbondreport.GeneralBondReportPosition{}, e.WrapIfErr("failed to get specifications from moex to buy now", err)
		}

		bondReport, err := s.Helpers.GeneralBondReportProcessor.GetGeneralBondReportPosition(
			ctx,
			resultBondPosition.CurrentPositions,
			totalAmount,
			moexBuyDateData,
			moexNowData, firstBuyDate)
		if err != nil {
			return generalbondreport.GeneralBondReportPosition{}, e.WrapIfErr("failed to get general bond report position", err)
		}
		return bondReport, nil
	}
}

func (s *Service) addBondReport(generalBondReports *generalbondreport.GeneralBondReports, bondReport generalbondreport.GeneralBondReportPosition) {
	switch {
	case bondReport.Replaced:
		tickerTimeKey := generalbondreport.NewTickerTimeKey(bondReport.Ticker, bondReport.BuyDate)
		generalBondReports.ReplacedBondsReport[tickerTimeKey] = bondReport
	case bondReport.Currencies != "rub":
		tickerTimeKey := generalbondreport.NewTickerTimeKey(bondReport.Ticker, bondReport.BuyDate)
		generalBondReports.EuroBondsReport[tickerTimeKey] = bondReport
	default:
		tickerTimeKey := generalbondreport.NewTickerTimeKey(bondReport.Ticker, bondReport.BuyDate)
		generalBondReports.RubBondsReport[tickerTimeKey] = bondReport
	}
}

func (s *Service) handleProduceError(
	ctx context.Context,
	reportKind, chatIDStr, traceID string,
	errCode dto.ErrorCode,
	userMsg string,
	logMsg string,
	cause error,
) error {
	wrappedErr := e.WrapIfErr(logMsg, cause)

	s.logger.ErrorContext(ctx, logMsg, slog.Any("error", wrappedErr))

	if err := s.Producer.PublishFailedBondReportWithPng(
		ctx,
		reportKind,
		chatIDStr,
		traceID,
		string(errCode),
		userMsg,
	); err != nil {
		return e.WrapIfErr("failed to publish failed data", err)
	}

	return nil
}
