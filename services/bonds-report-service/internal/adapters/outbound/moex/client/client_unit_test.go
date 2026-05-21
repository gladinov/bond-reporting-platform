//go:build unit

package moex

import (
	"bonds-report-service/internal/adapters/outbound/moex/dto"
	"bonds-report-service/internal/adapters/outbound/moex/models"
	"bonds-report-service/internal/adapters/outbound/moex/transport/mocks"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClient_GetSpecifications(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	ctx := context.Background()
	ticker := "TEST"
	date := time.Now()

	t.Run("Success", func(t *testing.T) {
		// создаем объект Values с данными
		wantValues := dto.Values{
			ShortName: dto.NullString{
				Value:  "Test Bond",
				IsSet:  true,
				IsNull: false,
			},
			YieldToMaturity: dto.NullFloat64{
				Value:  5.5,
				IsSet:  true,
				IsNull: false,
			},
		}
		body, _ := json.Marshal(wantValues)
		httpResp := models.NewHTTPResponse(200, body)

		transportMock := mocks.NewTransportClient(t)
		transportMock.On("DoRequest",
			ctx,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("url.Values"),
			mock.Anything,
		).Return(httpResp, nil)

		client := NewMoexClient(logger, transportMock)
		got, err := client.GetSpecifications(ctx, ticker, date)

		require.NoError(t, err)
		require.Equal(t, wantValues.ShortName.Value, got.ShortName.Value)
		require.Equal(t, wantValues.YieldToMaturity.Value, got.YieldToMaturity.Value)
	})

	t.Run("HTTP 400 Bad Request", func(t *testing.T) {
		httpResp := models.NewHTTPResponse(400, []byte(`bad request`))
		transportMock := mocks.NewTransportClient(t)
		transportMock.On("DoRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(httpResp, nil)

		client := NewMoexClient(logger, transportMock)
		_, err := client.GetSpecifications(ctx, ticker, date)

		require.Error(t, err)
		require.Contains(t, err.Error(), "bad request")
	})

	t.Run("HTTP 500 Internal Server Error", func(t *testing.T) {
		httpResp := models.NewHTTPResponse(500, []byte(`internal error`))
		transportMock := mocks.NewTransportClient(t)
		transportMock.On("DoRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(httpResp, nil)

		client := NewMoexClient(logger, transportMock)
		_, err := client.GetSpecifications(ctx, ticker, date)

		require.Error(t, err)
		require.Contains(t, err.Error(), "internal server error")
	})

	t.Run("DoRequest network error", func(t *testing.T) {
		transportMock := mocks.NewTransportClient(t)
		transportMock.On("DoRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("network unreachable"))

		client := NewMoexClient(logger, transportMock)
		_, err := client.GetSpecifications(ctx, ticker, date)

		require.Error(t, err)
		require.Contains(t, err.Error(), "network unreachable")
	})

	t.Run("JSON unmarshal error", func(t *testing.T) {
		httpResp := models.NewHTTPResponse(200, []byte(`{invalid json}`))
		transportMock := mocks.NewTransportClient(t)
		transportMock.On("DoRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(httpResp, nil)

		client := NewMoexClient(logger, transportMock)
		_, err := client.GetSpecifications(ctx, ticker, date)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to unmarshal")
	})
}
