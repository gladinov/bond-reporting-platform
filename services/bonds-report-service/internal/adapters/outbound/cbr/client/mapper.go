package cbr

import (
	"bonds-report-service/internal/adapters/outbound/cbr/dto"
	domain "bonds-report-service/internal/domain"
	"strconv"
	"strings"
	"time"
)

const layoutCurr = "02.01.2006"

func MapCurrenciesResponseToDomain(dtoResp dto.CurrenciesResponse) (domain.CurrenciesCBR, error) {
	var out domain.CurrenciesCBR
	currenciesMap := make(map[string]domain.CurrencyCBR)

	date, err := parseCBRDate(dtoResp.Date)
	if err != nil {
		return domain.CurrenciesCBR{}, err
	}

	for _, dtoCurr := range dtoResp.Currencies {
		domCurr, err := mapSingleCurrency(dtoCurr, date)
		if err != nil {
			return domain.CurrenciesCBR{}, err
		}
		key := normalizeCharCode(dtoCurr.CharCode)
		currenciesMap[key] = domCurr
	}

	out.CurrenciesMap = currenciesMap
	return out, nil
}

func mapSingleCurrency(dtoCurr dto.Currency, date time.Time) (domain.CurrencyCBR, error) {
	var domCurr domain.CurrencyCBR
	var err error

	domCurr.Date = date
	domCurr.NumCode = dtoCurr.NumCode
	domCurr.CharCode = normalizeCharCode(dtoCurr.CharCode)
	domCurr.Nominal, err = parseNominal(dtoCurr.Nominal)
	if err != nil {
		return domCurr, err
	}
	domCurr.Name = dtoCurr.Name
	domCurr.Value, err = parseFloat(dtoCurr.Value)
	if err != nil {
		return domCurr, err
	}
	domCurr.VunitRate, err = parseFloat(dtoCurr.VunitRate)
	if err != nil {
		return domCurr, err
	}

	return domCurr, nil
}

func parseCBRDate(s string) (time.Time, error) {
	return time.Parse(layoutCurr, s)
}

func parseNominal(s string) (int, error) {
	return strconv.Atoi(s)
}

func parseFloat(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}

func normalizeCharCode(code string) string {
	return strings.ToLower(code)
}
