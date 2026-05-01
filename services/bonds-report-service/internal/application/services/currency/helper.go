package cbrHelper

import (
	"bonds-report-service/internal/application/ports"
	"log/slog"
)

type CbrHelper struct {
	logger  *slog.Logger
	Cbr     ports.CbrClient
	Storage ports.Storage
}

func NewCbrHelper(
	logger *slog.Logger,
	cbr ports.CbrClient,
	storage ports.Storage,
) *CbrHelper {
	return &CbrHelper{
		logger:  logger,
		Cbr:     cbr,
		Storage: storage,
	}
}
