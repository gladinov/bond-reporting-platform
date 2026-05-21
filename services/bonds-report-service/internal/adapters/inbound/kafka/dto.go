package kafka

import "errors"

// TODO: создать отдельную библиотеку для хранения общих для двух сервисов костант
const (
	ReportGenerated = "report.generated"
	ReportFailed    = "report.failed"
	ReportRequested = "report.requested"
)

// TODO: Вынести в отдельную библиотеку для взаимодействия между сервисами
const (
	BondReportsWithPngKind = "bondReportsWithPng"
)

var (
	ErrEmptyChatID     = errors.New("chatID is empty")
	ErrEmptyTraceID    = errors.New("traceID is empty")
	ErrEmptyReportKind = errors.New("reportKind is empty")
	ErrEmptyErr        = errors.New("error is nil in failed report from kafka")
)

type kafkaRequest struct {
	ReportKind string `json:"reportkind"`
	TraceID    string `json:"traceID"`
	ChatID     string `json:"chatID"`
}

func (k *kafkaRequest) Validate() error {
	if k.ReportKind == "" {
		return ErrEmptyReportKind
	}
	if k.ChatID == "" {
		return ErrEmptyChatID
	}
	if k.TraceID == "" {
		return ErrEmptyTraceID
	}

	return nil
}
