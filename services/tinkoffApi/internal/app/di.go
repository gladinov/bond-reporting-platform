package app

import (
	"context"
	"log/slog"
	"tinkoffApi/internal/closer"
	"tinkoffApi/internal/configs"
	"tinkoffApi/internal/handlers"
	redisClient "tinkoffApi/internal/repository/redis"
	"tinkoffApi/internal/service"
	"tinkoffApi/internal/token"
	loggeradapter "tinkoffApi/lib/logger/loggerAdapter"

	"github.com/russianinvestments/invest-api-go-sdk/investgo"

	"github.com/labstack/echo/v4"
)

type appHandler interface {
	ContextHeaderTraceIdMiddleWare(next echo.HandlerFunc) echo.HandlerFunc
	LoggerMiddleWare(next echo.HandlerFunc) echo.HandlerFunc
	CheckTokenFromRedisByChatIDMiddleWare(next echo.HandlerFunc) echo.HandlerFunc
	CheckTokenFromHeadersMiddleWare(next echo.HandlerFunc) echo.HandlerFunc
	CheckToken(c echo.Context) error
	GetAccounts(c echo.Context) error
	GetPortfolio(c echo.Context) error
	GetOperations(c echo.Context) error
	GetAllAssetUids(c echo.Context) error
	GetFutureBy(c echo.Context) error
	GetBondBy(c echo.Context) error
	GetCurrencyBy(c echo.Context) error
	GetShareCurrencyBy(c echo.Context) error
	FindBy(c echo.Context) error
	GetBondsActions(c echo.Context) error
	GetLastPriceInPersentageToNominal(c echo.Context) error
}

type diContainer struct {
	logger            *slog.Logger
	configs           *configs.Configs
	loggerAdapter     investgo.Logger
	analyticsService  service.AnalyticsService
	portfolioService  service.PortfolioService
	instrumentService service.InstrumentService
	service           *service.Service
	tokenDecrypter    handlers.TokenDecrypter
	tokenStorage      handlers.TokenStorage
	handlers          appHandler
}

func newDIContainer(logger *slog.Logger, configs *configs.Configs) *diContainer {
	return &diContainer{
		logger:  logger,
		configs: configs,
	}
}

func (d *diContainer) LoggerAdapter() investgo.Logger {
	if d.loggerAdapter == nil {
		d.logger.Info("initialize logger adapter")
		d.loggerAdapter = loggeradapter.NewLoggerAdapter(d.logger)
	}

	return d.loggerAdapter
}

func (d *diContainer) AnalyticsService() service.AnalyticsService {
	if d.analyticsService == nil {
		d.logger.Info("initialize analytics service")
		d.analyticsService = service.NewAnalyticsServiceClient(d.configs.TinkoffApiConfig, d.LoggerAdapter())
	}

	return d.analyticsService
}

func (d *diContainer) PortfolioService() service.PortfolioService {
	if d.portfolioService == nil {
		d.logger.Info("initialize portfolio service")
		d.portfolioService = service.NewPortfolioServiceClient(d.configs.TinkoffApiConfig, d.LoggerAdapter())
	}

	return d.portfolioService
}

func (d *diContainer) InstrumentService() service.InstrumentService {
	if d.instrumentService == nil {
		d.logger.Info("initialize instrument service")
		d.instrumentService = service.NewInstrumentServiceClient(d.configs.TinkoffApiConfig, d.LoggerAdapter())
	}

	return d.instrumentService
}

func (d *diContainer) Service() *service.Service {
	if d.service == nil {
		d.logger.Info("initialize service client")
		d.service = service.NewService(
			d.AnalyticsService(),
			d.PortfolioService(),
			d.InstrumentService())
	}

	return d.service
}

func (d *diContainer) TokenDecrypter() handlers.TokenDecrypter {
	if d.tokenDecrypter == nil {
		d.logger.Info("initialize token decrypter")
		d.tokenDecrypter = token.NewDecrypter(d.configs.Config.Key)
	}

	return d.tokenDecrypter
}

func (d *diContainer) TokenStorage() handlers.TokenStorage {
	if d.tokenStorage == nil {
		d.logger.Info("initialize token storage", slog.String("address", d.configs.Config.RedisHTTPServer.GetAddress()))

		tokenStorage, err := redisClient.NewTokenStorage(context.Background(), d.configs.Config)
		if err != nil {
			d.logger.Error("haven't connect with redis", slog.String("error", err.Error()))
			return d.tokenStorage
		}

		d.tokenStorage = tokenStorage
		closer.Add("redis token storage", tokenStorage.Close)
	}

	return d.tokenStorage
}

func (d *diContainer) Handlers() appHandler {
	if d.handlers == nil {
		d.logger.Info("initialize handlers")
		d.handlers = handlers.NewHandlers(d.logger, d.Service(), d.TokenDecrypter(), d.TokenStorage())
	}

	return d.handlers
}
