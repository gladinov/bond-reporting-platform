package analyticsclient

import (
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/transport"
	"log/slog"
)

type AnalyticsTinkoffClient struct {
	logger    *slog.Logger
	transport transport.TransportClient
}

func NewAnalyticsTinkoffClient(logger *slog.Logger, transport transport.TransportClient) *AnalyticsTinkoffClient {
	return &AnalyticsTinkoffClient{
		logger:    logger,
		transport: transport,
	}
}
