package handlers

import (
	"cbr/internal/service"
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

var (
	errInvalidRequestBody error = errors.New("invalid request body")
	errGetData            error = errors.New("could not get data")
)

type Handlers struct {
	logger  *slog.Logger
	service service.CurrencyService
}

func NewHandlers(logger *slog.Logger, srvc service.CurrencyService) *Handlers {
	return &Handlers{
		logger:  logger,
		service: srvc,
	}
}

func (h *Handlers) GetAllCurrencies(c echo.Context) error {
	const op = "handlers.GetAllCurrencies"

	ctx := c.Request().Context()

	var currencyRequest CurrencyRequest

	err := c.Bind(&currencyRequest)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}

	currencies, err := h.service.GetAllCurrencies(ctx, currencyRequest.Date)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	return c.JSON(http.StatusOK, currencies)
}
