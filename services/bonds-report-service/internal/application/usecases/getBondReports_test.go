package usecases

// import (
// 	tinkoffHelper "bonds-report-service/internal/application/services/tinkoff"
// 	"bonds-report-service/internal/application/ports/mocks"
// 	"log/slog"
// 	"testing"
// )

// func newTestService(t *testing.T) *Service {
// 	logger := slog.New(slog.DiscardHandler)
// 	workersNumber := 5
// 	moexMock := mocks.NewMoexClient(t)
// 	cbrMock := mocks.NewCbrClient(t)
// 	storage := mocks.NewStorage(t)

// 	bondReportProcessorMock := mocks.NewBondReportProcessor(t)

// 	cbrCurrencyGetterMock := mocks.NewCbrCurrencyGetter(t)

// 	generalBondReporterMock := mocks.NewGeneralBondReportProcessor(t)

// 	moexSpecificationGetterMock := mocks.NewMoexSpecificationGetter(t)

// 	reportProcessorMock := mocks.NewReportProcessor(t)

// 	operationsUpdaterMock := mocks.NewOperationsUpdater(t)

// 	positionProcessorMock := mocks.NewPositionProcessor(t)

// 	reportLineBuilderMock := mocks.NewReportLineBuilder(t)

// 	dividerByAssetTypeMock := mocks.NewDividerByAssetType(t)

// 	instrumentsClientMock := mocks.NewTinkoffInstrumentsClient(t)

// 	analyticsClientMock := mocks.NewTinkoffAnalyticsClient(t)

// 	portfolioClientMock := mocks.NewTinkoffPortfolioClient(t)

// 	tinkoffHelper := tinkoffHelper.NewTinkoffHelper(logger, instrumentsClientMock, portfolioClientMock, analyticsClientMock)

// 	helpers := NewHelpers(bondReportProcessorMock,
// 		cbrCurrencyGetterMock,
// 		generalBondReporterMock,
// 		moexSpecificationGetterMock,
// 		reportProcessorMock,
// 		tinkoffHelper,
// 		operationsUpdaterMock,
// 		positionProcessorMock,
// 		reportLineBuilderMock,
// 		dividerByAssetTypeMock)

// 	externalApis := NewExternalApis(moexMock, cbrMock, nil)

// 	s := NewService(logger, workersNumber, externalApis, helpers, storage)
// 	return s
// }

// // func TestService_GetBondReports(t *testing.T) {
// // 	t.Run("success", func(t *testing.T) {
// // 		logger := slog.New(slog.DiscardHandler)
// // 		s := NewService(logger, workersNumber, externalApis, helpers, storage)
// // 		got, gotErr := s.GetBondReports(context.Background(), chatID)
// // 	})
// // }
