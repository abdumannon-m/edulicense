package web

import (
	"context"
	"errors"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"edu-license/pkg/app"
	"edu-license/pkg/fx"
	"edu-license/pkg/httpx"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

// FinanceStore is the persistence the finance pages need.
type FinanceStore interface {
	LoadFinData(ctx context.Context, today time.Time) (app.FinData, error)
	UpsertFinRate(ctx context.Context, date time.Time, rate float64, source string) error
	HasFinRate(ctx context.Context, date time.Time) (bool, error)
	CreateFinTransaction(ctx context.Context, input app.FinTxnInput, actorID string) (string, error)
	UpdateFinTransaction(ctx context.Context, id string, input app.FinTxnInput, actorID string) (string, error)
	DeleteFinTransaction(ctx context.Context, id, actorID string) error
	EnsureFinCounterparty(ctx context.Context, name, kind string) (string, error)
	SaveFinCounterparty(ctx context.Context, id string, input app.FinCounterpartyInput, actorID string) (string, error)
	SaveFinCategory(ctx context.Context, input app.FinCategoryInput, actorID string) error
	SaveFinContract(ctx context.Context, id string, input app.FinContractInput, actorID string) (string, error)
	DeleteFinContract(ctx context.Context, id, actorID string) error
	SetFinMilestone(ctx context.Context, contractID string, step int, date *time.Time, actorID string) error
	CreateFinRevenue(ctx context.Context, input app.FinRevenueInput, actorID string) error
	DeleteFinRevenue(ctx context.Context, id string) error
	CreateFinAccrual(ctx context.Context, input app.FinAccrualInput, actorID string) error
	DeleteFinAccrual(ctx context.Context, id string) error
	SaveFinOpening(ctx context.Context, input app.FinOpeningInput, actorID string) error
	DeleteFinOpening(ctx context.Context, id string) error
	SaveFinRecurring(ctx context.Context, id string, input app.FinRecurringInput) error
	DeleteFinRecurring(ctx context.Context, id string) error
	CreateFinForecastItem(ctx context.Context, input app.FinForecastInput) error
	DeleteFinForecastItem(ctx context.Context, id string) error
}

func (s *Server) financeRoutes(admin chi.Router) {
	area := func(h http.HandlerFunc) http.HandlerFunc { return s.requireArea("finance", h) }
	admin.Get("/admin/finance", area(s.financeDashboard))
	admin.Get("/admin/finance/transactions", area(s.financeTransactions))
	admin.Post("/admin/finance/transactions", area(s.financeTransactionCreate))
	admin.Get("/admin/finance/transactions/{id}", area(s.financeTransactionEdit))
	admin.Post("/admin/finance/transactions/{id}", area(s.financeTransactionUpdate))
	admin.Post("/admin/finance/transactions/{id}/delete", area(s.financeTransactionDelete))
	admin.Get("/admin/finance/contracts", area(s.financeContracts))
	admin.Post("/admin/finance/contracts", area(s.financeContractCreate))
	admin.Get("/admin/finance/contracts/{id}", area(s.financeContractShow))
	admin.Post("/admin/finance/contracts/{id}", area(s.financeContractUpdate))
	admin.Post("/admin/finance/contracts/{id}/delete", area(s.financeContractDelete))
	admin.Post("/admin/finance/contracts/{id}/steps/{step}", area(s.financeMilestone))
	admin.Post("/admin/finance/contracts/{id}/work", area(s.financeWorkCreate))
	admin.Post("/admin/finance/work/{id}/delete", area(s.financeWorkDelete))
	admin.Get("/admin/finance/payables", area(s.financePayables))
	admin.Post("/admin/finance/accruals", area(s.financeAccrualCreate))
	admin.Post("/admin/finance/accruals/{id}/delete", area(s.financeAccrualDelete))
	admin.Post("/admin/finance/openings", area(s.financeOpeningSave))
	admin.Post("/admin/finance/openings/{id}/delete", area(s.financeOpeningDelete))
	admin.Get("/admin/finance/pnl", area(s.financePL))
	admin.Get("/admin/finance/forecast", area(s.financeForecast))
	admin.Post("/admin/finance/recurring", area(s.financeRecurringSave))
	admin.Post("/admin/finance/recurring/{id}", area(s.financeRecurringSave))
	admin.Post("/admin/finance/recurring/{id}/delete", area(s.financeRecurringDelete))
	admin.Post("/admin/finance/forecast-items", area(s.financeForecastItemCreate))
	admin.Post("/admin/finance/forecast-items/{id}/delete", area(s.financeForecastItemDelete))
	admin.Get("/admin/finance/settings", area(s.financeSettings))
	admin.Post("/admin/finance/rates", area(s.financeRateSave))
	admin.Post("/admin/finance/rates/fetch", area(s.financeRatesFetch))
	admin.Post("/admin/finance/categories", area(s.financeCategorySave))
	admin.Post("/admin/finance/counterparties", area(s.financeCounterpartySave))
	admin.Post("/admin/finance/counterparties/{id}", area(s.financeCounterpartySave))
}

func (s *Server) finStore() (FinanceStore, bool) {
	store, ok := s.store.(FinanceStore)
	return store, ok
}

func (s *Server) finToday() time.Time {
	return app.DateOnly(s.nowInAppTimezone())
}

// finPage loads the finance data and prepares the common page fields.
func (s *Server) finPage(w http.ResponseWriter, r *http.Request, title, tab string) (app.AdminPageData, *app.FinPage, bool) {
	data := s.adminData(w, r, title)
	data.Success = queryMessage(r.URL.Query(), "success")
	data.Error = queryMessage(r.URL.Query(), "error")
	store, ok := s.finStore()
	if !ok {
		http.Error(w, "finance is not available", http.StatusInternalServerError)
		return data, nil, false
	}
	fin, err := store.LoadFinData(r.Context(), s.finToday())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return data, nil, false
	}
	page := &app.FinPage{Tab: tab, Today: fin.Today, Data: &fin, Charts: map[string]template.HTML{}}
	page.Rate, _ = app.RateOn(fin.Rates, fin.Today)
	for i := len(fin.Rates) - 1; i >= 0; i-- {
		if !fin.Rates[i].Date.After(fin.Today) {
			page.RateDate = fin.Rates[i].Date
			break
		}
	}
	data.Fin = page
	return data, page, true
}

func (s *Server) financeDashboard(w http.ResponseWriter, r *http.Request) {
	data, page, ok := s.finPage(w, r, "Finance", "dashboard")
	if !ok {
		return
	}
	dash := page.Data.Dashboard()
	page.Dashboard = dash
	months := dash.PL.ActiveMonths(page.Today)
	var labels []string
	var income, expenses, cash []float64
	for _, m := range months {
		labels = append(labels, app.ShortMonth(dash.PL.Months[m]))
		income = append(income, dash.PL.TotalIncome.Values[m])
		expenses = append(expenses, dash.PL.TotalExpenses.Values[m])
		cash = append(cash, dash.CashFlow[m].Closing)
	}
	page.Charts["pl"] = app.BarChart(labels, []app.ChartSeries{
		{Label: "Income", Class: "fin-chart__bar--income", Values: income},
		{Label: "Expenses", Class: "fin-chart__bar--expense", Values: expenses},
	}, "$")
	// Cash line: actual month ends, then the forecast (dashed).
	cashLabels := append([]string{}, labels...)
	dashedFrom := len(cash) - 1
	for i, month := range dash.Forecast {
		if i == 0 {
			continue // the current month is already the last actual point
		}
		cashLabels = append(cashLabels, app.ShortMonth(month.Month))
		cash = append(cash, month.Closing)
	}
	page.Charts["cash"] = app.LineChart(cashLabels, cash, dashedFrom, "", 640)
	s.renderer.Render(w, http.StatusOK, "finance_dashboard", data)
}

// ---------------------------------------------------------------------------
// Transactions

func (s *Server) financeTransactions(w http.ResponseWriter, r *http.Request) {
	data, page, ok := s.finPage(w, r, "Transactions", "transactions")
	if !ok {
		return
	}
	query := r.URL.Query()
	page.Filter = app.FinTxnFilter{
		Month:          query.Get("month"),
		Account:        query.Get("account"),
		Category:       query.Get("category"),
		CounterpartyID: query.Get("counterparty"),
		Query:          strings.TrimSpace(query.Get("q")),
	}
	if !page.Filter.Active() && query.Get("all") == "" {
		page.Filter.Month = app.MonthKey(page.Today)
	}
	page.Ledger = page.Data.Ledger(page.Filter)
	page.Cash = page.Data.CashPosition()
	page.CategoryGroups = page.Data.CategoryGroups(true)
	page.ActiveContracts = page.Data.ActiveContracts()
	s.renderer.Render(w, http.StatusOK, "finance_transactions", data)
}

func (s *Server) parseFinTxn(ctx context.Context, r *http.Request, fin *app.FinData) (app.FinTxnInput, error) {
	store, _ := s.finStore()
	date, err := time.Parse("2006-01-02", strings.TrimSpace(r.FormValue("date")))
	if err != nil {
		return app.FinTxnInput{}, errors.New("Choose a valid date.")
	}
	amount, err := app.ParseAmount(r.FormValue("amount"))
	if err != nil {
		return app.FinTxnInput{}, errors.New("Enter a valid amount.")
	}
	input := app.FinTxnInput{
		Date:          date,
		Account:       r.FormValue("account"),
		Direction:     r.FormValue("direction"),
		Amount:        amount,
		CategoryCode:  r.FormValue("category"),
		ContractID:    r.FormValue("contract_id"),
		Description:   r.FormValue("description"),
		CreateAccrual: r.FormValue("create_accrual") == "1",
		ToAccount:     r.FormValue("to_account"),
	}
	if raw := strings.TrimSpace(r.FormValue("fx_rate")); raw != "" {
		if input.FxRate, err = app.ParseAmount(raw); err != nil {
			return input, errors.New("Enter a valid exchange rate.")
		}
	}
	if raw := strings.TrimSpace(r.FormValue("to_amount")); raw != "" {
		if input.ToAmount, err = app.ParseAmount(raw); err != nil {
			return input, errors.New("Enter a valid received amount.")
		}
	}
	if raw := strings.TrimSpace(r.FormValue("settles_amount")); raw != "" {
		settles, err := app.ParseAmount(raw)
		if err != nil {
			return input, errors.New("Enter a valid contract amount.")
		}
		input.SettlesAmount = &settles
	}
	if !input.AccrualApplies(fin.Category(input.CategoryCode)) {
		input.CreateAccrual = false
	}
	if input.IsTransfer() {
		return input, nil
	}
	if input.ContractID != "" {
		if _, ok := fin.Contract(input.ContractID); !ok {
			return input, errors.New("That contract no longer exists.")
		}
	}
	name := strings.Join(strings.Fields(r.FormValue("counterparty")), " ")
	switch {
	case name != "":
		id, err := store.EnsureFinCounterparty(ctx, name, app.CounterpartyKindForCategory(fin.Category(input.CategoryCode)))
		if err != nil {
			return input, err
		}
		input.CounterpartyID = id
	case input.ContractID != "":
		contract, _ := fin.Contract(input.ContractID)
		input.CounterpartyID = contract.CounterpartyID
	}
	return input, nil
}

// ensureRate saves the CBU rate for the day if it is missing. It never fails
// the request: without it the latest saved rate is used.
func (s *Server) ensureRate(ctx context.Context, date time.Time) {
	store, ok := s.finStore()
	if !ok || date.After(s.finToday()) {
		return
	}
	if exists, err := store.HasFinRate(ctx, date); err != nil || exists {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	if rate, err := fx.NewClient().USDRate(ctx, date); err == nil {
		_ = store.UpsertFinRate(ctx, date, rate, "cbu")
	}
}

func actorID(r *http.Request) string {
	user, _ := httpx.CurrentUser(r)
	return user.ID
}

// financeBack returns the form's "back" target when it is a finance page on
// this site, else fallback.
func financeBack(r *http.Request, fallback string) string {
	return safeFinanceTarget(r.FormValue("back"), fallback)
}

func safeFinanceTarget(back, fallback string) string {
	if !strings.HasPrefix(back, "/") || strings.ContainsAny(back, "\\\r\n") {
		return fallback
	}
	target, err := url.Parse(back)
	if err != nil || target.Scheme != "" || target.Host != "" || target.User != nil {
		return fallback
	}
	clean := path.Clean(target.Path)
	if clean != "/admin/finance" && !strings.HasPrefix(clean, "/admin/finance/") {
		return fallback
	}
	if target.RawQuery != "" {
		clean += "?" + target.RawQuery
	}
	return clean
}

func (s *Server) financeTransactionCreate(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	back := financeBack(r, "/admin/finance/transactions")
	store, _ := s.finStore()
	fin, err := store.LoadFinData(r.Context(), s.finToday())
	if err != nil {
		redirectWithError(w, r, back, err)
		return
	}
	input, err := s.parseFinTxn(r.Context(), r, &fin)
	if err != nil {
		redirectWithError(w, r, back, err)
		return
	}
	s.ensureRate(r.Context(), input.Date)
	if _, err := store.CreateFinTransaction(r.Context(), input, actorID(r)); err != nil {
		redirectWithError(w, r, back, err)
		return
	}
	redirectWithSuccess(w, r, back, "Saved.")
}

func (s *Server) financeTransactionEdit(w http.ResponseWriter, r *http.Request) {
	data, page, ok := s.finPage(w, r, "Edit transaction", "transactions")
	if !ok {
		return
	}
	txn, found := page.Data.Transaction(chi.URLParam(r, "id"))
	if !found {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	page.Txn = txn
	page.TxnTransfer = page.Data.TransferLegs(txn.TransferGroup)
	page.CategoryGroups = page.Data.CategoryGroups(false)
	page.ActiveContracts = page.Data.ActiveContracts()
	// Keep a cancelled contract selectable so saving does not unlink it.
	if contract, ok := page.Data.Contract(txn.ContractID); ok && contract.Status == "cancelled" {
		page.ActiveContracts = append(page.ActiveContracts, contract)
	}
	s.renderer.Render(w, http.StatusOK, "finance_transaction_edit", data)
}

func (s *Server) financeTransactionUpdate(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	id := chi.URLParam(r, "id")
	target := "/admin/finance/transactions/" + id
	store, _ := s.finStore()
	fin, err := store.LoadFinData(r.Context(), s.finToday())
	if err != nil {
		redirectWithError(w, r, target, err)
		return
	}
	input, err := s.parseFinTxn(r.Context(), r, &fin)
	if err != nil {
		redirectWithError(w, r, target, err)
		return
	}
	if original, ok := fin.Transaction(id); ok {
		input.FxRate = editedRate(original, input)
	}
	s.ensureRate(r.Context(), input.Date)
	if _, err := store.UpdateFinTransaction(r.Context(), id, input, actorID(r)); err != nil {
		redirectWithError(w, r, target, notFoundMessage(err))
		return
	}
	redirectWithSuccess(w, r, financeBack(r, "/admin/finance/transactions?month="+app.MonthKey(input.Date)), "Transaction updated.")
}

// editedRate decides the rate for an edited transaction. The form shows the
// saved rate rounded to 2 decimals, so a field within half a tiyin of it was
// left untouched: keep the exact saved rate on the same day, and use the new
// day's rate when the date moved. A cleared field also means the day's rate.
func editedRate(original app.FinTransaction, input app.FinTxnInput) float64 {
	if input.FxRate <= 0 {
		return 0
	}
	if math.Abs(input.FxRate-original.FxRate) >= 0.005 {
		return input.FxRate
	}
	if app.DateOnly(input.Date).Equal(app.DateOnly(original.Date)) {
		return original.FxRate
	}
	return 0
}

func (s *Server) financeTransactionDelete(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	store, _ := s.finStore()
	back := financeBack(r, "/admin/finance/transactions")
	if err := store.DeleteFinTransaction(r.Context(), chi.URLParam(r, "id"), actorID(r)); err != nil {
		redirectWithError(w, r, back, notFoundMessage(err))
		return
	}
	redirectWithSuccess(w, r, back, "Transaction deleted.")
}

func notFoundMessage(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.New("That record no longer exists.")
	}
	return err
}

// ---------------------------------------------------------------------------
// Contracts and work done

func (s *Server) financeContracts(w http.ResponseWriter, r *http.Request) {
	data, page, ok := s.finPage(w, r, "Contracts", "contracts")
	if !ok {
		return
	}
	page.Contracts = page.Data.ContractTotals()
	s.renderer.Render(w, http.StatusOK, "finance_contracts", data)
}

func (s *Server) parseFinContract(ctx context.Context, r *http.Request) (app.FinContractInput, error) {
	store, _ := s.finStore()
	amount, err := app.ParseAmount(r.FormValue("amount"))
	if err != nil {
		return app.FinContractInput{}, errors.New("Enter the contract amount.")
	}
	signed, err := time.Parse("2006-01-02", r.FormValue("signed_on"))
	if err != nil {
		return app.FinContractInput{}, errors.New("Choose the signed date.")
	}
	input := app.FinContractInput{
		ApplicationID: r.FormValue("application_id"),
		Title:         r.FormValue("title"),
		Amount:        amount,
		Currency:      defaultString(r.FormValue("currency"), "USD"),
		SignedOn:      signed,
		Status:        r.FormValue("status"),
		Notes:         r.FormValue("notes"),
		Weights:       app.DefaultFinWeights(),
	}
	for i := 0; i < 4; i++ {
		raw := strings.TrimSpace(r.FormValue("w" + strconv.Itoa(i+1)))
		if raw == "" {
			continue
		}
		if input.Weights[i], err = app.ParseAmount(raw); err != nil {
			return input, errors.New("Step shares must be numbers.")
		}
	}
	name := strings.TrimSpace(r.FormValue("school"))
	if name == "" {
		return input, errors.New("Enter the school name.")
	}
	input.CounterpartyID, err = store.EnsureFinCounterparty(ctx, name, "school")
	return input, err
}

func (s *Server) financeContractCreate(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	store, _ := s.finStore()
	input, err := s.parseFinContract(r.Context(), r)
	if err != nil {
		redirectWithError(w, r, "/admin/finance/contracts", err)
		return
	}
	id, err := store.SaveFinContract(r.Context(), "", input, actorID(r))
	if err != nil {
		redirectWithError(w, r, "/admin/finance/contracts", err)
		return
	}
	redirectWithSuccess(w, r, "/admin/finance/contracts/"+id, "Contract added.")
}

func (s *Server) financeContractShow(w http.ResponseWriter, r *http.Request) {
	data, page, ok := s.finPage(w, r, "Contract", "contracts")
	if !ok {
		return
	}
	contract, found := page.Data.Contract(chi.URLParam(r, "id"))
	if !found {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	page.Contract = page.Data.ContractSummary(contract)
	data.Title = contract.CounterpartyName
	s.renderer.Render(w, http.StatusOK, "finance_contract", data)
}

func (s *Server) financeContractUpdate(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	id := chi.URLParam(r, "id")
	target := "/admin/finance/contracts/" + id
	store, _ := s.finStore()
	input, err := s.parseFinContract(r.Context(), r)
	if err != nil {
		redirectWithError(w, r, target, err)
		return
	}
	if _, err := store.SaveFinContract(r.Context(), id, input, actorID(r)); err != nil {
		redirectWithError(w, r, target, notFoundMessage(err))
		return
	}
	redirectWithSuccess(w, r, target, "Contract saved.")
}

func (s *Server) financeContractDelete(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	store, _ := s.finStore()
	if err := store.DeleteFinContract(r.Context(), chi.URLParam(r, "id"), actorID(r)); err != nil {
		redirectWithError(w, r, "/admin/finance/contracts", notFoundMessage(err))
		return
	}
	redirectWithSuccess(w, r, "/admin/finance/contracts", "Contract deleted.")
}

// financeMilestone records the day a step was reached and, by default, books
// that step's share of the contract as work done in that month.
func (s *Server) financeMilestone(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	id := chi.URLParam(r, "id")
	target := "/admin/finance/contracts/" + id
	step, err := strconv.Atoi(chi.URLParam(r, "step"))
	if err != nil || step < 1 || step > 4 {
		redirectWithError(w, r, target, errors.New("Unknown step."))
		return
	}
	store, _ := s.finStore()
	var date *time.Time
	if raw := strings.TrimSpace(r.FormValue("date")); raw != "" && r.FormValue("clear") == "" {
		parsed, err := time.Parse("2006-01-02", raw)
		if err != nil {
			redirectWithError(w, r, target, errors.New("Choose a valid date."))
			return
		}
		date = &parsed
	}
	if err := store.SetFinMilestone(r.Context(), id, step, date, actorID(r)); err != nil {
		redirectWithError(w, r, target, notFoundMessage(err))
		return
	}
	if date == nil {
		redirectWithSuccess(w, r, target, "Step cleared. Remove its work-done line below if it was booked.")
		return
	}
	if r.FormValue("record_work") == "1" {
		fin, err := store.LoadFinData(r.Context(), s.finToday())
		if err == nil {
			contract, _ := fin.Contract(id)
			already := false
			for _, rev := range fin.Revenue {
				if rev.ContractID == id && rev.Milestone == step {
					already = true
				}
			}
			amount := app.Round2(contract.Amount * contract.Weights[step-1] / 100)
			if !already && amount > 0 {
				s.ensureRate(r.Context(), *date)
				err = store.CreateFinRevenue(r.Context(), app.FinRevenueInput{ContractID: id, Date: *date, Amount: amount, Milestone: step}, actorID(r))
			}
		}
		if err != nil {
			redirectWithError(w, r, target, err)
			return
		}
	}
	redirectWithSuccess(w, r, target, "Step saved.")
}

func (s *Server) financeWorkCreate(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	id := chi.URLParam(r, "id")
	target := "/admin/finance/contracts/" + id
	month, err := app.ParseMonth(r.FormValue("month"))
	if err != nil {
		redirectWithError(w, r, target, errors.New("Choose the month the work was done."))
		return
	}
	amount, err := app.ParseAmount(r.FormValue("amount"))
	if err != nil {
		redirectWithError(w, r, target, errors.New("Enter the amount of work done."))
		return
	}
	step, _ := strconv.Atoi(r.FormValue("step"))
	// Book on the last day of the month unless that is in the future.
	date := month.AddDate(0, 1, -1)
	if today := s.finToday(); date.After(today) && !month.After(today) {
		date = today
	}
	s.ensureRate(r.Context(), date)
	store, _ := s.finStore()
	input := app.FinRevenueInput{ContractID: id, Date: date, Amount: amount, Milestone: step, Note: r.FormValue("note")}
	if err := store.CreateFinRevenue(r.Context(), input, actorID(r)); err != nil {
		redirectWithError(w, r, target, notFoundMessage(err))
		return
	}
	redirectWithSuccess(w, r, target, "Work done added to "+app.MonthLabel(month)+".")
}

func (s *Server) financeWorkDelete(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	store, _ := s.finStore()
	back := financeBack(r, "/admin/finance/contracts")
	if err := store.DeleteFinRevenue(r.Context(), chi.URLParam(r, "id")); err != nil {
		redirectWithError(w, r, back, notFoundMessage(err))
		return
	}
	redirectWithSuccess(w, r, back, "Work done line removed.")
}

// ---------------------------------------------------------------------------
// Payables, founders and lenders

func (s *Server) financePayables(w http.ResponseWriter, r *http.Request) {
	data, page, ok := s.finPage(w, r, "Who we owe", "payables")
	if !ok {
		return
	}
	page.Payables = page.Data.Payables()
	page.Founders = page.Data.FounderBalances()
	page.Lenders = page.Data.LenderBalances()
	for i := len(page.Data.Accruals) - 1; i >= 0 && len(page.RecentAccruals) < 12; i-- {
		if page.Data.Accruals[i].TransactionID == "" {
			page.RecentAccruals = append(page.RecentAccruals, page.Data.Accruals[i])
		}
	}
	page.CategoryGroups = page.Data.CategoryGroups(false)
	s.renderer.Render(w, http.StatusOK, "finance_payables", data)
}

func (s *Server) financeAccrualCreate(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	target := "/admin/finance/payables"
	store, _ := s.finStore()
	month, err := app.ParseMonth(r.FormValue("month"))
	if err != nil {
		redirectWithError(w, r, target, errors.New("Choose the month."))
		return
	}
	amount, err := app.ParseAmount(r.FormValue("amount"))
	if err != nil {
		redirectWithError(w, r, target, errors.New("Enter the amount owed."))
		return
	}
	name := strings.TrimSpace(r.FormValue("counterparty"))
	if name == "" {
		redirectWithError(w, r, target, errors.New("Enter who is owed."))
		return
	}
	fin, err := store.LoadFinData(r.Context(), s.finToday())
	if err != nil {
		redirectWithError(w, r, target, err)
		return
	}
	category := fin.Category(r.FormValue("category"))
	counterpartyID, err := store.EnsureFinCounterparty(r.Context(), name, app.CounterpartyKindForCategory(category))
	if err != nil {
		redirectWithError(w, r, target, err)
		return
	}
	date := month.AddDate(0, 1, -1)
	if today := s.finToday(); date.After(today) && !month.After(today) {
		date = today
	}
	s.ensureRate(r.Context(), date)
	input := app.FinAccrualInput{Date: date, CategoryCode: category.Code, CounterpartyID: counterpartyID, Amount: amount, Currency: r.FormValue("currency"), Note: r.FormValue("note")}
	if err := store.CreateFinAccrual(r.Context(), input, actorID(r)); err != nil {
		redirectWithError(w, r, target, err)
		return
	}
	redirectWithSuccess(w, r, target, "Amount owed added.")
}

func (s *Server) financeAccrualDelete(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	store, _ := s.finStore()
	if err := store.DeleteFinAccrual(r.Context(), chi.URLParam(r, "id")); err != nil {
		redirectWithError(w, r, "/admin/finance/payables", notFoundMessage(err))
		return
	}
	redirectWithSuccess(w, r, "/admin/finance/payables", "Removed.")
}

func (s *Server) financeOpeningSave(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	target := "/admin/finance/payables"
	store, _ := s.finStore()
	asOf, err := time.Parse("2006-01-02", r.FormValue("as_of"))
	if err != nil {
		redirectWithError(w, r, target, errors.New("Choose the date of the balance."))
		return
	}
	amount, err := app.ParseAmount(strings.TrimPrefix(strings.TrimSpace(r.FormValue("amount")), "-"))
	if err != nil {
		redirectWithError(w, r, target, errors.New("Enter the balance."))
		return
	}
	if r.FormValue("direction") == "they_owe" {
		amount = -amount
	}
	kind := defaultString(r.FormValue("kind"), "founder")
	counterpartyID, err := store.EnsureFinCounterparty(r.Context(), r.FormValue("counterparty"), kind)
	if err != nil {
		redirectWithError(w, r, target, err)
		return
	}
	input := app.FinOpeningInput{CounterpartyID: counterpartyID, AsOf: asOf, AmountUZS: amount, Note: r.FormValue("note")}
	if err := store.SaveFinOpening(r.Context(), input, actorID(r)); err != nil {
		redirectWithError(w, r, target, err)
		return
	}
	redirectWithSuccess(w, r, target, "Balance saved.")
}

func (s *Server) financeOpeningDelete(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	store, _ := s.finStore()
	if err := store.DeleteFinOpening(r.Context(), chi.URLParam(r, "id")); err != nil {
		redirectWithError(w, r, "/admin/finance/payables", notFoundMessage(err))
		return
	}
	redirectWithSuccess(w, r, "/admin/finance/payables", "Removed.")
}

// ---------------------------------------------------------------------------
// Reports

func (s *Server) financePL(w http.ResponseWriter, r *http.Request) {
	data, page, ok := s.finPage(w, r, "Profit and loss", "pnl")
	if !ok {
		return
	}
	page.Year = page.Today.Year()
	if year, err := strconv.Atoi(r.URL.Query().Get("year")); err == nil && year > 2000 && year < 2100 {
		page.Year = year
	}
	page.Currency = "USD"
	if r.URL.Query().Get("currency") == "UZS" {
		page.Currency = "UZS"
	}
	page.Years = page.Data.Years()
	page.PL = page.Data.PL(page.Year, page.Currency)
	page.PLMonths = page.PL.ActiveMonths(page.Today)
	page.CashFlow = page.Data.CashFlow(page.Year)
	s.renderer.Render(w, http.StatusOK, "finance_pnl", data)
}

func (s *Server) financeForecast(w http.ResponseWriter, r *http.Request) {
	data, page, ok := s.finPage(w, r, "Forecast", "forecast")
	if !ok {
		return
	}
	page.Cash = page.Data.CashPosition()
	page.Forecast = page.Data.Forecast(6, page.Cash.TotalUZS)
	page.MonthlyFixed = page.Data.MonthlyRecurringUZS()
	if len(page.Forecast) > 0 {
		page.ForecastEnd = page.Forecast[len(page.Forecast)-1]
	}
	start := app.MonthStart(page.Today)
	for _, item := range page.Data.ForecastItems {
		if !item.Month.Before(start) {
			page.UpcomingItems = append(page.UpcomingItems, item)
		}
	}
	var labels []string
	var values []float64
	for _, month := range page.Forecast {
		labels = append(labels, app.ShortMonth(month.Month))
		values = append(values, month.Closing)
	}
	page.Charts["forecast"] = app.LineChart(labels, values, 0, "", 1240)
	s.renderer.Render(w, http.StatusOK, "finance_forecast", data)
}

func (s *Server) financeRecurringSave(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	target := "/admin/finance/forecast"
	store, _ := s.finStore()
	amount, err := app.ParseAmount(r.FormValue("amount"))
	if err != nil {
		redirectWithError(w, r, target, errors.New("Enter the monthly amount."))
		return
	}
	input := app.FinRecurringInput{Label: r.FormValue("label"), Amount: amount, Currency: r.FormValue("currency"), Active: r.FormValue("active") != "0"}
	if err := store.SaveFinRecurring(r.Context(), chi.URLParam(r, "id"), input); err != nil {
		redirectWithError(w, r, target, notFoundMessage(err))
		return
	}
	redirectWithSuccess(w, r, target, "Monthly cost saved.")
}

func (s *Server) financeRecurringDelete(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	store, _ := s.finStore()
	if err := store.DeleteFinRecurring(r.Context(), chi.URLParam(r, "id")); err != nil {
		redirectWithError(w, r, "/admin/finance/forecast", notFoundMessage(err))
		return
	}
	redirectWithSuccess(w, r, "/admin/finance/forecast", "Removed.")
}

func (s *Server) financeForecastItemCreate(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	target := "/admin/finance/forecast"
	store, _ := s.finStore()
	month, err := app.ParseMonth(r.FormValue("month"))
	if err != nil {
		redirectWithError(w, r, target, errors.New("Choose the month."))
		return
	}
	amount, err := app.ParseAmount(r.FormValue("amount"))
	if err != nil {
		redirectWithError(w, r, target, errors.New("Enter the amount."))
		return
	}
	input := app.FinForecastInput{Month: month, Direction: r.FormValue("direction"), Label: r.FormValue("label"), Amount: amount, Currency: r.FormValue("currency")}
	if err := store.CreateFinForecastItem(r.Context(), input); err != nil {
		redirectWithError(w, r, target, err)
		return
	}
	redirectWithSuccess(w, r, target, "Added to the forecast.")
}

func (s *Server) financeForecastItemDelete(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	store, _ := s.finStore()
	if err := store.DeleteFinForecastItem(r.Context(), chi.URLParam(r, "id")); err != nil {
		redirectWithError(w, r, "/admin/finance/forecast", notFoundMessage(err))
		return
	}
	redirectWithSuccess(w, r, "/admin/finance/forecast", "Removed.")
}

// ---------------------------------------------------------------------------
// Settings: rates, categories, people

func (s *Server) financeSettings(w http.ResponseWriter, r *http.Request) {
	data, page, ok := s.finPage(w, r, "Finance settings", "settings")
	if !ok {
		return
	}
	for i := len(page.Data.Rates) - 1; i >= 0 && len(page.RecentRates) < 10; i-- {
		page.RecentRates = append(page.RecentRates, page.Data.Rates[i])
	}
	page.CategoryGroups = page.Data.CategoryGroups(true)
	s.renderer.Render(w, http.StatusOK, "finance_settings", data)
}

func (s *Server) financeRateSave(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	store, _ := s.finStore()
	date, err := time.Parse("2006-01-02", r.FormValue("date"))
	if err != nil {
		redirectWithError(w, r, "/admin/finance/settings", errors.New("Choose the date."))
		return
	}
	rate, err := app.ParseAmount(r.FormValue("rate"))
	if err != nil || rate <= 0 {
		redirectWithError(w, r, "/admin/finance/settings", errors.New("Enter the rate in so'm per dollar."))
		return
	}
	if err := store.UpsertFinRate(r.Context(), date, rate, "manual"); err != nil {
		redirectWithError(w, r, "/admin/finance/settings", err)
		return
	}
	redirectWithSuccess(w, r, "/admin/finance/settings", "Rate saved.")
}

// financeRatesFetch downloads missing CBU rates for the last days.
func (s *Server) financeRatesFetch(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	store, _ := s.finStore()
	days, _ := strconv.Atoi(r.FormValue("days"))
	if days <= 0 || days > 60 {
		days = 7
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	client := fx.NewClient()
	saved, failed := 0, 0
	today := s.finToday()
	for i := 0; i < days && ctx.Err() == nil; i++ {
		date := today.AddDate(0, 0, -i)
		if exists, err := store.HasFinRate(ctx, date); err == nil && exists && i > 0 {
			continue
		}
		rate, err := client.USDRate(ctx, date)
		if err != nil {
			failed++
			continue
		}
		if err := store.UpsertFinRate(ctx, date, rate, "cbu"); err == nil {
			saved++
		}
	}
	if saved == 0 && failed > 0 {
		redirectWithError(w, r, "/admin/finance/settings", errors.New("Could not reach cbu.uz. Enter the rate by hand."))
		return
	}
	redirectWithSuccess(w, r, "/admin/finance/settings", "Saved "+strconv.Itoa(saved)+" CBU rate(s).")
}

func (s *Server) financeCategorySave(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	store, _ := s.finStore()
	order, _ := strconv.Atoi(defaultString(r.FormValue("sort_order"), "100"))
	input := app.FinCategoryInput{
		Code:         r.FormValue("code"),
		Name:         r.FormValue("name"),
		Kind:         r.FormValue("kind"),
		PLGroup:      r.FormValue("pl_group"),
		AccrualBased: r.FormValue("accrual_based") == "1",
		SortOrder:    order,
		Active:       r.FormValue("active") != "0",
	}
	if err := store.SaveFinCategory(r.Context(), input, actorID(r)); err != nil {
		redirectWithError(w, r, "/admin/finance/settings", err)
		return
	}
	redirectWithSuccess(w, r, "/admin/finance/settings", "Category saved.")
}

func (s *Server) financeCounterpartySave(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	store, _ := s.finStore()
	input := app.FinCounterpartyInput{
		Name:          r.FormValue("name"),
		Kind:          r.FormValue("kind"),
		ApplicationID: r.FormValue("application_id"),
		Active:        r.FormValue("active") != "0",
		Notes:         r.FormValue("notes"),
	}
	if _, err := store.SaveFinCounterparty(r.Context(), chi.URLParam(r, "id"), input, actorID(r)); err != nil {
		redirectWithError(w, r, "/admin/finance/settings", notFoundMessage(err))
		return
	}
	redirectWithSuccess(w, r, "/admin/finance/settings", "Saved.")
}
