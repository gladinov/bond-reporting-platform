package dto

import "time"

type SpecificationsRequest struct {
	Ticker string    `json:"ticker"`
	Date   time.Time `json:"date"`
}

func NewSpecificationsRequest(ticker string, date time.Time) *SpecificationsRequest {
	return &SpecificationsRequest{
		Ticker: ticker,
		Date:   date,
	}
}

type Values struct {
	ShortName       NullString  `json:"SHORTNAME"`
	TradeDate       NullString  `json:"TRADEDATE"`    // Торговая дата(на момент которой рассчитаны остальные данные)
	MaturityDate    NullString  `json:"MATDATE"`      // Дата погашения
	OfferDate       NullString  `json:"OFFERDATE"`    // Дата Оферты
	BuybackDate     NullString  `json:"BUYBACKDATE"`  // дата обратного выкупа
	YieldToMaturity NullFloat64 `json:"YIELDCLOSE"`   // Доходность к погашению при покупке
	YieldToOffer    NullFloat64 `json:"YIELDTOOFFER"` // Доходность к оферте при покупке
	FaceValue       NullFloat64 `json:"FACEVALUE"`
	FaceUnit        NullString  `json:"FACEUNIT"` // номинальная стоимость облигации
	Duration        NullFloat64 `json:"DURATION"` // дюрация (средневзвешенный срок платежей)
}

type NullString struct {
	Value  string `json:"value"`
	IsSet  bool   `json:"isSet"`
	IsNull bool   `json:"isNull"`
}

type NullFloat64 struct {
	Value  float64 `json:"value"`
	IsSet  bool    `json:"isSet"`
	IsNull bool    `json:"isNull"`
}
