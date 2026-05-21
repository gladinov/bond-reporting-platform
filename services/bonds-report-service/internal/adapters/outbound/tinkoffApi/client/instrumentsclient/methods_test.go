//go:build unit

package instrumentsclient

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

func setupClient(t *testing.T) (*InstrumentsTinkoffClient, *mocks.TransportClient) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	transportMock := mocks.NewTransportClient(t)
	client := NewInstrumentsTinkoffClient(logger, transportMock)
	return client, transportMock
}

func TestGetFutureBy(t *testing.T) {
	ctx := context.Background()
	path := path.Join("tinkoff", "future")
	figi := "FIGI123"

	t.Run("Success", func(t *testing.T) {
		wantName := "Test Future"
		dtoResp := factories.NewFutureDTO()
		jsonData, _ := json.Marshal(dtoResp)

		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       jsonData,
			}, nil)

		res, err := client.GetFutureBy(ctx, figi)
		assert.NoError(t, err)
		assert.Equal(t, wantName, res.Name)
	})

	t.Run("DoRequest_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(nil, errors.New("network error"))

		_, err := client.GetFutureBy(ctx, figi)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DoRequest")
	})

	t.Run("HTTP_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       []byte("bad request"),
			}, nil)

		_, err := client.GetFutureBy(ctx, figi)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad request")
	})

	t.Run("JSON_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte("invalid json"),
			}, nil)

		_, err := client.GetFutureBy(ctx, figi)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})
}

func TestGetBondByUid(t *testing.T) {
	ctx := context.Background()
	path := path.Join("tinkoff", "bond")
	uid := "UID123"

	t.Run("Success", func(t *testing.T) {
		dtoResp := factories.NewBond()
		jsonData, _ := json.Marshal(dtoResp)

		client, transportMock := setupClient(t)
		transportMock.On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       jsonData,
			}, nil)

		res, err := client.GetBondByUid(ctx, uid)
		assert.NoError(t, err)
		assert.Equal(t, dtoResp.AciValue.Nano, res.AciValue.Nano)
	})
	t.Run("DoRequest_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(nil, errors.New("network error"))

		_, err := client.GetBondByUid(ctx, uid)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DoRequest")
	})

	t.Run("HTTP_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       []byte("bad request"),
			}, nil)

		_, err := client.GetBondByUid(ctx, uid)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad request")
	})

	t.Run("JSON_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte("invalid json"),
			}, nil)

		_, err := client.GetBondByUid(ctx, uid)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})
}

func TestGetCurrencyBy(t *testing.T) {
	ctx := context.Background()
	path := path.Join("tinkoff", "currency")
	isin := "USD000UTSTOM"

	t.Run("Success", func(t *testing.T) {
		dtoResp := dto.Currency{Isin: isin}
		jsonData, _ := json.Marshal(dtoResp)

		client, transportMock := setupClient(t)
		transportMock.On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       jsonData,
			}, nil)

		res, err := client.GetCurrencyBy(ctx, isin)
		assert.NoError(t, err)
		assert.Equal(t, isin, res.Isin)
	})

	t.Run("DoRequest_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(nil, errors.New("network error"))

		_, err := client.GetCurrencyBy(ctx, isin)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DoRequest")
	})

	t.Run("HTTP_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       []byte("bad request"),
			}, nil)

		_, err := client.GetCurrencyBy(ctx, isin)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad request")
	})

	t.Run("JSON_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte("invalid json"),
			}, nil)

		_, err := client.GetCurrencyBy(ctx, isin)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})
}

func TestFindBy(t *testing.T) {
	ctx := context.Background()
	path := path.Join("tinkoff", "findby")
	query := "FIGI123"

	t.Run("Success", func(t *testing.T) {
		dtoResp := []dto.InstrumentShort{factories.NewInstrumentShortTest()}
		jsonData, _ := json.Marshal(dtoResp)

		client, transportMock := setupClient(t)
		transportMock.On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       jsonData,
			}, nil)

		res, err := client.FindBy(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, query, res[0].Figi)
	})
	t.Run("DoRequest_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(nil, errors.New("network error"))

		_, err := client.FindBy(ctx, query)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DoRequest")
	})

	t.Run("HTTP_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       []byte("bad request"),
			}, nil)

		_, err := client.FindBy(ctx, query)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad request")
	})

	t.Run("JSON_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte("invalid json"),
			}, nil)

		_, err := client.FindBy(ctx, query)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})
}

func TestGetShareCurrencyBy(t *testing.T) {
	ctx := context.Background()
	figi := "FIGI123"
	path := path.Join("tinkoff", "share", "currency")

	t.Run("Success", func(t *testing.T) {
		dtoResp := dto.ShareCurrencyByResponse{
			Currency: "USD",
		}
		jsonData, _ := json.Marshal(dtoResp)

		client, transportMock := setupClient(t)

		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       jsonData,
			}, nil)

		result, err := client.GetShareCurrencyBy(ctx, figi)

		assert.NoError(t, err)
		assert.Equal(t, "USD", result.Currency)
	})

	t.Run("HTTP_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)

		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusUnauthorized,
				Body:       []byte("unauthorized"),
			}, nil)

		_, err := client.GetShareCurrencyBy(ctx, figi)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("JSON_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)

		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte("invalid json"),
			}, nil)

		_, err := client.GetShareCurrencyBy(ctx, figi)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})

	t.Run("DoRequest_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)

		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(nil, errors.New("network error"))

		_, err := client.GetShareCurrencyBy(ctx, figi)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DoRequest")
	})
}
