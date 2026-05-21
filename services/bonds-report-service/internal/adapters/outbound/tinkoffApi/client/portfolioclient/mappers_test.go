//go:build unit

package portfolioclient

import (
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/dto"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMapAccountToDomain(t *testing.T) {
	dto := dto.Account{
		ID:          "123",
		Type:        "broker",
		Name:        "Account1",
		Status:      1,
		OpenedDate:  time.Now(),
		ClosedDate:  time.Now(),
		AccessLevel: 2,
	}

	got := MapAccountToDomain(dto)
	require.Equal(t, dto.ID, got.ID)
	require.Equal(t, dto.Type, got.Type)
	require.Equal(t, dto.Name, got.Name)
	require.Equal(t, dto.Status, got.Status)
	require.Equal(t, dto.AccessLevel, got.AccessLevel)
}

func TestMapAccountsToDomain(t *testing.T) {
	dto := map[string]dto.Account{
		"a1": {ID: "1", Name: "A"},
		"a2": {ID: "2", Name: "B"},
	}

	got := MapAccountsToDomain(dto)
	require.Len(t, got, 2)
	require.Equal(t, "1", got["a1"].ID)
	require.Equal(t, "B", got["a2"].Name)
}

func TestMapQuotationToDomain(t *testing.T) {
	dto := dto.Quotation{Units: 5, Nano: 10}
	got := MapQuotationToDomain(dto)
	require.Equal(t, int64(5), got.Units)
	require.Equal(t, int32(10), got.Nano)
}

func TestMapMoneyValueToDomain(t *testing.T) {
	dto := dto.MoneyValue{Currency: "USD", Units: 100, Nano: 50}
	got := MapMoneyValueToDomain(dto)
	require.Equal(t, "USD", got.Currency)
	require.Equal(t, int64(100), got.Units)
	require.Equal(t, int32(50), got.Nano)
}

func TestMapOperationToDomain(t *testing.T) {
	dto := dto.Operation{
		BrokerAccountID: "acc1",
		Currency:        "USD",
		OperationID:     "op123",
		Name:            "Buy",
		Date:            time.Now(),
		Type:            1,
		Price:           dto.MoneyValue{Units: 10},
		YieldRelative:   dto.Quotation{Units: 2},
	}

	got := MapOperationToDomain(dto)
	require.Equal(t, dto.BrokerAccountID, got.BrokerAccountID)
	require.Equal(t, dto.Currency, got.Currency)
	require.Equal(t, dto.OperationID, got.OperationID)
	require.Equal(t, int64(10), got.Price.Units)
	require.Equal(t, int64(2), got.YieldRelative.Units)
}

func TestMapOperationsToDomain(t *testing.T) {
	dto := []dto.Operation{
		{OperationID: "op1"},
		{OperationID: "op2"},
	}

	got := MapOperationsToDomain(dto)
	require.Len(t, got, 2)
	require.Equal(t, "op1", got[0].OperationID)
	require.Equal(t, "op2", got[1].OperationID)
}

func TestMapPortfolioPositionToDomain(t *testing.T) {
	dto := dto.PortfolioPositions{
		Figi:                 "figi1",
		Quantity:             dto.Quotation{Units: 10},
		AveragePositionPrice: dto.MoneyValue{Units: 100},
		Blocked:              true,
		Ticker:               "TICK",
	}

	got := MapPortfolioPositionToDomain(dto)
	require.Equal(t, "figi1", got.Figi)
	require.Equal(t, int64(10), got.Quantity.Units)
	require.Equal(t, int64(100), got.AveragePositionPrice.Units)
	require.True(t, got.Blocked)
	require.Equal(t, "TICK", got.Ticker)
}

func TestMapPortfolioToDomain(t *testing.T) {
	dto := dto.Portfolio{
		Positions: []dto.PortfolioPositions{
			{Figi: "figi1", Quantity: dto.Quotation{Units: 5}},
			{Figi: "figi2", Quantity: dto.Quotation{Units: 10}},
		},
		TotalAmount: dto.MoneyValue{Currency: "USD", Units: 1000},
	}

	got := MapPortfolioToDomain(dto)
	require.Len(t, got.Positions, 2)
	require.Equal(t, "figi1", got.Positions[0].Figi)
	require.Equal(t, int64(5), got.Positions[0].Quantity.Units)
	require.Equal(t, "USD", got.TotalAmount.Currency)
	require.Equal(t, int64(1000), got.TotalAmount.Units)
}
