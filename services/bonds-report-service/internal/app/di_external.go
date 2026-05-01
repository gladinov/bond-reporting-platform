package app

import (
	cbr "bonds-report-service/internal/adapters/outbound/cbr/client"
	cbrtransport "bonds-report-service/internal/adapters/outbound/cbr/transport"
	moex "bonds-report-service/internal/adapters/outbound/moex/client"
	moextransport "bonds-report-service/internal/adapters/outbound/moex/transport"
	"bonds-report-service/internal/adapters/outbound/sber"
	"bonds-report-service/internal/application/ports"
	"bonds-report-service/internal/application/usecases"
	"log/slog"
)

func (d *diContainer) MoexClient() ports.MoexClient {
	if d.moexClient == nil {
		d.logger.Info("initialize Moex client")
		d.moexClient = moex.NewMoexClient(d.logger, d.MoexTransport())
	}

	return d.moexClient
}

func (d *diContainer) MoexTransport() moextransport.TransportClient {
	if d.moexTransport == nil {
		host := d.config.Clients.MoexClient.GetMoexAppAddress()
		d.logger.Info("initialize Moex transport", slog.String("address", host))
		if host == "" {
			panic("moex host is empty")
		}

		d.moexTransport = moextransport.NewTransport(d.logger, host, d.config.Timeouts.RequestTimeout)
	}

	return d.moexTransport
}

func (d *diContainer) CBRClient() ports.CbrClient {
	if d.cbrClient == nil {
		d.logger.Info("initialize CBR client")
		d.cbrClient = cbr.NewCbrClient(d.logger, d.CBRTransport())
	}

	return d.cbrClient
}

func (d *diContainer) CBRTransport() cbrtransport.TransportClient {
	if d.cbrTransport == nil {
		host := d.config.Clients.CBRClient.GetCBRAppAddress()
		d.logger.Info("initialize CBR transport", slog.String("address", host))
		if host == "" {
			panic("cbr host is empty")
		}

		d.cbrTransport = cbrtransport.NewTransport(d.logger, host, d.config.Timeouts.RequestTimeout)
	}

	return d.cbrTransport
}

func (d *diContainer) SberClient() ports.SberClient {
	if d.sberClient == nil {
		d.logger.Info("initialize Sber client", slog.String("address", d.config.SberConfigPath))

		sberClient, err := sber.NewClient(d.config.RootPath, d.config.SberConfigPath)
		if err != nil {
			d.logger.Error("could not create sber client", slog.String("error", err.Error()))
			panic(err)
		}

		d.sberClient = sberClient
	}

	return d.sberClient
}

func (d *diContainer) ExternalApis() *usecases.ExternalApis {
	if d.externalApis == nil {
		d.externalApis = usecases.NewExternalApis(d.MoexClient(), d.CBRClient(), d.SberClient())
	}

	return d.externalApis
}
