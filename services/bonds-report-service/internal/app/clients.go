package app

import (
	cbr "bonds-report-service/internal/adapters/outbound/cbr/client"
	cbrtransport "bonds-report-service/internal/adapters/outbound/cbr/transport"
	moex "bonds-report-service/internal/adapters/outbound/moex/client"
	moextransport "bonds-report-service/internal/adapters/outbound/moex/transport"
	"bonds-report-service/internal/adapters/outbound/sber"
	"log/slog"

	config "bonds-report-service/internal/configs"
)

func InitCBRClient(logger *slog.Logger, host string) *cbr.Client {
	logger.Info("initialize CBR client", slog.String("address", host))
	if host == "" {
		panic("cbr host is empty")
	}
	transport := cbrtransport.NewTransport(logger, host)
	client := cbr.NewCbrClient(logger, transport)
	return client
}

func InitTiMoexClient(logger *slog.Logger, host string) *moex.Client {
	logger.Info("initialize Moex client", slog.String("address", host))
	if host == "" {
		panic("moex host is empty")
	}
	transport := moextransport.NewTransport(logger, host)
	client := moex.NewMoexClient(logger, transport)
	return client
}

func InitSberClient(logger *slog.Logger, conf *config.Config) (*sber.Client, error) {
	logger.Info("initialize Sber client", slog.String("address", conf.SberConfigPath))
	return sber.NewClient(conf.RootPath, conf.SberConfigPath)
}
