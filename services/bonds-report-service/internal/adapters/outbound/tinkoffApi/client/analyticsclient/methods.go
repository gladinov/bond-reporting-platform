package analyticsclient

import (
	httperrors "bonds-report-service/internal/adapters/outbound/http"
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/dto"
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/utils/logging"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/gladinov/e"
)

func (c *AnalyticsTinkoffClient) GetAllAssetUids(ctx context.Context) (_ map[string]string, err error) {
	const op = "tinkoffApi.GetAllAssetUids"

	logg := c.logger.With()
	defer logging.LogOperation_Debug(ctx, logg, op, &err)()

	path := path.Join("tinkoff", "allassetsuid")
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

	var data map[string]string
	err = json.Unmarshal(httpResponse.Body, &data)
	if err != nil {
		return nil, e.WrapIfErr("failed to unmarshal response", err)
	}

	return data, nil
}

func (c *AnalyticsTinkoffClient) GetBondsActions(ctx context.Context, instrumentUid string) (_ domain.BondIdentIdentifiers, err error) {
	const op = "tinkoffApi.GetBondsActions"

	logg := c.logger.With()
	defer logging.LogOperation_Debug(ctx, logg, op, &err)()

	path := path.Join("tinkoff", "bondactions")
	query := url.Values{}

	requestBody := dto.NewBondsActionsReq(instrumentUid)

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return domain.BondIdentIdentifiers{}, e.WrapIfErr("failed json.Marshal", err)
	}
	formatRequestBody := bytes.NewBuffer(jsonData)

	httpResponse, err := c.transport.DoRequest(ctx, path, query, formatRequestBody)
	if err != nil {
		return domain.BondIdentIdentifiers{}, e.WrapIfErr("failed transport DoRequest", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return domain.BondIdentIdentifiers{}, httperrors.MapHTTPError(
			httpResponse.StatusCode,
			httpResponse.Body,
		)
	}

	var data dto.BondIdentIdentifiers
	err = json.Unmarshal(httpResponse.Body, &data)
	if err != nil {
		return domain.BondIdentIdentifiers{}, e.WrapIfErr("failed to unmarshal response", err)
	}

	domainData := MapBondIdentIdentifiers(data)

	return domainData, nil
}

func (c *AnalyticsTinkoffClient) GetLastPriceInPersentageToNominal(ctx context.Context, instrumentUid string) (_ domain.LastPrice, err error) {
	const op = "tinkoffApi.GetLastPriceInPersentageToNominal"

	logg := c.logger.With()
	defer logging.LogOperation_Debug(ctx, logg, op, &err)()

	path := path.Join("tinkoff", "lastprice")
	query := url.Values{}

	requestBody := dto.NewLastPriceReq(instrumentUid)

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return domain.LastPrice{}, e.WrapIfErr("failed json.Marshal", err)
	}

	formatRequestBody := bytes.NewBuffer(jsonData)

	httpResponse, err := c.transport.DoRequest(ctx, path, query, formatRequestBody)
	if err != nil {
		return domain.LastPrice{}, e.WrapIfErr("failed transport DoRequest", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return domain.LastPrice{}, httperrors.MapHTTPError(
			httpResponse.StatusCode,
			httpResponse.Body,
		)
	}

	var data dto.LastPriceResponse
	err = json.Unmarshal(httpResponse.Body, &data)
	if err != nil {
		return domain.LastPrice{}, e.WrapIfErr("failed to unmarshal response", err)
	}

	domainData := MapLastPriceResponseToDomain(data)

	return domainData, nil
}
