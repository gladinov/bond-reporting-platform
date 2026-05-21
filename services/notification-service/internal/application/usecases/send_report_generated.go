package usecases

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/gladinov/e"
)

func (s *Service) SendReportGenerated(ctx context.Context, body ReportGenerated) error {
	switch body.ReportKind {
	case BondReportsWithPngKind:
		chatID, err := strconv.Atoi(body.ChatID)
		if err != nil {
			return e.WrapIfErr("failed to convert chatID from string to int", err)
		}
		reportsInByteByAccount := body.BondReportsResponce.Media
		for _, reportsInByte := range reportsInByteByAccount {
			for _, v := range reportsInByte {
				switch len(v.Reports) {
				case 0:
					continue
				case 1:
					err = s.telegram.SendImageFromBuffer(ctx, chatID, v.Reports[0].Data, v.Reports[0].Caption)
					if err != nil {
						return e.WrapIfErr("can't get bond report with png", err)
					}
				default:
					s.telegram.SendMediaGroupFromBuffer(ctx, chatID, v.Reports)
					if err != nil {
						return e.WrapIfErr("can't get bond report with png", err)
					}
				}
			}
		}

	default:
		s.logger.WarnContext(ctx, "unknown report kind", slog.Any("reportKind", body.ReportKind), slog.Any("chatID", body.ChatID))
	}
	return nil
}
