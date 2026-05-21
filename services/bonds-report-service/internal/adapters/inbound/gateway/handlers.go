package handlers

import (
	"bonds-report-service/internal/application/usecases"
	"log/slog"
	"net/http"

	"github.com/gladinov/valuefromcontext"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	logger  *slog.Logger
	service *usecases.Service
}

func NewHandlers(logger *slog.Logger, service *usecases.Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) GetAccountsList(c *gin.Context) {
	const op = "handlers.GetAccountsList"
	ctx := c.Request.Context()
	accountsResponce, err := h.service.GetAccountsList(ctx)
	if err != nil {
		h.logger.Error("failed to get accounts list",
			slog.String("op", op),
			slog.Any("error", err),
			slog.String("path", c.Request.URL.Path),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not get accounts"})
		return
	}

	accountsHTTP := MapAccountListToHTTP(&accountsResponce)

	c.JSON(http.StatusOK, accountsHTTP)
}

func (h *Handler) GetBondReportsByFifo(c *gin.Context) {
	const op = "handlers.GetBondReportsByFifo"
	ctx := c.Request.Context()
	chatID, err := valuefromcontext.GetChatIDFromCtxInt(ctx)
	if err != nil {
		h.logger.Warn("incorrect X-ChatId header",
			slog.String("op", op),
			slog.Any("error", err),
			slog.String("path", c.Request.URL.Path),
		)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "incorrect X-ChatId header"})
		return
	}
	err = h.service.GetBondReportsByFifo(ctx, chatID)
	if err != nil {
		h.logger.Error("internal server error",
			slog.String("op", op),
			slog.Any("error", err),
			slog.String("path", c.Request.URL.Path),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GetUSD(c *gin.Context) {
	const op = "handlers.GetUSD"
	ctx := c.Request.Context()

	usdResponce, err := h.service.GetUsd(ctx)
	if err != nil {
		h.logger.Error("internal server error",
			slog.String("op", op),
			slog.Any("error", err),
			slog.String("path", c.Request.URL.Path),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	usdHTTP := MapUsdToHTTP(&usdResponce)

	c.JSON(http.StatusOK, usdHTTP)
}

func (h *Handler) GetBondReports(c *gin.Context) {
	const op = "handlers.GetBondReports"
	ctx := c.Request.Context()
	logg := h.logger.With(
		slog.String("op", op),
		slog.String("path", c.Request.URL.Path))

	chatID, err := valuefromcontext.GetChatIDFromCtxInt(ctx)
	if err != nil {
		logg.Warn(
			"incorrect X-ChatId header",
			slog.Any("error", err))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "incorrect X-ChatId header"})
		return
	}
	getBondReportsResponse, err := h.service.GetBondReports(ctx, chatID)
	if err != nil {
		logg.Error("GetBondReports err",
			slog.Any("error", err),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	bondResponceHTTP := MapBondReportsToHTTP(&getBondReportsResponse)

	c.JSON(http.StatusOK, bondResponceHTTP)
}

func (h *Handler) GetPortfolioStructure(c *gin.Context) {
	const op = "handlers.GetPortfolioStructure"
	ctx := c.Request.Context()

	portfolioStructuresResonce, err := h.service.GetPortfolioStructureForEachAccount(ctx)
	if err != nil {
		h.logger.Error("internal server error",
			slog.String("op", op),
			slog.Any("error", err),
			slog.String("path", c.Request.URL.Path),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	portfolioHTTP := MapPortfolioStructureForEachAccountToHTTP(&portfolioStructuresResonce)
	c.JSON(http.StatusOK, portfolioHTTP)
}

func (h *Handler) GetUnionPortfolioStructure(c *gin.Context) {
	const op = "handlers.GetUnionPortfolioStructure"
	ctx := c.Request.Context()

	portfolioStructure, err := h.service.GetUnionPortfolioStructureForEachAccount(ctx)
	if err != nil {
		h.logger.Error("internal server error",
			slog.String("op", op),
			slog.Any("error", err),
			slog.String("path", c.Request.URL.Path),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	portfolioHTTP := MapUnionPortfolioStructureToHTTP(&portfolioStructure)

	c.JSON(http.StatusOK, portfolioHTTP)
}

func (h *Handler) GetUnionPortfolioStructureWithSber(c *gin.Context) {
	const op = "handlers.GetUnionPortfolioStructureWithSber"
	ctx := c.Request.Context()

	portfolioStructure, err := h.service.GetUnionPortfolioStructureWithSber(ctx)
	if err != nil {
		h.logger.Error("internal server error",
			slog.String("op", op),
			slog.Any("error", err),
			slog.String("path", c.Request.URL.Path),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	portfolioHTTP := MapUnionPortfolioStructureWithSberToHTTP(&portfolioStructure)
	c.JSON(http.StatusOK, portfolioHTTP)
}
