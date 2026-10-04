package app

import (
	"math"
	"sort"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Account balances

type FinAccountBalance struct {
	Account    FinAccount
	Balance    float64 // in the account's own currency
	BalanceUZS float64 // valued at today's rate
	Count      int
}

type FinCashPosition struct {
	Accounts []FinAccountBalance
	TotalUZS float64
	TotalUSD float64
	Rate     float64
}

func (d *FinData) CashPosition() FinCashPosition {
	rate, _ := RateOn(d.Rates, d.Today)
	balances := make(map[string]*FinAccountBalance, len(FinAccounts))
	position := FinCashPosition{Rate: rate}
	for _, account := range FinAccounts {
		balances[account.Code] = &FinAccountBalance{Account: account}
	}
	for _, txn := range d.Transactions {
		balance, ok := balances[txn.Account]
		if !ok {
			continue
		}
		balance.Balance += txn.Signed() * txn.Amount
		balance.Count++
	}
	for _, account := range FinAccounts {
		balance := balances[account.Code]
		balance.Balance = Round2(balance.Balance)
		if account.Currency == "USD" {
			balance.BalanceUZS = Round2(balance.Balance * rate)
		} else {
			balance.BalanceUZS = balance.Balance
		}
		position.TotalUZS += balance.BalanceUZS
		position.Accounts = append(position.Accounts, *balance)
	}
	position.TotalUZS = Round2(position.TotalUZS)
	if rate > 0 {
		position.TotalUSD = Round2(position.TotalUZS / rate)
	}
	return position
}

// ---------------------------------------------------------------------------
// Contracts, work done and milestones

type FinMilestoneStatus struct {
	FinMilestone
	Weight   float64
	Done     bool
	On       *time.Time
	FromApp  bool
	Unlocked float64 // cumulative unlocked share after this step, 0-1
}

type FinContractSummary struct {
	Contract   FinContract
	Recognized float64 // work done, contract currency
	Paid       float64 // settled by payments, contract currency
	Balance    float64 // recognized - paid: >0 school owes us, <0 advance received
	Receivable float64
	Advance    float64
	Unbilled   float64 // contract amount not yet recognized
	Unpaid     float64 // contract amount not yet paid
	Unlocked   float64 // 0-1
	BlockedUZS float64 // money received that the milestone rule still locks
	Milestones []FinMilestoneStatus
	Payments   []FinTransaction
	Revenue    []FinRevenue
	PaidUZS    float64
}

func (s FinContractSummary) DonePercent() float64 {
	if s.Contract.Amount <= 0 {
		return 0
	}
	return math.Min(100, s.Recognized/s.Contract.Amount*100)
}

func (s FinContractSummary) PaidPercent() float64 {
	if s.Contract.Amount <= 0 {
		return 0
	}
	return math.Min(100, s.Paid/s.Contract.Amount*100)
}

// PaymentValue is how much of a contract a payment settles, in contract currency.
func PaymentValue(txn FinTransaction, contractCurrency string) float64 {
	if txn.SettlesAmount != nil {
		return txn.Signed() * *txn.SettlesAmount
	}
	if contractCurrency == "USD" {
		return txn.Signed() * txn.AmountUSD
	}
	return txn.Signed() * txn.AmountUZS
}

func (d *FinData) ContractSummary(contract FinContract) FinContractSummary {
	rate, _ := RateOn(d.Rates, d.Today)
	summary := FinContractSummary{Contract: contract}
	for _, rev := range d.Revenue {
		if rev.ContractID == contract.ID {
			summary.Recognized += rev.Amount
			summary.Revenue = append(summary.Revenue, rev)
		}
	}
	for _, txn := range d.Transactions {
		if txn.ContractID != contract.ID {
			continue
		}
		summary.Paid += PaymentValue(txn, contract.Currency)
		summary.PaidUZS += txn.Signed() * txn.AmountUZS
		summary.Payments = append(summary.Payments, txn)
	}
	summary.Recognized = Round2(summary.Recognized)
	summary.Paid = Round2(summary.Paid)
	summary.Balance = Round2(summary.Recognized - summary.Paid)
	if summary.Balance > 0 {
		summary.Receivable = summary.Balance
	} else {
		summary.Advance = -summary.Balance
	}
	if contract.Status != "cancelled" {
		summary.Unbilled = math.Max(0, Round2(contract.Amount-summary.Recognized))
		summary.Unpaid = math.Max(0, Round2(contract.Amount-summary.Paid))
	}
	summary.Unlocked = contract.UnlockedShare()
	cumulative := 0.0
	for i, milestone := range FinMilestones {
		status := FinMilestoneStatus{FinMilestone: milestone, Weight: contract.Weights[i]}
		status.Done = contract.MilestoneDone(milestone.Number)
		status.On = contract.MilestoneOn[i]
		status.FromApp = status.Done && status.On == nil && contract.Status != "completed"
		if status.Done {
			cumulative += contract.Weights[i]
		}
		status.Unlocked = math.Min(1, cumulative/100)
		summary.Milestones = append(summary.Milestones, status)
	}
	// Money received beyond the value of the steps already done stays locked.
	// A school that paid exactly for its finished steps has nothing locked.
	if summary.Paid > 0 && summary.Unlocked < 1 {
		locked := math.Max(0, summary.Paid-contract.Amount*summary.Unlocked)
		if contract.Currency == "USD" {
			summary.BlockedUZS = Round2(locked * rate)
		} else {
			summary.BlockedUZS = Round2(locked)
		}
	}
	sort.Slice(summary.Payments, func(i, j int) bool { return summary.Payments[i].Date.Before(summary.Payments[j].Date) })
	sort.Slice(summary.Revenue, func(i, j int) bool { return summary.Revenue[i].Date.Before(summary.Revenue[j].Date) })
	return summary
}

type FinContractTotals struct {
	Contracts   []FinContractSummary
	AmountUSD   float64
	Recognized  float64
	Paid        float64
	Receivable  float64
	Advance     float64
	Unpaid      float64
	BlockedUZS  float64
	Unallocated []FinTransaction
}

// ContractTotals summarises every contract. USD totals include only USD
// contracts; UZS contracts are converted at today's rate.
func (d *FinData) ContractTotals() FinContractTotals {
	rate, _ := RateOn(d.Rates, d.Today)
	totals := FinContractTotals{}
	toUSD := func(value float64, currency string) float64 {
		if currency == "USD" || rate <= 0 {
			return value
		}
		return value / rate
	}
	for _, contract := range d.Contracts {
		summary := d.ContractSummary(contract)
		totals.Contracts = append(totals.Contracts, summary)
		if contract.Status == "cancelled" {
			continue
		}
		totals.AmountUSD += toUSD(contract.Amount, contract.Currency)
		totals.Recognized += toUSD(summary.Recognized, contract.Currency)
		totals.Paid += toUSD(summary.Paid, contract.Currency)
		totals.Receivable += toUSD(summary.Receivable, contract.Currency)
		totals.Advance += toUSD(summary.Advance, contract.Currency)
		totals.Unpaid += toUSD(summary.Unpaid, contract.Currency)
		totals.BlockedUZS += summary.BlockedUZS
	}
	for _, txn := range d.Transactions {
		if txn.ContractID == "" && d.Category(txn.CategoryCode).Kind == KindRevenue {
			totals.Unallocated = append(totals.Unallocated, txn)
		}
	}
	sort.SliceStable(totals.Contracts, func(i, j int) bool {
		a, b := totals.Contracts[i].Contract, totals.Contracts[j].Contract
		if a.Status != b.Status {
			return statusRank(a.Status) < statusRank(b.Status)
		}
		return a.SignedOn.After(b.SignedOn)
	})
	return totals
}

func statusRank(status string) int {
	switch status {
	case "active":
		return 0
	case "completed":
		return 1
	}
	return 2
}

// ---------------------------------------------------------------------------
// Payables: salaries, bonuses and services owed

type FinPayable struct {
	CounterpartyID   string
	CounterpartyName string
	Currency         string // basis currency of the balance
	Accrued          float64
	Paid             float64
	Balance          float64 // >0 we owe them
	BalanceUZS       float64
	Categories       []string
	LastAccrual      time.Time
}

func (d *FinData) Payables() []FinPayable {
	rate, _ := RateOn(d.Rates, d.Today)
	type acc struct {
		payable            FinPayable
		accruedUZS, accUSD float64
		paidUZS, paidUSD   float64
		usdRows, uzsRows   int
		categories         map[string]bool
	}
	byCounterparty := map[string]*acc{}
	get := func(id, name string) *acc {
		if id == "" {
			id = "-"
		}
		entry, ok := byCounterparty[id]
		if !ok {
			entry = &acc{payable: FinPayable{CounterpartyID: id, CounterpartyName: name}, categories: map[string]bool{}}
			byCounterparty[id] = entry
		}
		if entry.payable.CounterpartyName == "" {
			entry.payable.CounterpartyName = name
		}
		return entry
	}
	for _, accrual := range d.Accruals {
		entry := get(accrual.CounterpartyID, accrual.CounterpartyName)
		entry.accruedUZS += accrual.AmountUZS
		entry.accUSD += accrual.AmountUSD
		if accrual.Currency == "USD" {
			entry.usdRows++
		} else {
			entry.uzsRows++
		}
		entry.categories[d.Category(accrual.CategoryCode).Name] = true
		if accrual.Date.After(entry.payable.LastAccrual) {
			entry.payable.LastAccrual = accrual.Date
		}
	}
	for _, txn := range d.Transactions {
		if !d.Category(txn.CategoryCode).AccrualBased || txn.CounterpartyID == "" {
			continue
		}
		entry, ok := byCounterparty[txn.CounterpartyID]
		if !ok {
			entry = get(txn.CounterpartyID, txn.CounterpartyName)
		}
		// Money paid out settles what we owe; money coming back reverses it.
		entry.paidUZS -= txn.Signed() * txn.AmountUZS
		entry.paidUSD -= txn.Signed() * txn.AmountUSD
	}
	var out []FinPayable
	for _, entry := range byCounterparty {
		payable := entry.payable
		if entry.usdRows > entry.uzsRows {
			payable.Currency = "USD"
			payable.Accrued, payable.Paid = Round2(entry.accUSD), Round2(entry.paidUSD)
			payable.Balance = Round2(payable.Accrued - payable.Paid)
			payable.BalanceUZS = Round2(payable.Balance * rate)
		} else {
			payable.Currency = "UZS"
			payable.Accrued, payable.Paid = Round2(entry.accruedUZS), Round2(entry.paidUZS)
			payable.Balance = Round2(payable.Accrued - payable.Paid)
			payable.BalanceUZS = payable.Balance
		}
		for name := range entry.categories {
			payable.Categories = append(payable.Categories, name)
		}
		sort.Strings(payable.Categories)
		out = append(out, payable)
	}
	sort.Slice(out, func(i, j int) bool {
		if (math.Abs(out[i].BalanceUZS) > 1000) != (math.Abs(out[j].BalanceUZS) > 1000) {
			return math.Abs(out[i].BalanceUZS) > 1000
		}
		return out[i].CounterpartyName < out[j].CounterpartyName
	})
	return out
}

// PayablesOwedUZS sums what the company owes on salaries and services,
// ignoring rounding noise below 1 000 so'm.
func PayablesOwedUZS(payables []FinPayable) float64 {
	total := 0.0
	for _, payable := range payables {
		if payable.BalanceUZS > 1000 {
			total += payable.BalanceUZS
		}
	}
	return Round2(total)
}

// ---------------------------------------------------------------------------
// Founders and lenders

type FinPartyBalance struct {
	Counterparty     FinCounterparty
	Opening          *FinOpening
	Movement         float64 // since opening (or all time), UZS; + increases what we owe
	Balance          float64 // >0 company owes them, <0 they owe the company
	LifetimeDraws    float64
	LoansOutstanding float64 // founders/staff: loans given minus returned
	Transactions     int
}

func (d *FinData) openingFor(counterpartyID string) *FinOpening {
	for i := range d.Openings {
		if d.Openings[i].CounterpartyID == counterpartyID {
			return &d.Openings[i]
		}
	}
	return nil
}

// FounderBalances: opening balance minus founder draws since the opening date.
// Charter capital is equity and does not change what the company owes.
func (d *FinData) FounderBalances() []FinPartyBalance {
	var out []FinPartyBalance
	for _, counterparty := range d.Counterparties {
		if counterparty.Kind != "founder" {
			continue
		}
		balance := FinPartyBalance{Counterparty: counterparty, Opening: d.openingFor(counterparty.ID)}
		for _, txn := range d.Transactions {
			if txn.CounterpartyID != counterparty.ID {
				continue
			}
			kind := d.Category(txn.CategoryCode).Kind
			switch kind {
			case KindOwnerDraw:
				balance.LifetimeDraws -= txn.Signed() * txn.AmountUZS
				if balance.Opening == nil || !txn.Date.Before(balance.Opening.AsOf) {
					balance.Movement += txn.Signed() * txn.AmountUZS
					balance.Transactions++
				}
			case KindLoanGiven, KindLoanReturned:
				balance.LoansOutstanding -= txn.Signed() * txn.AmountUZS
			}
		}
		opening := 0.0
		if balance.Opening != nil {
			opening = balance.Opening.AmountUZS
		}
		balance.LifetimeDraws = Round2(balance.LifetimeDraws)
		balance.LoansOutstanding = Round2(balance.LoansOutstanding)
		balance.Movement = Round2(balance.Movement)
		balance.Balance = Round2(opening + balance.Movement)
		out = append(out, balance)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Counterparty.Name < out[j].Counterparty.Name })
	return out
}

// LenderBalances covers everyone the company borrowed from, plus anyone with an
// opening balance who is not a founder.
func (d *FinData) LenderBalances() []FinPartyBalance {
	var out []FinPartyBalance
	for _, counterparty := range d.Counterparties {
		if counterparty.Kind == "founder" {
			continue
		}
		opening := d.openingFor(counterparty.ID)
		balance := FinPartyBalance{Counterparty: counterparty, Opening: opening}
		relevant := opening != nil || counterparty.Kind == "lender"
		for _, txn := range d.Transactions {
			if txn.CounterpartyID != counterparty.ID {
				continue
			}
			kind := d.Category(txn.CategoryCode).Kind
			switch kind {
			case KindLoanReceived, KindLoanRepaid:
				relevant = true
				if opening == nil || !txn.Date.Before(opening.AsOf) {
					balance.Movement += txn.Signed() * txn.AmountUZS
					balance.Transactions++
				}
			case KindLoanGiven, KindLoanReturned:
				relevant = true
				balance.LoansOutstanding -= txn.Signed() * txn.AmountUZS
			}
		}
		if !relevant {
			continue
		}
		start := 0.0
		if opening != nil {
			start = opening.AmountUZS
		}
		balance.Movement = Round2(balance.Movement)
		balance.LoansOutstanding = Round2(balance.LoansOutstanding)
		balance.Balance = Round2(start + balance.Movement)
		out = append(out, balance)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Counterparty.Name < out[j].Counterparty.Name })
	return out
}

// ---------------------------------------------------------------------------
// Monthly P&L (accrual basis, like Forma 2)

type FinPLRow struct {
	Label  string
	Code   string
	Values [12]float64
	Total  float64
	Strong bool
	Muted  bool
}

func (r *FinPLRow) add(month int, value float64) {
	r.Values[month] += value
	r.Total += value
}

func (r FinPLRow) IsZero() bool {
	for _, value := range r.Values {
		if math.Abs(value) >= 0.005 {
			return false
		}
	}
	return true
}

type FinPLSection struct {
	Label    string
	Group    string
	Rows     []FinPLRow
	Subtotal FinPLRow
}

type FinPL struct {
	Year          int
	Currency      string
	Months        []time.Time
	Revenue       FinPLRow
	OtherIncome   FinPLRow
	TotalIncome   FinPLRow
	Sections      []FinPLSection
	TotalExpenses FinPLRow
	NetProfit     FinPLRow
	Draws         FinPLRow
	Capex         FinPLRow
	Retained      FinPLRow
}

var finExpenseGroups = []string{"payroll", "bonus", "payroll_tax", "financial", "operating", "tax"}

func pick(currency string, uzs, usd float64) float64 {
	if currency == "USD" {
		return usd
	}
	return uzs
}

// PL builds the profit and loss for a calendar year. Revenue is work done by
// month. Salaries, bonuses, rent and services come from accruals (what was
// owed that month); other expenses are counted when paid.
func (d *FinData) PL(year int, currency string) FinPL {
	pl := FinPL{Year: year, Currency: currency}
	for m := 1; m <= 12; m++ {
		pl.Months = append(pl.Months, time.Date(year, time.Month(m), 1, 0, 0, 0, 0, time.UTC))
	}
	pl.Revenue = FinPLRow{Label: "Revenue (work done)", Strong: true}
	pl.OtherIncome = FinPLRow{Label: "Other income"}
	for _, rev := range d.Revenue {
		if rev.Date.Year() == year {
			pl.Revenue.add(int(rev.Date.Month())-1, pick(currency, rev.AmountUZS, rev.AmountUSD))
		}
	}
	categoryRows := map[string]*FinPLRow{}
	rowFor := func(category FinCategory) *FinPLRow {
		row, ok := categoryRows[category.Code]
		if !ok {
			row = &FinPLRow{Label: category.Name, Code: category.Code}
			categoryRows[category.Code] = row
		}
		return row
	}
	pl.Draws = FinPLRow{Label: "Founder draws (net)", Muted: true}
	pl.Capex = FinPLRow{Label: "Equipment bought (capex)", Muted: true}
	for _, txn := range d.Transactions {
		if txn.Date.Year() != year {
			continue
		}
		category := d.Category(txn.CategoryCode)
		month := int(txn.Date.Month()) - 1
		value := pick(currency, txn.AmountUZS, txn.AmountUSD)
		switch category.Kind {
		case KindOtherIncome:
			pl.OtherIncome.add(month, txn.Signed()*value)
		case KindExpense, KindTax:
			if category.AccrualBased || category.PLGroup == "" {
				continue
			}
			rowFor(category).add(month, -txn.Signed()*value)
		case KindOwnerDraw:
			pl.Draws.add(month, -txn.Signed()*value)
		case KindCapex:
			pl.Capex.add(month, -txn.Signed()*value)
		}
	}
	for _, accrual := range d.Accruals {
		if accrual.Date.Year() != year {
			continue
		}
		category := d.Category(accrual.CategoryCode)
		if category.PLGroup == "" {
			continue
		}
		rowFor(category).add(int(accrual.Date.Month())-1, pick(currency, accrual.AmountUZS, accrual.AmountUSD))
	}
	pl.TotalIncome = FinPLRow{Label: "Total income", Strong: true}
	for m := 0; m < 12; m++ {
		pl.TotalIncome.add(m, pl.Revenue.Values[m]+pl.OtherIncome.Values[m])
	}
	pl.TotalExpenses = FinPLRow{Label: "Total expenses", Strong: true}
	for _, group := range finExpenseGroups {
		section := FinPLSection{Label: FinPLGroupLabel(group), Group: group}
		section.Subtotal = FinPLRow{Label: FinPLGroupLabel(group), Strong: true}
		for _, category := range d.Categories {
			if category.PLGroup != group {
				continue
			}
			row, ok := categoryRows[category.Code]
			if !ok || row.IsZero() {
				continue
			}
			section.Rows = append(section.Rows, *row)
			for m := 0; m < 12; m++ {
				section.Subtotal.add(m, row.Values[m])
			}
		}
		if len(section.Rows) == 0 {
			continue
		}
		for m := 0; m < 12; m++ {
			pl.TotalExpenses.add(m, section.Subtotal.Values[m])
		}
		pl.Sections = append(pl.Sections, section)
	}
	pl.NetProfit = FinPLRow{Label: "Net profit", Strong: true}
	pl.Retained = FinPLRow{Label: "Profit left in company", Muted: true}
	for m := 0; m < 12; m++ {
		pl.NetProfit.add(m, pl.TotalIncome.Values[m]-pl.TotalExpenses.Values[m])
		pl.Retained.add(m, pl.NetProfit.Values[m]-pl.Draws.Values[m])
	}
	return pl
}

// ActiveMonths returns the months of the year that have any activity, so the
// P&L table can hide empty future months.
func (pl FinPL) ActiveMonths(today time.Time) []int {
	first, last := -1, -1
	rows := []FinPLRow{pl.TotalIncome, pl.TotalExpenses, pl.Draws, pl.Capex}
	for m := 0; m < 12; m++ {
		for _, row := range rows {
			if math.Abs(row.Values[m]) >= 0.005 {
				if first < 0 {
					first = m
				}
				last = m
			}
		}
	}
	if first < 0 {
		first = 0
	}
	if pl.Year == today.Year() && int(today.Month())-1 > last {
		last = int(today.Month()) - 1
	}
	if pl.Year < today.Year() {
		last = 11
	}
	var out []int
	for m := first; m <= last; m++ {
		out = append(out, m)
	}
	return out
}

// ---------------------------------------------------------------------------
// Monthly cash flow (UZS, from the ledger)

type FinCashFlowMonth struct {
	Month        time.Time
	Opening      float64
	OperatingIn  float64
	OperatingOut float64
	Investing    float64
	FinancingIn  float64
	FinancingOut float64
	Net          float64
	Closing      float64
}

func (m FinCashFlowMonth) OperatingNet() float64 { return m.OperatingIn - m.OperatingOut }

func (d *FinData) CashFlow(year int) []FinCashFlowMonth {
	months := make([]FinCashFlowMonth, 12)
	opening := 0.0
	for _, txn := range d.Transactions {
		if txn.Date.Year() < year && d.Category(txn.CategoryCode).Kind != KindTransfer {
			opening += txn.Signed() * txn.AmountUZS
		}
	}
	for i := range months {
		months[i].Month = time.Date(year, time.Month(i+1), 1, 0, 0, 0, 0, time.UTC)
	}
	for _, txn := range d.Transactions {
		if txn.Date.Year() != year {
			continue
		}
		m := &months[int(txn.Date.Month())-1]
		value := txn.Signed() * txn.AmountUZS
		switch d.Category(txn.CategoryCode).Kind {
		case KindRevenue, KindOtherIncome, KindExpense, KindTax:
			if value >= 0 {
				m.OperatingIn += value
			} else {
				m.OperatingOut -= value
			}
		case KindCapex:
			m.Investing -= value
		case KindOwnerCapital, KindLoanReceived, KindLoanReturned, KindOwnerDraw, KindLoanRepaid, KindLoanGiven:
			if value >= 0 {
				m.FinancingIn += value
			} else {
				m.FinancingOut -= value
			}
		case KindTransfer:
			// Moves between our own accounts are not cash flow.
		}
	}
	for i := range months {
		months[i].Opening = Round2(opening)
		months[i].Net = Round2(months[i].OperatingIn - months[i].OperatingOut - months[i].Investing + months[i].FinancingIn - months[i].FinancingOut)
		opening += months[i].Net
		months[i].Closing = Round2(opening)
	}
	return months
}

// ---------------------------------------------------------------------------
// Forecast

type FinForecastMonth struct {
	Month     time.Time
	Opening   float64
	Inflows   float64
	Recurring float64
	Payables  float64
	Other     float64
	Closing   float64
	Items     []FinForecastItem
}

func (m FinForecastMonth) Outflows() float64 { return m.Recurring + m.Payables + m.Other }

// MonthlyRecurringUZS is the planned fixed monthly spend at today's rate.
func (d *FinData) MonthlyRecurringUZS() float64 {
	rate, _ := RateOn(d.Rates, d.Today)
	total := 0.0
	for _, item := range d.Recurring {
		if !item.Active {
			continue
		}
		if item.Currency == "USD" {
			total += item.Amount * rate
		} else {
			total += item.Amount
		}
	}
	return Round2(total)
}

// Forecast projects cash for the next months. The current month pays what is
// owed right now (payables); following months pay the recurring costs.
func (d *FinData) Forecast(months int, startCash float64) []FinForecastMonth {
	rate, _ := RateOn(d.Rates, d.Today)
	recurring := d.MonthlyRecurringUZS()
	owed := PayablesOwedUZS(d.Payables())
	start := MonthStart(d.Today)
	out := make([]FinForecastMonth, 0, months)
	cash := startCash
	for i := 0; i < months; i++ {
		month := start.AddDate(0, i, 0)
		row := FinForecastMonth{Month: month, Opening: Round2(cash)}
		if i == 0 {
			row.Payables = owed
		} else {
			row.Recurring = recurring
		}
		for _, item := range d.ForecastItems {
			if !MonthStart(item.Month).Equal(month) {
				continue
			}
			value := item.Amount
			if item.Currency == "USD" {
				value *= rate
			}
			if item.Direction == "in" {
				row.Inflows += value
			} else {
				row.Other += value
			}
			row.Items = append(row.Items, item)
		}
		cash += row.Inflows - row.Outflows()
		row.Closing = Round2(cash)
		out = append(out, row)
	}
	return out
}

// ---------------------------------------------------------------------------
// Dashboard

type FinExpenseSlice struct {
	Label string
	Value float64
	Share float64
}

type FinAlert struct {
	Tone    string // warning, danger, info
	Message string
	Link    string
}

type FinDashboard struct {
	Cash               FinCashPosition
	BlockedUZS         float64
	FreeCashUZS        float64
	ReceivableUSD      float64
	AdvanceUSD         float64
	ExpectedUSD        float64
	PayablesUZS        float64
	FoundersOwedUZS    float64
	LendersOwedUZS     float64
	LoansToFoundersUZS float64
	MonthlyBurnUZS     float64
	RunwayMonths       float64
	HasRunway          bool
	PL                 FinPL
	YearProfitUSD      float64
	MonthProfitUSD     float64
	ExpenseMix         []FinExpenseSlice
	CashFlow           []FinCashFlowMonth
	Forecast           []FinForecastMonth
	Alerts             []FinAlert
	Contracts          FinContractTotals
	Founders           []FinPartyBalance
	Lenders            []FinPartyBalance
	Payables           []FinPayable
}

func (d *FinData) Dashboard() FinDashboard {
	dash := FinDashboard{Cash: d.CashPosition()}
	dash.Contracts = d.ContractTotals()
	dash.BlockedUZS = Round2(dash.Contracts.BlockedUZS)
	dash.FreeCashUZS = Round2(dash.Cash.TotalUZS - dash.BlockedUZS)
	dash.ReceivableUSD = Round2(dash.Contracts.Receivable)
	dash.AdvanceUSD = Round2(dash.Contracts.Advance)
	dash.ExpectedUSD = Round2(dash.Contracts.Unpaid)
	dash.Payables = d.Payables()
	dash.PayablesUZS = PayablesOwedUZS(dash.Payables)
	dash.Founders = d.FounderBalances()
	for _, founder := range dash.Founders {
		if founder.Opening != nil {
			dash.FoundersOwedUZS += founder.Balance
		}
		dash.LoansToFoundersUZS += founder.LoansOutstanding
	}
	dash.Lenders = d.LenderBalances()
	for _, lender := range dash.Lenders {
		if lender.Balance > 0 {
			dash.LendersOwedUZS += lender.Balance
		}
	}
	dash.MonthlyBurnUZS = d.MonthlyRecurringUZS()
	if dash.MonthlyBurnUZS > 0 {
		dash.HasRunway = true
		dash.RunwayMonths = math.Max(0, dash.FreeCashUZS) / dash.MonthlyBurnUZS
	}
	year := d.Today.Year()
	dash.PL = d.PL(year, "USD")
	dash.YearProfitUSD = Round2(dash.PL.NetProfit.Total)
	dash.MonthProfitUSD = Round2(dash.PL.NetProfit.Values[int(d.Today.Month())-1])
	totalExpenses := 0.0
	for _, section := range dash.PL.Sections {
		if section.Subtotal.Total > 0 {
			totalExpenses += section.Subtotal.Total
		}
	}
	for _, section := range dash.PL.Sections {
		if section.Subtotal.Total <= 0 {
			continue
		}
		slice := FinExpenseSlice{Label: section.Label, Value: section.Subtotal.Total}
		if totalExpenses > 0 {
			slice.Share = section.Subtotal.Total / totalExpenses
		}
		dash.ExpenseMix = append(dash.ExpenseMix, slice)
	}
	sort.Slice(dash.ExpenseMix, func(i, j int) bool { return dash.ExpenseMix[i].Value > dash.ExpenseMix[j].Value })
	dash.CashFlow = d.CashFlow(year)
	dash.Forecast = d.Forecast(4, dash.Cash.TotalUZS)
	dash.Alerts = d.alerts(dash)
	return dash
}

func (d *FinData) alerts(dash FinDashboard) []FinAlert {
	var alerts []FinAlert
	if dash.FreeCashUZS < 0 {
		alerts = append(alerts, FinAlert{
			Tone:    "danger",
			Message: "Free cash is " + FormatMoney(dash.FreeCashUZS, "UZS") + ". Part of the school prepayments is already spent before the steps were done.",
			Link:    "/admin/finance/contracts",
		})
	}
	if dash.HasRunway && dash.RunwayMonths < 2 && dash.FreeCashUZS >= 0 {
		alerts = append(alerts, FinAlert{Tone: "warning", Message: "Free cash covers less than two months of fixed costs.", Link: "/admin/finance/forecast"})
	}
	if len(dash.Contracts.Unallocated) > 0 {
		alerts = append(alerts, FinAlert{Tone: "info", Message: "A school payment is not linked to a contract.", Link: "/admin/finance/contracts#unallocated"})
	}
	rate, ok := RateOn(d.Rates, d.Today)
	if !ok || rate <= 0 {
		alerts = append(alerts, FinAlert{Tone: "danger", Message: "No exchange rate saved. Add today's CBU rate in Settings.", Link: "/admin/finance/settings"})
	}
	return alerts
}

// ---------------------------------------------------------------------------
// Ledger filtering

type FinTxnFilter struct {
	Month          string
	Account        string
	Category       string
	CounterpartyID string
	Query          string
}

func (f FinTxnFilter) Active() bool {
	return f.Month != "" || f.Account != "" || f.Category != "" || f.CounterpartyID != "" || f.Query != ""
}

type FinLedgerRow struct {
	Txn      FinTransaction
	Category FinCategory
	In       float64
	Out      float64
}

type FinLedger struct {
	Rows   []FinLedgerRow
	InUZS  float64
	OutUZS float64
	Total  int
	Months []string
}

func (d *FinData) Ledger(filter FinTxnFilter) FinLedger {
	ledger := FinLedger{Total: len(d.Transactions)}
	seenMonths := map[string]bool{}
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	for _, txn := range d.Transactions {
		seenMonths[MonthKey(txn.Date)] = true
		if filter.Month != "" && MonthKey(txn.Date) != filter.Month {
			continue
		}
		if filter.Account != "" && txn.Account != filter.Account {
			continue
		}
		if filter.Category != "" && txn.CategoryCode != filter.Category {
			continue
		}
		if filter.CounterpartyID != "" && txn.CounterpartyID != filter.CounterpartyID {
			continue
		}
		category := d.Category(txn.CategoryCode)
		if query != "" {
			haystack := strings.ToLower(txn.Description + " " + txn.CounterpartyName + " " + category.Name)
			if !strings.Contains(haystack, query) {
				continue
			}
		}
		row := FinLedgerRow{Txn: txn, Category: category}
		if txn.Direction == "in" {
			row.In = txn.Amount
			ledger.InUZS += txn.AmountUZS
		} else {
			row.Out = txn.Amount
			ledger.OutUZS += txn.AmountUZS
		}
		ledger.Rows = append(ledger.Rows, row)
	}
	sort.SliceStable(ledger.Rows, func(i, j int) bool {
		a, b := ledger.Rows[i].Txn, ledger.Rows[j].Txn
		if !a.Date.Equal(b.Date) {
			return a.Date.After(b.Date)
		}
		return a.CreatedAt.After(b.CreatedAt)
	})
	for month := range seenMonths {
		ledger.Months = append(ledger.Months, month)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ledger.Months)))
	ledger.InUZS = Round2(ledger.InUZS)
	ledger.OutUZS = Round2(ledger.OutUZS)
	return ledger
}

// Years returns every year with ledger or revenue activity, newest first.
func (d *FinData) Years() []int {
	seen := map[int]bool{d.Today.Year(): true}
	for _, txn := range d.Transactions {
		seen[txn.Date.Year()] = true
	}
	for _, rev := range d.Revenue {
		seen[rev.Date.Year()] = true
	}
	var years []int
	for year := range seen {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))
	return years
}
