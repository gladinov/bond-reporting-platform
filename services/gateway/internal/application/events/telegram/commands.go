package telegram

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	contextkeys "github.com/gladinov/contracts/context"
	"github.com/gladinov/contracts/trace"
	"github.com/gladinov/e"
	tokenauth "main.go/internal/application/tokenAuth"
	"main.go/internal/utils/logging"
)

const (
	HelpCmd                    = "/help"
	StartCmd                   = "/start"
	AccountsCmd                = "/accounts"
	GetBondReport              = "/bondfifo"
	GetGeneralBondReport       = "/bondreport"
	GetUSD                     = "/usd"
	GetPortfolioStructure      = "/portfoliostructure"
	GetUnionPortfolioStructure = "/unionportfoliostructure"
	GetUnionWithSber           = "/unionpswithsber"
)

const (
	BondReportsWithPngKind = "bondReportsWithPng"
)

const (
	ReportRequested = "report.requested"
)

type TokenStatus int

const (
	TokenError TokenStatus = iota
	TokenFound
	TokenInserted
)

var constList = []string{
	HelpCmd,
	StartCmd,
	AccountsCmd,
	GetBondReport,
	GetGeneralBondReport,
	GetUSD,
	GetPortfolioStructure,
	GetUnionPortfolioStructure,
	GetUnionWithSber,
}

func ContainsInConstantCommands(text string) bool {
	for _, v := range constList {
		if text == v {
			return true
		}
	}
	return false
}

var ErrIncorrectToken = errors.New("incorrect token")

func (p *Processor) doCmd(ctx context.Context, text string, chatID int, username string) (err error) {
	const op = "telegram.doCmd"

	logg := p.logger.With(
		slog.String("op", op),
		slog.String("username", username),
		slog.Int("chatID", chatID),
	)

	defer func() {
		logg.InfoContext(ctx, "finished")
		logging.LogOperation_Debug(ctx, logg, op, &err)()
	}()

	text = strings.TrimSpace(text)

	if ContainsInConstantCommands(text) {
		logg.InfoContext(ctx, "got new command",
			slog.String("msg", text),
		)
	} else {
		logg.InfoContext(ctx, "got new other command")
	}

	chatIDStr := strconv.Itoa(chatID)
	ctx = context.WithValue(ctx, contextkeys.ChatIDKey, chatIDStr)

	// TODO: Написать разные приветствия в зависимости от наличия токена
	if text == StartCmd {
		return p.sendHello(ctx, chatID)
	}

	err = p.handleTokenAuth(ctx, text, username, chatID)
	if err != nil {
		return err
	}

	switch text {
	case HelpCmd:
		return p.sendHelp(ctx, chatID)
	case AccountsCmd:
		return p.sendAccounts(ctx, chatID)
	case GetUSD:
		return p.getUSD(ctx, chatID)
	case GetBondReport:
		return p.getBondReports(ctx, chatID)
	case GetGeneralBondReport:
		return p.fetchBondRepotsWithPng(ctx, chatID)
	case GetPortfolioStructure:
		return p.GetPortfolioStructure(ctx, chatID)
	case GetUnionPortfolioStructure:
		return p.GetUnionPortfolioStructure(ctx, chatID)
	case GetUnionWithSber:
		return p.GetUnionPortfolioStructureWithSber(ctx, chatID)
	default:
		return p.tg.SendMessage(ctx, chatID, msgUnknownCommand)
	}
}

func (p *Processor) handleTokenAuth(ctx context.Context, text string, username string, chatID int) error {
	tokenStatus, err := p.tokenAuthService.Auth(ctx, text, username)

	switch {
	case err != nil && !errors.Is(err, tokenauth.ErrIncorrectToken):
		return err
	case errors.Is(err, tokenauth.ErrIncorrectToken):
		return p.tg.SendMessage(ctx, chatID, msgNoToken)
	case err == nil:
		switch tokenStatus {
		case tokenauth.TokenInserted:
			return p.tg.SendMessage(ctx, chatID, msgTrueToken)
		}
	}
	return nil
}

func (p *Processor) getUSD(ctx context.Context, chatId int) error {
	usdResponce, err := p.bondReportService.GetUsd(ctx)
	if err != nil {
		return e.WrapIfErr("can't get usd", err)
	}
	usd := strconv.FormatFloat(usdResponce.Usd, 'f', 5, 64)
	p.tg.SendMessage(ctx, chatId, usd)
	return nil
}

func (p *Processor) sendAccounts(ctx context.Context, chatID int) error {
	accountsResponce, err := p.bondReportService.GetAccountsList(ctx)
	if err != nil {
		return e.WrapIfErr("can't get account", err)
	}
	accounts := accountsResponce.Accounts
	p.tg.SendMessage(ctx, chatID, accounts)
	return nil
}

func (p *Processor) getBondReports(ctx context.Context, chatID int) (err error) {
	if err = p.bondReportService.GetBondReportsByFifo(ctx); err != nil {
		return e.WrapIfErr("getBondReport: can't get Bond reports", err)
	}

	p.tg.SendMessage(ctx, chatID, "Отчет по облигациям по методу FIFO успешно сохранен в базу данных")
	return nil
}

func (p *Processor) getBondRepotsWithPng(ctx context.Context, chatID int) (err error) {
	bondReportsResponce, err := p.bondReportService.GetBondReports(ctx)
	if err != nil {
		return e.WrapIfErr("can't get bond report with png", err)
	}
	reportsInByteByAccount := bondReportsResponce.Media
	for _, reportsInByte := range reportsInByteByAccount {
		for _, v := range reportsInByte {
			switch len(v.Reports) {
			case 0:
				continue
			case 1:
				err = p.tg.SendImageFromBuffer(ctx, chatID, v.Reports[0].Data, v.Reports[0].Caption)
				if err != nil {
					return e.WrapIfErr("can't get bond report with png", err)
				}
			default:
				err = p.tg.SendMediaGroupFromBuffer(ctx, chatID, v.Reports)
				if err != nil {
					return e.WrapIfErr("can't get bond report with png", err)
				}
			}
		}
	}
	return nil
}

func (p *Processor) fetchBondRepotsWithPng(ctx context.Context, chatID int) (err error) {
	const op = "processor.requestBondRepotsWithPng"
	logging.LogOperation_Debug(ctx, p.logger, op, &err)

	chatIDStr := strconv.Itoa(chatID)

	traceID, ok := trace.TraceIDFromContext(ctx)
	if !ok {
		p.logger.WarnContext(ctx, "hasn't traceID in ctx")
	}

	err = p.kafka.PublishRequest(ctx, BondReportsWithPngKind, chatIDStr, traceID)
	if err != nil {
		return p.tg.SendMessage(ctx, chatID, msgInternalErr)
	}
	// TODO: Добавить отдельный генератор ID запроса за место traceID
	return p.tg.SendMessage(ctx, chatID, msgKafka(traceID))
}

func (p *Processor) GetPortfolioStructure(ctx context.Context, chatID int) (err error) {
	portfolioStructures, err := p.bondReportService.GetPortfolioStructure(ctx)
	if err != nil {
		return e.WrapIfErr("can't get portfolio structure", err)
	}
	for _, report := range portfolioStructures.PortfolioStructures {
		p.tg.SendMessage(ctx, chatID, report)
	}
	return nil
}

func (p *Processor) GetUnionPortfolioStructure(ctx context.Context, chatID int) (err error) {
	unionPortfolioStructureResponce, err := p.bondReportService.GetUnionPortfolioStructure(ctx)
	if err != nil {
		return e.WrapIfErr("processor: can't get union portfolio structure", err)
	}
	unionPortfolioStructure := unionPortfolioStructureResponce.Report
	p.tg.SendMessage(ctx, chatID, unionPortfolioStructure)
	return nil
}

func (p *Processor) GetUnionPortfolioStructureWithSber(ctx context.Context, chatID int) (err error) {
	unionPortfolioStructureResponce, err := p.bondReportService.GetUnionPortfolioStructureWithSber(ctx)
	if err != nil {
		return e.WrapIfErr("processor: can't get union portfolio structure", err)
	}
	unionPortfolioStructure := unionPortfolioStructureResponce.Report
	p.tg.SendMessage(ctx, chatID, unionPortfolioStructure)
	return nil
}

func (p *Processor) sendHelp(ctx context.Context, chatID int) error {
	return p.tg.SendMessage(ctx, chatID, msgHelp)
}

func (p *Processor) sendHello(ctx context.Context, chatID int) error {
	return p.tg.SendMessage(ctx, chatID, msgHello)
}
