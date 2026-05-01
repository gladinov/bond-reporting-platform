package tinkoffHelper

import (
	"bonds-report-service/internal/application/ports"
	"log/slog"
	"time"
)

type TinkoffHelper struct {
	logger      *slog.Logger
	Instruments ports.TinkoffInstrumentsClient
	Portfolio   ports.TinkoffPortfolioClient
	Analytics   ports.TinkoffAnalyticsClient
	now         func() time.Time
}

func NewTinkoffHelper(
	logger *slog.Logger,
	instruments ports.TinkoffInstrumentsClient,
	portfolio ports.TinkoffPortfolioClient,
	analytics ports.TinkoffAnalyticsClient,
) *TinkoffHelper {
	return &TinkoffHelper{
		logger:      logger,
		Instruments: instruments,
		Portfolio:   portfolio,
		Analytics:   analytics,
		now:         time.Now,
	}
}
