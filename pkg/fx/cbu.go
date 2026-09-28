// Package fx fetches the official UZS/USD rate from the Central Bank of
// Uzbekistan (cbu.uz).
package fx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const cbuURL = "https://cbu.uz/uz/arkhiv-kursov-valyut/json/USD/%s/"

type Client struct {
	HTTP    *http.Client
	BaseURL string // format string with one %s for the date; empty = cbu.uz
}

func NewClient() *Client {
	return &Client{HTTP: &http.Client{Timeout: 4 * time.Second}}
}

type cbuRate struct {
	Rate string `json:"Rate"`
	Date string `json:"Date"`
}

// USDRate returns the CBU rate (so'm per dollar) valid on the given day.
func (c *Client) USDRate(ctx context.Context, day time.Time) (float64, error) {
	base := c.BaseURL
	if base == "" {
		base = cbuURL
	}
	url := fmt.Sprintf(base, day.Format("2006-01-02"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 4 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("cbu.uz returned %s", resp.Status)
	}
	var rows []cbuRate
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, errors.New("cbu.uz returned no rate")
	}
	rate, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(rows[0].Rate), ",", "."), 64)
	if err != nil || rate <= 0 {
		return 0, fmt.Errorf("cbu.uz returned an invalid rate %q", rows[0].Rate)
	}
	return rate, nil
}
