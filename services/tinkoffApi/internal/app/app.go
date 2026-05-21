package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"tinkoffApi/internal/closer"
	"tinkoffApi/internal/configs"

	sl "github.com/gladinov/mylogger"
	"github.com/gladinov/traceidgenerator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type App struct {
	configs    *configs.Configs
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
		a.initConfigs,
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

func (a *App) initConfigs() {
	a.configs = configs.MustInitConfigs()
}

func (a *App) initLogger() {
	a.logger = sl.NewLogger(a.configs.Config.Env)
}

func (a *App) initDiContainer() {
	a.di = newDIContainer(a.logger, a.configs)
}

func (a *App) initTraceIDGenerator() {
	_ = traceidgenerator.Must()
}

func (a *App) initRouter() {
	router := echo.New()
	handlers := a.di.Handlers()

	router.Use(middleware.CORS())
	router.Use(middleware.Recover())
	router.Use(middleware.ContextTimeout(a.configs.Config.Timeouts.RequestTimeout))
	router.Use(handlers.ContextHeaderTraceIdMiddleWare)
	router.Use(handlers.LoggerMiddleWare)
	router.Use(handlers.CheckTokenFromRedisByChatIDMiddleWare)

	router.GET("/tinkoff/checktoken", handlers.CheckToken, handlers.CheckTokenFromHeadersMiddleWare)
	router.GET("/tinkoff/accounts", handlers.GetAccounts)
	router.POST("/tinkoff/portfolio", handlers.GetPortfolio)
	router.POST("/tinkoff/operations", handlers.GetOperations)
	router.GET("/tinkoff/allassetsuid", handlers.GetAllAssetUids)
	router.POST("/tinkoff/future", handlers.GetFutureBy)
	router.POST("/tinkoff/bond", handlers.GetBondBy)
	router.POST("/tinkoff/currency", handlers.GetCurrencyBy)
	router.POST("/tinkoff/share/currency", handlers.GetShareCurrencyBy)
	router.POST("/tinkoff/findby", handlers.FindBy)
	router.POST("/tinkoff/bondactions", handlers.GetBondsActions)
	router.POST("/tinkoff/lastprice", handlers.GetLastPriceInPersentageToNominal)

	a.router = router
}

func (a *App) initHTTPServer() {
	timeouts := a.configs.Config.Timeouts

	a.httpServer = &http.Server{
		Addr:              a.configs.Config.GetTinkoffAppAddress(),
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
		slog.String("env", a.configs.Config.Env),
		slog.String("tinkoff_app_host", a.configs.Config.TinkoffApiAppHost),
		slog.String("tinkoff_app_port", a.configs.Config.TinkoffApiAppPort))

	address := a.configs.Config.GetTinkoffAppAddress()
	errCh := make(chan error, 1)

	go func() {
		a.logger.Info("run tinkoff api", slog.String("address", address))
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		a.logger.InfoContext(ctx, "shutdown signal received")
	case err := <-errCh:
		a.logger.ErrorContext(ctx, "server stopped with error", slog.Any("error", err))
	}

	stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.configs.Config.Timeouts.HTTPShutdownTimeout)
	defer shutdownCancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("shutdown server error", slog.Any("error", err))
	}

	a.logger.Info("server stop")

	closerCtx, closerCancel := context.WithTimeout(context.Background(), a.configs.Config.Timeouts.AppCloseTimeout)
	defer closerCancel()

	if err := closer.CloseAll(closerCtx); err != nil {
		a.logger.Error("resource close error", slog.Any("error", err))
	}

	a.logger.Info("server exited gracefully")

	return nil
}
