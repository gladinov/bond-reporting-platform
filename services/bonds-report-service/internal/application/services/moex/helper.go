package moexHelper

import (
	"bonds-report-service/internal/application/ports"
	"log/slog"
	"time"
)

type MoexHelper struct {
	logger *slog.Logger
	Moex   ports.MoexClient
	now    func() time.Time
}

func NewMoexHelper(
	logger *slog.Logger,
	Moex ports.MoexClient,
) *MoexHelper {
	return &MoexHelper{
		logger: logger,
		Moex:   Moex,
		now:    time.Now,
	}
}
