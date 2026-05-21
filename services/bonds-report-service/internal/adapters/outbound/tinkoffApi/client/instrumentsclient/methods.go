package instrumentsclient

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

func (c *InstrumentsTinkoffClient) GetFutureBy(ctx context.Context, figi string) (_ domain.Future, err error) {
	const op = "tinkoffApi.GetFutureBy"

	logg := c.logger.With()
	defer logging.LogOperation_Debug(ctx, logg, op, &err)()

	path := path.Join("tinkoff", "future")
	query := url.Values{}

	requestBody := dto.NewFutureReq(figi)

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return domain.Future{}, e.WrapIfErr("failed json.Marshal", err)
	}
	formatRequestBody := bytes.NewBuffer(jsonData)

	httpResponse, err := c.transport.DoRequest(ctx, path, query, formatRequestBody)
	if err != nil {
		return domain.Future{}, e.WrapIfErr("failed transport DoRequest", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return domain.Future{}, httperrors.MapHTTPError(
			httpResponse.StatusCode,
			httpResponse.Body,
		)
	}

	var data dto.Future
	err = json.Unmarshal(httpResponse.Body, &data)
	if err != nil {
		return domain.Future{}, e.WrapIfErr("failed to unmarshal response", err)
	}

	domainData := MapFutureToDomain(data)
	return domainData, nil
}

func (c *InstrumentsTinkoffClient) GetBondByUid(ctx context.Context, uid string) (_ domain.Bond, err error) {
	const op = "tinkoffApi.GetBondByUid"

	logg := c.logger.With()
	defer logging.LogOperation_Debug(ctx, logg, op, &err)()

	path := path.Join("tinkoff", "bond")
	query := url.Values{}

	requestBody := dto.NewBondReq(uid)

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return domain.Bond{}, e.WrapIfErr("failed json.Marshal", err)
	}

	formatRequestBody := bytes.NewBuffer(jsonData)

	httpResponse, err := c.transport.DoRequest(ctx, path, query, formatRequestBody)
	if err != nil {
		return domain.Bond{}, e.WrapIfErr("failed transport DoRequest", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return domain.Bond{}, httperrors.MapHTTPError(
			httpResponse.StatusCode,
			httpResponse.Body,
		)
	}

	var data dto.Bond
	err = json.Unmarshal(httpResponse.Body, &data)
	if err != nil {
		return domain.Bond{}, e.WrapIfErr("failed to unmarshal response", err)
	}

	domainData := MapBondToDomain(data)
	return domainData, nil
}

func (c *InstrumentsTinkoffClient) GetCurrencyBy(ctx context.Context, figi string) (_ domain.Currency, err error) {
	const op = "tinkoffApi.GetCurrencyBy"

	logg := c.logger.With()
	defer logging.LogOperation_Debug(ctx, logg, op, &err)()

	path := path.Join("tinkoff", "currency")
	query := url.Values{}

	requestBody := dto.NewCurrencyReq(figi)

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return domain.Currency{}, e.WrapIfErr("failed json.Marshal", err)
	}
	formatRequestBody := bytes.NewBuffer(jsonData)

	httpResponse, err := c.transport.DoRequest(ctx, path, query, formatRequestBody)
	if err != nil {
		return domain.Currency{}, e.WrapIfErr("failed transport DoRequest", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return domain.Currency{}, httperrors.MapHTTPError(
			httpResponse.StatusCode,
			httpResponse.Body,
		)
	}

	var data dto.Currency
	err = json.Unmarshal(httpResponse.Body, &data)
	if err != nil {
		return domain.Currency{}, e.WrapIfErr("failed to unmarshal response", err)
	}

	domainData := MapCurrencyToDomain(data)

	return domainData, nil
}

func (c *InstrumentsTinkoffClient) FindBy(ctx context.Context, findQuery string) (_ domain.InstrumentShortList, err error) {
	const op = "tinkoffApi.FindBy"

	logg := c.logger.With()
	defer logging.LogOperation_Debug(ctx, logg, op, &err)()

	path := path.Join("tinkoff", "findby")

	requestBody := dto.NewFindByReq(findQuery)
	query := url.Values{}

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

	var data []dto.InstrumentShort
	err = json.Unmarshal(httpResponse.Body, &data)
	if err != nil {
		return nil, e.WrapIfErr("failed to unmarshal response", err)
	}

	domainData := MapSliceInstrumentShortToDomain(data)
	return domainData, nil
}

func (c *InstrumentsTinkoffClient) GetShareCurrencyBy(ctx context.Context, figi string) (_ domain.ShareCurrency, err error) {
	const op = "tinkoffApi.GetShareCurrencyBy"

	logg := c.logger.With()
	defer logging.LogOperation_Debug(ctx, logg, op, &err)()

	path := path.Join("tinkoff", "share", "currency")
	query := url.Values{}

	requestBody := dto.NewShareCurrencyByReq(figi)

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return domain.ShareCurrency{}, e.WrapIfErr("failed json.Marshal", err)
	}
	formatRequestBody := bytes.NewBuffer(jsonData)

	httpResponse, err := c.transport.DoRequest(ctx, path, query, formatRequestBody)
	if err != nil {
		return domain.ShareCurrency{}, e.WrapIfErr("failed transport DoRequest", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return domain.ShareCurrency{}, httperrors.MapHTTPError(
			httpResponse.StatusCode,
			httpResponse.Body,
		)
	}

	var data dto.ShareCurrencyByResponse
	err = json.Unmarshal(httpResponse.Body, &data)
	if err != nil {
		return domain.ShareCurrency{}, e.WrapIfErr("failed to unmarshal response", err)
	}

	domainData := MapShareCurrencyByResponseToDomain(data)
	return domainData, nil
}
