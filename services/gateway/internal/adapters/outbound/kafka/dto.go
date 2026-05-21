package kafka

const (
	ReportGenerated = "report.generated"
	ReportFailed    = "report.failed"
	ReportRequested = "report.requested"
)

type Request struct {
	ReportKind string `json:"reportkind"`
	TraceID    string `json:"traceID"`
	ChatID     string `json:"chatID"`
}

func NewRequest(reportKind string, traceID string, chatID string) Request {
	return Request{
		ReportKind: reportKind,
		TraceID:    traceID,
		ChatID:     chatID,
	}
}
