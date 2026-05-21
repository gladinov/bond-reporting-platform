package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"telegram-mock/internal/config"
	"telegram-mock/internal/handlers"
	"telegram-mock/internal/service"
	"time"

	sl "github.com/gladinov/mylogger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	sendMessageMethod    = "sendMessage"
	sendPhotoMethod      = "sendPhoto"
	sendMediaGroupMethod = "sendMediaGroup"
	getUpdates           = "getUpdates"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	conf := config.MustInitConfig()

	logg := sl.NewLogger(conf.Env)

	logg.Info("start telegram mock app",
		slog.String("env", conf.Env),
		slog.String("telegram mock_host", conf.Clients.TelegramMock.Host),
		slog.String("telegram mock_port", conf.Clients.TelegramMock.Port))

	logg.Info("initialize handlers")

	logg.Info("initialize service")
	service := service.NewService(logg)

	handler := handlers.NewHandler(logg, service)

	address := conf.Clients.TelegramMock.GetTelegramMockAddress()

	router := echo.New()
	router.Use(middleware.CORS())
	// router.Use(handler.LoggerMiddleWare)
	router.HTTPErrorHandler = handlers.HTTPErrorHandler(logg)

	getUpdatesPath := newPath(conf.Clients.TelegramMock.Token, getUpdates)
	sendMessagePath := newPath(conf.Clients.TelegramMock.Token, sendMessageMethod)
	sendPhotoPath := newPath(conf.Clients.TelegramMock.Token, sendPhotoMethod)
	sendMediaGroupPath := newPath(conf.Clients.TelegramMock.Token, sendMediaGroupMethod)

	router.GET(getUpdatesPath, handler.GetUpdates)
	router.POST("/message", handler.PostMessage, handler.LoggerMiddleWare)
	router.GET(sendMessagePath, handler.SendMessage, handler.LoggerMiddleWare)
	router.POST(sendPhotoPath, handler.SendPhoto, handler.LoggerMiddleWare)
	router.POST(sendMediaGroupPath, handler.SendMediaGroup, handler.LoggerMiddleWare)

	httpSrv := &http.Server{
		Addr:         address,
		Handler:      router,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		logg.Info("run telegram mock server", slog.String("address", address))
		if err := httpSrv.ListenAndServeTLS(conf.CertPath, conf.KeyPath); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logg.InfoContext(ctx, "Shutdown signal received")
	case err := <-errCh:
		logg.ErrorContext(ctx, "server stopped with error", slog.Any("error", err))
	}
	gracefulShutdown(ctx, logg, httpSrv)
}

func gracefulShutdown(ctx context.Context, logg *slog.Logger, httpSrv *http.Server) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		logg.ErrorContext(ctx, "Forced shutdown", slog.Any("err", err))
	}
	logg.InfoContext(shutdownCtx, "Server exited gracefully")
}

func newPath(token, method string) string {
	res := "/" + "bot" + token + "/" + method
	return res
}
