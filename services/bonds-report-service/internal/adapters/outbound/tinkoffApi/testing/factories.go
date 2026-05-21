package factories

import (
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/dto"
	"bonds-report-service/internal/domain"
	"math"
	"time"
)

func NewTestLastPrice(numb float64) domain.LastPrice {
	return domain.LastPrice{
		LastPrice: QuotationFromFloat(numb),
	}
}

func QuotationFromFloat(v float64) domain.Quotation {
	const nanoFactor = 1_000_000_000

	v = math.Round(v*nanoFactor) / nanoFactor

	units := int64(v)
	nano := int32(math.Round((v - float64(units)) * nanoFactor))

	if nano == nanoFactor {
		units++
		nano = 0
	}

	return domain.Quotation{
		Units: units,
		Nano:  nano,
	}
}

func NewTestLastPriceDto(numb float64) dto.LastPriceResponse {
	return dto.LastPriceResponse{
		LastPrice: QuotationDtoFromFloat(numb),
	}
}

func QuotationDtoFromFloat(v float64) dto.Quotation {
	const nanoFactor = 1_000_000_000

	v = math.Round(v*nanoFactor) / nanoFactor

	units := int64(v)
	nano := int32(math.Round((v - float64(units)) * nanoFactor))

	if nano == nanoFactor {
		units++
		nano = 0
	}

	return dto.Quotation{
		Units: units,
		Nano:  nano,
	}
}

func NewFutureDTO() dto.Future {
	return dto.Future{
		Name: "Test Future",
		MinPriceIncrement: dto.Quotation{
			Units: 1,
			Nano:  0,
		},
		MinPriceIncrementAmount: dto.Quotation{
			Units: 10,
			Nano:  0,
		},
		AssetType:             "commodity",
		BasicAssetPositionUid: "basic-uid-123",
	}
}

func NewBond() dto.Bond {
	return dto.Bond{
		AciValue: dto.MoneyValue{
			Currency: "RUB",
			Units:    10,
			Nano:     500_000_000,
		},
		Currency: "RUB",
		Nominal: dto.MoneyValue{
			Currency: "RUB",
			Units:    1000,
			Nano:     0,
		},
	}
}

func NewInstrumentShortTest() dto.InstrumentShort {
	return dto.InstrumentShort{
		InstrumentType: "bond",
		Uid:            "UID123",
		Figi:           "FIGI123",
	}
}

func NewPortfolioPosition() dto.PortfolioPositions {
	return dto.PortfolioPositions{
		Figi:           "TEST_FIGI",
		InstrumentType: "share",
		Quantity: dto.Quotation{
			Units: 10,
			Nano:  0,
		},
		AveragePositionPrice: dto.MoneyValue{
			Currency: "RUB",
			Units:    100,
			Nano:     0,
		},
		ExpectedYield: dto.Quotation{
			Units: 5,
			Nano:  0,
		},
		CurrentNkd: dto.MoneyValue{
			Currency: "RUB",
			Units:    0,
			Nano:     0,
		},
		CurrentPrice: dto.MoneyValue{
			Currency: "RUB",
			Units:    110,
			Nano:     0,
		},
		AveragePositionPriceFifo: dto.MoneyValue{
			Currency: "RUB",
			Units:    100,
			Nano:     0,
		},
		Blocked: false,
		BlockedLots: dto.Quotation{
			Units: 0,
			Nano:  0,
		},
		PositionUid:   "POSITION_UID",
		InstrumentUid: "INSTRUMENT_UID",
		VarMargin: dto.MoneyValue{
			Currency: "RUB",
			Units:    0,
			Nano:     0,
		},
		ExpectedYieldFifo: dto.Quotation{
			Units: 5,
			Nano:  0,
		},
		DailyYield: dto.MoneyValue{
			Currency: "RUB",
			Units:    1,
			Nano:     0,
		},
		Ticker: "TEST",
	}
}

func NewPortfolio() dto.Portfolio {
	position := NewPortfolioPosition()

	return dto.Portfolio{
		Positions: []dto.PortfolioPositions{position},
		TotalAmount: dto.MoneyValue{
			Currency: "RUB",
			Units:    1000,
			Nano:     0,
		},
	}
}

func NewOperationDTO() dto.Operation {
	return dto.Operation{
		BrokerAccountID:   "ACC123",
		Currency:          "USD",
		OperationID:       "OP123",
		ParentOperationID: "POP123",
		Name:              "Buy Stock",
		Date:              time.Now(),
		Type:              1,
		Description:       "Test operation",
		InstrumentUid:     "INST123",
		Figi:              "FIGI123",
		InstrumentType:    "Stock",
		InstrumentKind:    "Common",
		PositionUid:       "POS123",
		Payment:           NewMoneyValueDTO(),
		Price:             NewMoneyValueDTO(),
		Commission:        NewMoneyValueDTO(),
		Yield:             NewMoneyValueDTO(),
		YieldRelative:     NewQuotationDTO(),
		AccruedInt:        NewMoneyValueDTO(),
		QuantityDone:      100,
		AssetUid:          "ASSET123",
	}
}

func NewMoneyValueDTO() dto.MoneyValue {
	return dto.MoneyValue{
		Currency: "USD",
		Units:    100,
		Nano:     500_000_000,
	}
}

func NewQuotationDTO() dto.Quotation {
	return dto.Quotation{
		Units: 1,
		Nano:  250_000_000,
	}
}
