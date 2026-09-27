package fx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUSDRate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "2026-09-26") {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Write([]byte(`[{"id":69,"Code":"840","Ccy":"USD","Rate":"11825.40","Date":"26.09.2026"}]`))
	}))
	defer server.Close()
	client := &Client{HTTP: server.Client(), BaseURL: server.URL + "/%s/"}
	rate, err := client.USDRate(context.Background(), time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC))
	if err != nil || rate != 11825.40 {
		t.Fatalf("rate = %v, %v", rate, err)
	}
}
