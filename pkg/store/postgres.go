package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"edu-license/pkg/app"
	"edu-license/pkg/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Postgres{pool: pool}, nil
}

func (s *Postgres) Close() {
	s.pool.Close()
}

func (s *Postgres) UserByEmail(ctx context.Context, email string) (auth.StoredUser, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, name, email, password_hash, role, active, created_at, updated_at
		FROM users
		WHERE lower(email) = lower($1)
	`, email)
	var user auth.StoredUser
	if err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return auth.StoredUser{}, err
	}
	return user, nil
}

func (s *Postgres) UserBySessionTokenHash(ctx context.Context, hash string, now time.Time) (app.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT u.id::text, u.name, u.email, u.role, u.active, u.created_at, u.updated_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1
			AND s.expires_at > $2
			AND u.active = true
	`, hash, now)
	return scanUser(row)
}

func (s *Postgres) CreateSession(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (s *Postgres) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	return err
}

func (s *Postgres) CreateUser(ctx context.Context, input auth.CreateUserInput) (app.User, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash, role, active)
		VALUES ($1, lower($2), $3, $4, $5)
		RETURNING id::text, name, email, role, active, created_at, updated_at
	`, input.Name, input.Email, input.PasswordHash, input.Role, input.Active)
	return scanUser(row)
}

func (s *Postgres) ListUsers(ctx context.Context) ([]app.User, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, name, email, role, active, created_at, updated_at
		FROM users
		ORDER BY active DESC, name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []app.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *Postgres) ListUsersByRoles(ctx context.Context, roles ...app.Role) ([]app.User, error) {
	values := make([]string, 0, len(roles))
	for _, role := range roles {
		values = append(values, string(role))
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, name, email, role, active, created_at, updated_at
		FROM users
		WHERE active = true AND role = ANY($1)
		ORDER BY name ASC
	`, values)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []app.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *Postgres) CertificateBySlug(ctx context.Context, slug string) (app.Certificate, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, slug, institution, designation, sat_administration_date,
			college_board_screenshot, verification_id, issue_date
		FROM certificates
		WHERE slug = $1
	`, slug)
	return scanCertificate(row)
}

func (s *Postgres) UpsertCertificate(ctx context.Context, input app.CertificateInput) (app.Certificate, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO certificates (
			slug, institution, designation, sat_administration_date,
			college_board_screenshot, verification_id, issue_date
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7
		)
		ON CONFLICT (slug) DO UPDATE
		SET updated_at = certificates.updated_at
		RETURNING id::text, slug, institution, designation, sat_administration_date,
			college_board_screenshot, verification_id, issue_date
	`, input.Slug, input.Institution, input.Designation, input.SATAdministrationDate,
		input.CollegeBoardScreenshot, input.VerificationID, input.IssueDate)
	return scanCertificate(row)
}

func (s *Postgres) ListApplications(ctx context.Context, stage, adminID, location string) ([]app.TestCenterApplication, error) {
	args := []any{}
	conditions := []string{"true"}
	if stage != "" {
		args = append(args, stage)
		conditions = append(conditions, fmt.Sprintf("a.stage = $%d", len(args)))
	}
	if adminID != "" {
		args = append(args, adminID)
		conditions = append(conditions, fmt.Sprintf("a.assigned_admin_id = $%d", len(args)))
	}
	if location != "" {
		args = append(args, "%"+location+"%")
		conditions = append(conditions, fmt.Sprintf("a.location ILIKE $%d", len(args)))
	}

	query := `
		SELECT a.id::text, a.institution_name, a.institution_type, a.location, a.website,
			a.responsible_person, a.phone, a.director_phone,
			COALESCE(a.assigned_admin_id::text, ''), COALESCE(admin.name, ''),
			COALESCE(a.sales_referral_id::text, ''), COALESCE(sd.school_name, ''),
			a.payment_option, a.payment_amount::text, a.payment_currency, COALESCE(a.profit_sharing, ''),
			a.ceeb_code, a.stage, a.checklist_status, COALESCE(a.created_by::text, ''),
			a.created_at, a.updated_at
		FROM test_center_applications a
		LEFT JOIN users admin ON admin.id = a.assigned_admin_id
		LEFT JOIN sales_deals sd ON sd.id = a.sales_referral_id
		WHERE ` + strings.Join(conditions, " AND ") + `
		ORDER BY a.updated_at DESC
	`
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applications []app.TestCenterApplication
	for rows.Next() {
		item, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}
		applications = append(applications, item)
	}
	return applications, rows.Err()
}

func (s *Postgres) ApplicationByID(ctx context.Context, id string) (app.TestCenterApplication, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT a.id::text, a.institution_name, a.institution_type, a.location, a.website,
			a.responsible_person, a.phone, a.director_phone,
			COALESCE(a.assigned_admin_id::text, ''), COALESCE(admin.name, ''),
			COALESCE(a.sales_referral_id::text, ''), COALESCE(sd.school_name, ''),
			a.payment_option, a.payment_amount::text, a.payment_currency, COALESCE(a.profit_sharing, ''),
			a.ceeb_code, a.stage, a.checklist_status, COALESCE(a.created_by::text, ''),
			a.created_at, a.updated_at
		FROM test_center_applications a
		LEFT JOIN users admin ON admin.id = a.assigned_admin_id
		LEFT JOIN sales_deals sd ON sd.id = a.sales_referral_id
		WHERE a.id = $1
	`, id)
	item, err := scanApplication(row)
	if err != nil {
		return app.TestCenterApplication{}, err
	}
	docs, err := s.ListApplicationDocuments(ctx, id)
	if err != nil {
		return app.TestCenterApplication{}, err
	}
	reminders, err := s.ListRemindersForEntity(ctx, "application", id)
	if err != nil {
		return app.TestCenterApplication{}, err
	}
	item.Documents = docs
	item.Reminders = reminders
	return item, nil
}

func (s *Postgres) CreateApplication(ctx context.Context, input app.ApplicationInput, actorID string) (app.TestCenterApplication, error) {
	checklistJSON, err := json.Marshal(input.ChecklistStatus)
	if err != nil {
		return app.TestCenterApplication{}, err
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO test_center_applications (
			institution_name, institution_type, location, website, responsible_person,
			phone, director_phone, assigned_admin_id, sales_referral_id, payment_option,
			payment_amount, payment_currency, profit_sharing, ceeb_code, stage,
			checklist_status, created_by
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, nullif($8, '')::uuid, nullif($9, '')::uuid, $10,
			nullif($11, '')::numeric, $12, nullif($13, ''), $14, $15,
			$16::jsonb, $17
		)
		RETURNING id::text
	`, input.InstitutionName, input.InstitutionType, input.Location, input.Website, input.ResponsiblePerson,
		input.Phone, input.DirectorPhone, input.AssignedAdminID, input.SalesReferralID, input.PaymentOption,
		input.PaymentAmount, input.PaymentCurrency, input.ProfitSharing, input.CEEBCode, input.Stage,
		string(checklistJSON), actorID)
	var id string
	if err := row.Scan(&id); err != nil {
		return app.TestCenterApplication{}, err
	}
	_ = s.LogActivity(ctx, actorID, "create", "application", id, "Created application "+input.InstitutionName)
	return s.ApplicationByID(ctx, id)
}

func (s *Postgres) UpdateApplication(ctx context.Context, id string, input app.ApplicationInput, actorID string) (app.TestCenterApplication, error) {
	checklistJSON, err := json.Marshal(input.ChecklistStatus)
	if err != nil {
		return app.TestCenterApplication{}, err
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE test_center_applications
		SET institution_name = $1,
			institution_type = $2,
			location = $3,
			website = $4,
			responsible_person = $5,
			phone = $6,
			director_phone = $7,
			assigned_admin_id = nullif($8, '')::uuid,
			sales_referral_id = nullif($9, '')::uuid,
			payment_option = $10,
			payment_amount = nullif($11, '')::numeric,
			payment_currency = $12,
			profit_sharing = nullif($13, ''),
			ceeb_code = $14,
			stage = $15,
			checklist_status = $16::jsonb,
			updated_at = now()
		WHERE id = $17
	`, input.InstitutionName, input.InstitutionType, input.Location, input.Website, input.ResponsiblePerson,
		input.Phone, input.DirectorPhone, input.AssignedAdminID, input.SalesReferralID, input.PaymentOption,
		input.PaymentAmount, input.PaymentCurrency, input.ProfitSharing, input.CEEBCode, input.Stage,
		string(checklistJSON), id)
	if err != nil {
		return app.TestCenterApplication{}, err
	}
	_ = s.LogActivity(ctx, actorID, "update", "application", id, "Updated application "+input.InstitutionName)
	return s.ApplicationByID(ctx, id)
}

func (s *Postgres) DeleteApplication(ctx context.Context, id string, actorID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM test_center_applications WHERE id = $1`, id)
	if err == nil {
		_ = s.LogActivity(ctx, actorID, "delete", "application", id, "Deleted application")
	}
	return err
}

func (s *Postgres) ListApplicationDocuments(ctx context.Context, applicationID string) ([]app.ApplicationDocument, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, application_id::text, doc_type, original_filename,
			content_type, size_bytes, storage_key, COALESCE(uploaded_by::text, ''), created_at
		FROM application_documents
		WHERE application_id = $1
		ORDER BY created_at DESC
	`, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var documents []app.ApplicationDocument
	for rows.Next() {
		var doc app.ApplicationDocument
		if err := rows.Scan(
			&doc.ID,
			&doc.ApplicationID,
			&doc.DocType,
			&doc.OriginalFilename,
			&doc.ContentType,
			&doc.SizeBytes,
			&doc.StorageKey,
			&doc.UploadedBy,
			&doc.CreatedAt,
		); err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}
	return documents, rows.Err()
}

func (s *Postgres) CreateApplicationDocument(ctx context.Context, input app.DocumentInput) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO application_documents (
			application_id, doc_type, original_filename, content_type,
			size_bytes, storage_key, uploaded_by
		) VALUES ($1, $2, $3, $4, $5, $6, nullif($7, '')::uuid)
	`, input.ApplicationID, input.DocType, input.OriginalFilename, input.ContentType, input.SizeBytes, input.StorageKey, input.UploadedBy)
	return err
}

func (s *Postgres) ListDeals(ctx context.Context) ([]app.SalesDeal, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d.id::text, d.school_name, d.capacity,
			COALESCE(d.assigned_sales_agent_id::text, ''), COALESCE(agent.name, ''),
			d.stage, d.negotiation_price::text, d.currency, COALESCE(d.profit_sharing, ''),
			d.notes, COALESCE(d.converted_application_id::text, ''), COALESCE(d.created_by::text, ''),
			d.created_at, d.updated_at
		FROM sales_deals d
		LEFT JOIN users agent ON agent.id = d.assigned_sales_agent_id
		ORDER BY d.updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var deals []app.SalesDeal
	for rows.Next() {
		deal, err := scanDeal(rows)
		if err != nil {
			return nil, err
		}
		deals = append(deals, deal)
	}
	return deals, rows.Err()
}

func (s *Postgres) DealsByStage(ctx context.Context) (map[string][]app.SalesDeal, error) {
	deals, err := s.ListDeals(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]app.SalesDeal, len(app.SalesStages))
	for _, stage := range app.SalesStages {
		out[stage] = []app.SalesDeal{}
	}
	for _, deal := range deals {
		out[deal.Stage] = append(out[deal.Stage], deal)
	}
	return out, nil
}

func (s *Postgres) DealByID(ctx context.Context, id string) (app.SalesDeal, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT d.id::text, d.school_name, d.capacity,
			COALESCE(d.assigned_sales_agent_id::text, ''), COALESCE(agent.name, ''),
			d.stage, d.negotiation_price::text, d.currency, COALESCE(d.profit_sharing, ''),
			d.notes, COALESCE(d.converted_application_id::text, ''), COALESCE(d.created_by::text, ''),
			d.created_at, d.updated_at
		FROM sales_deals d
		LEFT JOIN users agent ON agent.id = d.assigned_sales_agent_id
		WHERE d.id = $1
	`, id)
	return scanDeal(row)
}

func (s *Postgres) CreateDeal(ctx context.Context, input app.DealInput, actorID string) (app.SalesDeal, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO sales_deals (
			school_name, capacity, assigned_sales_agent_id, stage, negotiation_price,
			currency, profit_sharing, notes, created_by
		) VALUES ($1, $2, nullif($3, '')::uuid, $4, nullif($5, '')::numeric, $6, nullif($7, ''), $8, $9)
		RETURNING id::text
	`, input.SchoolName, input.Capacity, input.AssignedSalesAgentID, input.Stage, input.NegotiationPrice, input.Currency, input.ProfitSharing, input.Notes, actorID)
	var id string
	if err := row.Scan(&id); err != nil {
		return app.SalesDeal{}, err
	}
	_ = s.LogActivity(ctx, actorID, "create", "deal", id, "Created CRM deal "+input.SchoolName)
	return s.DealByID(ctx, id)
}

func (s *Postgres) UpdateDealStage(ctx context.Context, id, stage, actorID string) (app.SalesDeal, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return app.SalesDeal{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE sales_deals SET stage = $1, updated_at = now() WHERE id = $2`, stage, id)
	if err != nil {
		return app.SalesDeal{}, err
	}
	if stage == "won" {
		if err := s.convertWonDeal(ctx, tx, id, actorID); err != nil {
			return app.SalesDeal{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return app.SalesDeal{}, err
	}
	_ = s.LogActivity(ctx, actorID, "move", "deal", id, "Moved CRM deal to "+app.StageLabel(stage))
	return s.DealByID(ctx, id)
}

func (s *Postgres) convertWonDeal(ctx context.Context, tx pgx.Tx, dealID, actorID string) error {
	var applicationID sql.NullString
	var schoolName, profitSharing, currency string
	var price sql.NullString
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(converted_application_id::text, ''), school_name,
			COALESCE(profit_sharing, ''), currency, negotiation_price::text
		FROM sales_deals
		WHERE id = $1
	`, dealID).Scan(&applicationID, &schoolName, &profitSharing, &currency, &price)
	if err != nil {
		return err
	}
	if applicationID.Valid && applicationID.String != "" {
		return nil
	}
	paymentOption := "our_investment"
	if price.Valid && price.String != "" {
		paymentOption = "fixed_amount"
	}
	var newApplicationID string
	err = tx.QueryRow(ctx, `
		INSERT INTO test_center_applications (
			institution_name, institution_type, sales_referral_id, payment_option,
			payment_amount, payment_currency, profit_sharing, stage, created_by
		) VALUES ($1, 'school', $2, $3, nullif($4, '')::numeric, $5, nullif($6, ''), 'start', $7)
		RETURNING id::text
	`, schoolName, dealID, paymentOption, nullableStringValue(price), currency, profitSharing, actorID).Scan(&newApplicationID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE sales_deals SET converted_application_id = $1 WHERE id = $2`, newApplicationID, dealID)
	return err
}

func (s *Postgres) CreateReminder(ctx context.Context, input app.ReminderInput, actorID string) (app.Reminder, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO reminders (entity_type, entity_id, title, note, due_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text
	`, input.EntityType, input.EntityID, input.Title, input.Note, input.DueAt, actorID)
	var id string
	if err := row.Scan(&id); err != nil {
		return app.Reminder{}, err
	}
	_ = s.LogActivity(ctx, actorID, "create", "reminder", id, "Created reminder "+input.Title)
	return s.ReminderByID(ctx, id)
}

func (s *Postgres) ReminderByID(ctx context.Context, id string) (app.Reminder, error) {
	row := s.pool.QueryRow(ctx, reminderSelect()+` WHERE r.id = $1`, id)
	return scanReminder(row)
}

func (s *Postgres) ListRemindersForEntity(ctx context.Context, entityType, entityID string) ([]app.Reminder, error) {
	rows, err := s.pool.Query(ctx, reminderSelect()+`
		WHERE r.entity_type = $1 AND r.entity_id = $2
		ORDER BY r.due_at ASC
	`, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReminders(rows)
}

func (s *Postgres) ListDueReminders(ctx context.Context, now time.Time, limit int) ([]app.Reminder, error) {
	rows, err := s.pool.Query(ctx, reminderSelect()+`
		WHERE r.sent_at IS NULL AND r.due_at <= $1
		ORDER BY r.due_at ASC
		LIMIT $2
	`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReminders(rows)
}

func (s *Postgres) ListOpenReminders(ctx context.Context, limit int) ([]app.Reminder, error) {
	rows, err := s.pool.Query(ctx, reminderSelect()+`
		WHERE r.sent_at IS NULL
		ORDER BY r.due_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReminders(rows)
}

func (s *Postgres) MarkReminderSent(ctx context.Context, id string, sentAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE reminders
		SET sent_at = $1, updated_at = now()
		WHERE id = $2 AND sent_at IS NULL
	`, sentAt, id)
	return err
}

func (s *Postgres) DashboardStats(ctx context.Context, now time.Time) (app.DashboardStats, error) {
	stats := app.DashboardStats{
		ApplicationCounts: map[string]int{},
		DealCounts:        map[string]int{},
	}
	for _, stage := range app.ApplicationStages {
		stats.ApplicationCounts[stage] = 0
	}
	for _, stage := range app.SalesStages {
		stats.DealCounts[stage] = 0
	}

	rows, err := s.pool.Query(ctx, `SELECT stage, count(*) FROM test_center_applications GROUP BY stage`)
	if err != nil {
		return stats, err
	}
	for rows.Next() {
		var stage string
		var count int
		if err := rows.Scan(&stage, &count); err != nil {
			rows.Close()
			return stats, err
		}
		stats.ApplicationCounts[stage] = count
	}
	rows.Close()

	rows, err = s.pool.Query(ctx, `SELECT stage, count(*) FROM sales_deals GROUP BY stage`)
	if err != nil {
		return stats, err
	}
	for rows.Next() {
		var stage string
		var count int
		if err := rows.Scan(&stage, &count); err != nil {
			rows.Close()
			return stats, err
		}
		stats.DealCounts[stage] = count
	}
	rows.Close()

	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM reminders WHERE sent_at IS NULL AND due_at <= $1`, now).Scan(&stats.OverdueReminders)
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE active = true`).Scan(&stats.ActiveUsers)
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM sales_deals WHERE stage = 'won'`).Scan(&stats.WonDeals)
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM sales_deals WHERE converted_application_id IS NOT NULL`).Scan(&stats.ConvertedDeals)

	activity, err := s.RecentActivity(ctx, 10)
	if err != nil {
		return stats, err
	}
	stats.RecentActivity = activity
	reminders, err := s.ListOpenReminders(ctx, 8)
	if err != nil {
		return stats, err
	}
	stats.DueReminders = reminders
	return stats, nil
}

func (s *Postgres) RecentActivity(ctx context.Context, limit int) ([]app.ActivityLog, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT l.id::text, COALESCE(l.user_id::text, ''), COALESCE(u.name, ''),
			l.action, l.entity_type, COALESCE(l.entity_id::text, ''), l.summary, l.created_at
		FROM activity_logs l
		LEFT JOIN users u ON u.id = l.user_id
		ORDER BY l.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []app.ActivityLog
	for rows.Next() {
		var log app.ActivityLog
		if err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.UserName,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.Summary,
			&log.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (s *Postgres) LogActivity(ctx context.Context, userID, action, entityType, entityID, summary string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO activity_logs (user_id, action, entity_type, entity_id, summary)
		VALUES (nullif($1, '')::uuid, $2, $3, nullif($4, '')::uuid, $5)
	`, userID, action, entityType, entityID, summary)
	return err
}

func scanUser(row pgx.Row) (app.User, error) {
	var user app.User
	if err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return app.User{}, err
	}
	return user, nil
}

func scanCertificate(row pgx.Row) (app.Certificate, error) {
	var cert app.Certificate
	if err := row.Scan(
		&cert.ID,
		&cert.Slug,
		&cert.Institution,
		&cert.Designation,
		&cert.SATAdministrationDate,
		&cert.CollegeBoardScreenshot,
		&cert.VerificationID,
		&cert.IssueDate,
	); err != nil {
		return app.Certificate{}, err
	}
	cert.TestCenterCode = app.TestCenterCodeFromVerificationID(cert.VerificationID)
	cert.CountryRegion = "Uzbekistan"
	return cert, nil
}

func scanApplication(row pgx.Row) (app.TestCenterApplication, error) {
	var item app.TestCenterApplication
	var amount sql.NullString
	var checklist []byte
	if err := row.Scan(
		&item.ID,
		&item.InstitutionName,
		&item.InstitutionType,
		&item.Location,
		&item.Website,
		&item.ResponsiblePerson,
		&item.Phone,
		&item.DirectorPhone,
		&item.AssignedAdminID,
		&item.AssignedAdminName,
		&item.SalesReferralID,
		&item.SalesReferralName,
		&item.PaymentOption,
		&amount,
		&item.PaymentCurrency,
		&item.ProfitSharing,
		&item.CEEBCode,
		&item.Stage,
		&checklist,
		&item.CreatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return app.TestCenterApplication{}, err
	}
	if amount.Valid {
		item.PaymentAmount = &amount.String
	}
	item.ChecklistStatus = map[string]bool{}
	if len(checklist) > 0 {
		_ = json.Unmarshal(checklist, &item.ChecklistStatus)
	}
	return item, nil
}

func scanDeal(row pgx.Row) (app.SalesDeal, error) {
	var deal app.SalesDeal
	var price sql.NullString
	if err := row.Scan(
		&deal.ID,
		&deal.SchoolName,
		&deal.Capacity,
		&deal.AssignedSalesAgentID,
		&deal.AssignedSalesAgentName,
		&deal.Stage,
		&price,
		&deal.Currency,
		&deal.ProfitSharing,
		&deal.Notes,
		&deal.ConvertedApplicationID,
		&deal.CreatedBy,
		&deal.CreatedAt,
		&deal.UpdatedAt,
	); err != nil {
		return app.SalesDeal{}, err
	}
	if price.Valid {
		deal.NegotiationPrice = &price.String
	}
	return deal, nil
}

func reminderSelect() string {
	return `
		SELECT r.id::text, r.entity_type, r.entity_id::text,
			CASE
				WHEN r.entity_type = 'application' THEN COALESCE(a.institution_name, 'Application')
				WHEN r.entity_type = 'deal' THEN COALESCE(d.school_name, 'CRM deal')
				ELSE 'Reminder'
			END AS entity_name,
			CASE
				WHEN r.entity_type = 'application' THEN COALESCE(admin.name, '')
				WHEN r.entity_type = 'deal' THEN COALESCE(agent.name, '')
				ELSE ''
			END AS owner_name,
			CASE
				WHEN r.entity_type = 'application' THEN COALESCE(a.stage, '')
				WHEN r.entity_type = 'deal' THEN COALESCE(d.stage, '')
				ELSE ''
			END AS current_stage,
			r.title, r.note, r.due_at, COALESCE(r.created_by::text, ''), r.sent_at,
			r.created_at, r.updated_at
		FROM reminders r
		LEFT JOIN test_center_applications a ON r.entity_type = 'application' AND a.id = r.entity_id
		LEFT JOIN users admin ON admin.id = a.assigned_admin_id
		LEFT JOIN sales_deals d ON r.entity_type = 'deal' AND d.id = r.entity_id
		LEFT JOIN users agent ON agent.id = d.assigned_sales_agent_id
	`
}

func scanReminder(row pgx.Row) (app.Reminder, error) {
	var reminder app.Reminder
	var sent sql.NullTime
	if err := row.Scan(
		&reminder.ID,
		&reminder.EntityType,
		&reminder.EntityID,
		&reminder.EntityName,
		&reminder.OwnerName,
		&reminder.CurrentStage,
		&reminder.Title,
		&reminder.Note,
		&reminder.DueAt,
		&reminder.CreatedBy,
		&sent,
		&reminder.CreatedAt,
		&reminder.UpdatedAt,
	); err != nil {
		return app.Reminder{}, err
	}
	if sent.Valid {
		reminder.SentAt = &sent.Time
	}
	return reminder, nil
}

func scanReminders(rows pgx.Rows) ([]app.Reminder, error) {
	var reminders []app.Reminder
	for rows.Next() {
		reminder, err := scanReminder(rows)
		if err != nil {
			return nil, err
		}
		reminders = append(reminders, reminder)
	}
	return reminders, rows.Err()
}

func nullableStringValue(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
