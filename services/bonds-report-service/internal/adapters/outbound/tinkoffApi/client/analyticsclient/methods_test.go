//go:build unit

package analyticsclient

import (
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/dto"
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/models"
	factories "bonds-report-service/internal/adapters/outbound/tinkoffApi/testing"
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/transport/mocks"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupClient(t *testing.T) (*AnalyticsTinkoffClient, *mocks.TransportClient) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	transportMock := mocks.NewTransportClient(t)
	client := NewAnalyticsTinkoffClient(logger, transportMock)
	return client, transportMock
}

func TestGetAllAssetUids(t *testing.T) {
	ctx := context.Background()
	path := path.Join("tinkoff", "allassetsuid")

	t.Run("Success", func(t *testing.T) {
		respBody := map[string]string{"uid1": "asset1", "uid2": "asset2"}
		jsonData, _ := json.Marshal(respBody)
		client, transportMock := setupClient(t)
		transportMock.On("DoRequest", mock.Anything, path, mock.Anything, nil).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       jsonData,
			}, nil)

		result, err := client.GetAllAssetUids(ctx)
		assert.NoError(t, err)
		assert.Equal(t, respBody, result)
	})
	t.Run("HTTP_Error", func(t *testing.T) {
		want := "unauthorized"
		client, transportMock := setupClient(t)
		transportMock.On("DoRequest", mock.Anything, path, mock.Anything, nil).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusUnauthorized,
				Body:       []byte("unauthorized"),
			}, nil)

		_, err := client.GetAllAssetUids(ctx)
		assert.Error(t, err)
		assert.ErrorContains(t, err, want)
	})
	t.Run("JSON_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.On("DoRequest", mock.Anything, path, mock.Anything, nil).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte("invalid json"),
			}, nil)

		_, err := client.GetAllAssetUids(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})
	t.Run("JSON_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.On("DoRequest", mock.Anything, path, mock.Anything, nil).
			Return(nil, errors.New("mock error"))

		_, err := client.GetAllAssetUids(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock error")
		assert.Contains(t, err.Error(), "failed transport DoRequest")
	})
}

func TestGetBondsActions(t *testing.T) {
	ctx := context.Background()
	path := path.Join("tinkoff", "bondactions")
	instrumentUid := "bond-uid-123"

	t.Run("Success", func(t *testing.T) {
		dto := dto.BondIdentIdentifiers{
			Ticker:    "TICK",
			ClassCode: "TQBR",
			Name:      "Bond Name",
			Nominal: dto.MoneyValue{
				Currency: "RUB",
				Units:    1000,
				Nano:     0,
			},
			NominalCurrency: "RUB",
			Replaced:        false,
		}

		jsonData, _ := json.Marshal(dto)

		client, transportMock := setupClient(t)
		transportMock.On(
			"DoRequest",
			mock.Anything,
			path,
			mock.Anything,
			mock.Anything,
		).Return(&models.HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       jsonData,
		}, nil)

		result, err := client.GetBondsActions(ctx, instrumentUid)
		assert.NoError(t, err)

		assert.Equal(t, dto.Ticker, result.Ticker)
		assert.Equal(t, dto.ClassCode, result.ClassCode)
		assert.Equal(t, dto.Name, result.Name)
		assert.Equal(t, dto.NominalCurrency, result.NominalCurrency)
		assert.Equal(t, dto.Replaced, result.Replaced)
	})

	t.Run("HTTP_Error", func(t *testing.T) {
		want := "unauthorized"

		client, transportMock := setupClient(t)
		transportMock.On(
			"DoRequest",
			mock.Anything,
			path,
			mock.Anything,
			mock.Anything,
		).Return(&models.HTTPResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       []byte("unauthorized"),
		}, nil)

		_, err := client.GetBondsActions(ctx, instrumentUid)
		assert.Error(t, err)
		assert.ErrorContains(t, err, want)
	})

	t.Run("JSON_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.On(
			"DoRequest",
			mock.Anything,
			path,
			mock.Anything,
			mock.Anything,
		).Return(&models.HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       []byte("invalid json"),
		}, nil)

		_, err := client.GetBondsActions(ctx, instrumentUid)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})

	t.Run("DoRequest_Error", func(t *testing.T) {
		want := "failed transport DoRequest"

		client, transportMock := setupClient(t)
		transportMock.On(
			"DoRequest",
			mock.Anything,
			path,
			mock.Anything,
			mock.Anything,
		).Return(nil, errors.New("network error"))

		_, err := client.GetBondsActions(ctx, instrumentUid)
		assert.Error(t, err)
		assert.ErrorContains(t, err, want)
	})
}

func TestGetLastPriceInPersentageToNominal(t *testing.T) {
	ctx := context.Background()
	instrumentUid := "test-uid"
	path := path.Join("tinkoff", "lastprice")

	t.Run("Success", func(t *testing.T) {
		wantlastprice := 102.45
		respBody := factories.NewTestLastPriceDto(wantlastprice)
		jsonData, _ := json.Marshal(respBody)

		want := factories.NewTestLastPrice(wantlastprice)

		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       jsonData,
			}, nil)

		result, err := client.GetLastPriceInPersentageToNominal(ctx, instrumentUid)
		assert.NoError(t, err)
		assert.Equal(t, want, result)
	})

	t.Run("HTTP_Error", func(t *testing.T) {
		want := "unauthorized"

		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusUnauthorized,
				Body:       []byte("unauthorized"),
			}, nil)

		_, err := client.GetLastPriceInPersentageToNominal(ctx, instrumentUid)
		assert.Error(t, err)
		assert.ErrorContains(t, err, want)
	})

	t.Run("JSON_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte("invalid json"),
			}, nil)

		_, err := client.GetLastPriceInPersentageToNominal(ctx, instrumentUid)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})

	t.Run("DoRequest_Error", func(t *testing.T) {
		want := "failed transport DoRequest"

		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(nil, errors.New("network error"))

		_, err := client.GetLastPriceInPersentageToNominal(ctx, instrumentUid)
		assert.Error(t, err)
		assert.ErrorContains(t, err, want)
	})
}
