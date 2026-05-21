package dto

import (
	"math"
	"time"
)

type Portfolio struct {
	Positions   []PortfolioPositions `json:"positions,omitempty"`
	TotalAmount MoneyValue           `json:"totalAmount,omitempty"`
}

type PortfolioPositions struct {
	Figi                     string     `json:"figi,omitempty"`
	InstrumentType           string     `json:"instrumentType,omitempty"`
	Quantity                 Quotation  `json:"quantity,omitempty"`
	AveragePositionPrice     MoneyValue `json:"averagePositionPrice,omitempty"`
	ExpectedYield            Quotation  `json:"expectedYield,omitempty"`
	CurrentNkd               MoneyValue `json:"currentNkd,omitempty"`
	CurrentPrice             MoneyValue `json:"currentPrice,omitempty"`
	AveragePositionPriceFifo MoneyValue `json:"averagePositionPriceFifo,omitempty"`
	Blocked                  bool       `json:"blocked,omitempty"`
	BlockedLots              Quotation  `json:"blockedLots,omitempty"`
	PositionUid              string     `json:"positionUid,omitempty"`
	InstrumentUid            string     `json:"instrumentUid,omitempty"`
	VarMargin                MoneyValue `json:"varMargin,omitempty"`
	ExpectedYieldFifo        Quotation  `json:"expectedYieldFifo,omitempty"`
	DailyYield               MoneyValue `json:"dailyYield,omitempty"`
	Ticker                   string     `json:"ticker,omitempty"`
}

type MoneyValue struct {
	Currency string `json:"currency,omitempty"`
	Units    int64  `json:"units,omitempty"`
	Nano     int32  `json:"nano,omitempty"`
}

func (x *MoneyValue) GetCurrency() string {
	if x != nil {
		return x.Currency
	}
	return ""
}

func (x *MoneyValue) GetUnits() int64 {
	if x != nil {
		return x.Units
	}
	return 0
}

func (x *MoneyValue) GetNano() int32 {
	if x != nil {
		return x.Nano
	}
	return 0
}

func (mv *MoneyValue) ToFloat() float64 {
	num := float64(mv.Units) + float64(mv.Nano)*math.Pow10(-9)
	num = num * math.Pow10(9)
	num = math.Round(num)
	num = num / math.Pow10(9)
	return num
}

type Quotation struct {
	Units int64 `json:"units,omitempty"`
	Nano  int32 `json:"nano,omitempty"`
}

func (x *Quotation) GetUnits() int64 {
	if x != nil {
		return x.Units
	}
	return 0
}

func (x *Quotation) GetNano() int32 {
	if x != nil {
		return x.Nano
	}
	return 0
}

func (q *Quotation) ToFloat() float64 {
	num := float64(q.Units) + float64(q.Nano)*math.Pow10(-9)
	num = num * math.Pow10(9)
	num = math.Round(num)
	num = num / math.Pow10(9)
	return num
}

type Future struct {
	Name                    string    `json:"name,omitempty"`
	MinPriceIncrement       Quotation `json:"minPriceIncrement,omitempty"`
	MinPriceIncrementAmount Quotation `json:"minPriceIncrementAmount,omitempty"`
	AssetType               string    `json:"assetType,omitempty"`
	BasicAssetPositionUid   string    `json:"basicAssetPositionUid,omitempty"`
}

type Bond struct {
	AciValue MoneyValue `json:"aciValue,omitempty"`
	Currency string     `json:"currency,omitempty"`
	Nominal  MoneyValue `json:"nominal,omitempty"`
}

type Currency struct {
	Isin string `json:"isin,omitempty"`
}

type InstrumentShort struct {
	InstrumentType string `json:"instrumentType,omitempty"`
	Uid            string `json:"uid,omitempty"`
	Figi           string `json:"figi,omitempty"`
}

type Operation struct {
	BrokerAccountID   string     `json:"brokerAccountId,omitempty"`
	Currency          string     `json:"currency,omitempty"`
	OperationID       string     `json:"operationId,omitempty"`
	ParentOperationID string     `json:"parentOperationId,omitempty"`
	Name              string     `json:"name,omitempty"`
	Date              time.Time  `json:"date,omitempty"`
	Type              int64      `json:"type,omitempty"`
	Description       string     `json:"description,omitempty"`
	InstrumentUid     string     `json:"instrumentUid,omitempty"`
	Figi              string     `json:"figi,omitempty"`
	InstrumentType    string     `json:"instrumentType,omitempty"`
	InstrumentKind    string     `json:"instrumentKind,omitempty"`
	PositionUid       string     `json:"positionUid,omitempty"`
	Payment           MoneyValue `json:"payment,omitempty"`
	Price             MoneyValue `json:"price,omitempty"`
	Commission        MoneyValue `json:"commission,omitempty"`
	Yield             MoneyValue `json:"yield,omitempty"`
	YieldRelative     Quotation  `json:"yieldRelative,omitempty"`
	AccruedInt        MoneyValue `json:"accruedInt,omitempty"`
	QuantityDone      int64      `json:"quantityDone,omitempty"`
	AssetUid          string     `json:"assetUid,omitempty"`
}

type Account struct {
	ID          string    `json:"id,omitempty"`
	Type        string    `json:"type,omitempty"`
	Name        string    `json:"name,omitempty"`
	Status      int64     `json:"status,omitempty"`
	OpenedDate  time.Time `json:"openedDate,omitempty"`
	ClosedDate  time.Time `json:"closedDate,omitempty"`
	AccessLevel int64     `json:"accessLevel,omitempty"`
}

type BondIdentIdentifiers struct {
	Ticker          string     `json:"ticker,omitempty"`
	ClassCode       string     `json:"classCode,omitempty"`
	Name            string     `json:"name,omitempty"`
	Nominal         MoneyValue `json:"nominal,omitempty"`
	NominalCurrency string     `json:"nominalCurrency,omitempty"`
	Replaced        bool       `json:"replaced,omitempty"`
}

type PortfolioRequest struct {
	AccountID     string `json:"accountId,omitempty"`
	AccountStatus int64  `json:"accountStatus,omitempty"`
}

func NewPortfolioRequest(accountID string, accountStatus int64) PortfolioRequest {
	return PortfolioRequest{
		AccountID:     accountID,
		AccountStatus: accountStatus,
	}
}

type OperationsRequest struct {
	AccountID string    `json:"accountID,omitempty"`
	Date      time.Time `json:"date,omitempty"`
}

func NewOperationsRequest(accountID string, date time.Time) OperationsRequest {
	return OperationsRequest{
		AccountID: accountID,
		Date:      date,
	}
}

type FutureReq struct {
	Figi string `json:"figi,omitempty"`
}

func NewFutureReq(figi string) FutureReq {
	return FutureReq{Figi: figi}
}

type CurrencyReq struct {
	Figi string `json:"figi,omitempty"`
}

func NewCurrencyReq(figi string) *CurrencyReq {
	return &CurrencyReq{Figi: figi}
}

type BondReq struct {
	Uid string `json:"uid,omitempty"`
}

func NewBondReq(uid string) *BondReq {
	return &BondReq{Uid: uid}
}

type BaseShareFutureValuteReq struct {
	SharePositionUid string `json:"sharePositionUid,omitempty"`
}

func NewBaseShareFutureValuteReq(sharePositionUid string) *BaseShareFutureValuteReq {
	return &BaseShareFutureValuteReq{SharePositionUid: sharePositionUid}
}

type FindByReq struct {
	Query string `json:"query,omitempty"`
}

func NewFindByReq(query string) *FindByReq {
	return &FindByReq{Query: query}
}

type BondsActionsReq struct {
	InstrumentUid string `json:"instrumentUid,omitempty"`
}

func NewBondsActionsReq(instrumentUid string) BondsActionsReq {
	return BondsActionsReq{InstrumentUid: instrumentUid}
}

type LastPriceReq struct {
	InstrumentUid string `json:"instrumentUid,omitempty"`
}

func NewLastPriceReq(instrumentUid string) LastPriceReq {
	return LastPriceReq{InstrumentUid: instrumentUid}
}

type LastPriceResponse struct {
	LastPrice Quotation `json:"lastPrice,omitempty"`
}

type BaseShareFutureValuteResponse struct {
	Currency string `json:"currency,omitempty"`
}

type ShareCurrencyByResponse struct {
	Currency string `json:"currency,omitempty"`
}

type ShareCurrencyByReq struct {
	Figi string `json:"figi,omitempty"`
}

func NewShareCurrencyByReq(figi string) *ShareCurrencyByReq {
	return &ShareCurrencyByReq{Figi: figi}
}
