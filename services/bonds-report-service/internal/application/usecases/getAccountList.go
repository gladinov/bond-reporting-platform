package usecases

import (
	"bonds-report-service/internal/application/dto"
	"bonds-report-service/internal/application/presenter"
	"bonds-report-service/internal/utils/logging"
	"context"

	"github.com/gladinov/e"
)

func (s *Service) GetAccountsList(ctx context.Context) (answ dto.AccountListResponce, err error) {
	const op = "service.GetAccountsList"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	accs, err := s.Helpers.TinkoffProvider.TinkoffGetAccounts(ctx)
	if err != nil {
		return dto.AccountListResponce{}, e.WrapIfErr("can't get accounts from tinkoff", err)
	}

	accountResponce := presenter.GetAccount(accs)
	return accountResponce, nil
}
