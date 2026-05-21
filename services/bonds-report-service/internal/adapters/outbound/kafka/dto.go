package kafka

import (
	"bonds-report-service/internal/application/dto"
	"errors"
)

const (
	ReportGenerated = "report.generated"
	ReportFailed    = "report.failed"
	ReportRequested = "report.requested"
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

func NewRequestReportGenerated(reportKind, chatID, traceID string, bondReportRequestBody BondReportsRequestBody) RequestReportGenerated {
	return RequestReportGenerated{
		ReportKind:             reportKind,
		ChatID:                 chatID,
		TraceID:                traceID,
		BondReportsRequestBody: bondReportRequestBody,
	}
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

func mapBondReportsResponseToBondResponceBody(src dto.BondReportsResponce) BondReportsRequestBody {
	res := BondReportsRequestBody{
		Media: make([][]*MediaGroup, 0, len(src.Media)),
	}

	for _, mediaRow := range src.Media {
		res.Media = append(res.Media, mapMediaGroupRow(mediaRow))
	}

	return res
}

func mapMediaGroupRow(src []*dto.MediaGroup) []*MediaGroup {
	res := make([]*MediaGroup, 0, len(src))

	for _, group := range src {
		res = append(res, mapMediaGroup(group))
	}

	return res
}

func mapMediaGroup(src *dto.MediaGroup) *MediaGroup {
	if src == nil {
		return nil
	}

	res := NewMediaGroup()
	if len(src.Reports) == 0 {
		return res
	}

	for _, report := range src.Reports {
		res.Reports = append(res.Reports, mapImageData(report))
	}

	return res
}

func mapImageData(src *dto.ImageData) *ImageData {
	if src == nil {
		return nil
	}

	res := NewImageData()
	res.Name = src.Name
	res.Data = append([]byte(nil), src.Data...)
	res.Caption = src.Caption

	return res
}
