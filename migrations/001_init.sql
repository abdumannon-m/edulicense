-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	name text NOT NULL,
	email text NOT NULL UNIQUE,
	password_hash text NOT NULL,
	role text NOT NULL CHECK (role IN ('super_admin', 'admin', 'sales')),
	active boolean NOT NULL DEFAULT true,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	token_hash text NOT NULL UNIQUE,
	expires_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE certificates (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	slug text NOT NULL UNIQUE,
	institution text NOT NULL,
	designation text NOT NULL,
	sat_administration_date text NOT NULL,
	college_board_screenshot text NOT NULL,
	verification_id text NOT NULL UNIQUE,
	issue_date text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE sales_deals (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	school_name text NOT NULL,
	capacity integer NOT NULL DEFAULT 0,
	assigned_sales_agent_id uuid REFERENCES users(id) ON DELETE SET NULL,
	stage text NOT NULL DEFAULT 'new_lead' CHECK (stage IN ('new_lead', 'contacted', 'meeting_scheduled', 'proposal_sent', 'negotiating', 'won', 'lost')),
	negotiation_price numeric(12,2),
	currency text NOT NULL DEFAULT 'USD' CHECK (currency IN ('USD', 'UZS')),
	profit_sharing text CHECK (profit_sharing IN ('80-20', '70-30', '60-40', '50-50', '40-60', '30-70') OR profit_sharing IS NULL),
	notes text NOT NULL DEFAULT '',
	converted_application_id uuid,
	created_by uuid REFERENCES users(id) ON DELETE SET NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE test_center_applications (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	institution_name text NOT NULL,
	institution_type text NOT NULL DEFAULT 'school',
	location text NOT NULL DEFAULT '',
	website text NOT NULL DEFAULT '',
	responsible_person text NOT NULL DEFAULT '',
	phone text NOT NULL DEFAULT '',
	director_phone text NOT NULL DEFAULT '',
	assigned_admin_id uuid REFERENCES users(id) ON DELETE SET NULL,
	sales_referral_id uuid REFERENCES sales_deals(id) ON DELETE SET NULL,
	payment_option text NOT NULL DEFAULT 'our_investment' CHECK (payment_option IN ('investor', 'our_investment', 'fixed_amount')),
	payment_amount numeric(12,2),
	payment_currency text NOT NULL DEFAULT 'USD' CHECK (payment_currency IN ('USD', 'UZS')),
	profit_sharing text CHECK (profit_sharing IN ('80-20', '70-30', '60-40', '50-50', '40-60', '30-70') OR profit_sharing IS NULL),
	ceeb_code text NOT NULL DEFAULT '',
	stage text NOT NULL DEFAULT 'start' CHECK (stage IN ('start', 'document_gather', 'applied_to_ceeb', 'got_ceeb', 'applied_to_test_center', 'got_test_center', 'in_test_center_list')),
	checklist_status jsonb NOT NULL DEFAULT '{}'::jsonb,
	created_by uuid REFERENCES users(id) ON DELETE SET NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE sales_deals
	ADD CONSTRAINT sales_deals_converted_application_fk
	FOREIGN KEY (converted_application_id) REFERENCES test_center_applications(id) ON DELETE SET NULL;

CREATE TABLE application_documents (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	application_id uuid NOT NULL REFERENCES test_center_applications(id) ON DELETE CASCADE,
	doc_type text NOT NULL CHECK (doc_type IN ('license_english', 'domain_email', 'website_english', 'responsible_staff')),
	original_filename text NOT NULL,
	content_type text NOT NULL DEFAULT '',
	size_bytes bigint NOT NULL DEFAULT 0,
	storage_key text NOT NULL,
	uploaded_by uuid REFERENCES users(id) ON DELETE SET NULL,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE reminders (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	entity_type text NOT NULL CHECK (entity_type IN ('application', 'deal')),
	entity_id uuid NOT NULL,
	title text NOT NULL,
	note text NOT NULL DEFAULT '',
	due_at timestamptz NOT NULL,
	created_by uuid REFERENCES users(id) ON DELETE SET NULL,
	sent_at timestamptz,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE activity_logs (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id uuid REFERENCES users(id) ON DELETE SET NULL,
	action text NOT NULL,
	entity_type text NOT NULL,
	entity_id uuid,
	summary text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX sessions_token_hash_idx ON sessions(token_hash);
CREATE INDEX sessions_expires_at_idx ON sessions(expires_at);
CREATE INDEX sales_deals_stage_idx ON sales_deals(stage);
CREATE INDEX applications_stage_idx ON test_center_applications(stage);
CREATE INDEX reminders_due_idx ON reminders(due_at) WHERE sent_at IS NULL;
CREATE INDEX activity_logs_created_idx ON activity_logs(created_at DESC);

INSERT INTO certificates (
	slug,
	institution,
	designation,
	sat_administration_date,
	college_board_screenshot,
	verification_id,
	issue_date
) VALUES (
	'oriental-university-sat-center',
	'Oriental University',
	'Authorised SAT Test Centre',
	'8 March 2025',
	'/static/certificates/oriental-university-sat-center.svg',
	'EDL-2025-0312-OU',
	'12 March 2025'
) ON CONFLICT (slug) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS activity_logs;
DROP TABLE IF EXISTS reminders;
DROP TABLE IF EXISTS application_documents;
ALTER TABLE IF EXISTS sales_deals DROP CONSTRAINT IF EXISTS sales_deals_converted_application_fk;
DROP TABLE IF EXISTS test_center_applications;
DROP TABLE IF EXISTS sales_deals;
DROP TABLE IF EXISTS certificates;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
