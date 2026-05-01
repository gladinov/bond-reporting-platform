package reporting

import (
	"bonds-report-service/internal/domain"
	report "bonds-report-service/internal/domain/report"
	report_position "bonds-report-service/internal/domain/report_position"
	"time"

	"github.com/gladinov/e"
)

func CreateBondReport(
	now time.Time,
	currentPositions []report_position.PositionByFIFO,
	moexBuyDateData domain.ValuesMoex,
	moexNowData domain.ValuesMoex,
) (_ report.Report, err error) {
	var resultReports report.Report

	for i := range currentPositions {
		position := &currentPositions[i]
		err := resultReports.Add(position, moexBuyDateData, moexNowData, now)
		if err != nil {
			return report.Report{}, e.WrapIfErr("failed to add result reports", err)
		}
	}
	return resultReports, nil
}
