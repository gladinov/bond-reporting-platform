package app

import (
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/client/analyticsclient"
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/client/instrumentsclient"
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/client/portfolioclient"
	tinkofftransport "bonds-report-service/internal/adapters/outbound/tinkoffApi/transport"
	"bonds-report-service/internal/application/ports"
	tinkoffHelper "bonds-report-service/internal/application/services/tinkoff"
	"log/slog"
)

func (d *diContainer) TinkoffHelper() *tinkoffHelper.TinkoffHelper {
	if d.tinkoffHelper == nil {
		d.logger.Info("initialize Tinkoff helper")
		d.tinkoffHelper = tinkoffHelper.NewTinkoffHelper(
			d.logger,
			d.TinkoffInstrumentsClient(),
			d.TinkoffPortfolioClient(),
			d.TinkoffAnalyticsClient(),
		)
	}

	return d.tinkoffHelper
}

func (d *diContainer) TinkoffTransport() tinkofftransport.TransportClient {
	if d.tinkoffTransport == nil {
		host := d.config.Clients.TinkoffClient.GetTinkoffApiAddress()
		d.logger.Info("initialize Tinkoff transport", slog.String("address", host))
		if host == "" {
			panic("tinkoff host is empty")
		}

		d.tinkoffTransport = tinkofftransport.NewTransport(d.logger, host, d.config.Timeouts.RequestTimeout)
	}

	return d.tinkoffTransport
}

func (d *diContainer) TinkoffAnalyticsClient() ports.TinkoffAnalyticsClient {
	if d.tinkoffAnalyticsClient == nil {
		d.logger.Info("initialize Tinkoff analytics client")
		d.tinkoffAnalyticsClient = analyticsclient.NewAnalyticsTinkoffClient(d.logger, d.TinkoffTransport())
	}

	return d.tinkoffAnalyticsClient
}

func (d *diContainer) TinkoffInstrumentsClient() ports.TinkoffInstrumentsClient {
	if d.tinkoffInstrumentsClient == nil {
		d.logger.Info("initialize Tinkoff instruments client")
		d.tinkoffInstrumentsClient = instrumentsclient.NewInstrumentsTinkoffClient(d.logger, d.TinkoffTransport())
	}

	return d.tinkoffInstrumentsClient
}

func (d *diContainer) TinkoffPortfolioClient() ports.TinkoffPortfolioClient {
	if d.tinkoffPortfolioClient == nil {
		d.logger.Info("initialize Tinkoff portfolio client")
		d.tinkoffPortfolioClient = portfolioclient.NewPortfolioTinkoffClient(d.logger, d.TinkoffTransport())
	}

	return d.tinkoffPortfolioClient
}
