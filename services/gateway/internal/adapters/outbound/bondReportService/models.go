package bondreportservice

type MediaGroup struct {
	Reports []*ImageData `json:"reports"`
}

func NewMediaGroup() *MediaGroup {
	return &MediaGroup{
		Reports: make([]*ImageData, 0),
	}
}

type ImageData struct {
	Name    string `json:"name"`
	Data    []byte `json:"data"`
	Caption string `json:"caption"`
}

func NewImageData() *ImageData {
	return &ImageData{}
}

type AccountListResponce struct {
	Accounts string `json:"accounts,omitempty"`
}

type BondReportsRequest struct {
	ChatID int `json:"chatID,omitempty"`
}

type UsdResponce struct {
	Usd float64 `json:"usd,omitempty"`
}

type BondReportsResponce struct {
	Media [][]*MediaGroup `json:"media"`
}

type PortfolioStructureForEachAccountResponce struct {
	PortfolioStructures []string `json:"potftfolio"`
}

type UnionPortfolioStructureResponce struct {
	Report string `json:"report"`
}

type UnionPortfolioStructureWithSberResponce struct {
	Report string `json:"report"`
}
