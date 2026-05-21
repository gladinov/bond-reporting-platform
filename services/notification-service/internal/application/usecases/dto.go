package usecases

const (
	BondReportsWithPngKind = "bondReportsWithPng"
)

type ReportGenerated struct {
	ReportKind          string
	ChatID              string
	TraceID             string
	BondReportsResponce BondReports
}

type ReportFailed struct {
	ReportKind   string
	ChatID       string
	TraceID      string
	ErrorCode    string
	ErrorMessage string
}

type BondReports struct {
	Media [][]*MediaGroup
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
