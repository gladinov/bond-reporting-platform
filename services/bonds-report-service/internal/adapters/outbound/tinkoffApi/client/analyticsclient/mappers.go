package analyticsclient

import (
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/dto"
	"bonds-report-service/internal/domain"
)

func MapBondIdentIdentifiers(dtoBondId dto.BondIdentIdentifiers) domain.BondIdentIdentifiers {
	return domain.BondIdentIdentifiers{
		Ticker:          dtoBondId.Ticker,
		ClassCode:       dtoBondId.ClassCode,
		Name:            dtoBondId.Name,
		Nominal:         MapMoneyValue(dtoBondId.Nominal),
		NominalCurrency: dtoBondId.NominalCurrency,
		Replaced:        dtoBondId.Replaced,
	}
}

func MapMoneyValue(dtoMoneyValue dto.MoneyValue) domain.MoneyValue {
	return domain.MoneyValue{
		Currency: dtoMoneyValue.Currency,
		Units:    dtoMoneyValue.Units,
		Nano:     dtoMoneyValue.Nano,
	}
}

func MapLastPriceResponseToDomain(dtoResp dto.LastPriceResponse) domain.LastPrice {
	return domain.LastPrice{
		LastPrice: MapQuotationToDomain(dtoResp.LastPrice),
	}
}

func MapQuotationToDomain(dtoQuat dto.Quotation) domain.Quotation {
	return domain.Quotation{
		Units: dtoQuat.Units,
		Nano:  dtoQuat.Nano,
	}
}
