//go:build unit

package cbrHelper

import (
	"bonds-report-service/internal/application/ports"
	"bonds-report-service/internal/application/ports/mocks"
	factories "bonds-report-service/internal/application/testing"
	"bonds-report-service/internal/domain"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getCbrHelperForTest(logger *slog.Logger, cbrHelper ports.CbrClient, storage ports.Storage) *CbrHelper {
	return NewCbrHelper(logger, cbrHelper, storage)
}

func TestGetCurrencyFromCB(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()
	fixedDate := time.Date(2026, 2, 12, 0, 0, 0, 0, time.UTC)
	charCode := "USD"
	expectedRate := 74.5

	t.Run("Success from storage", func(t *testing.T) {
		mockStorage := mocks.NewStorage(t)
		mockCbr := mocks.NewCbrClient(t)
		srvs := getCbrHelperForTest(logger, mockCbr, mockStorage)

		mockStorage.On("GetCurrency", ctx, charCode, fixedDate).Return(expectedRate, nil)

		rate, err := srvs.GetCurrencyFromCB(ctx, charCode, fixedDate)
		assert.NoError(t, err)
		assert.Equal(t, expectedRate, rate)

		mockStorage.AssertExpectations(t)
		mockCbr.AssertExpectations(t)
	})

	t.Run("Success from CBR when storage returns ErrNoCurrency", func(t *testing.T) {
		currencies := factories.NewCurrenciesCBR()
		mockStorage := mocks.NewStorage(t)
		mockCbr := mocks.NewCbrClient(t)
		srvs := getCbrHelperForTest(logger, mockCbr, mockStorage)

		mockStorage.On("GetCurrency", ctx, charCode, fixedDate).Return(0.0, domain.ErrNoCurrency)
		mockCbr.On("GetAllCurrencies", ctx, fixedDate).Return(currencies, nil)
		mockStorage.On("SaveCurrency", ctx, currencies, fixedDate).Return(nil)

		rate, err := srvs.GetCurrencyFromCB(ctx, charCode, fixedDate)
		assert.NoError(t, err)
		assert.Equal(t, expectedRate, rate)

		mockStorage.AssertExpectations(t)
		mockCbr.AssertExpectations(t)
	})

	t.Run("Err: storage returns unexpected error", func(t *testing.T) {
		mockStorage := mocks.NewStorage(t)
		mockCbr := mocks.NewCbrClient(t)
		srvs := getCbrHelperForTest(logger, mockCbr, mockStorage)

		mockStorage.On("GetCurrency", ctx, charCode, fixedDate).Return(0.0, errors.New("db error"))

		_, err := srvs.GetCurrencyFromCB(ctx, charCode, fixedDate)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "db error")
	})

	t.Run("Err: CBR returns error", func(t *testing.T) {
		mockStorage := mocks.NewStorage(t)
		mockCbr := mocks.NewCbrClient(t)
		srvs := getCbrHelperForTest(logger, mockCbr, mockStorage)

		mockStorage.On("GetCurrency", ctx, charCode, fixedDate).Return(0.0, domain.ErrNoCurrency)
		mockCbr.On("GetAllCurrencies", ctx, fixedDate).Return(domain.CurrenciesCBR{}, errors.New("CB error"))

		_, err := srvs.GetCurrencyFromCB(ctx, charCode, fixedDate)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "CB error")
	})
	t.Run("Err: currencies not found", func(t *testing.T) {
		charCode := "fjhglk"
		wantErr := "not found"

		currencies := factories.NewCurrenciesCBR()
		mockStorage := mocks.NewStorage(t)
		mockCbr := mocks.NewCbrClient(t)
		srvs := getCbrHelperForTest(logger, mockCbr, mockStorage)

		mockStorage.On("GetCurrency", ctx, charCode, fixedDate).Return(0.0, domain.ErrNoCurrency)
		mockCbr.On("GetAllCurrencies", ctx, fixedDate).Return(currencies, nil)

		_, err := srvs.GetCurrencyFromCB(ctx, charCode, fixedDate)
		assert.Error(t, err)
		assert.ErrorContains(t, err, wantErr)
		assert.ErrorContains(t, err, charCode)

		mockStorage.AssertExpectations(t)
		mockCbr.AssertExpectations(t)
	})
	t.Run("Err: Save currency", func(t *testing.T) {
		currencies := factories.NewCurrenciesCBR()
		mockStorage := mocks.NewStorage(t)
		mockCbr := mocks.NewCbrClient(t)
		srvs := getCbrHelperForTest(logger, mockCbr, mockStorage)

		mockStorage.On("GetCurrency", ctx, charCode, fixedDate).Return(0.0, domain.ErrNoCurrency)
		mockCbr.On("GetAllCurrencies", ctx, fixedDate).Return(currencies, nil)
		mockStorage.On("SaveCurrency", ctx, currencies, fixedDate).Return(errors.New("err"))

		rate, err := srvs.GetCurrencyFromCB(ctx, charCode, fixedDate)
		assert.NoError(t, err)
		assert.Equal(t, expectedRate, rate)

		mockStorage.AssertExpectations(t)
		mockCbr.AssertExpectations(t)
	})
}
