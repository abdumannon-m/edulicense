package app

import (
	"html/template"
	"math"
	"time"
)

// FinPage is the view model shared by the finance templates.
type FinPage struct {
	Tab      string
	Today    time.Time
	Rate     float64
	RateDate time.Time
	Data     *FinData

	Dashboard FinDashboard
	Cash      FinCashPosition

	Ledger      FinLedger
	Filter      FinTxnFilter
	Txn         FinTransaction
	TxnTransfer []FinTransaction

	Contracts FinContractTotals
	Contract  FinContractSummary

	Payables       []FinPayable
	Founders       []FinPartyBalance
	Lenders        []FinPartyBalance
	RecentAccruals []FinAccrual

	PL       FinPL
	PLMonths []int
	Year     int
	Currency string
	Years    []int
	CashFlow []FinCashFlowMonth

	Forecast      []FinForecastMonth
	MonthlyFixed  float64
	ForecastEnd   FinForecastMonth
	UpcomingItems []FinForecastItem

	Charts map[string]template.HTML

	CategoryGroups  []FinCategoryGroup
	ActiveContracts []FinContract
	RecentRates     []FinRate
}

type FinCategoryGroup struct {
	Label      string
	Kind       string
	Categories []FinCategory
}

// CategoryGroups groups active categories by kind for <optgroup> selects.
func (d *FinData) CategoryGroups(includeTransfer bool) []FinCategoryGroup {
	var groups []FinCategoryGroup
	for _, kind := range FinCategoryKinds {
		if kind == KindTransfer && !includeTransfer {
			continue
		}
		group := FinCategoryGroup{Label: FinKindLabel(kind), Kind: kind}
		for _, category := range d.Categories {
			if category.Kind == kind && category.Active {
				group.Categories = append(group.Categories, category)
			}
		}
		if len(group.Categories) > 0 {
			groups = append(groups, group)
		}
	}
	return groups
}

// ActiveContracts returns contracts that can still receive payments.
func (d *FinData) ActiveContracts() []FinContract {
	var out []FinContract
	for _, contract := range d.Contracts {
		if contract.Status != "cancelled" {
			out = append(out, contract)
		}
	}
	return out
}

// CounterpartyKindForCategory guesses who a new name is from the category.
func CounterpartyKindForCategory(category FinCategory) string {
	switch category.Kind {
	case KindRevenue:
		return "school"
	case KindOwnerCapital, KindOwnerDraw:
		return "founder"
	case KindLoanReceived, KindLoanRepaid:
		return "lender"
	case KindTax:
		return "government"
	}
	switch category.PLGroup {
	case "payroll", "bonus":
		return "employee"
	case "payroll_tax":
		return "government"
	case "financial":
		return "bank"
	case "operating":
		return "supplier"
	}
	return "other"
}

// Transfers returns both legs of a transfer.
func (d *FinData) TransferLegs(group string) []FinTransaction {
	var legs []FinTransaction
	if group == "" {
		return legs
	}
	for _, txn := range d.Transactions {
		if txn.TransferGroup == group {
			legs = append(legs, txn)
		}
	}
	return legs
}

func (d *FinData) Transaction(id string) (FinTransaction, bool) {
	for _, txn := range d.Transactions {
		if txn.ID == id {
			return txn, true
		}
	}
	return FinTransaction{}, false
}

// ---------------------------------------------------------------------------
// Template helpers

var shortMonths = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

func MonthLabel(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return shortMonths[t.Month()-1] + " " + t.Format("2006")
}

func ShortMonth(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return shortMonths[t.Month()-1]
}

func ShortDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2") + " " + shortMonths[t.Month()-1] + " " + t.Format("2006")
}

// MoneyShort renders a compact amount such as "92.9M so'm" or "$7.9k".
func MoneyShort(value float64, currency string) string {
	if currency == "USD" {
		if value < 0 {
			return "−$" + compactNumber(-value)
		}
		return "$" + compactNumber(value)
	}
	return compactNumber(value) + " so'm"
}

func Percent(value float64) string {
	return GroupDigits(math.Round(value*100), 0) + "%"
}

func FinTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"money":        FormatMoney,
		"moneyShort":   MoneyShort,
		"uzs":          func(v float64) string { return FormatMoney(v, "UZS") },
		"usd":          func(v float64) string { return FormatMoney(v, "USD") },
		"num":          FormatNumber,
		"percent":      Percent,
		"pctWidth":     func(v float64) string { return GroupDigits(math.Max(0, math.Min(100, v)), 1) },
		"accountLabel": FinAccountLabel,
		"kindLabel":    FinKindLabel,
		"plGroupLabel": FinPLGroupLabel,
		"cpKindLabel":  FinCounterpartyKindLabel,
		"monthLabel":   MonthLabel,
		"shortMonth":   ShortMonth,
		"shortDate":    ShortDate,
		"monthKey":     MonthKey,
		"isNeg":        func(v float64) bool { return v < -0.005 },
		"isPos":        func(v float64) bool { return v > 0.005 },
		"isZero":       func(v float64) bool { return math.Abs(v) < 0.005 },
		"abs":          math.Abs,
		"sub":          func(a, b float64) float64 { return a - b },
		"mul":          func(a, b float64) float64 { return a * b },
		"ptrDate": func(t *time.Time) string {
			if t == nil {
				return ""
			}
			return t.Format("2006-01-02")
		},
		"ptrFloat": func(v *float64) string {
			if v == nil {
				return ""
			}
			return FormatNumber(*v)
		},
		"plain": FormatNumber,
		"milestoneLabel": func(n int) string {
			if n >= 1 && n <= len(FinMilestones) {
				return FinMilestones[n-1].Label
			}
			return ""
		},
		"monthValue": func(values [12]float64, month int) float64 {
			if month < 0 || month > 11 {
				return 0
			}
			return values[month]
		},
		"monthOf": func(pl FinPL, month int) time.Time { return pl.Months[month] },
		"weightAmount": func(contract FinContract, step int) float64 {
			if step < 1 || step > 4 {
				return 0
			}
			return Round2(contract.Amount * contract.Weights[step-1] / 100)
		},
		"settles": func(t FinTransaction) float64 {
			if t.SettlesAmount == nil {
				return 0
			}
			return *t.SettlesAmount
		},
		"monthKeyLabel": func(key string) string {
			if t, err := ParseMonth(key); err == nil {
				return MonthLabel(t)
			}
			return key
		},
		"addi":          func(a, b int) int { return a + b },
		"finAccounts":   func() []FinAccount { return FinAccounts },
		"finKinds":      func() []string { return FinCategoryKinds },
		"finPLGroups":   func() []string { return FinPLGroups },
		"finPartyKinds": func() []string { return FinCounterpartyKinds },
		"finMilestones": func() []FinMilestone { return FinMilestones },
		"seq": func(n int) []int {
			out := make([]int, n)
			for i := range out {
				out[i] = i
			}
			return out
		},
	}
}
