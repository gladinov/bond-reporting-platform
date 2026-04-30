package app

import (
	gateway "bonds-report-service/internal/adapters/inbound/gateway"
	"bonds-report-service/internal/closer"
	config "bonds-report-service/internal/configs"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	sl "github.com/gladinov/mylogger"
	"github.com/gladinov/traceidgenerator"
)

type App struct {
	config     config.Config
	logger     *slog.Logger
	di         *diContainer
	router     http.Handler
	httpServer *http.Server
}

func New() *App {
	a := &App{}
	a.initDeps()

	return a
}

func (a *App) initDeps() {
	inits := []func(){
		a.initConfig,
		a.initLogger,
		a.initDiContainer,
		a.initTraceIDGenerator,
		a.initRouter,
		a.initHTTPServer,
	}

	for _, fn := range inits {
		fn()
	}
}

func (a *App) initConfig() {
	a.config = config.MustInitConfig()
}

func (a *App) initLogger() {
	a.logger = sl.NewLogger(a.config.Env)
}

func (a *App) initDiContainer() {
	a.di = newDIContainer(a.logger, a.config)
}

func (a *App) initTraceIDGenerator() {
	_ = traceidgenerator.Must()
}

func (a *App) initRouter() {
	handler := a.di.Handler()
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(requestTimeout(a.config.Timeouts.RequestTimeout))
	router.Use(handler.ContextHeaderTraceIdMiddleWare())
	router.Use(handler.LoggerMiddleware())
	router.Use(handler.AuthMiddleware())

	router.GET("/bondReportService/accounts", handler.GetAccountsList)
	router.GET("/bondReportService/getBondReportsByFifo", handler.GetBondReportsByFifo)
	router.GET("/bondReportService/getUSD", handler.GetUSD)
	router.GET("/bondReportService/getBondReports", handler.GetBondReports)
	router.GET("/bondReportService/getPortfolioStructure", handler.GetPortfolioStructure)
	router.GET("/bondReportService/getUnionPortfolioStructure", handler.GetUnionPortfolioStructure)
	router.GET("/bondReportService/getUnionPortfolioStructureWithSber", handler.GetUnionPortfolioStructureWithSber)

	a.router = router
}

func (a *App) initHTTPServer() {
	timeouts := a.config.Timeouts

	a.httpServer = &http.Server{
		Addr:              a.config.Clients.BondReportService.GetBondReportServiceAppAddress(),
		Handler:           a.router,
		ReadHeaderTimeout: timeouts.HTTPReadHeaderTimeout,
		WriteTimeout:      timeouts.HTTPWriteTimeout,
		ReadTimeout:       timeouts.HTTPReadTimeout,
		IdleTimeout:       timeouts.HTTPIdleTimeout,
	}
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a.logger.Info("start app",
		slog.String("env", a.config.Env),
		slog.String("bond_report_service_host", a.config.Clients.BondReportService.Host),
		slog.String("bond_report_service_port", a.config.Clients.BondReportService.Port))

	address := a.config.Clients.BondReportService.GetBondReportServiceAppAddress()
	errCh := make(chan error, 1)
	errChConsumer := make(chan error, 1)

	go func() {
		a.logger.Info("run bond-report-service", slog.String("address", address))
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	go func() {
		a.logger.InfoContext(ctx, "run kafka consumer")
		if err := a.di.Consumer().Run(ctx); err != nil {
			errChConsumer <- err
		}
	}()

	select {
	case <-ctx.Done():
		a.logger.InfoContext(ctx, "shutdown signal received")
	case err := <-errChConsumer:
		a.logger.ErrorContext(ctx, "consumer stopped with error", slog.Any("error", err))
	case err := <-errCh:
		a.logger.ErrorContext(ctx, "server stopped with error", slog.Any("error", err))
	}

	stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.config.Timeouts.HTTPShutdownTimeout)
	defer shutdownCancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("shutdown server error", slog.Any("error", err))
	}

	a.logger.Info("server stop")

	closerCtx, closerCancel := context.WithTimeout(context.Background(), a.config.Timeouts.AppCloseTimeout)
	defer closerCancel()

	if err := closer.CloseAll(closerCtx); err != nil {
		a.logger.Error("resource close error", slog.Any("error", err))
	}

	a.logger.Info("server exited gracefully")

	return nil
}

func requestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

type appHandler interface {
	ContextHeaderTraceIdMiddleWare() gin.HandlerFunc
	LoggerMiddleware() gin.HandlerFunc
	AuthMiddleware() gin.HandlerFunc
	GetAccountsList(c *gin.Context)
	GetBondReportsByFifo(c *gin.Context)
	GetUSD(c *gin.Context)
	GetBondReports(c *gin.Context)
	GetPortfolioStructure(c *gin.Context)
	GetUnionPortfolioStructure(c *gin.Context)
	GetUnionPortfolioStructureWithSber(c *gin.Context)
}

var _ appHandler = (*gateway.Handler)(nil)
