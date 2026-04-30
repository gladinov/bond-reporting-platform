package postgreSQL

import (
	"bonds-report-service/internal/domain"
	"bonds-report-service/internal/domain/generalbondreport"
	"bonds-report-service/internal/domain/report"
	"bonds-report-service/internal/utils/logging"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gladinov/e"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) Close() error {
	s.db.Close()
	return nil
}

func (s *Storage) LastOperationTime(ctx context.Context, chatID int, accountID string) (_ time.Time, err error) {
	const op = "postgreSQL.LastOperationTime"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	q := "SELECT date FROM operations WHERE chatId = $1 AND broker_account_id = $2 ORDER BY DATE DESC LIMIT 1"

	var date time.Time

	err = s.db.QueryRow(ctx, q, chatID, accountID).Scan(&date)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, domain.ErrNoOpperations
		}
		return time.Time{}, err
	}
	return date, nil
}

func (s *Storage) SaveOperations(ctx context.Context, chatID int, accountId string, operations []domain.OperationWithoutCustomTypes) (err error) {
	const op = "postgreSQL.SaveOperations"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	batch := &pgx.Batch{}
	for _, op := range operations {
		batch.Queue(`
        INSERT INTO operations (
            chatId,
            broker_account_id,
            currency,
            operation_id,
            parent_operation_id,
            name,
            date,
            type,
            description,
            instrument_uid,
            figi,
            instrument_type,
            instrument_kind,
            position_uid,
            payment,
            price,
            commission,
            yield,
            yield_relative,
            accrued_int,
            quantity_done,
            asset_uid
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 
            $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, 
            $21, $22
        )`,
			chatID,
			accountId,
			op.Currency,
			op.OperationID,
			op.ParentOperationID,
			op.Name,
			op.Date,
			op.Type,
			op.Description,
			op.InstrumentUid,
			op.Figi,
			op.InstrumentType,
			op.InstrumentKind,
			op.PositionUid,
			op.Payment,
			op.Price,
			op.Commission,
			op.Yield,
			op.YieldRelative,
			op.AccruedInt,
			op.QuantityDone,
			op.AssetUid,
		)
	}
	br := tx.SendBatch(ctx, batch)

	for range operations {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert operation error: %w", err)
		}
	}
	err = br.Close()
	if err != nil {
		return fmt.Errorf("could not close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (s *Storage) GetOperations(ctx context.Context, chatId int, assetUid string, accountId string) (_ []domain.OperationWithoutCustomTypes, err error) {
	const op = "postgreSql.GetOperations"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	q := `select 
    name,
    date,
    type,
    figi,
    operation_id,
    quantity_done,
    instrument_type,
    instrument_uid,
    price,currency,
    accrued_int,
    commission,
    payment
from operations where chatId = $1 and broker_account_id = $2 and asset_uid = $3 order by date`

	var operationRes []domain.OperationWithoutCustomTypes

	rows, err := s.db.Query(ctx, q, chatId, accountId, assetUid)
	if err != nil {
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}
	defer rows.Close()
	for rows.Next() {
		var operation domain.OperationWithoutCustomTypes
		err := rows.Scan(&operation.Name,
			&operation.Date,
			&operation.Type,
			&operation.Figi,
			&operation.OperationID,
			&operation.QuantityDone,
			&operation.InstrumentType,
			&operation.InstrumentUid,
			&operation.Price,
			&operation.Currency,
			&operation.AccruedInt,
			&operation.Commission,
			&operation.Payment)
		if err != nil {
			return nil, fmt.Errorf("%s, row scan: %w", op, err)
		}
		operationRes = append(operationRes, operation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s,rows iteration failed: %w", op, err)
	}

	return operationRes, nil
}

func (s *Storage) GetAllOperations(ctx context.Context, chatId int, accountId string) (_ []domain.OperationWithoutCustomTypes, err error) {
	const op = "postgreSql.GetOperations"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	q := `select 
    name,
	asset_uid,
    date,
    type,
    figi,
    operation_id,
    quantity_done,
    instrument_type,
    instrument_uid,
    price,currency,
    accrued_int,
    commission,
    payment
from operations where chatId = $1 and broker_account_id = $2 order by date`

	var operationRes []domain.OperationWithoutCustomTypes

	rows, err := s.db.Query(ctx, q, chatId, accountId)
	if err != nil {
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}
	defer rows.Close()
	for rows.Next() {
		var operation domain.OperationWithoutCustomTypes
		err := rows.Scan(&operation.Name,
			&operation.AssetUid,
			&operation.Date,
			&operation.Type,
			&operation.Figi,
			&operation.OperationID,
			&operation.QuantityDone,
			&operation.InstrumentType,
			&operation.InstrumentUid,
			&operation.Price,
			&operation.Currency,
			&operation.AccruedInt,
			&operation.Commission,
			&operation.Payment)
		if err != nil {
			return nil, fmt.Errorf("%s, row scan: %w", op, err)
		}
		operationRes = append(operationRes, operation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s,rows iteration failed: %w", op, err)
	}

	return operationRes, nil
}

func (s *Storage) DeleteBondReport(ctx context.Context, chatID int, accountId string) (err error) {
	const op = "postgreSql.DeleteBondReport"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	q := `DELETE FROM bond_reports WHERE chatId = $1 AND broker_account_id = $2`
	if _, err := s.db.Exec(ctx,
		q,
		chatID, accountId); err != nil {
		return err
	}
	return nil
}

func (s *Storage) SaveBondReport(ctx context.Context, chatID int, accountId string, bondReport []report.BondReport) (err error) {
	const op = "postgreSql.SaveBondReport"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	batch := &pgx.Batch{}
	for _, report := range bondReport {
		batch.Queue(`
    INSERT INTO bond_reports (
		chatId,
		broker_account_id,
        name,
        ticker,
        maturity_date,
        offer_date,
        duration,
        buy_date,
        buy_price,
        yield_to_maturity_on_purchase,
        yield_to_offer_on_purchase,
        yield_to_maturity,
        yield_to_offer,
        current_price,
        nominal,
        profit,
        annualized_return
    ) VALUES ($1, $2, $3, $4, NULLIF($5,'')::date, NULLIF($6,'')::date, $7, NULLIF($8,'')::date, $9, $10, $11, $12, $13, $14, $15,$16,$17)
	`,
			chatID,
			accountId,
			report.Name,
			report.Ticker,
			report.MaturityDate,
			report.OfferDate,
			report.Duration,
			report.BuyDate,
			report.BuyPrice,
			report.YieldToMaturityOnPurchase,
			report.YieldToOfferOnPurchase,
			report.YieldToMaturity,
			report.YieldToOffer,
			report.CurrentPrice,
			report.Nominal,
			report.Profit,
			report.AnnualizedReturn)
	}
	br := tx.SendBatch(ctx, batch)

	for range bondReport {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert bond report failed: %w", err)
		}
	}
	err = br.Close()
	if err != nil {
		return fmt.Errorf("could not close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (s *Storage) DeleteGeneralBondReport(ctx context.Context, chatID int, accountId string) (err error) {
	const op = "postgreSql.SaveBondReport"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	q := `DELETE FROM general_bond_report WHERE chatId = $1 AND broker_account_id = $2`

	_, err = s.db.Exec(ctx, q, chatID, accountId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) SaveGeneralBondReport(ctx context.Context, chatID int, accountId string, positions []generalbondreport.GeneralBondReportPosition) (err error) {
	const op = "postgreSql.SaveGeneralBondReport"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	batch := &pgx.Batch{}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for _, pos := range positions {
		batch.Queue(`
			INSERT INTO general_bond_report (
				chatId,
				broker_account_id,
				name,
				ticker,
				currencies,
				quantity,
				percent_of_portfolio,
				maturity_date,
				duration,
				buy_date,
				position_price,
				yield_to_maturity_on_purchase,
				yield_to_maturity,
				current_price,
				nominal,
				profit,
				profit_in_percentage
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		`,
			chatID,
			accountId,
			pos.Name,
			pos.Ticker,
			pos.Currencies,
			pos.Quantity,
			pos.PercentOfPortfolio,
			pos.MaturityDate,
			pos.Duration,
			pos.BuyDate,
			pos.PositionPrice,
			pos.YieldToMaturityOnPurchase,
			pos.YieldToMaturity,
			pos.CurrentPrice,
			pos.Nominal,
			pos.Profit,
			pos.ProfitInPercentage,
		)
	}

	br := tx.SendBatch(ctx, batch)

	for range positions {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert general bond report failed: %w", err)
		}
	}
	err = br.Close()
	if err != nil {
		return fmt.Errorf("could not close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (s *Storage) SaveUids(ctx context.Context, uids map[string]string) (err error) {
	const op = "postgreSql.SaveUids"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	if len(uids) == 0 {
		return errors.New("empty uids map")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q_delete := "DELETE FROM uids"
	_, err = tx.Exec(ctx, q_delete)
	if err != nil {
		return err
	}
	q_insert := `
    INSERT INTO uids (
		instrument_uid,
		asset_uid
		) VALUES ($1, $2)
	`

	batch := &pgx.Batch{}
	for instrument_uid, asset_uid := range uids {
		batch.Queue(q_insert, instrument_uid, asset_uid)
	}

	br := tx.SendBatch(ctx, batch)

	for range uids {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert uids failed: %w", err)
		}
	}
	err = br.Close()
	if err != nil {
		return fmt.Errorf("could not close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (s *Storage) IsUpdatedUids(ctx context.Context) (_ time.Time, err error) {
	const op = "postgreSql.IsUpdatedUids"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	q := `SELECT update_time FROM uids LIMIT 1`

	var date time.Time

	err = s.db.QueryRow(ctx, q).Scan(&date)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, domain.ErrEmptyUids
	}
	if err != nil {
		return time.Time{}, e.WrapIfErr("can't check update uids", err)
	}

	return date, nil
}

func (s *Storage) GetUid(ctx context.Context, instrumentUid string) (_ string, err error) {
	const op = "postgreSql.GetUid"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	q := `SELECT asset_uid FROM uids WHERE instrument_uid = $1`
	var asset_uid string
	err = s.db.QueryRow(ctx, q, instrumentUid).Scan(&asset_uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrEmptyUids
		} else {
			return "", e.WrapIfErr("can't get uid from DB", err)
		}
	}
	return asset_uid, nil
}

func (s *Storage) SaveCurrency(ctx context.Context, currencies domain.CurrenciesCBR, date time.Time) (err error) {
	const op = "postgreSql.SaveCurrency"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	queryInsert := `
        INSERT INTO currencies 
        (date, num_code, char_code, nominal, name, value, vunit_rate)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	batch := &pgx.Batch{}

	for _, c := range currencies.CurrenciesMap {
		batch.Queue(queryInsert, date, c.NumCode, c.CharCode, c.Nominal, c.Name, c.Value, c.VunitRate)
	}

	br := tx.SendBatch(ctx, batch)
	defer func() { _ = br.Close() }()

	for range currencies.CurrenciesMap {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert currencies failed: %w", err)
		}
	}

	err = br.Close()
	if err != nil {
		return fmt.Errorf("could not close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (s *Storage) GetCurrency(ctx context.Context, charCode string, date time.Time) (_ float64, err error) {
	const op = "postgreSql.SaveCurrency"
	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()
	q := `
        SELECT vunit_rate
        FROM currencies
        WHERE char_code = $1 
          AND date::date = $2::date
        LIMIT 1
    `
	var vunit_rate float64
	err = s.db.QueryRow(ctx, q, charCode, date).Scan(&vunit_rate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return vunit_rate, domain.ErrNoCurrency
		}
		return vunit_rate, e.WrapIfErr("can't get currency from DB", err)
	}

	return vunit_rate, nil
}
