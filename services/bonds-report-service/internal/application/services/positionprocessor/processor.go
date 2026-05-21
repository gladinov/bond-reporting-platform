package positionProcessor

import (
	"bonds-report-service/internal/domain"
	"context"
	"errors"
	"log/slog"

	"github.com/gladinov/e"
)

type Processor struct {
	logger      *slog.Logger
	UidProvider UidProvider
}

func NewProcessor(logger *slog.Logger, uidProvider UidProvider) *Processor {
	return &Processor{
		logger:      logger,
		UidProvider: uidProvider,
	}
}

func (p *Processor) ProcessPositionsToPositionsWithAssetUid(ctx context.Context, portffolioPositions []domain.PortfolioPosition) (_ []domain.PortfolioPositionsWithAssetUid, err error) {
	const op = "positionProcessor.TransformPositions"

	portfolio := make([]domain.PortfolioPositionsWithAssetUid, 0, len(portffolioPositions))
	for _, position := range portffolioPositions {
		transformPosition, err := p.processPosition(ctx, position)
		if err != nil {
			return nil, e.WrapIfErr("can't process position", err)
		}
		portfolio = append(portfolio, transformPosition)
	}

	return portfolio, nil
}

func (p *Processor) processPosition(ctx context.Context, position domain.PortfolioPosition) (domain.PortfolioPositionsWithAssetUid, error) {
	const op = "positionProcessor.processPosition"

	assetUid, err := p.UidProvider.GetUid(ctx, position.InstrumentUid)
	if err != nil && !errors.Is(err, domain.ErrEmptyUidAfterUpdate) {
		return domain.PortfolioPositionsWithAssetUid{}, e.WrapIfErr("failed to get uid by instrument uid", err)
	}
	if errors.Is(err, domain.ErrEmptyUidAfterUpdate) {
		assetUid = ""
	}
	transformPosition := domain.PortfolioPositionsWithAssetUid{
		InstrumentType: position.InstrumentType,
		AssetUid:       assetUid,
		InstrumentUid:  position.InstrumentUid,
	}
	return transformPosition, nil
}
