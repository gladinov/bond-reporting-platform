package cbrHelper

import (
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/utils/logging"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gladinov/e"
)

func (h *CbrHelper) GetCurrencyFromCB(ctx context.Context, charCode string, date time.Time) (vunit_rate float64, err error) {
	const op = "service.GetCurrencyFromCB"

	defer logging.LogOperation_Debug(ctx, h.logger, op, &err)()

	vunit_rate, err = h.Storage.GetCurrency(ctx, charCode, date)
	if err == nil {
		return vunit_rate, nil
	}
	if err != nil && !errors.Is(err, domain.ErrNoCurrency) {
		return vunit_rate, e.WrapIfErr("filed to get currency from storage", err)
	}

	currencies, err := h.Cbr.GetAllCurrencies(ctx, date)
	if err != nil {
		return vunit_rate, e.WrapIfErr("filed to get all currencies from cbr", err)
	}
	charCode = strings.ToLower(charCode)
	if val, ok := currencies.CurrenciesMap[charCode]; ok {
		vunit_rate = val.VunitRate
	} else {
		return 0, fmt.Errorf("currency %s not found", charCode)
	}

	err = h.Storage.SaveCurrency(ctx, currencies, date)
	if err != nil {
		h.logger.Warn("failed to save currency", "error", err)
	}
	return vunit_rate, nil
}
