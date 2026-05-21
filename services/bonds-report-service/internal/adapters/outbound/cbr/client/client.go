package cbr

import (
	"bonds-report-service/internal/adapters/outbound/cbr/dto"
	"bonds-report-service/internal/adapters/outbound/cbr/transport"
	httperrors "bonds-report-service/internal/adapters/outbound/http"
	"bonds-report-service/internal/utils/logging"
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"time"

	domain "bonds-report-service/internal/domain"

	"github.com/gladinov/e"
)

type Client struct {
	logger    *slog.Logger
	transport transport.TransportClient
}

func NewCbrClient(logger *slog.Logger, transport transport.TransportClient) *Client {
	return &Client{
		logger:    logger,
		transport: transport,
	}
}

func (c *Client) GetAllCurrencies(ctx context.Context, date time.Time) (_ domain.CurrenciesCBR, err error) {
	const op = "cbr.GetAllCurrencies"
	logg := c.logger.With()
	defer logging.LogOperation_Debug(ctx, logg, op, &err)()

	request := dto.NewCurrencyRequest(date)
	Path := path.Join("cbr", "currencies")
	params := url.Values{}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return domain.CurrenciesCBR{}, e.WrapIfErr("failed json.Marshal", err)
	}
	formatRequestBody := bytes.NewBuffer(requestBody)

	httpResponse, err := c.transport.DoRequest(ctx, Path, params, formatRequestBody)
	if err != nil {
		return domain.CurrenciesCBR{}, e.WrapIfErr("failed transport DoRequest", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return domain.CurrenciesCBR{}, httperrors.MapHTTPError(
			httpResponse.StatusCode,
			httpResponse.Body,
		)
	}

	var res dto.CurrenciesResponse
	err = json.Unmarshal(httpResponse.Body, &res)
	if err != nil {
		return domain.CurrenciesCBR{}, e.WrapIfErr("failed to unmarshal response", err)
	}

	domainRes, err := MapCurrenciesResponseToDomain(res)
	if err != nil {
		return domain.CurrenciesCBR{}, e.WrapIfErr("failed map currencies response to domain", err)
	}

	return domainRes, nil
}
