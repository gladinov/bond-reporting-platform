package kafka

import (
	"errors"

	"github.com/gladinov/notification-service/internal/application/usecases"
)

const (
	ReportGenerated = "report.generated"
	ReportFailed    = "report.failed"
)

var (
	ErrEmptyChatID     = errors.New("chatID is empty")
	ErrEmptyTraceID    = errors.New("traceID is empty")
	ErrEmptyReportKind = errors.New("reportKind is empty")
	ErrEmptyErr        = errors.New("error is nil in failed report from kafka")
)

type ImageData struct {
	Name    string `json:"name"`
	Data    []byte `json:"data"`
	Caption string `json:"caption"`
}

func NewImageData() *ImageData {
	return &ImageData{}
}

type BondReportsRequestBody struct {
	Media [][]*MediaGroup `json:"media"`
}

type MediaGroup struct {
	Reports []*ImageData `json:"reports"`
}

func NewMediaGroup() *MediaGroup {
	return &MediaGroup{
		Reports: make([]*ImageData, 0),
	}
}

type RequestReportGenerated struct {
	ReportKind             string                 `json:"reportkind"`
	ChatID                 string                 `json:"chatid"` // TODO: Можно ли передавать заголовки в контексте?
	TraceID                string                 `json:"traceid"`
	BondReportsRequestBody BondReportsRequestBody `json:"bondreportresponce"`
}

func (r *RequestReportGenerated) Validate() error {
	if r.ReportKind == "" {
		return ErrEmptyReportKind
	}
	if r.ChatID == "" {
		return ErrEmptyChatID
	}
	if r.TraceID == "" {
		return ErrEmptyTraceID
	}
	return nil
}

func (r *RequestReportGenerated) ResponceReportGeneratedToUsecaseDto() usecases.ReportGenerated {
	return usecases.ReportGenerated{
		ReportKind:          r.ReportKind,
		ChatID:              r.ChatID,
		TraceID:             r.TraceID,
		BondReportsResponce: mapBondReportsResponseToUsecase(r.BondReportsRequestBody),
	}
}

type ResponceReportFailed struct {
	ReportKind   string `json:"reportkind"`
	ChatID       string `json:"chatid"` // TODO: Можно ли передавать заголовки в контексте?
	TraceID      string `json:"traceid"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	// Retraible bool // TODO: Добавить поле о ретрае
}

func NewRepsponceReportFailed(reportKind, chatID, traceID, errorCode, errorMessage string) ResponceReportFailed {
	return ResponceReportFailed{
		ReportKind:   reportKind,
		ChatID:       chatID,
		TraceID:      traceID,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	}
}

func (r *ResponceReportFailed) Validate() error {
	if r.ReportKind == "" {
		return ErrEmptyReportKind
	}
	if r.ChatID == "" {
		return ErrEmptyChatID
	}
	if r.TraceID == "" {
		return ErrEmptyTraceID
	}
	if r.ErrorCode == "" {
		return ErrEmptyErr
	}
	if r.ErrorMessage == "" {
		return ErrEmptyErr
	}
	return nil
}

func (r *ResponceReportFailed) ResponceReportFailedToUsecaseDto() usecases.ReportFailed {
	return usecases.ReportFailed{
		ReportKind:   r.ReportKind,
		ChatID:       r.ChatID,
		TraceID:      r.TraceID,
		ErrorCode:    r.ErrorCode,
		ErrorMessage: r.ErrorMessage,
	}
}

func mapBondReportsResponseToUsecase(src BondReportsRequestBody) usecases.BondReports {
	res := usecases.BondReports{
		Media: make([][]*usecases.MediaGroup, 0, len(src.Media)),
	}

	for _, mediaRow := range src.Media {
		res.Media = append(res.Media, mapMediaGroupRow(mediaRow))
	}

	return res
}

func mapMediaGroupRow(src []*MediaGroup) []*usecases.MediaGroup {
	res := make([]*usecases.MediaGroup, 0, len(src))

	for _, group := range src {
		res = append(res, mapMediaGroup(group))
	}

	return res
}

func mapMediaGroup(src *MediaGroup) *usecases.MediaGroup {
	if src == nil {
		return nil
	}

	res := usecases.NewMediaGroup()
	if len(src.Reports) == 0 {
		return res
	}

	for _, report := range src.Reports {
		res.Reports = append(res.Reports, mapImageData(report))
	}

	return res
}

func mapImageData(src *ImageData) *usecases.ImageData {
	if src == nil {
		return nil
	}

	res := usecases.NewImageData()
	res.Name = src.Name
	res.Data = append([]byte(nil), src.Data...)
	res.Caption = src.Caption

	return res
}
