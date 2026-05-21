package dto

import "time"

type CurrencyRequest struct {
	Date time.Time `json:"date,omitempty"`
}

func NewCurrencyRequest(date time.Time) *CurrencyRequest {
	return &CurrencyRequest{Date: date}
}

type Currency struct {
	NumCode   string `json:"numCode,omitempty"`
	CharCode  string `json:"charCode,omitempty"`
	Nominal   string `json:"nominal,omitempty"`
	Name      string `json:"name,omitempty"`
	Value     string `json:"value,omitempty"`
	VunitRate string `json:"vunitRate,omitempty"`
}

type CurrenciesResponse struct {
	Date       string     `json:"date,omitempty"`
	Currencies []Currency `json:"valute,omitempty"`
}

func NewCurrenciesResponce(date string, currs []Currency) *CurrenciesResponse {
	return &CurrenciesResponse{
		Date:       date,
		Currencies: currs,
	}
}
