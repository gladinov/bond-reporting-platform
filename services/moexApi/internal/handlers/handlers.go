package handlers

import (
	"errors"
	"log/slog"
	"moex/internal/models"
	"moex/internal/service"
	"net/http"

	"github.com/gladinov/e"
	"github.com/labstack/echo/v4"
)

var (
	errGetData            error = errors.New("could not get data")
	errInvalidRequestBody error = errors.New("invalid request body")
)

type Handlers struct {
	logger  *slog.Logger
	service service.ServiceClient
}

func NewHandlers(logger *slog.Logger, service service.ServiceClient) *Handlers {
	return &Handlers{
		logger:  logger,
		service: service,
	}
}

func (h *Handlers) GetSpecifications(c echo.Context) error {
	const op = "handlers.GetSpecifications"
	ctx := c.Request().Context()

	logg := h.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")

	var req models.SpecificationsRequest
	if err := c.Bind(&req); err != nil {
		return newHTTPError(http.StatusBadRequest, errInvalidRequestBody, err)
	}

	if err := validateRequest(req); err != nil {
		return newHTTPError(http.StatusBadRequest, errInvalidRequestBody, err)
	}

	resp, err := h.service.GetSpecifications(ctx, req)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, errGetData, err)
	}

	return c.JSON(http.StatusOK, resp)
}

func validateRequest(req models.SpecificationsRequest) error {
	const op = "handlers.requestValidate"
	if req.Date.IsZero() {
		return errors.New("date is required")
	}
	if req.Ticker == "" {
		return errors.New("ticker is required")
	}
	return nil
}

func newHTTPError(code int, public error, cause error) *echo.HTTPError {
	return echo.NewHTTPError(code, public).
		SetInternal(e.WrapIfErr(public.Error(), cause))
}
