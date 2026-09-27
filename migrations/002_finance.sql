-- +goose Up
-- Finance module: ledger, contracts with monthly work done, payroll and supplier
-- accruals, founder and lender balances, forecast. Every money row stores the
-- original amount and currency plus its UZS and USD value at that day's CBU rate,
-- so reports never re-convert history.

CREATE TABLE fin_fx_rates (
	rate_date date PRIMARY KEY,
	uzs_per_usd numeric(14,4) NOT NULL CHECK (uzs_per_usd > 0),
	source text NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'cbu', 'import')),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE fin_categories (
	code text PRIMARY KEY,
	name text NOT NULL,
	kind text NOT NULL CHECK (kind IN (
		'revenue', 'other_income', 'expense', 'tax', 'capex',
		'owner_capital', 'owner_draw',
		'loan_received', 'loan_repaid', 'loan_given', 'loan_returned',
		'transfer'
	)),
	pl_group text NOT NULL DEFAULT '' CHECK (pl_group IN (
		'', 'revenue', 'other_income', 'payroll', 'bonus', 'payroll_tax',
		'financial', 'operating', 'tax'
	)),
	accrual_based boolean NOT NULL DEFAULT false,
	sort_order integer NOT NULL DEFAULT 100,
	active boolean NOT NULL DEFAULT true,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE fin_counterparties (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	name text NOT NULL UNIQUE,
	kind text NOT NULL DEFAULT 'other' CHECK (kind IN (
		'school', 'employee', 'founder', 'supplier', 'lender', 'government', 'bank', 'other'
	)),
	application_id uuid REFERENCES test_center_applications(id) ON DELETE SET NULL,
	active boolean NOT NULL DEFAULT true,
	notes text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE fin_contracts (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	counterparty_id uuid NOT NULL REFERENCES fin_counterparties(id) ON DELETE RESTRICT,
	application_id uuid REFERENCES test_center_applications(id) ON DELETE SET NULL,
	title text NOT NULL DEFAULT '',
	amount numeric(16,2) NOT NULL CHECK (amount >= 0),
	currency text NOT NULL DEFAULT 'USD' CHECK (currency IN ('USD', 'UZS')),
	signed_on date NOT NULL,
	status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'cancelled')),
	-- Share of the prepayment that may be spent after each step (percent).
	w1 numeric(5,2) NOT NULL DEFAULT 15,
	w2 numeric(5,2) NOT NULL DEFAULT 25,
	w3 numeric(5,2) NOT NULL DEFAULT 20,
	w4 numeric(5,2) NOT NULL DEFAULT 40,
	-- Date each step was reached: documents submitted, CEEB code, test center
	-- application, test center code.
	m1_on date,
	m2_on date,
	m3_on date,
	m4_on date,
	notes text NOT NULL DEFAULT '',
	created_by uuid REFERENCES users(id) ON DELETE SET NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CHECK (w1 + w2 + w3 + w4 = 100)
);

CREATE TABLE fin_transactions (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	txn_date date NOT NULL,
	account text NOT NULL CHECK (account IN ('cash_uzs', 'card_uzs', 'bank_uzs', 'cash_usd', 'bank_usd')),
	direction text NOT NULL CHECK (direction IN ('in', 'out')),
	amount numeric(16,2) NOT NULL CHECK (amount > 0),
	currency text NOT NULL CHECK (currency IN ('USD', 'UZS')),
	fx_rate numeric(14,4) NOT NULL CHECK (fx_rate > 0),
	amount_uzs numeric(18,2) NOT NULL,
	amount_usd numeric(16,2) NOT NULL,
	category_code text NOT NULL REFERENCES fin_categories(code) ON UPDATE CASCADE,
	counterparty_id uuid REFERENCES fin_counterparties(id) ON DELETE RESTRICT,
	contract_id uuid REFERENCES fin_contracts(id) ON DELETE RESTRICT,
	-- How much of the contract (in contract currency) this payment settles.
	settles_amount numeric(16,2),
	description text NOT NULL DEFAULT '',
	transfer_group uuid,
	created_by uuid REFERENCES users(id) ON DELETE SET NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

-- Expenses owed for a month (salaries, bonuses, rent, services). Payments come
-- from fin_transactions in the same category and counterparty.
CREATE TABLE fin_accruals (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	accrual_date date NOT NULL,
	category_code text NOT NULL REFERENCES fin_categories(code) ON UPDATE CASCADE,
	counterparty_id uuid NOT NULL REFERENCES fin_counterparties(id) ON DELETE RESTRICT,
	amount numeric(16,2) NOT NULL CHECK (amount > 0),
	currency text NOT NULL CHECK (currency IN ('USD', 'UZS')),
	fx_rate numeric(14,4) NOT NULL CHECK (fx_rate > 0),
	amount_uzs numeric(18,2) NOT NULL,
	amount_usd numeric(16,2) NOT NULL,
	note text NOT NULL DEFAULT '',
	transaction_id uuid REFERENCES fin_transactions(id) ON DELETE CASCADE,
	created_by uuid REFERENCES users(id) ON DELETE SET NULL,
	created_at timestamptz NOT NULL DEFAULT now()
);

-- Work done for a school, recorded in the month it was done.
CREATE TABLE fin_revenue (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	contract_id uuid NOT NULL REFERENCES fin_contracts(id) ON DELETE RESTRICT,
	rec_date date NOT NULL,
	amount numeric(16,2) NOT NULL CHECK (amount > 0),
	fx_rate numeric(14,4) NOT NULL CHECK (fx_rate > 0),
	amount_uzs numeric(18,2) NOT NULL,
	amount_usd numeric(16,2) NOT NULL,
	milestone smallint CHECK (milestone BETWEEN 1 AND 4),
	note text NOT NULL DEFAULT '',
	created_by uuid REFERENCES users(id) ON DELETE SET NULL,
	created_at timestamptz NOT NULL DEFAULT now()
);

-- Balance the company owed a founder or lender at the start of as_of
-- (positive = company owes them). Later movements come from the ledger.
CREATE TABLE fin_openings (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	counterparty_id uuid NOT NULL UNIQUE REFERENCES fin_counterparties(id) ON DELETE RESTRICT,
	as_of date NOT NULL,
	amount_uzs numeric(18,2) NOT NULL,
	note text NOT NULL DEFAULT '',
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE fin_recurring (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	label text NOT NULL,
	counterparty_id uuid REFERENCES fin_counterparties(id) ON DELETE SET NULL,
	amount numeric(16,2) NOT NULL CHECK (amount > 0),
	currency text NOT NULL CHECK (currency IN ('USD', 'UZS')),
	active boolean NOT NULL DEFAULT true,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE fin_forecast_items (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	month date NOT NULL,
	direction text NOT NULL CHECK (direction IN ('in', 'out')),
	label text NOT NULL,
	amount numeric(16,2) NOT NULL CHECK (amount > 0),
	currency text NOT NULL CHECK (currency IN ('USD', 'UZS')),
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX fin_transactions_date_idx ON fin_transactions(txn_date DESC, created_at DESC);
CREATE INDEX fin_transactions_category_idx ON fin_transactions(category_code);
CREATE INDEX fin_transactions_counterparty_idx ON fin_transactions(counterparty_id);
CREATE INDEX fin_transactions_contract_idx ON fin_transactions(contract_id);
CREATE INDEX fin_transactions_transfer_idx ON fin_transactions(transfer_group) WHERE transfer_group IS NOT NULL;
CREATE INDEX fin_accruals_date_idx ON fin_accruals(accrual_date);
CREATE INDEX fin_revenue_contract_idx ON fin_revenue(contract_id, rec_date);

INSERT INTO fin_categories (code, name, kind, pl_group, accrual_based, sort_order) VALUES
	('daromad', 'Daromadlar (maktab to''lovlari)', 'revenue', 'revenue', false, 10),
	('boshqa_tushum', 'Boshqa tushumlar', 'other_income', 'other_income', false, 20),
	('ish_haqi_mamuriy', 'Ma''muriy xodimlarning ish haqi', 'expense', 'payroll', true, 30),
	('ish_haqi_sotuv', 'Sotuv bo''limi xodimlarining ish haqi', 'expense', 'payroll', true, 31),
	('ish_haqi_xizmat', 'Xizmat ko''rsatish xodimlarining ish haqi', 'expense', 'payroll', true, 32),
	('bonus_sotuv', 'Sotuv bo''limi xodimlarining bonusi', 'expense', 'bonus', true, 40),
	('bonus_mamuriy', 'Ma''muriy xodimlarning bonusi', 'expense', 'bonus', true, 41),
	('bonus_xizmat', 'Xizmat ko''rsatish xodimlarining bonusi', 'expense', 'bonus', true, 42),
	('ish_haqi_soliq', 'Ish haqidan soliqlar (ESP, NDFL, INPS)', 'expense', 'payroll_tax', false, 50),
	('bank_komissiya', 'Bank komissiyalari', 'expense', 'financial', false, 60),
	('penya', 'Kredit foizlari, penya va jarimalar', 'expense', 'financial', false, 61),
	('arenda', 'Arenda', 'expense', 'operating', true, 70),
	('yuridik', 'Yuridik va konsalting xizmatlari', 'expense', 'operating', true, 71),
	('komunal', 'Komunal to''lovlar', 'expense', 'operating', false, 72),
	('pitaniya', 'Pitaniya', 'expense', 'operating', false, 73),
	('rasxodniy', 'Rasxodniy material', 'expense', 'operating', false, 74),
	('internet', 'Internet va aloqa', 'expense', 'operating', false, 75),
	('transport', 'Transport', 'expense', 'operating', false, 76),
	('komandirovka', 'Komandirovka', 'expense', 'operating', false, 77),
	('xoztovar', 'Xoztovar / Kanstovar', 'expense', 'operating', false, 78),
	('marketing', 'Marketing', 'expense', 'operating', false, 79),
	('homiylik', 'Homiylik xarajatlari', 'expense', 'operating', false, 80),
	('boshqa_xarajat', 'Boshqa xarajatlar', 'expense', 'operating', false, 81),
	('foyda_soligi', 'Foyda solig''i', 'tax', 'tax', false, 90),
	('aylanma_soliq', 'Aylanma soliq', 'tax', 'tax', false, 91),
	('qqs', 'Qo''shilgan qiymat solig''i (QQS)', 'tax', 'tax', false, 92),
	('boshqa_soliq', 'Boshqa soliqlar', 'tax', 'tax', false, 93),
	('asosiy_vositalar', 'Asosiy vositalar (jihozlar)', 'capex', '', false, 100),
	('ustav', 'Ustav kapitali', 'owner_capital', '', false, 110),
	('divident', 'Dividentlar (ta''sischilar olgan)', 'owner_draw', '', false, 111),
	('qarz_olindi', 'Qarz olindi', 'loan_received', '', false, 120),
	('qarz_qaytarish', 'Qarz qaytarish', 'loan_repaid', '', false, 121),
	('qarz_berildi', 'Ta''sischi/xodimga qarz berildi', 'loan_given', '', false, 122),
	('qarz_qaytdi', 'Berilgan qarz qaytdi', 'loan_returned', '', false, 123),
	('otkazma', 'Hisoblar o''rtasida o''tkazma', 'transfer', '', false, 130);

-- +goose Down
DROP TABLE IF EXISTS fin_forecast_items;
DROP TABLE IF EXISTS fin_recurring;
DROP TABLE IF EXISTS fin_openings;
DROP TABLE IF EXISTS fin_revenue;
DROP TABLE IF EXISTS fin_accruals;
DROP TABLE IF EXISTS fin_transactions;
DROP TABLE IF EXISTS fin_contracts;
DROP TABLE IF EXISTS fin_counterparties;
DROP TABLE IF EXISTS fin_categories;
DROP TABLE IF EXISTS fin_fx_rates;
