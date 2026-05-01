package tinkoffHelper

import (
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/utils/logging"
	"context"

	"github.com/gladinov/e"
)

func (h *TinkoffHelper) TinkoffGetAccounts(ctx context.Context) (_ map[string]domain.Account, err error) {
	const op = "service.TinkoffGetAccounts"

	defer logging.LogOperation_Debug(ctx, h.logger, op, &err)()

	accs, err := h.Portfolio.GetAccounts(ctx)
	if err != nil {
		return nil, e.WrapIfErr("failed to get accounts", err)
	}
	return accs, nil
}

func (h *TinkoffHelper) TinkoffGetPortfolio(ctx context.Context, account domain.Account) (_ domain.Portfolio, err error) {
	const op = "service.TinkoffGetPortfolio"

	defer logging.LogOperation_Debug(ctx, h.logger, op, &err)()

	err = account.ValidateForPortfolio()
	if err != nil {
		return domain.Portfolio{}, e.WrapIfErr("failed to validate account", err)
	}
	portfolio, err := h.Portfolio.GetPortfolio(ctx, account.ID, account.Status)
	if err != nil {
		return domain.Portfolio{}, e.WrapIfErr("failed to get portfolio", err)
	}
	return portfolio, nil
}

func (h *TinkoffHelper) TinkoffGetOperations(ctx context.Context, opRequest domain.OperationsRequest) (_ []domain.Operation, err error) {
	const op = "service.TinkoffGetOperations"

	defer logging.LogOperation_Debug(ctx, h.logger, op, &err)()

	if err := opRequest.Validate(h.now()); err != nil {
		return nil, e.WrapIfErr("failed to validate", err)
	}

	tinkoffOperations, err := h.Portfolio.GetOperations(ctx, opRequest.AccountID, opRequest.FromDate)
	if err != nil {
		return nil, e.WrapIfErr("failed get operations from tinkoff", err)
	}
	return tinkoffOperations, nil
}

func (h *TinkoffHelper) TinkoffGetBondActions(ctx context.Context, instrumentUid string) (_ domain.BondIdentIdentifiers, err error) {
	const op = "service.TinkoffGetBondActions"

	defer logging.LogOperation_Debug(ctx, h.logger, op, &err)()

	if instrumentUid == "" {
		return domain.BondIdentIdentifiers{}, domain.ErrEmptyInstrumentUid
	}
	bondActions, err := h.Analytics.GetBondsActions(ctx, instrumentUid)
	if err != nil {
		return domain.BondIdentIdentifiers{}, e.WrapIfErr("failed get bond actions from tinkoff", err)
	}
	return bondActions, nil
}

func (h *TinkoffHelper) TinkoffGetFutureBy(ctx context.Context, figi string) (_ domain.Future, err error) {
	const op = "service.TinkoffGetFutureBy"

	defer logging.LogOperation_Debug(ctx, h.logger, op, &err)()

	if figi == "" {
		return domain.Future{}, domain.ErrEmptyFigi
	}
	future, err := h.Instruments.GetFutureBy(ctx, figi)
	if err != nil {
		return domain.Future{}, e.WrapIfErr("failed get future by from tinkoff", err)
	}
	return future, nil
}

func (h *TinkoffHelper) TinkoffGetBondByUid(ctx context.Context, uid string) (_ domain.Bond, err error) {
	const op = "service.TinkoffGetBondByUid"

	defer logging.LogOperation_Debug(ctx, h.logger, op, &err)()

	if uid == "" {
		return domain.Bond{}, domain.ErrEmptyUid
	}
	bond, err := h.Instruments.GetBondByUid(ctx, uid)
	if err != nil {
		return domain.Bond{}, e.WrapIfErr("failed get bond by uid from tinkoff", err)
	}
	return bond, nil
}

func (h *TinkoffHelper) TinkoffGetCurrencyBy(ctx context.Context, figi string) (_ domain.Currency, err error) {
	const op = "service.TinkoffGetCurrencyBy"

	defer logging.LogOperation_Debug(ctx, h.logger, op, &err)()

	if figi == "" {
		return domain.Currency{}, domain.ErrEmptyFigi
	}
	currency, err := h.Instruments.GetCurrencyBy(ctx, figi)
	if err != nil {
		return domain.Currency{}, e.WrapIfErr("failed get currency by from tinkoff", err)
	}
	return currency, nil
}

func (h *TinkoffHelper) TinkoffGetBaseShareFutureValute(ctx context.Context, positionUid string) (_ domain.ShareCurrency, err error) {
	const op = "service.TinkoffGetBaseShareFutureValute"

	defer logging.LogOperation_Debug(ctx, h.logger, op, &err)()

	if positionUid == "" {
		return domain.ShareCurrency{}, domain.ErrEmptyPositionUid
	}

	instrumentsShortResponce, err := h.Instruments.FindBy(ctx, positionUid)
	if err != nil {
		return domain.ShareCurrency{}, e.WrapIfErr("failed find by from tinkoff", err)
	}

	instrument, err := instrumentsShortResponce.ValidateAndGetFirstShare()
	if err != nil {
		return domain.ShareCurrency{}, e.WrapIfErr("failed to validate and get first share", err)
	}

	currency, err := h.Instruments.GetShareCurrencyBy(ctx, instrument.Figi)
	if err != nil {
		return domain.ShareCurrency{}, e.WrapIfErr("failed get share future valute by from tinkoff", err)
	}

	return currency, nil
}

func (h *TinkoffHelper) TinkoffFindBy(ctx context.Context, query string) (_ domain.InstrumentShortList, err error) {
	const op = "service.TinkoffFindBy"

	defer logging.LogOperation_Debug(ctx, h.logger, op, &err)()

	if query == "" {
		return nil, domain.ErrEmptyQuery
	}
	resp, err := h.Instruments.FindBy(ctx, query)
	if err != nil {
		return nil, e.WrapIfErr("failed find by from tinkoff", err)
	}
	return resp, nil
}

func (h *TinkoffHelper) TinkoffGetLastPriceInPersentageToNominal(ctx context.Context, instrumentUid string) (_ domain.LastPrice, err error) {
	const op = "service.TinkoffGetLastPriceInPersentageToNominal"

	defer logging.LogOperation_Debug(ctx, h.logger, op, &err)()

	if instrumentUid == "" {
		return domain.LastPrice{}, domain.ErrEmptyInstrumentUid
	}
	lastPrice, err := h.Analytics.GetLastPriceInPersentageToNominal(ctx, instrumentUid)
	if err != nil {
		return domain.LastPrice{}, e.WrapIfErr("failed get last price in persentage to nominal from tinkoff", err)
	}
	return lastPrice, nil
}
