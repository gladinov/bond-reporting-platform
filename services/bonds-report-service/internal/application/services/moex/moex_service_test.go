//go:build unit

package moexHelper

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

func getMoexHelperForTestMoex(logger *slog.Logger, Moex ports.MoexClient) *MoexHelper {
	return NewMoexHelper(logger, Moex)
}

func TestGetSpecificationsFromMoex(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()
	fixedTime := time.Date(2026, 2, 12, 0, 0, 0, 0, time.UTC)

	t.Run("Success", func(t *testing.T) {
		ticker := "test_ticker"
		want := factories.NewValuesMoex()
		mockMoex := mocks.NewMoexClient(t)

		srvs := getMoexHelperForTestMoex(logger, mockMoex)
		mockMoex.On("GetSpecifications", ctx, ticker, fixedTime).
			Return(want, nil)

		got, err := srvs.GetSpecificationsFromMoex(ctx, ticker, fixedTime)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
	t.Run("Err: Empty Ticker", func(t *testing.T) {
		ticker := ""

		mockMoex := mocks.NewMoexClient(t)
		srvs := getMoexHelperForTestMoex(logger, mockMoex)

		_, err := srvs.GetSpecificationsFromMoex(ctx, ticker, fixedTime)
		assert.Error(t, err)
	})
	t.Run("Err: GetSpecifications", func(t *testing.T) {
		ticker := "test_ticker"
		wanterr := "failed get spec"
		mockMoex := mocks.NewMoexClient(t)
		srvs := getMoexHelperForTestMoex(logger, mockMoex)
		mockMoex.On("GetSpecifications", ctx, ticker, fixedTime).
			Return(domain.ValuesMoex{}, errors.New("failed get spec"))

		_, err := srvs.GetSpecificationsFromMoex(ctx, ticker, fixedTime)
		assert.Error(t, err)
		assert.ErrorContains(t, err, wanterr)
	})
}
