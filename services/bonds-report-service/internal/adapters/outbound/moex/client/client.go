package moex

import (
	httperrors "bonds-report-service/internal/adapters/outbound/http"
	"bonds-report-service/internal/adapters/outbound/moex/dto"
	"bonds-report-service/internal/adapters/outbound/moex/transport"
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/utils/logging"
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/gladinov/e"
)

type Client struct {
	logger    *slog.Logger
	transport transport.TransportClient
}

func NewMoexClient(logger *slog.Logger, transport transport.TransportClient) *Client {
	return &Client{
		logger:    logger,
		transport: transport,
	}
}

func (c *Client) GetSpecifications(ctx context.Context, ticker string, date time.Time) (data domain.ValuesMoex, err error) {
	const op = "moex.GetSpecifications"
	logg := c.logger.With()
	defer logging.LogOperation_Debug(ctx, logg, op, &err)()

	request := dto.NewSpecificationsRequest(ticker, date)
	Path := path.Join("moex", "specifications")
	params := url.Values{}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return domain.ValuesMoex{}, e.WrapIfErr("failed json.Marshal", err)
	}
	formatRequestBody := bytes.NewBuffer(requestBody)

	httpResponse, err := c.transport.DoRequest(ctx, Path, params, formatRequestBody)
	if err != nil {
		return domain.ValuesMoex{}, e.WrapIfErr("failed transport DoRequest", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return domain.ValuesMoex{}, httperrors.MapHTTPError(
			httpResponse.StatusCode,
			httpResponse.Body,
		)
	}

	var res dto.Values
	err = json.Unmarshal(httpResponse.Body, &res)
	if err != nil {
		return domain.ValuesMoex{}, e.WrapIfErr("failed to unmarshal response", err)
	}

	domainRes := MapValueFromDTOToDomain(res)

	return domainRes, nil
}
