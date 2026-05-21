package app

import (
	cbrclient "cbr/internal/clients/cbr"
	"cbr/internal/configs"
	"cbr/internal/handlers"
	"cbr/internal/service"
	"cbr/internal/utils"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

type appHandler interface {
	ContextHeaderTraceIdMiddleWare(next echo.HandlerFunc) echo.HandlerFunc
	LoggerMiddleWare(next echo.HandlerFunc) echo.HandlerFunc
	GetAllCurrencies(c echo.Context) error
}

type diContainer struct {
	logger       *slog.Logger
	cfg          configs.Config
	timeLocation *time.Location
	transport    cbrclient.TransportClient
	client       cbrclient.CbrClient
	service      service.CurrencyService
	handler      appHandler
}

func newDIContainer(logger *slog.Logger, cfg configs.Config) *diContainer {
	return &diContainer{
		logger: logger,
		cfg:    cfg,
	}
}

func (d *diContainer) TimeLocation() *time.Location {
	if d.timeLocation == nil {
		d.timeLocation = utils.MustGetMoscowLocation()
	}

	return d.timeLocation
}

func (d *diContainer) Transport() cbrclient.TransportClient {
	if d.transport == nil {
		d.logger.Info("initialize cbr transport")
		d.transport = cbrclient.NewTransport(d.logger, d.cfg.CbrHost)
	}

	return d.transport
}

func (d *diContainer) Client() cbrclient.CbrClient {
	if d.client == nil {
		d.logger.Info("initialize cbr client")
		d.client = cbrclient.NewClient(d.logger, d.Transport())
	}

	return d.client
}

func (d *diContainer) Service() service.CurrencyService {
	if d.service == nil {
		d.logger.Info("initialize service")
		d.service = service.NewService(d.logger, d.Client(), d.TimeLocation())
	}

	return d.service
}

func (d *diContainer) Handler() appHandler {
	if d.handler == nil {
		d.logger.Info("initialize handlers")
		d.handler = handlers.NewHandlers(d.logger, d.Service())
	}

	return d.handler
}
