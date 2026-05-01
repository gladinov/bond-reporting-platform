package handlers

import (
	httpmodels "bonds-report-service/internal/adapters/inbound/gateway/http"
	gatewaypresenter "bonds-report-service/internal/adapters/inbound/gateway/presenter"
	"bonds-report-service/internal/application/dto"
	"bonds-report-service/internal/domain"
)

func MapMediaGroupToHTTP(mg *dto.MediaGroup) *httpmodels.MediaGroup {
	if mg == nil {
		return nil
	}

	httpMG := httpmodels.NewMediaGroup()
	for _, r := range mg.Reports {
		httpMG.Reports = append(httpMG.Reports, MapImageDataToHTTP(r))
	}
	return httpMG
}

func MapImageDataToHTTP(img *dto.ImageData) *httpmodels.ImageData {
	if img == nil {
		return nil
	}
	return &httpmodels.ImageData{
		Name:    img.Name,
		Data:    img.Data,
		Caption: img.Caption,
	}
}

func MapAccountListToHTTP(acc *dto.AccountListResponce) *httpmodels.AccountListResponce {
	if acc == nil {
		return nil
	}
	return &httpmodels.AccountListResponce{
		Accounts: acc.Accounts,
	}
}

func MapBondReportsToHTTP(br *dto.BondReportsResponce) *httpmodels.BondReportsResponce {
	if br == nil {
		return nil
	}

	httpBR := &httpmodels.BondReportsResponce{
		Media: make([][]*httpmodels.MediaGroup, len(br.Media)),
	}

	for i, row := range br.Media {
		httpBR.Media[i] = make([]*httpmodels.MediaGroup, len(row))
		for j, mg := range row {
			httpBR.Media[i][j] = MapMediaGroupToHTTP(mg)
		}
	}

	return httpBR
}

func MapPortfolioStructureForEachAccountToHTTP(pf *domain.PortfolioStructureForEachAccountResponce) *httpmodels.PortfolioStructureForEachAccountResponce {
	if pf == nil {
		return nil
	}
	portfolioStructures := make([]string, 0, len(pf.PortfolioStructures))
	for _, portfolioStructure := range pf.PortfolioStructures {
		portfolioStructures = append(portfolioStructures, gatewaypresenter.ResponsePortfolioStructure(
			portfolioStructure.Portfolio,
			gatewaypresenter.EachPortfolio,
			portfolioStructure.AccountName,
		))
	}

	return &httpmodels.PortfolioStructureForEachAccountResponce{
		PortfolioStructures: portfolioStructures,
	}
}

func MapUnionPortfolioStructureToHTTP(u *domain.UnionPortfolioStructureResponce) *httpmodels.UnionPortfolioStructureResponce {
	if u == nil {
		return nil
	}
	return &httpmodels.UnionPortfolioStructureResponce{
		Report: gatewaypresenter.ResponsePortfolioStructure(u.Portfolio, gatewaypresenter.UnionPortfolio, ""),
	}
}

func MapUnionPortfolioStructureWithSberToHTTP(u *domain.UnionPortfolioStructureWithSberResponce) *httpmodels.UnionPortfolioStructureWithSberResponce {
	if u == nil {
		return nil
	}
	return &httpmodels.UnionPortfolioStructureWithSberResponce{
		Report: gatewaypresenter.ResponsePortfolioStructure(u.Portfolio, gatewaypresenter.UnionPortfolioWithSber, ""),
	}
}

func MapUsdToHTTP(u *domain.UsdResponce) *httpmodels.UsdResponce {
	if u == nil {
		return nil
	}
	return &httpmodels.UsdResponce{
		Usd: u.Usd,
	}
}
