package web

import (
	"testing"
	"time"

	"edu-license/pkg/app"
)

func TestSafeFinanceTarget(t *testing.T) {
	const fallback = "/admin/finance/transactions"
	cases := map[string]string{
		"/admin/finance": "/admin/finance",
		"/admin/finance/transactions?month=2026-09": "/admin/finance/transactions?month=2026-09",
		"/admin/finance/contracts/abc":              "/admin/finance/contracts/abc",
		"":                                          fallback,
		"https://evil.com/admin/finance":            fallback,
		"//evil.com/admin/finance":                  fallback,
		"/admin/finance/../../\\evil.com":           fallback,
		"/admin/finance/../../evil":                 fallback,
		"/admin/financex":                           fallback,
		"/admin/finance\r\nSet-Cookie: x=1":         fallback,
		"/admin/crm":                                fallback,
	}
	for back, want := range cases {
		if got := safeFinanceTarget(back, fallback); got != want {
			t.Fatalf("safeFinanceTarget(%q) = %q, want %q", back, got, want)
		}
	}
}

func TestEditedRate(t *testing.T) {
	sep9 := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)
	sep10 := sep9.AddDate(0, 0, 1)
	original := app.FinTransaction{Date: sep9, FxRate: 11813.2134}
	cases := []struct {
		name string
		date time.Time
		rate float64
		want float64
	}{
		{"untouched, same day keeps exact rate", sep9, 11813.21, 11813.2134},
		{"untouched, new day uses that day's rate", sep10, 11813.21, 0},
		{"typed a new rate", sep10, 11900, 11900},
		{"cleared the field", sep9, 0, 0},
	}
	for _, tc := range cases {
		got := editedRate(original, app.FinTxnInput{Date: tc.date, FxRate: tc.rate})
		if got != tc.want {
			t.Fatalf("%s: editedRate = %v, want %v", tc.name, got, tc.want)
		}
	}
}
