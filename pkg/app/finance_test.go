package app

import (
	"math"
	"strings"
	"testing"
	"time"
)

func day(value string) time.Time {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		panic(err)
	}
	return t
}

func near(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Fatalf("%s = %.2f, want %.2f", name, got, want)
	}
}

func testCategories() []FinCategory {
	return []FinCategory{
		{Code: "daromad", Name: "Daromad", Kind: KindRevenue, PLGroup: "revenue"},
		{Code: "ish_haqi_sotuv", Name: "Ish haqi sotuv", Kind: KindExpense, PLGroup: "payroll", AccrualBased: true},
		{Code: "komunal", Name: "Kommunal", Kind: KindExpense, PLGroup: "operating"},
		{Code: "aylanma_soliq", Name: "Aylanma soliq", Kind: KindTax, PLGroup: "tax"},
		{Code: "divident", Name: "Divident", Kind: KindOwnerDraw},
		{Code: "ustav", Name: "Ustav", Kind: KindOwnerCapital},
		{Code: "otkazma", Name: "O'tkazma", Kind: KindTransfer},
		{Code: "qarz_berildi", Name: "Qarz berildi", Kind: KindLoanGiven},
	}
}

func txn(date, account, dir string, amount float64, currency string, rate float64, category, counterparty, contract string) FinTransaction {
	uzs, usd := Convert(amount, currency, rate)
	return FinTransaction{
		ID: date + category + dir, Date: day(date), Account: account, Direction: dir, Amount: amount,
		Currency: currency, FxRate: rate, AmountUZS: uzs, AmountUSD: usd,
		CategoryCode: category, CounterpartyID: counterparty, ContractID: contract,
	}
}

func sampleData() *FinData {
	settle := 6000.0
	data := &FinData{
		Today:      day("2026-09-27"),
		Rates:      []FinRate{{Date: day("2026-04-01"), Rate: 12000}, {Date: day("2026-09-01"), Rate: 12000}},
		Categories: testCategories(),
		Counterparties: []FinCounterparty{
			{ID: "school", Name: "English Life", Kind: "school"},
			{ID: "emp", Name: "Sevinchxon", Kind: "employee"},
			{ID: "founder", Name: "Maqsadbek", Kind: "founder"},
		},
		Contracts: []FinContract{{
			ID: "c1", CounterpartyID: "school", CounterpartyName: "English Life", Amount: 15000, Currency: "USD",
			SignedOn: day("2026-09-01"), Status: "active", Weights: [4]float64{15, 25, 20, 40},
		}},
	}
	m1 := day("2026-09-26")
	data.Contracts[0].MilestoneOn[0] = &m1
	payment := txn("2026-09-26", "bank_uzs", "in", 72000000, "UZS", 12000, "daromad", "school", "c1")
	payment.SettlesAmount = &settle
	salary := txn("2026-09-10", "cash_uzs", "out", 3000000, "UZS", 12000, "ish_haqi_sotuv", "emp", "")
	data.Transactions = []FinTransaction{
		payment,
		salary,
		txn("2026-09-05", "cash_usd", "in", 1000, "USD", 12000, "ustav", "founder", ""),
		txn("2026-09-12", "bank_uzs", "out", 1200000, "UZS", 12000, "komunal", "", ""),
		txn("2026-09-15", "bank_uzs", "out", 2400000, "UZS", 12000, "divident", "founder", ""),
		txn("2026-09-16", "cash_uzs", "out", 1200000, "UZS", 12000, "qarz_berildi", "founder", ""),
		txn("2026-09-20", "bank_uzs", "out", 5000000, "UZS", 12000, "otkazma", "", ""),
		txn("2026-09-20", "cash_uzs", "in", 5000000, "UZS", 12000, "otkazma", "", ""),
	}
	uzs, usd := Convert(4200000, "UZS", 12000)
	data.Accruals = []FinAccrual{{ID: "a1", Date: day("2026-09-10"), CategoryCode: "ish_haqi_sotuv", CounterpartyID: "emp", CounterpartyName: "Sevinchxon", Amount: 4200000, Currency: "UZS", FxRate: 12000, AmountUZS: uzs, AmountUSD: usd}}
	ruzs, rusd := Convert(2250, "USD", 12000)
	data.Revenue = []FinRevenue{{ID: "r1", ContractID: "c1", Date: day("2026-09-26"), Amount: 2250, FxRate: 12000, AmountUZS: ruzs, AmountUSD: rusd, Milestone: 1}}
	data.Openings = []FinOpening{{ID: "o1", CounterpartyID: "founder", AsOf: day("2026-08-31"), AmountUZS: 17657000}}
	data.Recurring = []FinRecurring{{ID: "rc", Label: "Sevinchxon", Amount: 4200000, Currency: "UZS", Active: true}, {ID: "rent", Label: "Rent", Amount: 1000, Currency: "USD", Active: true}}
	data.ForecastItems = []FinForecastItem{{ID: "f1", Month: day("2026-10-01"), Direction: "in", Label: "Turon", Amount: 5000, Currency: "USD"}}
	return data
}

func TestCashPosition(t *testing.T) {
	data := sampleData()
	cash := data.CashPosition()
	balances := map[string]float64{}
	for _, account := range cash.Accounts {
		balances[account.Account.Code] = account.Balance
	}
	near(t, "cash_uzs", balances["cash_uzs"], -3000000-1200000+5000000)
	near(t, "bank_uzs", balances["bank_uzs"], 72000000-1200000-2400000-5000000)
	near(t, "cash_usd", balances["cash_usd"], 1000)
	near(t, "total", cash.TotalUZS, 800000+63400000+12000000)
}

func TestContractSummaryAndBlockedCash(t *testing.T) {
	data := sampleData()
	summary := data.ContractSummary(data.Contracts[0])
	near(t, "paid", summary.Paid, 6000)
	near(t, "recognized", summary.Recognized, 2250)
	near(t, "advance", summary.Advance, 3750)
	near(t, "unlocked", summary.Unlocked, 0.15)
	near(t, "blocked", summary.BlockedUZS, 6000*0.85*12000)
	if !summary.Milestones[0].Done || summary.Milestones[1].Done {
		t.Fatalf("milestones wrong: %+v", summary.Milestones)
	}
	// Application stage implies milestone 2.
	data.Contracts[0].ApplicationStage = "got_ceeb"
	near(t, "unlocked via app", data.Contracts[0].UnlockedShare(), 0.40)
	data.Contracts[0].Status = "completed"
	near(t, "completed unlock", data.Contracts[0].UnlockedShare(), 1)
}

func TestPayables(t *testing.T) {
	data := sampleData()
	payables := data.Payables()
	if len(payables) != 1 {
		t.Fatalf("payables = %d", len(payables))
	}
	near(t, "owed", payables[0].Balance, 1200000)
	near(t, "owed total", PayablesOwedUZS(payables), 1200000)
}

func TestFounderBalance(t *testing.T) {
	data := sampleData()
	founders := data.FounderBalances()
	if len(founders) != 1 {
		t.Fatalf("founders = %d", len(founders))
	}
	near(t, "founder balance", founders[0].Balance, 17657000-2400000)
	near(t, "founder loan", founders[0].LoansOutstanding, 1200000)
}

func TestPLUsesWorkDoneAndAccruals(t *testing.T) {
	data := sampleData()
	pl := data.PL(2026, "USD")
	sep := 8
	near(t, "revenue", pl.Revenue.Values[sep], 2250)
	near(t, "april revenue", pl.Revenue.Values[3], 0)
	near(t, "expenses", pl.TotalExpenses.Values[sep], 350+100) // accrued salary + utilities
	near(t, "profit", pl.NetProfit.Values[sep], 1800)
	near(t, "draws", pl.Draws.Values[sep], 200)
	uzs := data.PL(2026, "UZS")
	near(t, "uzs revenue", uzs.Revenue.Total, 27000000)
}

func TestCashFlowIgnoresTransfers(t *testing.T) {
	data := sampleData()
	flow := data.CashFlow(2026)
	sep := flow[8]
	near(t, "operating in", sep.OperatingIn, 72000000)
	near(t, "operating out", sep.OperatingOut, 4200000)
	near(t, "financing in", sep.FinancingIn, 12000000)
	near(t, "financing out", sep.FinancingOut, 3600000)
	near(t, "closing", sep.Closing, data.CashPosition().TotalUZS)
}

func TestForecast(t *testing.T) {
	data := sampleData()
	forecast := data.Forecast(3, 100000000)
	near(t, "sep payables", forecast[0].Payables, 1200000)
	near(t, "oct recurring", forecast[1].Recurring, 4200000+12000000)
	near(t, "oct inflow", forecast[1].Inflows, 60000000)
	near(t, "oct closing", forecast[1].Closing, 100000000-1200000+60000000-16200000)
}

func TestDashboard(t *testing.T) {
	data := sampleData()
	dash := data.Dashboard()
	near(t, "free cash", dash.FreeCashUZS, dash.Cash.TotalUZS-6000*0.85*12000)
	if dash.MonthlyBurnUZS != 16200000 {
		t.Fatalf("burn = %v", dash.MonthlyBurnUZS)
	}
	if len(dash.Alerts) == 0 {
		t.Fatal("expected alerts")
	}
}

func TestLedgerFilter(t *testing.T) {
	data := sampleData()
	ledger := data.Ledger(FinTxnFilter{Account: "cash_uzs"})
	if len(ledger.Rows) != 3 {
		t.Fatalf("rows = %d", len(ledger.Rows))
	}
	if !ledger.Rows[0].Txn.Date.Equal(day("2026-09-20")) {
		t.Fatalf("not newest first: %v", ledger.Rows[0].Txn.Date)
	}
	if got := data.Ledger(FinTxnFilter{Query: "english"}); len(got.Rows) != 0 {
		// counterparty name is empty in the sample transactions
		t.Fatalf("query rows = %d", len(got.Rows))
	}
}

func TestParseAndFormat(t *testing.T) {
	cases := map[string]float64{
		"1 200 000": 1200000, "72,287,930.99": 72287930.99, "1762,3": 1762.3, "15000": 15000,
		"1.234.567": 1234567, "1.234.567,89": 1234567.89, "72 287 930,99": 72287930.99, "1,234": 1234,
		"11789,3345": 11789.3345, "4655.81": 4655.81, "$7 860": 7860, "-45 000 000": -45000000, ".5": 0.5,
		"1\u00a0200\u00a0000": 1200000,
	}
	for input, want := range cases {
		got, err := ParseAmount(input)
		if err != nil || math.Abs(got-want) > 0.001 {
			t.Fatalf("ParseAmount(%q) = %v, %v", input, got, err)
		}
	}
	for _, input := range []string{"", "12abc", "1e5", "0x1p4", "NaN", "Inf", "1.23.4", "1,23,4", "1234,567", "1,234.5.6", "1-2", "--5"} {
		if got, err := ParseAmount(input); err == nil {
			t.Fatalf("ParseAmount(%q) = %v, want an error", input, got)
		}
	}
	if got := FormatMoney(92948931, "UZS"); got != "92 948 931 so'm" {
		t.Fatalf("FormatMoney = %q", got)
	}
	if got := FormatMoney(7860, "USD"); got != "$7 860" {
		t.Fatalf("FormatMoney usd = %q", got)
	}
	if rate, _ := RateOn([]FinRate{{Date: day("2026-09-01"), Rate: 11815}, {Date: day("2026-09-10"), Rate: 11797.92}}, day("2026-09-09")); rate != 11815 {
		t.Fatalf("RateOn = %v", rate)
	}
}

func TestCharts(t *testing.T) {
	svg := string(BarChart([]string{"Apr", "May"}, []ChartSeries{{Label: "Income", Class: "x", Values: []float64{100, -50}}}, "$"))
	if !strings.Contains(svg, "<svg") || strings.Count(svg, "<rect") != 2 {
		t.Fatalf("bar chart: %s", svg)
	}
	line := string(LineChart([]string{"a", "b", "c"}, []float64{1, -2, 3}, 1, "", 640))
	if !strings.Contains(line, "fin-chart__line--dashed") {
		t.Fatal("expected dashed forecast line")
	}
}

func TestAccrualApplies(t *testing.T) {
	salary := FinCategory{Code: "ish_haqi_sotuv", Kind: KindExpense, AccrualBased: true}
	office := FinCategory{Code: "komunal", Kind: KindExpense}
	cases := []struct {
		name     string
		input    FinTxnInput
		category FinCategory
		want     bool
	}{
		{"salary paid", FinTxnInput{Direction: "out"}, salary, true},
		{"salary refunded", FinTxnInput{Direction: "in"}, salary, false},
		{"office expense", FinTxnInput{Direction: "out"}, office, false},
		{"transfer", FinTxnInput{Direction: "out", ToAccount: "cash_uzs"}, salary, false},
	}
	for _, tc := range cases {
		if got := tc.input.AccrualApplies(tc.category); got != tc.want {
			t.Fatalf("%s: AccrualApplies = %v, want %v", tc.name, got, tc.want)
		}
	}
}
