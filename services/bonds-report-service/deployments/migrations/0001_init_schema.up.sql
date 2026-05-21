CREATE TABLE IF NOT EXISTS bond_reports (
    id SERIAL PRIMARY KEY,
    chatId BIGINT,
    broker_account_id TEXT NOT NULL,
    name TEXT NOT NULL,
    ticker TEXT NOT NULL,
    maturity_date DATE,
    offer_date DATE,
    duration INTEGER,
    buy_date DATE,
    buy_price NUMERIC(12, 4),
    yield_to_maturity_on_purchase NUMERIC(7, 4),
    yield_to_offer_on_purchase NUMERIC(7, 4),
    yield_to_maturity NUMERIC(7, 4),
    yield_to_offer NUMERIC(7, 4),
    current_price NUMERIC(12, 4),
    nominal NUMERIC(12, 2),
    profit NUMERIC(14, 2),
    annualized_return NUMERIC(7, 4)
);

CREATE TABLE IF NOT EXISTS general_bond_report (
    id SERIAL PRIMARY KEY,
    chatId BIGINT,
    broker_account_id TEXT,
    name TEXT,
    ticker TEXT,
    currencies TEXT,
    quantity INTEGER,
    percent_of_portfolio NUMERIC(10, 2),
    maturity_date TIMESTAMP,
    duration INTEGER,
    buy_date TIMESTAMP,
    position_price NUMERIC(12, 4),
    yield_to_maturity_on_purchase NUMERIC(6, 2),
    yield_to_maturity NUMERIC(6, 2),
    current_price NUMERIC(12, 4),
    nominal NUMERIC(12, 4),
    profit NUMERIC(14, 4),
    profit_in_percentage NUMERIC(6, 2)
);

CREATE TABLE IF NOT EXISTS uids (
    update_time TIMESTAMP DEFAULT current_timestamp,
    instrument_uid TEXT,
    asset_uid TEXT
);

CREATE TABLE IF NOT EXISTS operations (
    id SERIAL PRIMARY KEY,
    chatId BIGINT,
    broker_account_id TEXT,
    currency TEXT,
    operation_id TEXT,
    parent_operation_id TEXT,
    name TEXT,
    date TIMESTAMP,
    type INTEGER,
    description TEXT,
    instrument_uid TEXT,
    figi TEXT,
    instrument_type TEXT,
    instrument_kind TEXT,
    position_uid TEXT,
    payment NUMERIC(14, 4),
    price NUMERIC(14, 4),
    commission NUMERIC(14, 4),
    yield NUMERIC(14, 4),
    yield_relative NUMERIC(6, 2),
    accrued_int NUMERIC(14, 4),
    quantity_done INTEGER,
    asset_uid TEXT
);

CREATE TABLE IF NOT EXISTS currencies (
    date DATE,
    num_code TEXT,
    char_code TEXT,
    nominal INTEGER,
    name TEXT,
    value NUMERIC(14, 4),
    vunit_rate NUMERIC(14, 4)
);
