package presenter

import (
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/utils"
	"fmt"
)

type PortfolioStructureTitle int

const (
	UnionPortfolio PortfolioStructureTitle = iota
	UnionPortfolioWithSber
	EachPortfolio
)

func ResponsePortfolioStructure(portfolio *domain.PortfolioByTypeAndCurrency, title PortfolioStructureTitle, accountName string) string {
	var accountTitle string
	switch title {
	case UnionPortfolio:
		accountTitle = "Струтура всех открытых счетов в Тинькофф Инвестициях\n"
	case UnionPortfolioWithSber:
		accountTitle = "Струтура всех инвестиций\n"
	case EachPortfolio:
		accountTitle = fmt.Sprintf("Струтура брокерского счета: %s\n", accountName)
	default:
		accountTitle = ""
	}

	var output string
	totalAmount := portfolio.AllAssets
	totalBonds := portfolio.BondsAssets.SumOfAssets
	totalShares := portfolio.SharesAssets.SumOfAssets
	totalEtfs := portfolio.EtfsAssets.SumOfAssets
	totalFutures := portfolio.FuturesAssets.SumOfAssets
	totalCurrencies := portfolio.CurrenciesAssets.SumOfAssets
	totalLeverageRatio := (totalBonds + totalShares + totalEtfs + totalFutures) / totalAmount

	totalBondsInPerc := totalBonds / totalAmount * 100
	totalSharesInPerc := totalShares / totalAmount * 100
	totalEtfsInPerc := totalEtfs / totalAmount * 100
	totalFuturesInPerc := totalFutures / totalAmount * 100
	totalCurrenciesInPerc := totalCurrencies / totalAmount * 100
	totalLeverageRatioInPers := totalLeverageRatio * 100

	totalAmountOut := fmt.Sprintf("Общая стоимость портфеля составляет: %.2f\n", totalAmount)
	totalBondsOut := fmt.Sprintf("  Стоимость облигаций: %.2f (%.2f%%от портфеля)\n", totalBonds, totalBondsInPerc)
	totalSharesOut := fmt.Sprintf("  Стоимость акций: %.2f (%.2f%%от портфеля)\n", totalShares, totalSharesInPerc)
	totalEtfsOut := fmt.Sprintf("  Стоимость ETF: %.2f (%.2f%%от портфеля)\n", totalEtfs, totalEtfsInPerc)
	totalFuturesOut := fmt.Sprintf("  Стоимость фьючерсов: %.2f (%.2f%%от портфеля)\n", totalFutures, totalFuturesInPerc)
	totalCurrenciesOut := fmt.Sprintf("  Стоимость валют: %.2f (%.2f%%от портфеля)\n", totalCurrencies, totalCurrenciesInPerc)
	totalLeverageRatioOut := fmt.Sprintf("  Общий коэффициент левериджа: %.2f (%.2f%%от портфеля)\n", totalLeverageRatio, totalLeverageRatioInPers)

	var bondsByCurrencies string
	for k, v := range portfolio.BondsAssets.AssetsByCurrency {
		bondsByCurr := utils.RoundFloat(v.SumOfAssets, 2)
		bondsByCurrInPers := utils.RoundFloat(bondsByCurr/totalAmount*100, 2)
		bondsByCurrencies += "    " + fmt.Sprintf("Стоимость облигаций в %s: %.2f (%.2f%%от портфеля)\n", k, bondsByCurr, bondsByCurrInPers)
	}

	var sharesByCurrencies string
	for k, v := range portfolio.SharesAssets.AssetsByCurrency {
		AssetByCurr := utils.RoundFloat(v.SumOfAssets, 2)
		AssetByCurrInPers := utils.RoundFloat(AssetByCurr/totalAmount*100, 2)
		sharesByCurrencies += "    " + fmt.Sprintf("Стоимость акций в %s: %.2f (%.2f%%от портфеля)\n", k, AssetByCurr, AssetByCurrInPers)
	}

	var etfsByCurrencies string
	for k, v := range portfolio.EtfsAssets.AssetsByCurrency {
		AssetByCurr := utils.RoundFloat(v.SumOfAssets, 2)
		AssetByCurrInPers := utils.RoundFloat(AssetByCurr/totalAmount*100, 2)
		etfsByCurrencies += "    " + fmt.Sprintf("Стоимость Etf в %s: %.2f (%.2f%%от портфеля)\n", k, AssetByCurr, AssetByCurrInPers)
	}

	var futuresByCurrencies string
	if portfolio.FuturesAssets.AssetsByType.Commodity.SumOfAssets != 0 {
		futuresWithCommodityBase := portfolio.FuturesAssets.AssetsByType.Commodity.SumOfAssets
		futuresWithCommodityBaseInPers := futuresWithCommodityBase / totalAmount * 100

		futuresByCurrencies += "    " + fmt.Sprintf("Фьючерсы на товары стоят: %.2f (%.2f%%от портфеля)\n", futuresWithCommodityBase, futuresWithCommodityBaseInPers)
		for k, v := range portfolio.FuturesAssets.AssetsByType.Commodity.AssetsByCurrency {
			AssetByCurr := utils.RoundFloat(v.SumOfAssets, 2)
			AssetByCurrInPers := utils.RoundFloat(AssetByCurr/totalAmount*100, 2)
			futuresByCurrencies += "      " + fmt.Sprintf("Стоимость фьючерса %s: %.2f (%.2f%%от портфеля)\n", k, AssetByCurr, AssetByCurrInPers)
		}
	}

	if portfolio.FuturesAssets.AssetsByType.Currency.SumOfAssets != 0 {
		futuresWithcurrencyBase := portfolio.FuturesAssets.AssetsByType.Currency.SumOfAssets
		futuresWithcurrencyBaseInPers := futuresWithcurrencyBase / totalAmount * 100

		futuresByCurrencies += "    " + fmt.Sprintf("Фьючерсы на валюты стоят: %.2f (%.2f%%от портфеля)\n", futuresWithcurrencyBase, futuresWithcurrencyBaseInPers)
		for k, v := range portfolio.FuturesAssets.AssetsByType.Currency.AssetsByCurrency {
			AssetByCurr := utils.RoundFloat(v.SumOfAssets, 2)
			AssetByCurrInPers := utils.RoundFloat(AssetByCurr/totalAmount*100, 2)
			futuresByCurrencies += "      " + fmt.Sprintf("Стоимость фьючерса %s: %.2f (%.2f%%от портфеля)\n", k, AssetByCurr, AssetByCurrInPers)
		}
	}

	if portfolio.FuturesAssets.AssetsByType.Security.SumOfAssets != 0 {
		futuresWithcurrencyBase := portfolio.FuturesAssets.AssetsByType.Security.SumOfAssets
		futuresWithcurrencyBaseInPers := futuresWithcurrencyBase / totalAmount * 100

		futuresByCurrencies += "    " + fmt.Sprintf("Фьючерсы на акции стоят: %.2f (%.2f%%от портфеля)\n", futuresWithcurrencyBase, futuresWithcurrencyBaseInPers)
		for k, v := range portfolio.FuturesAssets.AssetsByType.Security.AssetsByCurrency {
			AssetByCurr := utils.RoundFloat(v.SumOfAssets, 2)
			AssetByCurrInPers := utils.RoundFloat(AssetByCurr/totalAmount*100, 2)
			futuresByCurrencies += "      " + fmt.Sprintf("Стоимость фьючерсов в %s: %.2f (%.2f%%от портфеля)\n", k, AssetByCurr, AssetByCurrInPers)
		}
	}

	if portfolio.FuturesAssets.AssetsByType.Index.SumOfAssets != 0 {
		futuresWithcurrencyBase := portfolio.FuturesAssets.AssetsByType.Index.SumOfAssets
		futuresWithcurrencyBaseInPers := futuresWithcurrencyBase / totalAmount * 100

		futuresByCurrencies += "    " + fmt.Sprintf("Фьючерсы на индексы стоят: %.2f (%.2f%%от портфеля)\n", futuresWithcurrencyBase, futuresWithcurrencyBaseInPers)
		futuresByCurrencies += "    " + "Фьючерсы на индексы\n"
		for k, v := range portfolio.FuturesAssets.AssetsByType.Index.AssetsByCurrency {
			AssetByCurr := utils.RoundFloat(v.SumOfAssets, 2)
			AssetByCurrInPers := utils.RoundFloat(AssetByCurr/totalAmount*100, 2)
			futuresByCurrencies += "      " + fmt.Sprintf("Стоимость фьючерса %s: %.2f (%.2f%%от портфеля)\n", k, AssetByCurr, AssetByCurrInPers)
		}
	}

	var currenciesByCurrencies string
	for k, v := range portfolio.CurrenciesAssets.AssetsByCurrency {
		AssetByCurr := utils.RoundFloat(v.SumOfAssets, 2)
		AssetByCurrInPers := utils.RoundFloat(AssetByCurr/totalAmount*100, 2)
		currenciesByCurrencies += "      " + fmt.Sprintf("Стоимость валюты %s: %.2f (%.2f%%от портфеля)\n", k, AssetByCurr, AssetByCurrInPers)
	}

	output += totalAmountOut +
		totalBondsOut +
		bondsByCurrencies +
		totalSharesOut +
		sharesByCurrencies +
		totalEtfsOut +
		etfsByCurrencies +
		totalFuturesOut +
		futuresByCurrencies +
		totalCurrenciesOut +
		currenciesByCurrencies +
		totalLeverageRatioOut

	return accountTitle + output
}
