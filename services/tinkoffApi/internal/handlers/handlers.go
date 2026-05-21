package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"tinkoffApi/internal/service"

	"github.com/labstack/echo/v4"
)

type TokenDecrypter interface {
	DecryptTokenFromBase64(tokenInBase64 string) (string, error)
}

type TokenStorage interface {
	GetToken(ctx context.Context, chatID string) (string, error)
}

type Handlers struct {
	logger         *slog.Logger
	service        *service.Service
	tokenDecrypter TokenDecrypter
	tokenStorage   TokenStorage
}

func NewHandlers(logger *slog.Logger, service *service.Service, tokenDecrypter TokenDecrypter, tokenStorage TokenStorage) *Handlers {
	return &Handlers{
		logger:         logger,
		service:        service,
		tokenDecrypter: tokenDecrypter,
		tokenStorage:   tokenStorage,
	}
}

var (
	ErrTokenNotFound      error = errors.New("token not found")
	errHeaderRequired     error = errors.New("header auth requierd")
	errInvalidAuthFormat  error = errors.New("invalid Authorization format, expected: Bearer <token>")
	errEmptyToken         error = errors.New("empty token")
	errIncorrectToken     error = errors.New("incorrect token")
	errNoTokenInRedis     error = errors.New("token not found for the provided user id")
	errRedisDoNotAnswer   error = errors.New("token storage temporarily unavailable")
	errGetData            error = errors.New("could not get data")
	errInvalidRequestBody error = errors.New("invalid request body")
)

func (h *Handlers) CheckToken(c echo.Context) (err error) {
	const op = "handlers.CheckToken"

	logg := h.logger.With(slog.String("op", op))
	logg.Debug("start")

	ctx := c.Request().Context()

	client, err := h.service.PortfolioService.GetClient(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errIncorrectToken)
	}
	defer func() {
		if err := client.Stop(); err != nil {
			logg.WarnContext(ctx, "client stop failed", slog.Any("error", err))
		}
	}()
	_, err = h.service.PortfolioService.GetAccounts(client)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) GetAccounts(c echo.Context) (err error) {
	const op = "handlers.GetAccount"
	ctx := c.Request().Context()

	logg := h.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")

	client, err := h.service.PortfolioService.GetClient(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errIncorrectToken)
	}

	defer func() {
		if err := client.Stop(); err != nil {
			logg.WarnContext(ctx, "client stop failed", slog.Any("error", err))
		}
	}()

	accs, err := h.service.PortfolioService.GetAccounts(client)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	return c.JSON(http.StatusOK, accs)
}

func (h *Handlers) GetPortfolio(c echo.Context) (err error) {
	const op = "handlers.GetPortfolio"
	ctx := c.Request().Context()

	logg := h.logger.With(slog.String("op", op))
	logg.Debug("start")

	var portffolioReq service.PortfolioRequest
	err = c.Bind(&portffolioReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}

	client, err := h.service.PortfolioService.GetClient(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errIncorrectToken)
	}

	defer func() {
		if err := client.Stop(); err != nil {
			logg.WarnContext(ctx, "client stop failed", slog.Any("error", err))
		}
	}()

	portf, err := h.service.PortfolioService.GetPortfolio(client, portffolioReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	return c.JSON(http.StatusOK, portf)
}

func (h *Handlers) GetOperations(c echo.Context) error {
	const op = "handlers.GetOperations"
	ctx := c.Request().Context()

	logg := h.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")

	var operationReq service.OperationsRequest
	err := c.Bind(&operationReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}

	client, err := h.service.PortfolioService.GetClient(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errIncorrectToken)
	}

	defer func() {
		if err := client.Stop(); err != nil {
			logg.WarnContext(ctx, "client stop failed", slog.Any("error", err))
		}
	}()

	operations, err := h.service.PortfolioService.MakeSafeGetOperationsRequest(client, operationReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	return c.JSON(http.StatusOK, operations)
}

func (h *Handlers) GetAllAssetUids(c echo.Context) error {
	const op = "handlers.GetAllAssetUids"
	ctx := c.Request().Context()

	logg := h.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")

	client, err := h.service.AnalyticsService.GetClient(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errIncorrectToken)
	}
	defer func() {
		if err := client.Stop(); err != nil {
			logg.WarnContext(ctx, "client stop failed", slog.Any("error", err))
		}
	}()

	allAssetUids, err := h.service.AnalyticsService.GetAllAssetUids(client)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	return c.JSON(http.StatusOK, allAssetUids)
}

func (h *Handlers) GetFutureBy(c echo.Context) error {
	const op = "handlers.GetFutureBy"

	ctx := c.Request().Context()

	logg := h.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")

	var body service.FutureReq
	err := c.Bind(&body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}

	client, err := h.service.InstrumentService.GetClient(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errIncorrectToken)
	}
	defer func() {
		if err := client.Stop(); err != nil {
			logg.WarnContext(ctx, "client stop failed", slog.Any("error", err))
		}
	}()

	future, err := h.service.InstrumentService.GetFutureBy(client, body.Figi)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	return c.JSON(http.StatusOK, future)
}

func (h *Handlers) GetBondBy(c echo.Context) error {
	const op = "handlers.GetBondBy"

	ctx := c.Request().Context()

	logg := h.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")

	var body service.BondReq
	err := c.Bind(&body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}

	client, err := h.service.InstrumentService.GetClient(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errIncorrectToken)
	}
	defer func() {
		if err := client.Stop(); err != nil {
			logg.WarnContext(ctx, "client stop failed", slog.Any("error", err))
		}
	}()

	bond, err := h.service.InstrumentService.GetBondByUid(client, body.Uid)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	return c.JSON(http.StatusOK, bond)
}

func (h *Handlers) GetCurrencyBy(c echo.Context) error {
	const op = "handlers.GetCurrencyBy"

	ctx := c.Request().Context()

	logg := h.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")

	var body service.CurrencyReq
	err := c.Bind(&body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}

	client, err := h.service.InstrumentService.GetClient(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errIncorrectToken)
	}
	defer func() {
		if err := client.Stop(); err != nil {
			logg.WarnContext(ctx, "client stop failed", slog.Any("error", err))
		}
	}()

	currency, err := h.service.InstrumentService.GetCurrencyBy(client, body.Figi)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	return c.JSON(http.StatusOK, currency)
}

func (h *Handlers) GetShareCurrencyBy(c echo.Context) error {
	const op = "handlers.GetShareCurrencyBy"

	ctx := c.Request().Context()

	logg := h.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")

	var body service.ShareCurrencyByRequest
	err := c.Bind(&body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}

	client, err := h.service.InstrumentService.GetClient(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errIncorrectToken)
	}
	defer func() {
		if err := client.Stop(); err != nil {
			logg.WarnContext(ctx, "client stop failed", slog.Any("error", err))
		}
	}()

	currency, err := h.service.InstrumentService.GetShareCurrencyBy(client, body.Figi)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	return c.JSON(http.StatusOK, currency)
}

func (h *Handlers) FindBy(c echo.Context) error {
	const op = "handlers.FindBy"

	ctx := c.Request().Context()

	logg := h.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")

	var body service.FindByReq
	err := c.Bind(&body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}

	client, err := h.service.InstrumentService.GetClient(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errIncorrectToken)
	}
	defer func() {
		if err := client.Stop(); err != nil {
			logg.WarnContext(ctx, "client stop failed", slog.Any("error", err))
		}
	}()

	instruments, err := h.service.InstrumentService.FindBy(client, body.Query)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	return c.JSON(http.StatusOK, instruments)
}

func (h *Handlers) GetBondsActions(c echo.Context) error {
	const op = "handlers.GetBondsActions"

	ctx := c.Request().Context()

	logg := h.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")

	var body service.BondsActionsReq
	err := c.Bind(&body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}

	client, err := h.service.AnalyticsService.GetClient(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errIncorrectToken)
	}
	defer func() {
		if err := client.Stop(); err != nil {
			logg.WarnContext(ctx, "client stop failed", slog.Any("error", err))
		}
	}()

	bondIdentificators, err := h.service.AnalyticsService.GetBondsActions(client, body.InstrumentUid)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	return c.JSON(http.StatusOK, bondIdentificators)
}

func (h *Handlers) GetLastPriceInPersentageToNominal(c echo.Context) error {
	const op = "handlers.GetLastPriceInPersentageToNominal"

	ctx := c.Request().Context()

	logg := h.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")

	var body service.LastPriceReq
	err := c.Bind(&body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}

	client, err := h.service.AnalyticsService.GetClient(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errIncorrectToken)
	}
	defer func() {
		if err := client.Stop(); err != nil {
			logg.WarnContext(ctx, "client stop failed", slog.Any("error", err))
		}
	}()

	lastPrice, err := h.service.AnalyticsService.GetLastPriceInPersentageToNominal(client, body.InstrumentUid)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	return c.JSON(http.StatusOK, lastPrice)
}
