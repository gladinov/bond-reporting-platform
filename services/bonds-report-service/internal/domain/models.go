package domain

import (
	"math"
	"time"
)

type PortfolioStructureForEachAccountResponce struct {
	PortfolioStructures []PortfolioStructure
}

type UnionPortfolioStructureResponce struct {
	Portfolio *PortfolioByTypeAndCurrency
}

type UnionPortfolioStructureWithSberResponce struct {
	Portfolio *PortfolioByTypeAndCurrency
}

type PortfolioStructure struct {
	AccountName string
	Portfolio   *PortfolioByTypeAndCurrency
}

type UsdResponce struct {
	Usd float64
}

type PortfolioByTypeAndCurrency struct {
	AllAssets        float64
	BondsAssets      BondsAssets
	SharesAssets     SharesAssets
	EtfsAssets       EtfsAssets
	FuturesAssets    FuturesAssets
	CurrenciesAssets CurrenciesAssets
}

func NewPortfolioByTypeAndCurrency() *PortfolioByTypeAndCurrency {
	return &PortfolioByTypeAndCurrency{
		BondsAssets:      NewBondAssets(),
		SharesAssets:     NewSharesAssets(),
		EtfsAssets:       NewEtfsAssets(),
		FuturesAssets:    NewFutureAssets(),
		CurrenciesAssets: NewCurrenciesAssets(),
	}
}

func (first *PortfolioByTypeAndCurrency) SumWithPortfolio(second *PortfolioByTypeAndCurrency) {
	if second == nil {
		return
	}
	first.AllAssets += second.AllAssets
	first.BondsAssets.SumWithBondAssets(second.BondsAssets)
	first.SharesAssets.SumWithSharesAssets(second.SharesAssets)
	first.CurrenciesAssets.SumWithCurrenciesAssets(second.CurrenciesAssets)
	first.EtfsAssets.SumWithEtfsAssets(second.EtfsAssets)
	first.FuturesAssets.SumWithFutureAssets(second.FuturesAssets)
}

type BondsAssets struct {
	SumOfAssets      float64
	AssetsByCurrency map[string]*AssetByParam
}

func NewBondAssets() BondsAssets {
	return BondsAssets{
		AssetsByCurrency: NewAssetsByCurrency(),
	}
}

func (first *BondsAssets) SumWithBondAssets(second BondsAssets) {
	first.SumOfAssets += second.SumOfAssets
	first.AssetsByCurrency = SumOfAssetsByCurrency(first.AssetsByCurrency, second.AssetsByCurrency)
}

func SumOfAssetsByCurrency(first, second map[string]*AssetByParam) map[string]*AssetByParam {
	if second == nil {
		return first
	}
	if first == nil {
		return second
	}

	for key, value := range second {
		if value != nil {
			AddToMap(first, key, value.SumOfAssets)
		}
	}
	return first
}

type AssetByParam struct {
	SumOfAssets float64
}

func NewAssetsByParam() *AssetByParam {
	return &AssetByParam{}
}

func AddToMap(assets map[string]*AssetByParam, field string, addValue float64) {
	assetByParam, exist := assets[field]
	if exist {
		assetByParam.SumOfAssets += addValue
	} else {
		assets[field] = &AssetByParam{SumOfAssets: addValue}
	}
}

type SharesAssets struct {
	SumOfAssets      float64
	AssetsByCurrency map[string]*AssetByParam
}

func (first *SharesAssets) SumWithSharesAssets(second SharesAssets) {
	first.SumOfAssets += second.SumOfAssets
	first.AssetsByCurrency = SumOfAssetsByCurrency(first.AssetsByCurrency, second.AssetsByCurrency)
}

func NewSharesAssets() SharesAssets {
	return SharesAssets{
		AssetsByCurrency: NewAssetsByCurrency(),
	}
}

type EtfsAssets struct {
	SumOfAssets      float64
	AssetsByCurrency map[string]*AssetByParam
}

func (first *EtfsAssets) SumWithEtfsAssets(second EtfsAssets) {
	first.SumOfAssets += second.SumOfAssets
	first.AssetsByCurrency = SumOfAssetsByCurrency(first.AssetsByCurrency, second.AssetsByCurrency)
}

func NewEtfsAssets() EtfsAssets {
	return EtfsAssets{
		AssetsByCurrency: NewAssetsByCurrency(),
	}
}

type FuturesAssets struct {
	SumOfAssets  float64
	AssetsByType AssetsByType
}

func (first *FuturesAssets) SumWithFutureAssets(second FuturesAssets) {
	first.SumOfAssets += second.SumOfAssets
	first.AssetsByType.Commodity.SumWithFuturesType(second.AssetsByType.Commodity)
	first.AssetsByType.Currency.SumWithFuturesType(second.AssetsByType.Currency)
	first.AssetsByType.Security.SumWithFuturesType(second.AssetsByType.Security)
	first.AssetsByType.Index.SumWithFuturesType(second.AssetsByType.Index)
}

func NewFutureAssets() FuturesAssets {
	return FuturesAssets{
		AssetsByType: NewAssetByType(),
	}
}

type AssetsByType struct {
	Commodity FuturesType
	Currency  FuturesType
	Security  FuturesType
	Index     FuturesType
}

func NewAssetByType() AssetsByType {
	return AssetsByType{
		Commodity: NewFuturesType(),
		Currency:  NewFuturesType(),
		Security:  NewFuturesType(),
		Index:     NewFuturesType(),
	}
}

type FuturesType struct {
	SumOfAssets      float64
	AssetsByCurrency map[string]*AssetByParam
}

func (first *FuturesType) SumWithFuturesType(second FuturesType) {
	first.SumOfAssets += second.SumOfAssets
	first.AssetsByCurrency = SumOfAssetsByCurrency(first.AssetsByCurrency, second.AssetsByCurrency)
}

func NewFuturesType() FuturesType {
	return FuturesType{
		AssetsByCurrency: NewAssetsByCurrency(),
	}
}

type CurrenciesAssets struct {
	SumOfAssets      float64
	AssetsByCurrency map[string]*AssetByParam
}

func (first *CurrenciesAssets) SumWithCurrenciesAssets(second CurrenciesAssets) {
	first.SumOfAssets += second.SumOfAssets
	first.AssetsByCurrency = SumOfAssetsByCurrency(first.AssetsByCurrency, second.AssetsByCurrency)
}

func NewCurrenciesAssets() CurrenciesAssets {
	return CurrenciesAssets{
		AssetsByCurrency: NewAssetsByCurrency(),
	}
}

func NewAssetsByCurrency() map[string]*AssetByParam {
	return make(map[string]*AssetByParam)
}

type OperationsRequest struct {
	AccountID string
	FromDate  time.Time
}

func NewOperationsRequest(accountID string, fromDate time.Time) OperationsRequest {
	return OperationsRequest{
		AccountID: accountID,
		FromDate:  fromDate,
	}
}

func (r OperationsRequest) Validate(now time.Time) error {
	if r.AccountID == "" {
		return ErrEmptyAccountID
	}
	if r.FromDate.After(now) {
		return ErrInvalidFromDate
	}
	return nil
}

type ReportLine struct {
	Operation  []OperationWithoutCustomTypes
	Bond       BondIdentIdentifiers
	LastPrice  LastPrice
	Vunit_rate Rate
}

func NewReportLine(op []OperationWithoutCustomTypes, bond BondIdentIdentifiers, price LastPrice, vunit_rate Rate) ReportLine {
	return ReportLine{
		Operation:  op,
		Bond:       bond,
		LastPrice:  price,
		Vunit_rate: vunit_rate,
	}
}

type Rate struct {
	IsoCurrencyName string
	Vunit_Rate      NullFloat64
}

type CurrencyCBR struct {
	Date      time.Time
	NumCode   string
	CharCode  string
	Nominal   int
	Name      string
	Value     float64
	VunitRate float64
}

type CurrenciesCBR struct {
	CurrenciesMap map[string]CurrencyCBR
}

func NewCurrencies(mapCurr map[string]CurrencyCBR) *CurrenciesCBR {
	return &CurrenciesCBR{
		CurrenciesMap: mapCurr,
	}
}

type ValuesMoex struct {
	ShortName       NullString
	TradeDate       NullString
	MaturityDate    NullString
	OfferDate       NullString
	BuybackDate     NullString
	YieldToMaturity NullFloat64
	YieldToOffer    NullFloat64
	FaceValue       NullFloat64
	FaceUnit        NullString
	Duration        NullFloat64
}

type NullString struct {
	Value  string
	IsSet  bool
	IsNull bool
}

func NewNullString(value string, isSet bool, isNull bool) NullString {
	return NullString{
		Value:  value,
		IsSet:  isSet,
		IsNull: isNull,
	}
}

func (ns NullString) GetValue() string {
	return ns.Value
}

func (ns NullString) GetIsSet() bool {
	return ns.IsSet
}

func (ns NullString) GetIsNull() bool {
	return ns.IsNull
}

type NullFloat64 struct {
	Value  float64
	IsSet  bool
	IsNull bool
}

func NewNullFloat64(value float64, isSet bool, isNull bool) NullFloat64 {
	return NullFloat64{
		Value:  value,
		IsSet:  isSet,
		IsNull: isNull,
	}
}

func (nf NullFloat64) GetValue() float64 {
	return nf.Value
}

func (nf NullFloat64) GetIsSet() bool {
	return nf.IsSet
}

func (nf NullFloat64) GetIsNull() bool {
	return nf.IsNull
}

func (nf NullFloat64) IsHasValue() bool {
	if !nf.IsSet || nf.IsNull {
		return false
	}
	return true
}

func (ns NullString) IsHasValue() bool {
	if !ns.IsSet || ns.IsNull {
		return false
	}
	return true
}

type BondIdentIdentifiers struct {
	Ticker          string
	ClassCode       string
	Name            string
	Nominal         MoneyValue
	NominalCurrency string
	Replaced        bool
}

type LastPrice struct {
	LastPrice Quotation
}

type MoneyValue struct {
	Currency string
	Units    int64
	Nano     int32
}

func NewMoneyValue(currency string, units int64, nano int32) MoneyValue {
	return MoneyValue{
		Units:    units,
		Nano:     nano,
		Currency: currency,
	}
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
	Units int64
	Nano  int32
}

func NewQuotation(units int64, nano int32) Quotation {
	return Quotation{
		Units: units,
		Nano:  nano,
	}
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
	Name                    string
	MinPriceIncrement       Quotation
	MinPriceIncrementAmount Quotation
	AssetType               string
	BasicAssetPositionUid   string
}

type Portfolio struct {
	Positions   []PortfolioPosition
	TotalAmount MoneyValue
}

type PortfolioPosition struct {
	Figi                     string
	InstrumentType           string
	Quantity                 Quotation
	AveragePositionPrice     MoneyValue
	ExpectedYield            Quotation
	CurrentNkd               MoneyValue
	CurrentPrice             MoneyValue
	AveragePositionPriceFifo MoneyValue
	Blocked                  bool
	BlockedLots              Quotation
	PositionUid              string
	InstrumentUid            string
	VarMargin                MoneyValue
	ExpectedYieldFifo        Quotation
	DailyYield               MoneyValue
	Ticker                   string
}

type PortfolioPositionsWithAssetUid struct {
	InstrumentType string
	AssetUid       string
	InstrumentUid  string
}

type Bond struct {
	AciValue MoneyValue
	Currency string
	Nominal  MoneyValue
}

type Currency struct {
	Isin string
}

type InstrumentShortList []InstrumentShort

func (list InstrumentShortList) ValidateAndGetFirstShare() (InstrumentShort, error) {
	if len(list) == 0 {
		return InstrumentShort{}, ErrEmptyInstrumentShortResponce
	}

	inst := list[0]

	if inst.InstrumentType != share {
		return InstrumentShort{}, ErrInstrumentNotShare
	}

	if inst.Figi == "" {
		return InstrumentShort{}, ErrEmptyFigi
	}

	return inst, nil
}

type InstrumentShort struct {
	InstrumentType string
	Uid            string
	Figi           string
}

type Operation struct {
	BrokerAccountID   string
	Currency          string
	OperationID       string
	ParentOperationID string
	Name              string
	Date              time.Time
	Type              int64
	Description       string
	InstrumentUid     string
	Figi              string
	InstrumentType    string
	InstrumentKind    string
	PositionUid       string
	Payment           MoneyValue
	Price             MoneyValue
	Commission        MoneyValue
	Yield             MoneyValue
	YieldRelative     Quotation
	AccruedInt        MoneyValue
	QuantityDone      int64
	AssetUid          string
}

type OperationWithoutCustomTypes struct {
	BrokerAccountID   string
	Currency          string
	OperationID       string
	ParentOperationID string
	Name              string
	Date              time.Time
	Type              int64
	Description       string
	InstrumentUid     string
	Figi              string
	InstrumentType    string
	InstrumentKind    string
	PositionUid       string
	Payment           float64
	Price             float64
	Commission        float64
	Yield             float64
	YieldRelative     float64
	AccruedInt        float64
	QuantityDone      float64
	AssetUid          string
}

// TODO: Подумать как назвать адекватней.
// TODO: Подумать стоит ли хранить струкутру и метод здесб
func (op *OperationWithoutCustomTypes) ApplyValuesIfCurrentQuantityLessThanSellQuantity(proportion float64, currQuantity float64) {
	// НКД
	op.AccruedInt -= op.AccruedInt * proportion
	// Плюсуем комиссию за продажу бумаг
	op.Commission -= op.Commission * proportion

	// Изменяем значение Quantity.Operation
	op.QuantityDone -= currQuantity
}

type Account struct {
	ID          string
	Type        string
	Name        string
	Status      int64
	OpenedDate  time.Time
	ClosedDate  time.Time
	AccessLevel int64
}

func (a Account) ValidateForPortfolio() error {
	switch a.Status {
	case 0:
		return ErrUnspecifiedAccount
	case 1:
		return ErrNewNotOpenYetAccount
	case 3:
		return ErrCloseAccount
	}

	if a.AccessLevel == 3 {
		return ErrNoAcces
	}

	if a.ID == "" {
		return ErrEmptyAccountIdInRequest
	}

	return nil
}

type ShareCurrency struct {
	Currency string
}
