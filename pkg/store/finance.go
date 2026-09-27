package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"edu-license/pkg/app"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrNoRate is returned when a money row is saved but no USD rate is known.
var ErrNoRate = errors.New("No exchange rate is saved yet. Add the CBU rate in Finance settings.")

// LoadFinData reads every finance table in one round trip.
func (s *Postgres) LoadFinData(ctx context.Context, today time.Time) (app.FinData, error) {
	data := app.FinData{Today: app.DateOnly(today)}
	batch := &pgx.Batch{}
	batch.Queue(`SELECT rate_date, uzs_per_usd::float8, source FROM fin_fx_rates ORDER BY rate_date`)
	batch.Queue(`SELECT code, name, kind, pl_group, accrual_based, sort_order, active FROM fin_categories ORDER BY sort_order, name`)
	batch.Queue(`
		SELECT c.id::text, c.name, c.kind, coalesce(c.application_id::text, ''), coalesce(a.institution_name, ''), c.active, c.notes
		FROM fin_counterparties c
		LEFT JOIN test_center_applications a ON a.id = c.application_id
		ORDER BY lower(c.name)`)
	batch.Queue(`
		SELECT k.id::text, k.counterparty_id::text, p.name, coalesce(k.application_id::text, ''),
			coalesce(a.institution_name, ''), coalesce(a.stage, ''), k.title, k.amount::float8, k.currency,
			k.signed_on, k.status, k.w1::float8, k.w2::float8, k.w3::float8, k.w4::float8,
			k.m1_on, k.m2_on, k.m3_on, k.m4_on, k.notes
		FROM fin_contracts k
		JOIN fin_counterparties p ON p.id = k.counterparty_id
		LEFT JOIN test_center_applications a ON a.id = k.application_id
		ORDER BY k.signed_on DESC, p.name`)
	batch.Queue(`
		SELECT t.id::text, t.txn_date, t.account, t.direction, t.amount::float8, t.currency, t.fx_rate::float8,
			t.amount_uzs::float8, t.amount_usd::float8, t.category_code, coalesce(t.counterparty_id::text, ''),
			coalesce(p.name, ''), coalesce(t.contract_id::text, ''), t.settles_amount::float8, t.description,
			coalesce(t.transfer_group::text, ''), t.created_at,
			EXISTS (SELECT 1 FROM fin_accruals x WHERE x.transaction_id = t.id)
		FROM fin_transactions t
		LEFT JOIN fin_counterparties p ON p.id = t.counterparty_id
		ORDER BY t.txn_date, t.created_at`)
	batch.Queue(`
		SELECT x.id::text, x.accrual_date, x.category_code, x.counterparty_id::text, p.name, x.amount::float8,
			x.currency, x.fx_rate::float8, x.amount_uzs::float8, x.amount_usd::float8, x.note, coalesce(x.transaction_id::text, '')
		FROM fin_accruals x
		JOIN fin_counterparties p ON p.id = x.counterparty_id
		ORDER BY x.accrual_date, x.created_at`)
	batch.Queue(`
		SELECT id::text, contract_id::text, rec_date, amount::float8, fx_rate::float8, amount_uzs::float8,
			amount_usd::float8, coalesce(milestone, 0), note
		FROM fin_revenue ORDER BY rec_date, created_at`)
	batch.Queue(`
		SELECT o.id::text, o.counterparty_id::text, p.name, p.kind, o.as_of, o.amount_uzs::float8, o.note
		FROM fin_openings o JOIN fin_counterparties p ON p.id = o.counterparty_id
		ORDER BY p.name`)
	batch.Queue(`
		SELECT r.id::text, r.label, coalesce(r.counterparty_id::text, ''), coalesce(p.name, ''), r.amount::float8, r.currency, r.active
		FROM fin_recurring r LEFT JOIN fin_counterparties p ON p.id = r.counterparty_id
		ORDER BY r.active DESC, r.label`)
	batch.Queue(`SELECT id::text, month, direction, label, amount::float8, currency FROM fin_forecast_items ORDER BY month, created_at`)
	batch.Queue(`SELECT id::text, institution_name, stage FROM test_center_applications ORDER BY lower(institution_name)`)

	results := s.pool.SendBatch(ctx, batch)
	defer results.Close()

	readAll := func(scan func(pgx.Rows) error) error {
		rows, err := results.Query()
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			if err := scan(rows); err != nil {
				return err
			}
		}
		return rows.Err()
	}

	steps := []func(pgx.Rows) error{
		func(rows pgx.Rows) error {
			var rate app.FinRate
			if err := rows.Scan(&rate.Date, &rate.Rate, &rate.Source); err != nil {
				return err
			}
			data.Rates = append(data.Rates, rate)
			return nil
		},
		func(rows pgx.Rows) error {
			var c app.FinCategory
			if err := rows.Scan(&c.Code, &c.Name, &c.Kind, &c.PLGroup, &c.AccrualBased, &c.SortOrder, &c.Active); err != nil {
				return err
			}
			data.Categories = append(data.Categories, c)
			return nil
		},
		func(rows pgx.Rows) error {
			var c app.FinCounterparty
			if err := rows.Scan(&c.ID, &c.Name, &c.Kind, &c.ApplicationID, &c.ApplicationName, &c.Active, &c.Notes); err != nil {
				return err
			}
			data.Counterparties = append(data.Counterparties, c)
			return nil
		},
		func(rows pgx.Rows) error {
			var k app.FinContract
			if err := rows.Scan(&k.ID, &k.CounterpartyID, &k.CounterpartyName, &k.ApplicationID, &k.ApplicationName,
				&k.ApplicationStage, &k.Title, &k.Amount, &k.Currency, &k.SignedOn, &k.Status,
				&k.Weights[0], &k.Weights[1], &k.Weights[2], &k.Weights[3],
				&k.MilestoneOn[0], &k.MilestoneOn[1], &k.MilestoneOn[2], &k.MilestoneOn[3], &k.Notes); err != nil {
				return err
			}
			data.Contracts = append(data.Contracts, k)
			return nil
		},
		func(rows pgx.Rows) error {
			var t app.FinTransaction
			if err := rows.Scan(&t.ID, &t.Date, &t.Account, &t.Direction, &t.Amount, &t.Currency, &t.FxRate,
				&t.AmountUZS, &t.AmountUSD, &t.CategoryCode, &t.CounterpartyID, &t.CounterpartyName, &t.ContractID,
				&t.SettlesAmount, &t.Description, &t.TransferGroup, &t.CreatedAt, &t.HasAccrual); err != nil {
				return err
			}
			data.Transactions = append(data.Transactions, t)
			return nil
		},
		func(rows pgx.Rows) error {
			var a app.FinAccrual
			if err := rows.Scan(&a.ID, &a.Date, &a.CategoryCode, &a.CounterpartyID, &a.CounterpartyName, &a.Amount,
				&a.Currency, &a.FxRate, &a.AmountUZS, &a.AmountUSD, &a.Note, &a.TransactionID); err != nil {
				return err
			}
			data.Accruals = append(data.Accruals, a)
			return nil
		},
		func(rows pgx.Rows) error {
			var r app.FinRevenue
			if err := rows.Scan(&r.ID, &r.ContractID, &r.Date, &r.Amount, &r.FxRate, &r.AmountUZS, &r.AmountUSD, &r.Milestone, &r.Note); err != nil {
				return err
			}
			data.Revenue = append(data.Revenue, r)
			return nil
		},
		func(rows pgx.Rows) error {
			var o app.FinOpening
			if err := rows.Scan(&o.ID, &o.CounterpartyID, &o.CounterpartyName, &o.CounterpartyKind, &o.AsOf, &o.AmountUZS, &o.Note); err != nil {
				return err
			}
			data.Openings = append(data.Openings, o)
			return nil
		},
		func(rows pgx.Rows) error {
			var r app.FinRecurring
			if err := rows.Scan(&r.ID, &r.Label, &r.CounterpartyID, &r.CounterpartyName, &r.Amount, &r.Currency, &r.Active); err != nil {
				return err
			}
			data.Recurring = append(data.Recurring, r)
			return nil
		},
		func(rows pgx.Rows) error {
			var f app.FinForecastItem
			if err := rows.Scan(&f.ID, &f.Month, &f.Direction, &f.Label, &f.Amount, &f.Currency); err != nil {
				return err
			}
			data.ForecastItems = append(data.ForecastItems, f)
			return nil
		},
		func(rows pgx.Rows) error {
			var a app.FinApplicationRef
			if err := rows.Scan(&a.ID, &a.Name, &a.Stage); err != nil {
				return err
			}
			data.Applications = append(data.Applications, a)
			return nil
		},
	}
	for i, step := range steps {
		if err := readAll(step); err != nil {
			return app.FinData{}, fmt.Errorf("finance query %d: %w", i+1, err)
		}
	}
	return data, nil
}

// ---------------------------------------------------------------------------
// Rates

func (s *Postgres) UpsertFinRate(ctx context.Context, date time.Time, rate float64, source string) error {
	if rate <= 0 {
		return errors.New("Rate must be greater than zero.")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO fin_fx_rates (rate_date, uzs_per_usd, source) VALUES ($1, $2, $3)
		ON CONFLICT (rate_date) DO UPDATE SET uzs_per_usd = EXCLUDED.uzs_per_usd, source = EXCLUDED.source, updated_at = now()
	`, app.DateOnly(date), rate, source)
	return err
}

func (s *Postgres) HasFinRate(ctx context.Context, date time.Time) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM fin_fx_rates WHERE rate_date = $1)`, app.DateOnly(date)).Scan(&exists)
	return exists, err
}

func rateFor(ctx context.Context, tx pgx.Tx, date time.Time, override float64) (float64, error) {
	if override > 0 {
		return override, nil
	}
	var rate float64
	err := tx.QueryRow(ctx, `
		SELECT uzs_per_usd::float8 FROM (
			(SELECT uzs_per_usd, 0 AS pref, rate_date FROM fin_fx_rates WHERE rate_date <= $1 ORDER BY rate_date DESC LIMIT 1)
			UNION ALL
			(SELECT uzs_per_usd, 1 AS pref, rate_date FROM fin_fx_rates ORDER BY rate_date ASC LIMIT 1)
		) r ORDER BY pref LIMIT 1
	`, app.DateOnly(date)).Scan(&rate)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNoRate
	}
	return rate, err
}

func (s *Postgres) withTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func nullUUID(value string) any {
	if value == "" {
		return nil
	}
	return value
}

// ---------------------------------------------------------------------------
// Transactions

func (s *Postgres) CreateFinTransaction(ctx context.Context, input app.FinTxnInput, actorID string) (string, error) {
	if err := input.Validate(); err != nil {
		return "", err
	}
	var id string
	err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		id, err = insertFinTransaction(ctx, tx, input, actorID)
		return err
	})
	if err != nil {
		return "", err
	}
	_ = s.LogActivity(ctx, actorID, "create", "finance_transaction", id, "Recorded "+app.FinAccountLabel(input.Account)+" "+app.FormatNumber(input.Amount))
	return id, nil
}

func insertFinTransaction(ctx context.Context, tx pgx.Tx, input app.FinTxnInput, actorID string) (string, error) {
	rate, err := rateFor(ctx, tx, input.Date, input.FxRate)
	if err != nil {
		return "", err
	}
	date := app.DateOnly(input.Date)
	insert := func(account, direction string, amount float64, group any) (string, error) {
		acct, _ := app.FinAccountByCode(account)
		uzs, usd := app.Convert(amount, acct.Currency, rate)
		var id string
		err := tx.QueryRow(ctx, `
			INSERT INTO fin_transactions (txn_date, account, direction, amount, currency, fx_rate, amount_uzs, amount_usd,
				category_code, counterparty_id, contract_id, settles_amount, description, transfer_group, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			RETURNING id::text
		`, date, account, direction, amount, acct.Currency, rate, uzs, usd, input.CategoryCode,
			nullUUID(input.CounterpartyID), nullUUID(input.ContractID), input.SettlesAmount, input.Description,
			group, nullUUID(actorID)).Scan(&id)
		return id, err
	}
	if input.IsTransfer() {
		var group string
		if err := tx.QueryRow(ctx, `SELECT gen_random_uuid()::text`).Scan(&group); err != nil {
			return "", err
		}
		id, err := insert(input.Account, "out", input.Amount, group)
		if err != nil {
			return "", err
		}
		if _, err := insert(input.ToAccount, "in", input.ToAmount, group); err != nil {
			return "", err
		}
		return id, nil
	}
	id, err := insert(input.Account, input.Direction, input.Amount, nil)
	if err != nil {
		return "", err
	}
	if input.CreateAccrual && input.Direction == "out" {
		var accrualBased bool
		if err := tx.QueryRow(ctx, `SELECT accrual_based FROM fin_categories WHERE code = $1`, input.CategoryCode).Scan(&accrualBased); err != nil {
			return "", err
		}
		if accrualBased {
			acct, _ := app.FinAccountByCode(input.Account)
			uzs, usd := app.Convert(input.Amount, acct.Currency, rate)
			if _, err := tx.Exec(ctx, `
				INSERT INTO fin_accruals (accrual_date, category_code, counterparty_id, amount, currency, fx_rate, amount_uzs, amount_usd, note, transaction_id, created_by)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			`, date, input.CategoryCode, input.CounterpartyID, input.Amount, acct.Currency, rate, uzs, usd, "Paid right away", id, nullUUID(actorID)); err != nil {
				return "", err
			}
		}
	}
	return id, nil
}

// UpdateFinTransaction replaces a transaction (and its transfer partner and
// automatic accrual) with the new values.
func (s *Postgres) UpdateFinTransaction(ctx context.Context, id string, input app.FinTxnInput, actorID string) (string, error) {
	if err := input.Validate(); err != nil {
		return "", err
	}
	var newID string
	err := s.withTx(ctx, func(tx pgx.Tx) error {
		if err := deleteFinTransaction(ctx, tx, id); err != nil {
			return err
		}
		var err error
		newID, err = insertFinTransaction(ctx, tx, input, actorID)
		return err
	})
	if err != nil {
		return "", err
	}
	_ = s.LogActivity(ctx, actorID, "update", "finance_transaction", newID, "Edited "+app.FinAccountLabel(input.Account)+" "+app.FormatNumber(input.Amount))
	return newID, nil
}

func deleteFinTransaction(ctx context.Context, tx pgx.Tx, id string) error {
	tag, err := tx.Exec(ctx, `
		DELETE FROM fin_transactions
		WHERE id = $1
			OR (transfer_group IS NOT NULL AND transfer_group = (SELECT transfer_group FROM fin_transactions WHERE id = $1))
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Postgres) DeleteFinTransaction(ctx context.Context, id, actorID string) error {
	if err := s.withTx(ctx, func(tx pgx.Tx) error { return deleteFinTransaction(ctx, tx, id) }); err != nil {
		return err
	}
	_ = s.LogActivity(ctx, actorID, "delete", "finance_transaction", "", "Deleted a finance transaction")
	return nil
}

// ---------------------------------------------------------------------------
// Counterparties and categories

func (s *Postgres) EnsureFinCounterparty(ctx context.Context, name, kind string) (string, error) {
	input := app.FinCounterpartyInput{Name: name, Kind: kind, Active: true}
	if err := input.Validate(); err != nil {
		return "", err
	}
	var id string
	err := s.pool.QueryRow(ctx, `
		WITH found AS (SELECT id FROM fin_counterparties WHERE lower(name) = lower($1) LIMIT 1),
		inserted AS (
			INSERT INTO fin_counterparties (name, kind)
			SELECT $1, $2 WHERE NOT EXISTS (SELECT 1 FROM found)
			RETURNING id
		)
		SELECT id::text FROM found UNION ALL SELECT id::text FROM inserted
	`, input.Name, input.Kind).Scan(&id)
	return id, err
}

func (s *Postgres) SaveFinCounterparty(ctx context.Context, id string, input app.FinCounterpartyInput, actorID string) (string, error) {
	if err := input.Validate(); err != nil {
		return "", err
	}
	var err error
	if id == "" {
		err = s.pool.QueryRow(ctx, `
			INSERT INTO fin_counterparties (name, kind, application_id, active, notes) VALUES ($1, $2, $3, $4, $5)
			RETURNING id::text
		`, input.Name, input.Kind, nullUUID(input.ApplicationID), input.Active, input.Notes).Scan(&id)
	} else {
		var tag pgconn.CommandTag
		tag, err = s.pool.Exec(ctx, `
			UPDATE fin_counterparties SET name = $2, kind = $3, application_id = $4, active = $5, notes = $6 WHERE id = $1
		`, id, input.Name, input.Kind, nullUUID(input.ApplicationID), input.Active, input.Notes)
		if err == nil && tag.RowsAffected() == 0 {
			err = pgx.ErrNoRows
		}
	}
	if err != nil {
		return "", uniqueMessage(err, "Someone with that name already exists.")
	}
	_ = s.LogActivity(ctx, actorID, "save", "finance_counterparty", id, "Saved "+input.Name)
	return id, nil
}

func (s *Postgres) SaveFinCategory(ctx context.Context, input app.FinCategoryInput, actorID string) error {
	if err := input.Validate(); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO fin_categories (code, name, kind, pl_group, accrual_based, sort_order, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, kind = EXCLUDED.kind, pl_group = EXCLUDED.pl_group,
			accrual_based = EXCLUDED.accrual_based, sort_order = EXCLUDED.sort_order, active = EXCLUDED.active
	`, input.Code, input.Name, input.Kind, input.PLGroup, input.AccrualBased, input.SortOrder, input.Active)
	if err == nil {
		_ = s.LogActivity(ctx, actorID, "save", "finance_category", "", "Saved category "+input.Name)
	}
	return err
}

// ---------------------------------------------------------------------------
// Contracts, milestones and work done

func (s *Postgres) SaveFinContract(ctx context.Context, id string, input app.FinContractInput, actorID string) (string, error) {
	if err := input.Validate(); err != nil {
		return "", err
	}
	w := input.Weights
	var err error
	if id == "" {
		err = s.pool.QueryRow(ctx, `
			INSERT INTO fin_contracts (counterparty_id, application_id, title, amount, currency, signed_on, status, w1, w2, w3, w4, notes, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			RETURNING id::text
		`, input.CounterpartyID, nullUUID(input.ApplicationID), input.Title, input.Amount, input.Currency,
			app.DateOnly(input.SignedOn), input.Status, w[0], w[1], w[2], w[3], input.Notes, nullUUID(actorID)).Scan(&id)
	} else {
		var tag pgconn.CommandTag
		tag, err = s.pool.Exec(ctx, `
			UPDATE fin_contracts SET counterparty_id = $2, application_id = $3, title = $4, amount = $5, currency = $6,
				signed_on = $7, status = $8, w1 = $9, w2 = $10, w3 = $11, w4 = $12, notes = $13, updated_at = now()
			WHERE id = $1
		`, id, input.CounterpartyID, nullUUID(input.ApplicationID), input.Title, input.Amount, input.Currency,
			app.DateOnly(input.SignedOn), input.Status, w[0], w[1], w[2], w[3], input.Notes)
		if err == nil && tag.RowsAffected() == 0 {
			err = pgx.ErrNoRows
		}
	}
	if err != nil {
		return "", err
	}
	_ = s.LogActivity(ctx, actorID, "save", "finance_contract", id, "Saved contract")
	return id, nil
}

func (s *Postgres) DeleteFinContract(ctx context.Context, id, actorID string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM fin_contracts WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return errors.New("This contract still has payments or work done. Remove its work-done lines and unlink its payments first, or mark it cancelled.")
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	_ = s.LogActivity(ctx, actorID, "delete", "finance_contract", "", "Deleted a contract")
	return nil
}

// SetFinMilestone records (or clears, when date is nil) the date a step was reached.
func (s *Postgres) SetFinMilestone(ctx context.Context, contractID string, step int, date *time.Time, actorID string) error {
	if step < 1 || step > 4 {
		return errors.New("Unknown step.")
	}
	var value any
	if date != nil {
		value = app.DateOnly(*date)
	}
	tag, err := s.pool.Exec(ctx, fmt.Sprintf(`UPDATE fin_contracts SET m%d_on = $2, updated_at = now() WHERE id = $1`, step), contractID, value)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	_ = s.LogActivity(ctx, actorID, "milestone", "finance_contract", contractID, fmt.Sprintf("Updated step %d", step))
	return nil
}

func (s *Postgres) CreateFinRevenue(ctx context.Context, input app.FinRevenueInput, actorID string) error {
	if err := input.Validate(); err != nil {
		return err
	}
	return s.withTx(ctx, func(tx pgx.Tx) error {
		var currency string
		if err := tx.QueryRow(ctx, `SELECT currency FROM fin_contracts WHERE id = $1`, input.ContractID).Scan(&currency); err != nil {
			return err
		}
		rate, err := rateFor(ctx, tx, input.Date, input.FxRate)
		if err != nil {
			return err
		}
		uzs, usd := app.Convert(input.Amount, currency, rate)
		var milestone any
		if input.Milestone > 0 {
			milestone = input.Milestone
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO fin_revenue (contract_id, rec_date, amount, fx_rate, amount_uzs, amount_usd, milestone, note, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, input.ContractID, app.DateOnly(input.Date), input.Amount, rate, uzs, usd, milestone, input.Note, nullUUID(actorID))
		return err
	})
}

func (s *Postgres) DeleteFinRevenue(ctx context.Context, id string) error {
	return execOne(ctx, s, `DELETE FROM fin_revenue WHERE id = $1`, id)
}

// ---------------------------------------------------------------------------
// Accruals, openings, recurring costs and forecast items

func (s *Postgres) CreateFinAccrual(ctx context.Context, input app.FinAccrualInput, actorID string) error {
	if err := input.Validate(); err != nil {
		return err
	}
	return s.withTx(ctx, func(tx pgx.Tx) error {
		rate, err := rateFor(ctx, tx, input.Date, input.FxRate)
		if err != nil {
			return err
		}
		uzs, usd := app.Convert(input.Amount, input.Currency, rate)
		_, err = tx.Exec(ctx, `
			INSERT INTO fin_accruals (accrual_date, category_code, counterparty_id, amount, currency, fx_rate, amount_uzs, amount_usd, note, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, app.DateOnly(input.Date), input.CategoryCode, input.CounterpartyID, input.Amount, input.Currency, rate, uzs, usd, input.Note, nullUUID(actorID))
		return err
	})
}

func (s *Postgres) DeleteFinAccrual(ctx context.Context, id string) error {
	return execOne(ctx, s, `DELETE FROM fin_accruals WHERE id = $1`, id)
}

func (s *Postgres) SaveFinOpening(ctx context.Context, input app.FinOpeningInput, actorID string) error {
	if err := input.Validate(); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO fin_openings (counterparty_id, as_of, amount_uzs, note) VALUES ($1, $2, $3, $4)
		ON CONFLICT (counterparty_id) DO UPDATE SET as_of = EXCLUDED.as_of, amount_uzs = EXCLUDED.amount_uzs, note = EXCLUDED.note, updated_at = now()
	`, input.CounterpartyID, app.DateOnly(input.AsOf), input.AmountUZS, input.Note)
	if err == nil {
		_ = s.LogActivity(ctx, actorID, "save", "finance_opening", input.CounterpartyID, "Saved opening balance")
	}
	return err
}

func (s *Postgres) DeleteFinOpening(ctx context.Context, id string) error {
	return execOne(ctx, s, `DELETE FROM fin_openings WHERE id = $1`, id)
}

func (s *Postgres) SaveFinRecurring(ctx context.Context, id string, input app.FinRecurringInput) error {
	if err := input.Validate(); err != nil {
		return err
	}
	if id == "" {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO fin_recurring (label, counterparty_id, amount, currency, active) VALUES ($1, $2, $3, $4, $5)
		`, input.Label, nullUUID(input.CounterpartyID), input.Amount, input.Currency, input.Active)
		return err
	}
	return execOne(ctx, s, `
		UPDATE fin_recurring SET label = $2, counterparty_id = $3, amount = $4, currency = $5, active = $6 WHERE id = $1
	`, id, input.Label, nullUUID(input.CounterpartyID), input.Amount, input.Currency, input.Active)
}

func (s *Postgres) DeleteFinRecurring(ctx context.Context, id string) error {
	return execOne(ctx, s, `DELETE FROM fin_recurring WHERE id = $1`, id)
}

func (s *Postgres) CreateFinForecastItem(ctx context.Context, input app.FinForecastInput) error {
	if err := input.Validate(); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO fin_forecast_items (month, direction, label, amount, currency) VALUES ($1, $2, $3, $4, $5)
	`, input.Month, input.Direction, input.Label, input.Amount, input.Currency)
	return err
}

func (s *Postgres) DeleteFinForecastItem(ctx context.Context, id string) error {
	return execOne(ctx, s, `DELETE FROM fin_forecast_items WHERE id = $1`, id)
}

// FinanceIsEmpty reports whether the ledger has no rows yet (used by import).
func (s *Postgres) FinanceIsEmpty(ctx context.Context) (bool, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM fin_transactions) + (SELECT count(*) FROM fin_contracts)`).Scan(&count)
	return count == 0, err
}

// ExecScript runs a multi-statement SQL script inside one transaction.
func (s *Postgres) ExecScript(ctx context.Context, script string) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, script)
		return err
	})
}

func execOne(ctx context.Context, s *Postgres, sql string, args ...any) error {
	tag, err := s.pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func uniqueMessage(err error, message string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return errors.New(message)
	}
	return err
}
