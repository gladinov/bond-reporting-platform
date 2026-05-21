package instrumentsclient

import (
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/transport"
	"log/slog"
)

type InstrumentsTinkoffClient struct {
	logger    *slog.Logger
	transport transport.TransportClient
}

func NewInstrumentsTinkoffClient(logger *slog.Logger, transport transport.TransportClient) *InstrumentsTinkoffClient {
	return &InstrumentsTinkoffClient{
		logger:    logger,
		transport: transport,
	}
}
