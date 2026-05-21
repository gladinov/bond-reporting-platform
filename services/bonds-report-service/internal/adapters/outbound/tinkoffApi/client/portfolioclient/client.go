package portfolioclient

import (
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/transport"
	"log/slog"
)

type PortfolioTinkoffClient struct {
	logger    *slog.Logger
	transport transport.TransportClient
}

func NewPortfolioTinkoffClient(logger *slog.Logger, transport transport.TransportClient) *PortfolioTinkoffClient {
	return &PortfolioTinkoffClient{
		logger:    logger,
		transport: transport,
	}
}
