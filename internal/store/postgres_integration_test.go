package store

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"testing"
	"time"

	"edu-license/internal/app"
	"edu-license/internal/auth"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func TestPostgresIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run Postgres integration tests")
	}
	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(db, "../../migrations"); err != nil {
		t.Fatal(err)
	}

	st, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	suffix := time.Now().UnixNano()
	hash, err := auth.HashPassword("integration-password")
	if err != nil {
		t.Fatal(err)
	}
	user, err := st.CreateUser(ctx, auth.CreateUserInput{
		Name:         "Integration Super Admin",
		Email:        "integration-super-admin-" + strconv.FormatInt(suffix, 10) + "@example.test",
		PasswordHash: hash,
		Role:         app.RoleSuperAdmin,
		Active:       true,
	})
	if err != nil {
		t.Fatal(err)
	}

	deal, err := st.CreateDeal(ctx, app.DealInput{
		SchoolName:       "Integration School",
		Capacity:         240,
		Stage:            "new_lead",
		NegotiationPrice: "12000",
		Currency:         "USD",
		ProfitSharing:    "50-50",
	}, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	deal, err = st.UpdateDealStage(ctx, deal.ID, "won", user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if deal.ConvertedApplicationID == "" {
		t.Fatal("won deal should create a linked application")
	}

	reminder, err := st.CreateReminder(ctx, app.ReminderInput{
		EntityType: "application",
		EntityID:   deal.ConvertedApplicationID,
		Title:      "Integration due reminder",
		DueAt:      time.Now().Add(-time.Hour),
	}, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	due, err := st.ListDueReminders(ctx, time.Now(), 10)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range due {
		if item.ID == reminder.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("created overdue reminder should be returned")
	}
	if err := st.MarkReminderSent(ctx, reminder.ID, time.Now()); err != nil {
		t.Fatal(err)
	}
	due, err = st.ListDueReminders(ctx, time.Now(), 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range due {
		if item.ID == reminder.ID {
			t.Fatal("sent reminder should not be returned again")
		}
	}
}
