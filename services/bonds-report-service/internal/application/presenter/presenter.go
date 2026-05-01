package presenter

import (
	"bonds-report-service/internal/application/dto"
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/domain/generalbondreport"
	"bonds-report-service/internal/utils"
	"bonds-report-service/internal/utils/logging"
	"bytes"
	"context"
	"fmt"
	"image/color"
	"image/png"
	"log/slog"
	"sort"
	"strings"
	"sync"

	"github.com/fogleman/gg"
	"github.com/gladinov/e"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

const (
	layout = "2006-01-02"
)

const (
	unionPortf = iota
	unionPortfWithSber
	eachPortf
)

func EncodePNGToBuffer(dc *gg.Context) ([]byte, error) {
	img := dc.Image()
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GenerateTablePNG(ctx context.Context, logger *slog.Logger,
	generalBondReports *generalbondreport.GeneralBondReports,
) (_ []*dto.MediaGroup, err error) {
	const op = "presenter.GenerateTablePNG"

	defer logging.LogOperation_Debug(ctx, logger, op, &err)()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		prepearingReports := prepareToGenerateTablePNG(ctx, logger, generalBondReports)

		reportsInByte := make([]*dto.MediaGroup, 3)
		for i, report := range prepearingReports {
			reportsInByte[i] = dto.NewMediaGroup()
			mediaGroup := reportsInByte[i]
			if len(report) == 0 {
				continue
			}

			var typeOfBonds string
			switch {
			case report[0].Replaced:
				typeOfBonds = domain.ReplacedBonds
			case report[0].Currencies != "rub":
				typeOfBonds = domain.EuroBonds
			default:
				typeOfBonds = domain.RubBonds
			}
			count := 1
			for start := 0; start < len(report); start += 10 {
				end := start + 10
				if end > len(report) {
					end = len(report)
				}
				pngData, err := generateTablePNGInByte(ctx, logger, report[start:end], typeOfBonds)
				if err != nil {
					return nil, e.WrapIfErr("vizualize error", err)
				}
				// TODO: Есть желание вынести все, что связано с отображением в отдельный микросервис. А из этого отправлять только струкутуры данных
				imageData := dto.NewImageData()
				imageData.Name = fmt.Sprintf("file%s_%v", typeOfBonds, count)
				imageData.Data = pngData
				imageData.Caption = typeOfBonds

				mediaGroup.Reports = append(mediaGroup.Reports, imageData)
				count += 1
			}
		}
		return reportsInByte, nil
	}
}

func generateTablePNGInByte(ctx context.Context,
	logger *slog.Logger,
	reports []generalbondreport.GeneralBondReportPosition,
	typeOfBonds string,
) (_ []byte, err error) {
	const op = "presenter.GenerateTablePNG"

	defer logging.LogOperation_Debug(ctx, logger, op, &err)()

	const (
		width        = 1200 // Увеличим ширину для лучшего отображения
		heightPerRow = 40
		margin       = 20.0
		headerHeight = 60
		rowHeight    = 30
		colCount     = 15 // Количество отображаемых колонок
	)

	// Рассчитываем высоту изображения
	height := headerHeight + len(reports)*rowHeight + margin*2
	if height < 300 {
		height = 300
	}

	// Создаем контекст изображения
	dc := gg.NewContext(width, height)

	// Загружаем шрифт
	font, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	face, err := opentype.NewFace(font, &opentype.FaceOptions{
		Size: 12,
		DPI:  72,
	})
	if err != nil {
		return nil, err
	}
	dc.SetFontFace(face)

	// Заливаем фон
	dc.SetColor(color.White)
	dc.Clear()

	// Рисуем заголовок
	dc.SetColor(color.Black)
	var name string
	switch typeOfBonds {
	case domain.ReplacedBonds:
		name = "Отчет по замещающим облигациям"
	case domain.RubBonds:
		name = "Отчет по рублевым облигациям"
	case domain.EuroBonds:
		name = "Отчет по валютным облигациям"
	}
	dc.DrawStringAnchored(name, width/2, margin, 0.5, 0.5)
	dc.SetFontFace(face)

	// Определяем колонки для отображения с более подходящими ширинами
	columns := []struct {
		Title string
		Width float64
		Flex  float64
	}{
		{"Тикер", 120, 0},              // 1
		{"Название", 180, 2},           // 2
		{"Валюта", 60, 0},              // 3
		{"Кол-во", 70, 0},              // 4
		{"% от портфеля", 90, 0},       // 5
		{"Дата погашения", 110, 0},     // 6
		{"Дюрация", 80, 0},             // 7
		{"Дата покупки", 90, 0},        // 8
		{"Ср.цена", 80, 0},             // 9
		{"Дох-ть при покупке", 120, 0}, // 10
		{"Дох-ть тек.", 90, 0},         // 11 (сокращенный заголовок)
		{"Тек.цена", 80, 0},            // 12 (сокращенный заголовок)
		{"Номинал", 80, 0},             // 13
		{"Доход", 80, 0},               // 14
		{"Дох-ть в %", 80, 0},          // 15
	}

	// Проверяем, что сумма минимальных ширин не превышает доступную ширину
	totalMinWidth := 0.0
	for _, col := range columns {
		totalMinWidth += col.Width
	}

	availableWidth := float64(width) - 2*margin
	if totalMinWidth > availableWidth {
		// Если минимальные ширины не помещаются, масштабируем их
		scaleFactor := availableWidth / totalMinWidth
		for i := range columns {
			columns[i].Width *= scaleFactor
		}
	} else {
		// Распределяем оставшееся пространство пропорционально Flex
		remainingWidth := availableWidth - totalMinWidth
		totalFlex := 0.0
		for _, col := range columns {
			totalFlex += col.Flex
		}

		if totalFlex > 0 {
			flexUnit := remainingWidth / totalFlex
			for i := range columns {
				if columns[i].Flex > 0 {
					columns[i].Width += flexUnit * columns[i].Flex
				}
			}
		}
	}

	// Рисование таблицы
	y := float64(headerHeight)
	currentX := margin

	// Заголовки
	for _, col := range columns {
		dc.SetColor(color.RGBA{200, 200, 200, 255})
		dc.DrawRectangle(currentX, y, col.Width, rowHeight)
		dc.Fill()

		dc.SetColor(color.Black)
		// Разбиваем длинные заголовки на несколько строк
		if len(col.Title) > 10 && strings.Contains(col.Title, " ") {
			parts := strings.Split(col.Title, " ")
			if len(parts) == 2 {
				dc.DrawStringAnchored(parts[0], currentX+col.Width/2, y+rowHeight/2-7, 0.5, 0.5)
				dc.DrawStringAnchored(parts[1], currentX+col.Width/2, y+rowHeight/2+7, 0.5, 0.5)
				currentX += col.Width
				continue
			}
		}
		dc.DrawStringAnchored(col.Title, currentX+col.Width/2, y+rowHeight/2, 0.5, 0.5)
		currentX += col.Width
	}

	// Данные
	for _, report := range reports {
		y += rowHeight
		currentX = margin
		dc.SetColor(color.Black)

		// Форматируем данные
		values := []string{
			report.Ticker,
			report.Name,
			report.Currencies,
			formatInt(report.Quantity),
			formatPercent(report.PercentOfPortfolio),
			formatTime(report.MaturityDate),
			formatInt(report.Duration),
			formatTime(report.BuyDate),
			formatFloat(report.PositionPrice),
			formatPercent(report.YieldToMaturityOnPurchase),
			formatPercent(report.YieldToMaturity),
			formatFloat(report.CurrentPrice),
			formatFloat(report.Nominal),
			formatFloat(report.Profit),
			formatPercent(report.ProfitInPercentage),
		}

		for i, col := range columns {
			dc.DrawStringAnchored(
				values[i],
				currentX+col.Width/2,
				y+rowHeight/2,
				0.5, 0.5,
			)
			currentX += col.Width
		}
	}
	pngData, err := EncodePNGToBuffer(dc)
	if err != nil {
		return nil, err
	}

	return pngData, nil
}

func prepareToGenerateTablePNG(ctx context.Context,
	logger *slog.Logger,
	generalBondReports *generalbondreport.GeneralBondReports,
) (_ [][]generalbondreport.GeneralBondReportPosition) {
	const op = "service.PrepareToGenerateTablePNG"

	defer logging.LogOperation_Debug(ctx, logger, op, nil)()
	var wg sync.WaitGroup
	reports := make([][]generalbondreport.GeneralBondReportPosition, 3)

	wg.Add(1)
	go func() {
		defer wg.Done()
		reports[0] = sortGeneralBondReports(ctx, logger, generalBondReports.RubBondsReport)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		reports[1] = sortGeneralBondReports(ctx, logger, generalBondReports.ReplacedBondsReport)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		reports[2] = sortGeneralBondReports(ctx, logger, generalBondReports.EuroBondsReport)
	}()

	wg.Wait()

	return reports
}

func sortGeneralBondReports(ctx context.Context,
	logger *slog.Logger,
	report map[generalbondreport.TickerTimeKey]generalbondreport.GeneralBondReportPosition,
) (_ []generalbondreport.GeneralBondReportPosition) {
	const op = "service.sortGeneralBondReports"

	defer logging.LogOperation_Debug(ctx, logger, op, nil)()

	keys := make([]generalbondreport.TickerTimeKey, 0, len(report))
	for k := range report {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Time.Equal(keys[j].Time) {
			return keys[i].Ticker < keys[j].Ticker
		}
		return keys[i].Time.Before(keys[j].Time)
	})
	result := make([]generalbondreport.GeneralBondReportPosition, len(keys))
	for i, k := range keys {
		result[i] = report[k]
	}

	return result
}

func ResponsePortfolioStructure(ctx context.Context, logger *slog.Logger, portfolio *domain.PortfolioByTypeAndCurrency, title int, accountName string) string {
	const op = "service.ResponsePortfolioStructure"
	// TODO: возможно! Корректней будет собирать строку через strings.Builder{}
	defer logging.LogOperation_Debug(ctx, logger, op, nil)()
	var accountTitle string
	switch title {
	case unionPortf:
		accountTitle = "Струтура всех открытых счетов в Тинькофф Инвестициях\n"
	case unionPortfWithSber:
		accountTitle = "Струтура всех инвестиций\n"
	case eachPortf:
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

	response := accountTitle + output

	return response
}

func GetAccount(accs map[string]domain.Account) dto.AccountListResponce {
	var accStr string = "По данному аккаунту доступны следующие счета:"
	for _, account := range accs {
		accStr += fmt.Sprintf("\n ID:%s, Type: %s, Name: %s, Status: %v \n", account.ID, account.Type, account.Name, account.Status)
	}
	accountResponce := dto.AccountListResponce{Accounts: accStr}
	return accountResponce
}
