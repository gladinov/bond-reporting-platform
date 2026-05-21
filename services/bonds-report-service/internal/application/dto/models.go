package dto

type BondReportsResponce struct {
	Media [][]*MediaGroup
}

type AccountListResponce struct {
	Accounts string
}

type MediaGroup struct {
	Reports []*ImageData
}

func NewMediaGroup() *MediaGroup {
	return &MediaGroup{
		Reports: make([]*ImageData, 0),
	}
}

type ImageData struct {
	Name    string
	Data    []byte
	Caption string
}

func NewImageData() *ImageData {
	return &ImageData{}
}

// type ResponceReportFailed struct {
// 	ReportKind   string
// 	ChatID       string
// 	TraceID      string
// 	ErrorCode    string
// 	ErrorMessage string
// 	// Retraible bool // TODO: Добавить поле о ретрае
// }

// func NewRepsponceReportFailed(reportKind, chatID, traceID, errorCode, errorMessage string) ResponceReportFailed {
// 	return ResponceReportFailed{
// 		ReportKind:   reportKind,
// 		ChatID:       chatID,
// 		TraceID:      traceID,
// 		ErrorCode:    errorCode,
// 		ErrorMessage: errorMessage,
// 	}
// }
