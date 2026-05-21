package factories

import (
	"bonds-report-service/internal/adapters/outbound/cbr/dto"
	"bonds-report-service/internal/adapters/outbound/cbr/models"
	"encoding/json"
	"time"
)

func NewCurrencyRequest(date ...time.Time) *dto.CurrencyRequest {
	d := time.Now()
	if len(date) > 0 {
		d = date[0]
	}
	return dto.NewCurrencyRequest(d)
}

func NewCurrency() dto.Currency {
	c := dto.Currency{
		NumCode:   "840",
		CharCode:  "USD",
		Nominal:   "1",
		Name:      "US Dollar",
		Value:     "75.12",
		VunitRate: "1.0",
	}
	return c
}

func NewCurrenciesResponse() *dto.CurrenciesResponse {
	return &dto.CurrenciesResponse{
		Date: time.Now().Format("02.01.2006"),
		Currencies: []dto.Currency{
			NewCurrency(),
		},
	}
}

func NewHTTPResponse(statusCode int, body interface{}) *models.HTTPResponse {
	var bodyBytes []byte
	switch v := body.(type) {
	case string:
		bodyBytes = []byte(v)
	default:
		b, _ := json.Marshal(v)
		bodyBytes = b
	}
	return models.NewHTTPResponse(statusCode, bodyBytes)
}
