package app

import (
	"errors"
	"math"
	"strings"
	"time"
)

// Inputs for the finance forms. Handlers parse forms into these, the store
// validates with Validate() and writes them.

type FinTxnInput struct {
	Date           time.Time
	Account        string
	Direction      string // in / out; ignored for transfers
	Amount         float64
	CategoryCode   string
	CounterpartyID string
	ContractID     string
	SettlesAmount  *float64
	Description    string
	CreateAccrual  bool
	FxRate         float64 // optional override; 0 = rate saved for the date

	// Transfers between our own accounts.
	ToAccount string
	ToAmount  float64 // required only when the two accounts differ in currency
}

func (in FinTxnInput) IsTransfer() bool { return in.ToAccount != "" }

// AccrualApplies reports whether the "salary or bill for this month" box means
// anything: only for money paid out on a category tracked as owed per month.
func (in FinTxnInput) AccrualApplies(category FinCategory) bool {
	return !in.IsTransfer() && in.Direction == "out" && category.AccrualBased
}

func (in *FinTxnInput) Validate() error {
	in.Description = strings.TrimSpace(in.Description)
	if in.Date.IsZero() {
		return errors.New("Date is required.")
	}
	from, ok := FinAccountByCode(in.Account)
	if !ok {
		return errors.New("Choose an account.")
	}
	if in.Amount <= 0 || math.IsNaN(in.Amount) || math.IsInf(in.Amount, 0) {
		return errors.New("Amount must be greater than zero.")
	}
	if in.FxRate < 0 {
		return errors.New("Exchange rate cannot be negative.")
	}
	if in.IsTransfer() {
		to, ok := FinAccountByCode(in.ToAccount)
		if !ok || in.ToAccount == in.Account {
			return errors.New("Choose two different accounts for a transfer.")
		}
		if from.Currency == to.Currency {
			in.ToAmount = in.Amount
		} else if in.ToAmount <= 0 {
			return errors.New("Enter the amount received in " + to.Currency + ".")
		}
		in.CategoryCode = "otkazma"
		in.CounterpartyID, in.ContractID, in.SettlesAmount, in.CreateAccrual = "", "", nil, false
		return nil
	}
	if in.Direction != "in" && in.Direction != "out" {
		return errors.New("Choose money in or money out.")
	}
	if strings.TrimSpace(in.CategoryCode) == "" {
		return errors.New("Choose a category.")
	}
	if in.SettlesAmount != nil && *in.SettlesAmount <= 0 {
		in.SettlesAmount = nil
	}
	if in.ContractID == "" {
		in.SettlesAmount = nil
	}
	if in.CreateAccrual && in.CounterpartyID == "" {
		return errors.New("Choose who was paid so the salary or bill can be tracked.")
	}
	return nil
}

type FinContractInput struct {
	CounterpartyID string
	ApplicationID  string
	Title          string
	Amount         float64
	Currency       string
	SignedOn       time.Time
	Status         string
	Weights        [4]float64
	Notes          string
}

func (in *FinContractInput) Validate() error {
	in.Title = strings.TrimSpace(in.Title)
	in.Notes = strings.TrimSpace(in.Notes)
	if in.CounterpartyID == "" {
		return errors.New("Choose the school.")
	}
	if in.Amount <= 0 {
		return errors.New("Contract amount must be greater than zero.")
	}
	if in.Currency != "USD" && in.Currency != "UZS" {
		return errors.New("Currency must be USD or UZS.")
	}
	if in.SignedOn.IsZero() {
		return errors.New("Signed date is required.")
	}
	switch in.Status {
	case "active", "completed", "cancelled":
	case "":
		in.Status = "active"
	default:
		return errors.New("Unknown contract status.")
	}
	sum := 0.0
	for _, weight := range in.Weights {
		if weight < 0 {
			return errors.New("Step shares cannot be negative.")
		}
		sum += weight
	}
	if math.Abs(sum-100) > 0.001 {
		return errors.New("Step shares must add up to 100%.")
	}
	return nil
}

func DefaultFinWeights() [4]float64 { return [4]float64{15, 25, 20, 40} }

type FinRevenueInput struct {
	ContractID string
	Date       time.Time
	Amount     float64
	Milestone  int
	Note       string
	FxRate     float64
}

func (in *FinRevenueInput) Validate() error {
	in.Note = strings.TrimSpace(in.Note)
	if in.ContractID == "" {
		return errors.New("Contract is missing.")
	}
	if in.Date.IsZero() {
		return errors.New("Choose the month the work was done.")
	}
	if in.Amount <= 0 {
		return errors.New("Work done amount must be greater than zero.")
	}
	if in.Milestone < 0 || in.Milestone > 4 {
		return errors.New("Unknown step.")
	}
	return nil
}

type FinAccrualInput struct {
	Date           time.Time
	CategoryCode   string
	CounterpartyID string
	Amount         float64
	Currency       string
	Note           string
	FxRate         float64
}

func (in *FinAccrualInput) Validate() error {
	in.Note = strings.TrimSpace(in.Note)
	if in.Date.IsZero() {
		return errors.New("Choose the month.")
	}
	if in.CategoryCode == "" || in.CounterpartyID == "" {
		return errors.New("Choose the category and who is owed.")
	}
	if in.Amount <= 0 {
		return errors.New("Amount must be greater than zero.")
	}
	if in.Currency != "USD" && in.Currency != "UZS" {
		return errors.New("Currency must be USD or UZS.")
	}
	return nil
}

type FinCounterpartyInput struct {
	Name          string
	Kind          string
	ApplicationID string
	Active        bool
	Notes         string
}

func (in *FinCounterpartyInput) Validate() error {
	in.Name = strings.Join(strings.Fields(in.Name), " ")
	in.Notes = strings.TrimSpace(in.Notes)
	if in.Name == "" {
		return errors.New("Name is required.")
	}
	for _, kind := range FinCounterpartyKinds {
		if kind == in.Kind {
			return nil
		}
	}
	return errors.New("Unknown type.")
}

type FinCategoryInput struct {
	Code         string
	Name         string
	Kind         string
	PLGroup      string
	AccrualBased bool
	SortOrder    int
	Active       bool
}

func (in *FinCategoryInput) Validate() error {
	in.Code = strings.ToLower(strings.TrimSpace(in.Code))
	in.Name = strings.TrimSpace(in.Name)
	if in.Code == "" || in.Name == "" {
		return errors.New("Code and name are required.")
	}
	for _, r := range in.Code {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_') {
			return errors.New("Code may use only a-z, 0-9 and _.")
		}
	}
	kindOK := false
	for _, kind := range FinCategoryKinds {
		if kind == in.Kind {
			kindOK = true
		}
	}
	if !kindOK {
		return errors.New("Unknown category type.")
	}
	groupOK := false
	for _, group := range FinPLGroups {
		if group == in.PLGroup {
			groupOK = true
		}
	}
	if !groupOK {
		return errors.New("Unknown P&L line.")
	}
	return nil
}

type FinRecurringInput struct {
	Label          string
	CounterpartyID string
	Amount         float64
	Currency       string
	Active         bool
}

func (in *FinRecurringInput) Validate() error {
	in.Label = strings.TrimSpace(in.Label)
	if in.Label == "" {
		return errors.New("Name is required.")
	}
	if in.Amount <= 0 {
		return errors.New("Amount must be greater than zero.")
	}
	if in.Currency != "USD" && in.Currency != "UZS" {
		return errors.New("Currency must be USD or UZS.")
	}
	return nil
}

type FinForecastInput struct {
	Month     time.Time
	Direction string
	Label     string
	Amount    float64
	Currency  string
}

func (in *FinForecastInput) Validate() error {
	in.Label = strings.TrimSpace(in.Label)
	if in.Month.IsZero() {
		return errors.New("Choose the month.")
	}
	in.Month = MonthStart(in.Month)
	if in.Direction != "in" && in.Direction != "out" {
		return errors.New("Choose money in or out.")
	}
	if in.Label == "" {
		return errors.New("Describe the item.")
	}
	if in.Amount <= 0 {
		return errors.New("Amount must be greater than zero.")
	}
	if in.Currency != "USD" && in.Currency != "UZS" {
		return errors.New("Currency must be USD or UZS.")
	}
	return nil
}

type FinOpeningInput struct {
	CounterpartyID string
	AsOf           time.Time
	AmountUZS      float64
	Note           string
}

func (in *FinOpeningInput) Validate() error {
	in.Note = strings.TrimSpace(in.Note)
	if in.CounterpartyID == "" {
		return errors.New("Choose the person.")
	}
	if in.AsOf.IsZero() {
		return errors.New("Choose the date.")
	}
	return nil
}
