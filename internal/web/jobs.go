package web

import (
	"context"
	"fmt"
	"time"
)

func (s *Server) SendDueReminders(ctx context.Context, now time.Time) (int, error) {
	reminders, err := s.store.ListDueReminders(ctx, now, 100)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, reminder := range reminders {
		if s.telegram == nil || !s.telegram.Configured() {
			return sent, fmt.Errorf("telegram is not configured")
		}
		if err := s.telegram.SendReminder(ctx, reminder); err != nil {
			return sent, err
		}
		if err := s.store.MarkReminderSent(ctx, reminder.ID, now); err != nil {
			return sent, err
		}
		sent++
	}
	return sent, nil
}
