package reporting

import (
	"bonds-report-service/internal/domain"
	report "bonds-report-service/internal/domain/report_position"
	"context"
	"errors"

	"github.com/gladinov/e"
)

func ProcessOperations(ctx context.Context, reportLine *domain.ReportLine) (*report.ReportPositions, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		processPosition := report.NewReportPositons()

		for _, operation := range reportLine.Operation {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				if err := processPosition.Apply(
					operation,
					reportLine.Bond,
					reportLine.LastPrice,
					reportLine.Vunit_rate); err != nil {
					if errors.Is(err, report.ErrUnknownOpp) {
						continue
					}
					if errors.Is(err, report.ErrZeroQuantity) {
						continue
					}
					return nil, e.WrapIfErr("failed to apply operation to processPosition", err)
				}
			}
		}
		return processPosition, nil
	}
}
