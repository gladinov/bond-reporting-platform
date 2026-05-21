//go:build unit

package portfolioclient

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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupClient(t *testing.T) (*PortfolioTinkoffClient, *mocks.TransportClient) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	transportMock := mocks.NewTransportClient(t)
	client := NewPortfolioTinkoffClient(logger, transportMock)
	return client, transportMock
}

func TestGetAccounts(t *testing.T) {
	ctx := context.Background()
	path := path.Join("tinkoff", "accounts")

	t.Run("Success", func(t *testing.T) {
		dtoResp := map[string]dto.Account{
			"acc1": {
				ID:   "acc1",
				Type: "broker",
				Name: "Main",
			},
		}

		jsonData, _ := json.Marshal(dtoResp)

		client, transportMock := setupClient(t)

		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, nil).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       jsonData,
			}, nil)

		result, err := client.GetAccounts(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "acc1", result["acc1"].ID)
		assert.Equal(t, "broker", result["acc1"].Type)
		assert.Equal(t, "Main", result["acc1"].Name)
	})

	t.Run("HTTP_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)

		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, nil).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusUnauthorized,
				Body:       []byte("unauthorized"),
			}, nil)

		_, err := client.GetAccounts(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("JSON_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)

		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, nil).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte("invalid json"),
			}, nil)

		_, err := client.GetAccounts(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})

	t.Run("DoRequest_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)

		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, nil).
			Return(nil, errors.New("network error"))

		_, err := client.GetAccounts(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DoRequest")
	})
}

func TestGetPortfolio(t *testing.T) {
	ctx := context.Background()
	path := path.Join("tinkoff", "portfolio")

	accountID := "acc123"
	accountStatus := int64(1)

	t.Run("Success", func(t *testing.T) {
		dtoResp := factories.NewPortfolio()

		jsonResp, _ := json.Marshal(dtoResp)

		client, transportMock := setupClient(t)

		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       jsonResp,
			}, nil)

		result, err := client.GetPortfolio(ctx, accountID, accountStatus)
		want := MapPortfolioToDomain(dtoResp)

		assert.NoError(t, err)
		assert.Equal(t, want, result)
	})

	t.Run("HTTP_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)

		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusForbidden,
				Body:       []byte("forbidden"),
			}, nil)

		_, err := client.GetPortfolio(ctx, accountID, accountStatus)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "forbidden")
	})

	t.Run("JSON_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)

		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte("invalid json"),
			}, nil)

		_, err := client.GetPortfolio(ctx, accountID, accountStatus)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})

	t.Run("DoRequest_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)

		transportMock.
			On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(nil, errors.New("network error"))

		_, err := client.GetPortfolio(ctx, accountID, accountStatus)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DoRequest")
	})
}

func TestGetOperations(t *testing.T) {
	ctx := context.Background()
	accountID := "ACC123"
	date := time.Now()
	path := path.Join("tinkoff", "operations")

	t.Run("Success", func(t *testing.T) {
		want := factories.NewOperationDTO()
		respBody := []dto.Operation{want}
		jsonData, _ := json.Marshal(respBody)

		client, transportMock := setupClient(t)
		transportMock.On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       jsonData,
			}, nil)

		result, err := client.GetOperations(ctx, accountID, date)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, want.OperationID, result[0].OperationID)
	})

	t.Run("HTTP_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusUnauthorized,
				Body:       []byte("unauthorized"),
			}, nil)

		_, err := client.GetOperations(ctx, accountID, date)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "unauthorized")
	})

	t.Run("JSON_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(&models.HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte("invalid json"),
			}, nil)

		_, err := client.GetOperations(ctx, accountID, date)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})

	t.Run("DoRequest_Error", func(t *testing.T) {
		client, transportMock := setupClient(t)
		transportMock.On("DoRequest", mock.Anything, path, mock.Anything, mock.Anything).
			Return(nil, errors.New("network error"))

		_, err := client.GetOperations(ctx, accountID, date)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DoRequest")
	})
}
