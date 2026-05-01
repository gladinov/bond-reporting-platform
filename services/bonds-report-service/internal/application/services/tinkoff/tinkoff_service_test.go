//go:build unit

package tinkoffHelper

import (
	"bonds-report-service/internal/application/ports/mocks"
	"bonds-report-service/internal/domain"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	factories "bonds-report-service/internal/application/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTinkoffGetPortfolio(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		account := factories.NewOpenAccount()
		wantPortfolio := factories.NewPortfolio()

		mockPortfolioClient := mocks.NewTinkoffPortfolioClient(t)
		mockPortfolioClient.On("GetPortfolio", ctx, account.ID, account.Status).
			Return(wantPortfolio, nil)

		srv := &TinkoffHelper{
			logger:    logger,
			Portfolio: mockPortfolioClient,
		}

		got, err := srv.TinkoffGetPortfolio(ctx, account)
		assert.NoError(t, err)
		assert.Equal(t, wantPortfolio, got)

		mockPortfolioClient.AssertExpectations(t)
	})

	t.Run("Err: Validate account fails", func(t *testing.T) {
		account := factories.NewOpenAccount()
		account.ID = "" // пустой ID вызовет ошибку в ValidateForPortfolio

		srv := &TinkoffHelper{
			logger: logger,
		}

		_, err := srv.TinkoffGetPortfolio(ctx, account)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to validate account")
	})

	t.Run("Err: Tinkoff Portfolio client fails", func(t *testing.T) {
		account := factories.NewOpenAccount()

		mockPortfolioClient := mocks.NewTinkoffPortfolioClient(t)
		mockPortfolioClient.On("GetPortfolio", ctx, account.ID, account.Status).
			Return(domain.Portfolio{}, errors.New("client error"))

		srv := &TinkoffHelper{
			logger:    logger,
			Portfolio: mockPortfolioClient,
		}

		_, err := srv.TinkoffGetPortfolio(ctx, account)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get portfolio")

		mockPortfolioClient.AssertExpectations(t)
	})
}

func TestTinkoffGetOperations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := factories.NewOperationsRequest()
		wantOperations := []domain.Operation{factories.NewOperation(), factories.NewOperation()}

		mockPortfolioClient := mocks.NewTinkoffPortfolioClient(t)
		mockPortfolioClient.On("GetOperations", ctx, req.AccountID, req.FromDate).
			Return(wantOperations, nil)

		srv := &TinkoffHelper{
			logger:    logger,
			Portfolio: mockPortfolioClient,
			now:       time.Now,
		}

		got, err := srv.TinkoffGetOperations(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, wantOperations, got)

		mockPortfolioClient.AssertExpectations(t)
	})

	t.Run("Err: Validate request fails", func(t *testing.T) {
		req := factories.NewOperationsRequest()
		req.AccountID = ""
		srv := &TinkoffHelper{
			logger: logger,
			now:    time.Now,
		}

		_, err := srv.TinkoffGetOperations(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to validate")
	})

	t.Run("Err: Tinkoff Portfolio client fails", func(t *testing.T) {
		req := factories.NewOperationsRequest()

		mockPortfolioClient := mocks.NewTinkoffPortfolioClient(t)
		mockPortfolioClient.On("GetOperations", ctx, req.AccountID, req.FromDate).
			Return(nil, errors.New("client error"))

		srv := &TinkoffHelper{
			logger:    logger,
			Portfolio: mockPortfolioClient,
			now:       time.Now,
		}

		_, err := srv.TinkoffGetOperations(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed get operations from tinkoff")

		mockPortfolioClient.AssertExpectations(t)
	})
}

func TestTinkoffGetBondActions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		instrumentUid := "instr-123"
		wantBondActions := factories.NewBondIdentIdentifiers()

		mockAnalyticsClient := mocks.NewTinkoffAnalyticsClient(t)
		mockAnalyticsClient.On("GetBondsActions", ctx, instrumentUid).
			Return(wantBondActions, nil)

		srv := &TinkoffHelper{
			logger:    logger,
			Analytics: mockAnalyticsClient,
		}

		got, err := srv.TinkoffGetBondActions(ctx, instrumentUid)
		assert.NoError(t, err)
		assert.Equal(t, wantBondActions, got)

		mockAnalyticsClient.AssertExpectations(t)
	})

	t.Run("Err: Empty instrument UID", func(t *testing.T) {
		srv := &TinkoffHelper{
			logger: logger,
		}

		_, err := srv.TinkoffGetBondActions(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrEmptyInstrumentUid, err)
	})

	t.Run("Err: Analytics client fails", func(t *testing.T) {
		instrumentUid := "instr-123"

		mockAnalyticsClient := mocks.NewTinkoffAnalyticsClient(t)
		mockAnalyticsClient.On("GetBondsActions", ctx, instrumentUid).
			Return(domain.BondIdentIdentifiers{}, errors.New("client error"))

		srv := &TinkoffHelper{
			logger:    logger,
			Analytics: mockAnalyticsClient,
		}

		_, err := srv.TinkoffGetBondActions(ctx, instrumentUid)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed get bond actions from tinkoff")

		mockAnalyticsClient.AssertExpectations(t)
	})
}

func TestTinkoffGetFutureBy(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		figi := "FIGI123"
		want := factories.NewFuture()

		mockInstruments := mocks.NewTinkoffInstrumentsClient(t)
		mockInstruments.
			On("GetFutureBy", ctx, figi).
			Return(want, nil)

		srv := &TinkoffHelper{
			logger:      logger,
			Instruments: mockInstruments,
		}

		got, err := srv.TinkoffGetFutureBy(ctx, figi)

		assert.NoError(t, err)
		assert.Equal(t, want, got)

		mockInstruments.AssertExpectations(t)
	})

	t.Run("Err: empty figi", func(t *testing.T) {
		srv := &TinkoffHelper{
			logger: logger,
		}

		_, err := srv.TinkoffGetFutureBy(ctx, "")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrEmptyFigi, err)
	})

	t.Run("Err: instruments client fails", func(t *testing.T) {
		figi := "FIGI123"

		mockInstruments := mocks.NewTinkoffInstrumentsClient(t)
		mockInstruments.
			On("GetFutureBy", ctx, figi).
			Return(domain.Future{}, errors.New("client error"))

		srv := &TinkoffHelper{
			logger:      logger,
			Instruments: mockInstruments,
		}
		_, err := srv.TinkoffGetFutureBy(ctx, figi)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed get future by from tinkoff")

		mockInstruments.AssertExpectations(t)
	})
}

func TestTinkoffGetBondByUid(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		uid := "bond-uid-123"
		want := factories.NewBond()

		mockInstruments := mocks.NewTinkoffInstrumentsClient(t)
		mockInstruments.
			On("GetBondByUid", ctx, uid).
			Return(want, nil)

		srv := &TinkoffHelper{
			logger:      logger,
			Instruments: mockInstruments,
		}
		got, err := srv.TinkoffGetBondByUid(ctx, uid)

		assert.NoError(t, err)
		assert.Equal(t, want, got)

		mockInstruments.AssertExpectations(t)
	})

	t.Run("Err: empty uid", func(t *testing.T) {
		srv := &TinkoffHelper{
			logger: logger,
		}

		_, err := srv.TinkoffGetBondByUid(ctx, "")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrEmptyUid, err)
	})

	t.Run("Err: instruments client fails", func(t *testing.T) {
		uid := "bond-uid-123"

		mockInstruments := mocks.NewTinkoffInstrumentsClient(t)
		mockInstruments.
			On("GetBondByUid", ctx, uid).
			Return(domain.Bond{}, errors.New("client error"))

		srv := &TinkoffHelper{
			logger: logger,

			Instruments: mockInstruments,
		}

		_, err := srv.TinkoffGetBondByUid(ctx, uid)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed get bond by uid from tinkoff")

		mockInstruments.AssertExpectations(t)
	})
}

func TestTinkoffGetCurrencyBy(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		figi := "FIGI123"
		want := factories.NewCurrency()

		mockInstruments := mocks.NewTinkoffInstrumentsClient(t)
		mockInstruments.
			On("GetCurrencyBy", ctx, figi).
			Return(want, nil)

		srv := &TinkoffHelper{
			logger:      logger,
			Instruments: mockInstruments,
		}

		got, err := srv.TinkoffGetCurrencyBy(ctx, figi)

		assert.NoError(t, err)
		assert.Equal(t, want, got)

		mockInstruments.AssertExpectations(t)
	})

	t.Run("Err: empty figi", func(t *testing.T) {
		srv := &TinkoffHelper{
			logger: logger,
		}

		_, err := srv.TinkoffGetCurrencyBy(ctx, "")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrEmptyFigi, err)
	})

	t.Run("Err: instruments client fails", func(t *testing.T) {
		figi := "FIGI123"

		mockInstruments := mocks.NewTinkoffInstrumentsClient(t)
		mockInstruments.
			On("GetCurrencyBy", ctx, figi).
			Return(domain.Currency{}, errors.New("client error"))

		srv := &TinkoffHelper{
			logger:      logger,
			Instruments: mockInstruments,
		}

		_, err := srv.TinkoffGetCurrencyBy(ctx, figi)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed get currency by from tinkoff")

		mockInstruments.AssertExpectations(t)
	})
}

func TestService_TinkoffGetBaseShareFutureValute(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		positionUid := "pos-123"
		figi := "figi-777"

		instrument := factories.NewInstrumentShort()
		instrument.Figi = figi

		list := factories.NewInstrumentShortList(instrument)

		wantCurrency := domain.ShareCurrency{
			Currency: "USD",
		}

		mockInstruments := mocks.NewTinkoffInstrumentsClient(t)

		mockInstruments.
			On("FindBy", ctx, positionUid).
			Return(list, nil)

		mockInstruments.
			On("GetShareCurrencyBy", ctx, figi).
			Return(wantCurrency, nil)

		srv := &TinkoffHelper{
			logger:      logger,
			Instruments: mockInstruments,
		}

		got, err := srv.TinkoffGetBaseShareFutureValute(ctx, positionUid)

		assert.NoError(t, err)
		assert.Equal(t, wantCurrency, got)

		mockInstruments.AssertExpectations(t)
	})

	t.Run("empty positionUid", func(t *testing.T) {
		srv := &TinkoffHelper{
			logger: logger,
		}

		_, err := srv.TinkoffGetBaseShareFutureValute(ctx, "")

		assert.ErrorIs(t, err, domain.ErrEmptyPositionUid)
	})

	t.Run("FindBy returns error", func(t *testing.T) {
		positionUid := "pos-123"

		mockInstruments := mocks.NewTinkoffInstrumentsClient(t)
		mockInstruments.
			On("FindBy", ctx, positionUid).
			Return(domain.InstrumentShortList{}, errors.New("client error"))

		srv := &TinkoffHelper{
			logger:      logger,
			Instruments: mockInstruments,
		}

		_, err := srv.TinkoffGetBaseShareFutureValute(ctx, positionUid)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed find by from tinkoff")

		mockInstruments.AssertExpectations(t)
	})

	t.Run("ValidateAndGetFirstShare fails (empty list)", func(t *testing.T) {
		positionUid := "pos-123"

		list := factories.NewInstrumentShortList()

		mockInstruments := mocks.NewTinkoffInstrumentsClient(t)
		mockInstruments.
			On("FindBy", ctx, positionUid).
			Return(list, nil)

		srv := &TinkoffHelper{
			logger:      logger,
			Instruments: mockInstruments,
		}

		_, err := srv.TinkoffGetBaseShareFutureValute(ctx, positionUid)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to validate and get first share")

		mockInstruments.AssertExpectations(t)
	})

	t.Run("GetShareCurrencyBy returns error", func(t *testing.T) {
		positionUid := "pos-123"
		figi := "figi-999"

		instrument := factories.NewInstrumentShort()
		instrument.Figi = figi

		list := factories.NewInstrumentShortList(instrument)

		mockInstruments := mocks.NewTinkoffInstrumentsClient(t)

		mockInstruments.
			On("FindBy", ctx, positionUid).
			Return(list, nil)

		mockInstruments.
			On("GetShareCurrencyBy", ctx, figi).
			Return(domain.ShareCurrency{}, errors.New("client error"))

		srv := &TinkoffHelper{
			logger:      logger,
			Instruments: mockInstruments,
		}

		_, err := srv.TinkoffGetBaseShareFutureValute(ctx, positionUid)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed get share future valute by from tinkoff")

		mockInstruments.AssertExpectations(t)
	})
}

func TestService_TinkoffFindBy(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		query := "AAPL"

		instrument := factories.NewInstrumentShort()
		want := factories.NewInstrumentShortList(instrument)

		mockInstruments := mocks.NewTinkoffInstrumentsClient(t)

		mockInstruments.
			On("FindBy", ctx, query).
			Return(want, nil)

		srv := &TinkoffHelper{
			logger:      logger,
			Instruments: mockInstruments,
		}

		got, err := srv.TinkoffFindBy(ctx, query)

		require.NoError(t, err)
		assert.Equal(t, want, got)

		mockInstruments.AssertExpectations(t)
	})

	t.Run("empty query", func(t *testing.T) {
		srv := &TinkoffHelper{
			logger: logger,
		}

		got, err := srv.TinkoffFindBy(ctx, "")

		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrEmptyQuery)
		assert.Nil(t, got)
	})

	t.Run("client returns error", func(t *testing.T) {
		query := "AAPL"

		mockInstruments := mocks.NewTinkoffInstrumentsClient(t)

		mockInstruments.
			On("FindBy", ctx, query).
			Return(nil, errors.New("client error"))

		srv := &TinkoffHelper{
			logger:      logger,
			Instruments: mockInstruments,
		}

		got, err := srv.TinkoffFindBy(ctx, query)

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "failed find by from tinkoff")

		mockInstruments.AssertExpectations(t)
	})
}

func TestService_TinkoffGetLastPriceInPersentageToNominal(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		instrumentUid := "uid-123"

		want := factories.NewLastPrice()

		mockAnalytics := mocks.NewTinkoffAnalyticsClient(t)

		mockAnalytics.
			On("GetLastPriceInPersentageToNominal", ctx, instrumentUid).
			Return(want, nil)

		srv := &TinkoffHelper{
			logger:    logger,
			Analytics: mockAnalytics,
			now:       time.Now,
		}

		got, err := srv.TinkoffGetLastPriceInPersentageToNominal(ctx, instrumentUid)

		require.NoError(t, err)
		assert.Equal(t, want, got)

		mockAnalytics.AssertExpectations(t)
	})

	t.Run("empty instrumentUid", func(t *testing.T) {
		srv := &TinkoffHelper{
			logger: logger,
			now:    time.Now,
		}

		got, err := srv.TinkoffGetLastPriceInPersentageToNominal(ctx, "")

		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrEmptyInstrumentUid)
		assert.Equal(t, domain.LastPrice{}, got)
	})

	t.Run("analytics returns error", func(t *testing.T) {
		instrumentUid := "uid-123"

		mockAnalytics := mocks.NewTinkoffAnalyticsClient(t)

		mockAnalytics.
			On("GetLastPriceInPersentageToNominal", ctx, instrumentUid).
			Return(domain.LastPrice{}, errors.New("client error"))

		srv := &TinkoffHelper{
			logger:    logger,
			Analytics: mockAnalytics,
			now:       time.Now,
		}

		got, err := srv.TinkoffGetLastPriceInPersentageToNominal(ctx, instrumentUid)

		require.Error(t, err)
		assert.Equal(t, domain.LastPrice{}, got)
		assert.Contains(t, err.Error(), "failed get last price in persentage to nominal from tinkoff")

		mockAnalytics.AssertExpectations(t)
	})
}
