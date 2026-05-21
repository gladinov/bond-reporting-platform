package usecases

import (
	"context"
	"strconv"
)

func (s *Service) SendReportFailed(ctx context.Context, body ReportFailed) error {
	chatID, err := strconv.Atoi(body.ChatID)
	if err != nil {
		return err
	}
	return s.telegram.SendMessage(ctx, chatID, body.ErrorMessage)
}
