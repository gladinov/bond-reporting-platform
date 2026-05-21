//go:build unit

package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"telegram-mock/internal/handlers/mocks"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestLoggerMiddleware_PassesThrough(t *testing.T) {
	e := echo.New()

	srvc := mocks.NewService(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := NewHandler(logger, srvc)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		c.Response().WriteHeader(http.StatusOK)
		return nil
	}

	mw := h.LoggerMiddleWare(next)
	err := mw(c)

	require.NoError(t, err)
	require.True(t, nextCalled)
}

func TestLoggerMiddleware_Error(t *testing.T) {
	e := echo.New()

	srvc := mocks.NewService(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := NewHandler(logger, srvc)

	req := httptest.NewRequest(http.MethodGet, "/err", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}

	mw := h.LoggerMiddleWare(next)
	err := mw(c)

	require.Error(t, err)
}
