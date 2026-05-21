package instrumentsclient

import (
	"bonds-report-service/internal/adapters/outbound/tinkoffApi/dto"
	"bonds-report-service/internal/domain"
)

func MapFutureToDomain(dto dto.Future) domain.Future {
	return domain.Future{
		Name:                    dto.Name,
		MinPriceIncrement:       MapQuotationToDomain(dto.MinPriceIncrement),
		MinPriceIncrementAmount: MapQuotationToDomain(dto.MinPriceIncrementAmount),
		AssetType:               dto.AssetType,
		BasicAssetPositionUid:   dto.BasicAssetPositionUid,
	}
}

func MapQuotationToDomain(dtoQuat dto.Quotation) domain.Quotation {
	return domain.Quotation{
		Units: dtoQuat.Units,
		Nano:  dtoQuat.Nano,
	}
}

func MapMoneyValueToDomain(dto dto.MoneyValue) domain.MoneyValue {
	return domain.MoneyValue{
		Currency: dto.Currency,
		Units:    dto.Units,
		Nano:     dto.Nano,
	}
}

func MapBondToDomain(dto dto.Bond) domain.Bond {
	return domain.Bond{
		AciValue: MapMoneyValueToDomain(dto.AciValue),
		Currency: dto.Currency,
		Nominal:  MapMoneyValueToDomain(dto.Nominal),
	}
}

func MapCurrencyToDomain(dto dto.Currency) domain.Currency {
	return domain.Currency{
		Isin: dto.Isin,
	}
}

func MapInstrumentShortToDomain(dto dto.InstrumentShort) domain.InstrumentShort {
	return domain.InstrumentShort{
		InstrumentType: dto.InstrumentType,
		Uid:            dto.Uid,
		Figi:           dto.Figi,
	}
}

func MapSliceInstrumentShortToDomain(dto []dto.InstrumentShort) domain.InstrumentShortList {
	out := make([]domain.InstrumentShort, 0, len(dto))
	for _, v := range dto {
		dom := MapInstrumentShortToDomain(v)
		out = append(out, dom)
	}
	return out
}

func MapShareCurrencyByResponseToDomain(dto dto.ShareCurrencyByResponse) domain.ShareCurrency {
	return domain.ShareCurrency{
		Currency: dto.Currency,
	}
}
