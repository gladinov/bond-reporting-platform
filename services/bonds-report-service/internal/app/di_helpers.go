package app

import (
	cbrHelper "bonds-report-service/internal/application/services/currency"
	moexHelper "bonds-report-service/internal/application/services/moex"
	updateoperations "bonds-report-service/internal/application/services/operations"
	dividerbyassettype "bonds-report-service/internal/application/services/portfolio"
	positionProcessor "bonds-report-service/internal/application/services/positionprocessor"
	reportlinebuiler "bonds-report-service/internal/application/services/reportline"
	"bonds-report-service/internal/application/services/uids"
	"bonds-report-service/internal/application/usecases"
)

func (d *diContainer) CBRCurrencyGetter() usecases.CbrCurrencyGetter {
	if d.cbrCurrencyGetter == nil {
		d.logger.Info("initialize cbr currency getter")
		d.cbrCurrencyGetter = cbrHelper.NewCbrHelper(d.logger, d.CBRClient(), d.Storage())
	}

	return d.cbrCurrencyGetter
}

func (d *diContainer) MoexSpecificationGetter() usecases.MoexSpecificationGetter {
	if d.moexSpecificationGetter == nil {
		d.logger.Info("initialize moex specification getter")
		d.moexSpecificationGetter = moexHelper.NewMoexHelper(d.logger, d.MoexClient())
	}

	return d.moexSpecificationGetter
}

func (d *diContainer) UidProvider() positionProcessor.UidProvider {
	if d.uidProvider == nil {
		d.logger.Info("initialize uid provider")
		d.uidProvider = uidprovider.NewUidProvider(d.Storage(), d.TinkoffAnalyticsClient())
	}

	return d.uidProvider
}

func (d *diContainer) OperationsUpdater() usecases.OperationsUpdater {
	if d.operationsUpdater == nil {
		d.logger.Info("initialize operations updater")
		d.operationsUpdater = updateoperations.NewUpdater(d.logger, d.Storage(), d.TinkoffHelper())
	}

	return d.operationsUpdater
}

func (d *diContainer) PositionProcessor() usecases.PositionProcessor {
	if d.positionProcessor == nil {
		d.logger.Info("initialize position processor")
		d.positionProcessor = positionProcessor.NewProcessor(d.logger, d.UidProvider())
	}

	return d.positionProcessor
}

func (d *diContainer) ReportLineBuilder() usecases.ReportLineBuilder {
	if d.reportLineBuilder == nil {
		d.logger.Info("initialize report line builder")
		d.reportLineBuilder = reportlinebuiler.NewReportLineBuilder(d.logger, d.TinkoffHelper(), d.CBRCurrencyGetter())
	}

	return d.reportLineBuilder
}

func (d *diContainer) DividerByAssetType() usecases.DividerByAssetType {
	if d.dividerByAssetType == nil {
		d.logger.Info("initialize divider by asset type")
		d.dividerByAssetType = dividerbyassettype.NewDividerByAssetType(
			d.logger,
			d.TinkoffHelper(),
			d.CBRCurrencyGetter(),
			d.config.WorkersNubmer,
		)
	}

	return d.dividerByAssetType
}

func (d *diContainer) Helpers() *usecases.Helpers {
	if d.helpers == nil {
		d.helpers = usecases.NewHelpers(
			d.CBRCurrencyGetter(),
			d.MoexSpecificationGetter(),
			d.TinkoffHelper(),
			d.OperationsUpdater(),
			d.PositionProcessor(),
			d.ReportLineBuilder(),
			d.DividerByAssetType(),
		)
	}

	return d.helpers
}
