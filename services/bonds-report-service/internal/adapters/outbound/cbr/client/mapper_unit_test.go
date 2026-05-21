//go:build unit

package cbr

import (
	"bonds-report-service/internal/adapters/outbound/cbr/dto"
	"testing"
	"time"

	domain "bonds-report-service/internal/domain"

	"github.com/stretchr/testify/require"
)

func TestMapCurrenciesResponseToDomain(t *testing.T) {
	layout := "02.01.2006"

	t.Run("success mapping", func(t *testing.T) {
		dtoResp := dto.CurrenciesResponse{
			Date: "05.02.2026",
			Currencies: []dto.Currency{
				{
					NumCode:   "840",
					CharCode:  "USD",
					Nominal:   "1",
					Name:      "US Dollar",
					Value:     "74,50",
					VunitRate: "1,00",
				},
			},
		}

		got, err := MapCurrenciesResponseToDomain(dtoResp)
		require.NoError(t, err)
		require.Len(t, got.CurrenciesMap, 1)

		expectedDate, _ := time.Parse(layout, "05.02.2026")
		expected := domain.CurrencyCBR{
			Date:      expectedDate,
			NumCode:   "840",
			CharCode:  "usd",
			Nominal:   1,
			Name:      "US Dollar",
			Value:     74.50,
			VunitRate: 1.00,
		}
		c := got.CurrenciesMap["usd"]
		require.Equal(t, expected, c)
	})

	t.Run("invalid date", func(t *testing.T) {
		dtoResp := dto.CurrenciesResponse{
			Date: "invalid-date",
			Currencies: []dto.Currency{
				{NumCode: "840", CharCode: "USD", Nominal: "1", Name: "US Dollar", Value: "74,50", VunitRate: "1,00"},
			},
		}
		_, err := MapCurrenciesResponseToDomain(dtoResp)
		require.Error(t, err)
	})

	t.Run("invalid nominal", func(t *testing.T) {
		dtoResp := dto.CurrenciesResponse{
			Date: "05.02.2026",
			Currencies: []dto.Currency{
				{NumCode: "840", CharCode: "USD", Nominal: "abc", Name: "US Dollar", Value: "74,50", VunitRate: "1,00"},
			},
		}
		_, err := MapCurrenciesResponseToDomain(dtoResp)
		require.Error(t, err)
	})

	t.Run("invalid value", func(t *testing.T) {
		dtoResp := dto.CurrenciesResponse{
			Date: "05.02.2026",
			Currencies: []dto.Currency{
				{NumCode: "840", CharCode: "USD", Nominal: "1", Name: "US Dollar", Value: "invalid", VunitRate: "1,00"},
			},
		}
		_, err := MapCurrenciesResponseToDomain(dtoResp)
		require.Error(t, err)
	})

	t.Run("invalid vunitRate", func(t *testing.T) {
		dtoResp := dto.CurrenciesResponse{
			Date: "05.02.2026",
			Currencies: []dto.Currency{
				{NumCode: "840", CharCode: "USD", Nominal: "1", Name: "US Dollar", Value: "74,50", VunitRate: "x"},
			},
		}
		_, err := MapCurrenciesResponseToDomain(dtoResp)
		require.Error(t, err)
	})
}
