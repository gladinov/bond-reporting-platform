package portfolioclient

import (
	httperrors "bonds-report-service/internal/adapters/outbound/http"
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/dto"
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/utils/logging"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/gladinov/e"
)

func (c *PortfolioTinkoffClient) GetAccounts(ctx context.Context) (_ map[string]domain.Account, err error) {
	const op = "tinkoffApi.GetAccounts"

	logg := c.logger.With()
	defer logging.LogOperation_Debug(ctx, logg, op, &err)()

	path := path.Join("tinkoff", "accounts")
	query := url.Values{}

	httpResponse, err := c.transport.DoRequest(ctx, path, query, nil)
	if err != nil {
		return nil, e.WrapIfErr("failed transport DoRequest", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return nil, httperrors.MapHTTPError(
			httpResponse.StatusCode,
			httpResponse.Body,
		)
	}

	data := make(map[string]dto.Account, 0)
	err = json.Unmarshal(httpResponse.Body, &data)
	if err != nil {
		return nil, e.WrapIfErr("failed to unmarshal response", err)
	}

	domainData := MapAccountsToDomain(data)

	return domainData, nil
}

func (c *PortfolioTinkoffClient) GetPortfolio(ctx context.Context, accountID string, accountStatus int64) (_ domain.Portfolio, err error) {
	const op = "tinkoffApi.GetPortfolio"

	logg := c.logger.With()
	defer logging.LogOperation_Debug(ctx, logg, op, &err)()

	path := path.Join("tinkoff", "portfolio")
	query := url.Values{}

	requestBody := dto.NewPortfolioRequest(accountID, accountStatus)

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return domain.Portfolio{}, fmt.Errorf("op:%s, could not marshall JSON", op)
	}
	formatRequestBody := bytes.NewBuffer(jsonData)

	httpResponse, err := c.transport.DoRequest(ctx, path, query, formatRequestBody)
	if err != nil {
		return domain.Portfolio{}, e.WrapIfErr("failed transport DoRequest", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return domain.Portfolio{}, httperrors.MapHTTPError(
			httpResponse.StatusCode,
			httpResponse.Body,
		)
	}

	var data dto.Portfolio
	err = json.Unmarshal(httpResponse.Body, &data)
	if err != nil {
		return domain.Portfolio{}, e.WrapIfErr("failed to unmarshal response", err)
	}

	domainData := MapPortfolioToDomain(data)

	return domainData, nil
}

func (c *PortfolioTinkoffClient) GetOperations(ctx context.Context, accountId string, date time.Time) (_ []domain.Operation, err error) {
	const op = "tinkoffApi.GetOperations"

	logg := c.logger.With()
	defer logging.LogOperation_Debug(ctx, logg, op, &err)()

	path := path.Join("tinkoff", "operations")
	query := url.Values{}

	requestBody := dto.NewOperationsRequest(accountId, date)

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, e.WrapIfErr("failed json.Marshal", err)
	}
	formatRequestBody := bytes.NewBuffer(jsonData)

	httpResponse, err := c.transport.DoRequest(ctx, path, query, formatRequestBody)
	if err != nil {
		return nil, e.WrapIfErr("failed transport DoRequest", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return nil, httperrors.MapHTTPError(
			httpResponse.StatusCode,
			httpResponse.Body,
		)
	}

	var data []dto.Operation
	err = json.Unmarshal(httpResponse.Body, &data)
	if err != nil {
		return nil, e.WrapIfErr("failed to unmarshal response", err)
	}

	domainData := MapOperationsToDomain(data)
	return domainData, nil
}
