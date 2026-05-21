package app

import (
	"cbr/internal/configs"
	"cbr/internal/handlers"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	sl "github.com/gladinov/mylogger"
	"github.com/gladinov/traceidgenerator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type App struct {
	config     configs.Config
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
	a.config = configs.MustInitConfig()
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
	router := echo.New()
	handler := a.di.Handler()

	router.Use(middleware.CORS())
	router.Use(middleware.Recover())
	router.Use(middleware.ContextTimeout(a.config.Timeouts.RequestTimeout))
	router.Use(handler.ContextHeaderTraceIdMiddleWare)
	router.Use(handler.LoggerMiddleWare)
	router.HTTPErrorHandler = handlers.HTTPErrorHandler(a.logger)

	router.POST("/cbr/currencies", handler.GetAllCurrencies)

	a.router = router
}

func (a *App) initHTTPServer() {
	timeouts := a.config.Timeouts

	a.httpServer = &http.Server{
		Addr:              a.config.Clients.CbrAppApiClient.GetCbrAppServer(),
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
		slog.String("host", a.config.CbrHost),
		slog.String("cbr_app_host", a.config.Clients.CbrAppApiClient.Host),
		slog.String("cbr_app_port", a.config.Clients.CbrAppApiClient.Port))

	address := a.config.Clients.CbrAppApiClient.GetCbrAppServer()
	errCh := make(chan error, 1)

	go func() {
		a.logger.Info("run cbr server", slog.String("address", address))
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

	// First Ctrl+C starts graceful shutdown, second interrupts immediately.
	stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.config.Timeouts.HTTPShutdownTimeout)
	defer shutdownCancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		a.logger.ErrorContext(ctx, "forced shutdown", slog.Any("error", err))
		return err
	}

	a.logger.InfoContext(ctx, "server exited gracefully")

	return nil
}
