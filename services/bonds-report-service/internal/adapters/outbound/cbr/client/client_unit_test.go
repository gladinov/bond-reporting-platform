//go:build unit

package cbr

import (
	"bonds-report-service/internal/adapters/outbound/cbr/models"
	factories "bonds-report-service/internal/adapters/outbound/cbr/testing"
	"bonds-report-service/internal/adapters/outbound/cbr/transport/mocks"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClient_GetAllCurrencies(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	ctx := context.Background()
	date := time.Now()

	t.Run("Success", func(t *testing.T) {
		currResp := factories.NewCurrenciesResponse()
		wantMock := factories.NewHTTPResponse(200, currResp)
		transportMock := mocks.NewTransportClient(t)
		transportMock.On("DoRequest",
			ctx,
			mock.MatchedBy(func(s string) bool { return s != "" }),
			mock.Anything,
			mock.Anything).
			Return(wantMock, nil)
		CbrClient := NewCbrClient(logger, transportMock)
		got, err := CbrClient.GetAllCurrencies(ctx, date)
		require.NoError(t, err)
		curr := currResp.Currencies[0]
		key := strings.ToLower(curr.CharCode)
		domainCurr, ok := got.CurrenciesMap[key]
		require.True(t, ok)

		require.Equal(t, strings.ToLower(curr.CharCode), strings.ToLower(domainCurr.CharCode))
	})

	t.Run("Transport error", func(t *testing.T) {
		transportMock := &mocks.TransportClient{}
		transportMock.On(
			"DoRequest",
			ctx,
			path.Join("cbr", "currencies"),
			url.Values{},
			mock.AnythingOfType("*bytes.Buffer"),
		).Return(nil, errors.New("network failure")).Once()

		client := NewCbrClient(logger, transportMock)
		_, err := client.GetAllCurrencies(ctx, date)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "network failure")

		transportMock.AssertExpectations(t)
	})

	t.Run("HTTP 400 Bad Request", func(t *testing.T) {
		transportMock := &mocks.TransportClient{}
		transportMock.On(
			"DoRequest",
			ctx,
			path.Join("cbr", "currencies"),
			url.Values{},
			mock.AnythingOfType("*bytes.Buffer"),
		).Return(models.NewHTTPResponse(http.StatusBadRequest, []byte(`{}`)), nil).Once()

		client := NewCbrClient(logger, transportMock)
		_, err := client.GetAllCurrencies(ctx, date)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bad request")

		transportMock.AssertExpectations(t)
	})

	t.Run("HTTP 500 Internal Server Error", func(t *testing.T) {
		transportMock := &mocks.TransportClient{}
		transportMock.On(
			"DoRequest",
			ctx,
			path.Join("cbr", "currencies"),
			url.Values{},
			mock.AnythingOfType("*bytes.Buffer"),
		).Return(models.NewHTTPResponse(http.StatusInternalServerError, []byte(`{}`)), nil).Once()

		client := NewCbrClient(logger, transportMock)
		_, err := client.GetAllCurrencies(ctx, date)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "internal server error")

		transportMock.AssertExpectations(t)
	})

	t.Run("Invalid JSON response", func(t *testing.T) {
		transportMock := &mocks.TransportClient{}
		transportMock.On(
			"DoRequest",
			ctx,
			path.Join("cbr", "currencies"),
			url.Values{},
			mock.AnythingOfType("*bytes.Buffer"),
		).Return(models.NewHTTPResponse(http.StatusOK, []byte(`invalid json`)), nil).Once()

		client := NewCbrClient(logger, transportMock)
		_, err := client.GetAllCurrencies(ctx, date)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal response")

		transportMock.AssertExpectations(t)
	})
	t.Run("Mapping error (invalid nominal)", func(t *testing.T) {
		brokenResp := factories.NewCurrenciesResponse()

		// Ломаем Nominal — strconv.Atoi гарантированно упадёт
		brokenResp.Currencies[0].Nominal = "not-a-number"

		httpResp := factories.NewHTTPResponse(http.StatusOK, brokenResp)

		transportMock := &mocks.TransportClient{}
		transportMock.On(
			"DoRequest",
			ctx,
			path.Join("cbr", "currencies"),
			url.Values{},
			mock.AnythingOfType("*bytes.Buffer"),
		).Return(httpResp, nil).Once()

		client := NewCbrClient(logger, transportMock)

		_, err := client.GetAllCurrencies(ctx, date)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "failed map currencies response to domain")

		transportMock.AssertExpectations(t)
	})

	t.Run("Mapping error (invalid date)", func(t *testing.T) {
		brokenResp := factories.NewCurrenciesResponse()

		// Ломаем дату: не соответствует layout "02.01.2006"
		brokenResp.Date = "2026-02-05"

		httpResp := factories.NewHTTPResponse(http.StatusOK, brokenResp)

		transportMock := &mocks.TransportClient{}
		transportMock.On(
			"DoRequest",
			ctx,
			path.Join("cbr", "currencies"),
			url.Values{},
			mock.AnythingOfType("*bytes.Buffer"),
		).Return(httpResp, nil).Once()

		client := NewCbrClient(logger, transportMock)

		_, err := client.GetAllCurrencies(ctx, date)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "failed map currencies response to domain")

		transportMock.AssertExpectations(t)
	})

	t.Run("Mapping error (invalid float)", func(t *testing.T) {
		brokenResp := factories.NewCurrenciesResponse()

		// Ломаем float: strconv.ParseFloat гарантированно упадёт
		brokenResp.Currencies[0].Value = "12,3,4"

		httpResp := factories.NewHTTPResponse(http.StatusOK, brokenResp)

		transportMock := &mocks.TransportClient{}
		transportMock.On(
			"DoRequest",
			ctx,
			path.Join("cbr", "currencies"),
			url.Values{},
			mock.AnythingOfType("*bytes.Buffer"),
		).Return(httpResp, nil).Once()

		client := NewCbrClient(logger, transportMock)

		_, err := client.GetAllCurrencies(ctx, date)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "failed map currencies response to domain")

		transportMock.AssertExpectations(t)
	})
}
