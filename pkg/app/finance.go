package app

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Accounts where money is held. Each account has a single currency.
var FinAccounts = []FinAccount{
	{Code: "cash_uzs", Label: "Naqd (so'm)", Currency: "UZS"},
	{Code: "card_uzs", Label: "Karta (so'm)", Currency: "UZS"},
	{Code: "bank_uzs", Label: "Bank (so'm)", Currency: "UZS"},
	{Code: "cash_usd", Label: "Naqd ($)", Currency: "USD"},
	{Code: "bank_usd", Label: "Bank ($)", Currency: "USD"},
}

type FinAccount struct {
	Code     string
	Label    string
	Currency string
}

func FinAccountByCode(code string) (FinAccount, bool) {
	for _, account := range FinAccounts {
		if account.Code == code {
			return account, true
		}
	}
	return FinAccount{}, false
}

func FinAccountLabel(code string) string {
	if account, ok := FinAccountByCode(code); ok {
		return account.Label
	}
	return code
}

// Category kinds decide where a transaction lands in the reports.
const (
	KindRevenue      = "revenue"
	KindOtherIncome  = "other_income"
	KindExpense      = "expense"
	KindTax          = "tax"
	KindCapex        = "capex"
	KindOwnerCapital = "owner_capital"
	KindOwnerDraw    = "owner_draw"
	KindLoanReceived = "loan_received"
	KindLoanRepaid   = "loan_repaid"
	KindLoanGiven    = "loan_given"
	KindLoanReturned = "loan_returned"
	KindTransfer     = "transfer"
)

var FinCategoryKinds = []string{
	KindRevenue, KindOtherIncome, KindExpense, KindTax, KindCapex,
	KindOwnerCapital, KindOwnerDraw,
	KindLoanReceived, KindLoanRepaid, KindLoanGiven, KindLoanReturned,
	KindTransfer,
}

var FinPLGroups = []string{"", "revenue", "other_income", "payroll", "bonus", "payroll_tax", "financial", "operating", "tax"}

var FinCounterpartyKinds = []string{"school", "employee", "founder", "supplier", "lender", "government", "bank", "other"}

func FinKindLabel(kind string) string {
	switch kind {
	case KindRevenue:
		return "Revenue"
	case KindOtherIncome:
		return "Other income"
	case KindExpense:
		return "Expense"
	case KindTax:
		return "Tax"
	case KindCapex:
		return "Equipment (capex)"
	case KindOwnerCapital:
		return "Founder capital in"
	case KindOwnerDraw:
		return "Founder draw"
	case KindLoanReceived:
		return "Loan received"
	case KindLoanRepaid:
		return "Loan repaid"
	case KindLoanGiven:
		return "Loan given"
	case KindLoanReturned:
		return "Loan returned to us"
	case KindTransfer:
		return "Transfer"
	}
	return kind
}

func FinPLGroupLabel(group string) string {
	switch group {
	case "revenue":
		return "Revenue"
	case "other_income":
		return "Other income"
	case "payroll":
		return "Salaries"
	case "bonus":
		return "Bonuses"
	case "payroll_tax":
		return "Payroll taxes"
	case "financial":
		return "Bank fees & penalties"
	case "operating":
		return "Operating expenses"
	case "tax":
		return "Taxes"
	case "":
		return "Not in P&L"
	}
	return group
}

func FinCounterpartyKindLabel(kind string) string {
	switch kind {
	case "school":
		return "School"
	case "employee":
		return "Employee"
	case "founder":
		return "Founder"
	case "supplier":
		return "Supplier"
	case "lender":
		return "Lender"
	case "government":
		return "Government"
	case "bank":
		return "Bank"
	}
	return "Other"
}

// Milestones of the SAT test center process. Money prepaid by a school is
// unlocked step by step as each milestone is reached.
var FinMilestones = []FinMilestone{
	{Number: 1, Label: "Documents submitted", Stage: "applied_to_ceeb"},
	{Number: 2, Label: "CEEB code received", Stage: "got_ceeb"},
	{Number: 3, Label: "Applied to test center", Stage: "applied_to_test_center"},
	{Number: 4, Label: "Test center code received", Stage: "got_test_center"},
}

type FinMilestone struct {
	Number int
	Label  string
	Stage  string
}

type FinCategory struct {
	Code         string
	Name         string
	Kind         string
	PLGroup      string
	AccrualBased bool
	SortOrder    int
	Active       bool
}

// IsIncomeLike reports whether money normally comes in for this category.
func (c FinCategory) IsIncomeLike() bool {
	switch c.Kind {
	case KindRevenue, KindOtherIncome, KindOwnerCapital, KindLoanReceived, KindLoanReturned:
		return true
	}
	return false
}

type FinCounterparty struct {
	ID              string
	Name            string
	Kind            string
	ApplicationID   string
	ApplicationName string
	Active          bool
	Notes           string
}

type FinRate struct {
	Date   time.Time
	Rate   float64
	Source string
}

type FinTransaction struct {
	ID               string
	Date             time.Time
	Account          string
	Direction        string
	Amount           float64
	Currency         string
	FxRate           float64
	AmountUZS        float64
	AmountUSD        float64
	CategoryCode     string
	CounterpartyID   string
	CounterpartyName string
	ContractID       string
	SettlesAmount    *float64
	Description      string
	TransferGroup    string
	CreatedAt        time.Time
	HasAccrual       bool
}

// Signed returns +1 for money in and -1 for money out.
func (t FinTransaction) Signed() float64 {
	if t.Direction == "in" {
		return 1
	}
	return -1
}

type FinAccrual struct {
	ID               string
	Date             time.Time
	CategoryCode     string
	CounterpartyID   string
	CounterpartyName string
	Amount           float64
	Currency         string
	FxRate           float64
	AmountUZS        float64
	AmountUSD        float64
	Note             string
	TransactionID    string
}

type FinRevenue struct {
	ID         string
	ContractID string
	Date       time.Time
	Amount     float64
	FxRate     float64
	AmountUZS  float64
	AmountUSD  float64
	Milestone  int
	Note       string
}

type FinContract struct {
	ID               string
	CounterpartyID   string
	CounterpartyName string
	ApplicationID    string
	ApplicationName  string
	ApplicationStage string
	Title            string
	Amount           float64
	Currency         string
	SignedOn         time.Time
	Status           string
	Weights          [4]float64
	MilestoneOn      [4]*time.Time
	Notes            string
}

// MilestoneDone reports whether step n (1-4) is reached, either recorded on
// the contract or implied by the linked application's stage.
func (c FinContract) MilestoneDone(n int) bool {
	if n < 1 || n > 4 {
		return false
	}
	if c.Status == "completed" {
		return true
	}
	if c.MilestoneOn[n-1] != nil {
		return true
	}
	return ApplicationStageReached(c.ApplicationStage, FinMilestones[n-1].Stage)
}

// UnlockedShare is the fraction (0-1) of money received that may be spent.
func (c FinContract) UnlockedShare() float64 {
	if c.Status == "completed" || c.Status == "cancelled" {
		return 1
	}
	share := 0.0
	for i := 1; i <= 4; i++ {
		if c.MilestoneDone(i) {
			share += c.Weights[i-1]
		}
	}
	return math.Min(1, share/100)
}

func (c FinContract) Label() string {
	if strings.TrimSpace(c.Title) != "" {
		return c.CounterpartyName + " · " + c.Title
	}
	return c.CounterpartyName
}

// ApplicationStageReached reports whether current is at or past target in the
// application pipeline.
func ApplicationStageReached(current, target string) bool {
	if current == "" || target == "" {
		return false
	}
	ci := slices.Index(ApplicationStages, current)
	ti := slices.Index(ApplicationStages, target)
	return ci >= 0 && ti >= 0 && ci >= ti
}

type FinOpening struct {
	ID               string
	CounterpartyID   string
	CounterpartyName string
	CounterpartyKind string
	AsOf             time.Time
	AmountUZS        float64
	Note             string
}

type FinRecurring struct {
	ID               string
	Label            string
	CounterpartyID   string
	CounterpartyName string
	Amount           float64
	Currency         string
	Active           bool
}

type FinForecastItem struct {
	ID        string
	Month     time.Time
	Direction string
	Label     string
	Amount    float64
	Currency  string
}

// FinData is everything the finance pages compute from. It is loaded in one
// round trip and all reports are pure functions of it.
type FinData struct {
	Today          time.Time
	Rates          []FinRate
	Categories     []FinCategory
	Counterparties []FinCounterparty
	Contracts      []FinContract
	Transactions   []FinTransaction
	Accruals       []FinAccrual
	Revenue        []FinRevenue
	Openings       []FinOpening
	Recurring      []FinRecurring
	ForecastItems  []FinForecastItem
	Applications   []FinApplicationRef

	categoryIndex map[string]FinCategory
}

type FinApplicationRef struct {
	ID    string
	Name  string
	Stage string
}

func (d *FinData) Category(code string) FinCategory {
	if d.categoryIndex == nil {
		d.categoryIndex = make(map[string]FinCategory, len(d.Categories))
		for _, category := range d.Categories {
			d.categoryIndex[category.Code] = category
		}
	}
	if category, ok := d.categoryIndex[code]; ok {
		return category
	}
	return FinCategory{Code: code, Name: code, Kind: KindExpense}
}

func (d *FinData) Contract(id string) (FinContract, bool) {
	for _, contract := range d.Contracts {
		if contract.ID == id {
			return contract, true
		}
	}
	return FinContract{}, false
}

func (d *FinData) Counterparty(id string) (FinCounterparty, bool) {
	for _, counterparty := range d.Counterparties {
		if counterparty.ID == id {
			return counterparty, true
		}
	}
	return FinCounterparty{}, false
}

// RateOn returns the UZS-per-USD rate for the date: the latest rate on or
// before it, else the earliest known rate. ok is false when no rates exist.
func RateOn(rates []FinRate, date time.Time) (float64, bool) {
	if len(rates) == 0 {
		return 0, false
	}
	day := DateOnly(date)
	best := -1
	for i, rate := range rates {
		if !rate.Date.After(day) && (best < 0 || rate.Date.After(rates[best].Date)) {
			best = i
		}
	}
	if best >= 0 {
		return rates[best].Rate, true
	}
	earliest := 0
	for i, rate := range rates {
		if rate.Date.Before(rates[earliest].Date) {
			earliest = i
		}
	}
	return rates[earliest].Rate, true
}

// HasExactRate reports whether a rate is stored for that exact day.
func HasExactRate(rates []FinRate, date time.Time) bool {
	day := DateOnly(date)
	for _, rate := range rates {
		if rate.Date.Equal(day) {
			return true
		}
	}
	return false
}

// Convert returns the UZS and USD value of an amount at the given rate.
func Convert(amount float64, currency string, rate float64) (uzs, usd float64) {
	if currency == "USD" {
		return Round2(amount * rate), Round2(amount)
	}
	return Round2(amount), Round2(amount / rate)
}

func Round2(value float64) float64 {
	return math.Round(value*100) / 100
}

func DateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func MonthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func MonthKey(t time.Time) string {
	return t.Format("2006-01")
}

// ParseMonth parses "2006-01" into the first day of that month.
func ParseMonth(value string) (time.Time, error) {
	parsed, err := time.Parse("2006-01", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month %q, use YYYY-MM", value)
	}
	return parsed, nil
}

// ParseAmount accepts "1 234 567,89", "1,234,567.89", "1.234.567" and
// "1234567.89". Anything else, including trailing text, is rejected rather
// than read partially.
func ParseAmount(raw string) (float64, error) {
	invalid := fmt.Errorf("invalid amount %q", raw)
	value := strings.TrimSpace(raw)
	for _, strip := range []string{" ", "\u00a0", "\u202f", "$"} {
		value = strings.ReplaceAll(value, strip, "")
	}
	if value == "" {
		return 0, errors.New("amount is required")
	}
	sign := ""
	if strings.HasPrefix(value, "-") {
		sign, value = "-", value[1:]
	}
	lastComma := strings.LastIndex(value, ",")
	lastDot := strings.LastIndex(value, ".")
	whole, frac := value, ""
	switch {
	case lastComma >= 0 && lastDot >= 0:
		// The later separator is the decimal point, the other groups thousands.
		decimal, thousands := ".", ","
		if lastComma > lastDot {
			decimal, thousands = ",", "."
		}
		split := strings.LastIndex(value, decimal)
		whole, frac = value[:split], value[split+1:]
		if strings.Contains(whole, decimal) || !thousandGroups(whole, thousands) {
			return 0, invalid
		}
		whole = strings.ReplaceAll(whole, thousands, "")
	case lastComma >= 0 || lastDot >= 0:
		sep := ","
		if lastDot >= 0 {
			sep = "."
		}
		split := strings.LastIndex(value, sep)
		// One dot is always a decimal point. One comma is a decimal comma
		// unless exactly three digits follow it ("1,234"). Repeated
		// separators group thousands ("1.234.567").
		single := strings.Count(value, sep) == 1
		if single && (sep == "." || len(value)-split-1 != 3) {
			whole, frac = value[:split], value[split+1:]
		} else {
			if !thousandGroups(value, sep) {
				return 0, invalid
			}
			whole = strings.ReplaceAll(value, sep, "")
		}
	}
	if whole == "" {
		whole = "0"
	}
	if !allDigits(whole) || (frac != "" && !allDigits(frac)) {
		return 0, invalid
	}
	number := sign + whole
	if frac != "" {
		number += "." + frac
	}
	amount, err := strconv.ParseFloat(number, 64)
	if err != nil || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, invalid
	}
	return amount, nil
}

// thousandGroups reports whether value is digits split by sep into a leading
// group of 1-3 digits followed by groups of exactly three.
func thousandGroups(value, sep string) bool {
	groups := strings.Split(value, sep)
	if len(groups[0]) < 1 || len(groups[0]) > 3 {
		return len(groups) == 1
	}
	for _, group := range groups[1:] {
		if len(group) != 3 {
			return false
		}
	}
	return true
}

func allDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// FormatMoney renders an amount the way the team writes it: "92 948 931 so'm"
// or "$7 860".
func FormatMoney(value float64, currency string) string {
	decimals := 0
	if currency == "USD" && math.Abs(value) < 1000 && math.Abs(value-math.Round(value)) > 0.004 {
		decimals = 2
	}
	text := GroupDigits(value, decimals)
	if currency == "USD" {
		if strings.HasPrefix(text, "-") {
			return "-$" + text[1:]
		}
		return "$" + text
	}
	return text + " so'm"
}

// GroupDigits formats a number with spaces between thousands.
func GroupDigits(value float64, decimals int) string {
	negative := value < 0
	value = math.Abs(value)
	pow := math.Pow(10, float64(decimals))
	value = math.Round(value*pow) / pow
	whole := math.Floor(value)
	frac := value - whole
	digits := fmt.Sprintf("%.0f", whole)
	var builder strings.Builder
	for i, r := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			builder.WriteByte(' ')
		}
		builder.WriteRune(r)
	}
	out := builder.String()
	if decimals > 0 {
		out += fmt.Sprintf("%.*f", decimals, frac)[1:]
	}
	if negative && (whole > 0 || frac > 0) {
		out = "-" + out
	}
	return out
}

// FormatNumber renders a plain amount without currency, for inputs.
func FormatNumber(value float64) string {
	if value == math.Trunc(value) {
		return fmt.Sprintf("%.0f", value)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}

func sortStrings(values []string) []string {
	sort.Strings(values)
	return values
}
