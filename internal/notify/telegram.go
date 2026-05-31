package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"edu-license/internal/app"
)

type Telegram struct {
	token   string
	chatID  string
	baseURL string
	client  *http.Client
}

func NewTelegram(token, chatID, baseURL string) *Telegram {
	return &Telegram{
		token:   token,
		chatID:  chatID,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *Telegram) Configured() bool {
	return t != nil && t.token != "" && t.chatID != ""
}

func (t *Telegram) SendReminder(ctx context.Context, reminder app.Reminder) error {
	if !t.Configured() {
		return fmt.Errorf("telegram bot token or operations chat id is not configured")
	}
	adminPath := "/admin/applications/" + reminder.EntityID
	if reminder.EntityType == "deal" {
		adminPath = "/admin/crm"
	}
	text := fmt.Sprintf(
		"Reminder: %s\nEntity: %s\nOwner: %s\nStage: %s\nDue: %s\nNote: %s\n%s%s",
		reminder.Title,
		reminder.EntityName,
		emptyFallback(reminder.OwnerName, "Unassigned"),
		app.StageLabel(reminder.CurrentStage),
		reminder.DueAt.Format("2006-01-02"),
		emptyFallback(reminder.Note, "-"),
		t.baseURL,
		adminPath,
	)
	payload := map[string]string{
		"chat_id": t.chatID,
		"text":    text,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.telegram.org/bot"+t.token+"/sendMessage", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram sendMessage returned %s", resp.Status)
	}
	return nil
}

func emptyFallback(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
