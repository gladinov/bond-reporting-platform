package app

import (
	"log/slog"
	moexclient "moex/internal/clients/moex"
	"moex/internal/configs"
	"moex/internal/handlers"
	"moex/internal/service"

	"github.com/labstack/echo/v4"
)

type appHandler interface {
	ContextHeaderTraceIdMiddleWare(next echo.HandlerFunc) echo.HandlerFunc
	LoggerMiddleWare(next echo.HandlerFunc) echo.HandlerFunc
	GetSpecifications(c echo.Context) error
}

type diContainer struct {
	logger    *slog.Logger
	cfg       *configs.Config
	transport moexclient.TransportClient
	client    moexclient.MoexClient
	service   service.ServiceClient
	handler   appHandler
}

func newDIContainer(logger *slog.Logger, cfg *configs.Config) *diContainer {
	return &diContainer{
		logger: logger,
		cfg:    cfg,
	}
}

func (d *diContainer) Transport() moexclient.TransportClient {
	if d.transport == nil {
		d.logger.Info("initialize MOEX transport")
		d.transport = moexclient.NewTransport(d.logger, d.cfg.MoexHost)
	}

	return d.transport
}

func (d *diContainer) Client() moexclient.MoexClient {
	if d.client == nil {
		d.logger.Info("initialize MOEX client")
		d.client = moexclient.NewMoexClient(d.logger, d.Transport())
	}

	return d.client
}

func (d *diContainer) Service() service.ServiceClient {
	if d.service == nil {
		d.logger.Info("initialize service")
		d.service = service.NewServiceClient(d.logger, d.Client())
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
