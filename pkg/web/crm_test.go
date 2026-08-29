package web

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"edu-license/pkg/app"
	"edu-license/pkg/auth"
	"edu-license/pkg/config"
	"edu-license/pkg/httpx"
)

type crmStore struct {
	Store
	deal        app.SalesDeal
	dealErr     error
	movedStage  string
	movedActor  string
	deletedID   string
	deletedByID string
}

func (s *crmStore) DealByID(ctx context.Context, id string) (app.SalesDeal, error) {
	if s.dealErr != nil {
		return app.SalesDeal{}, s.dealErr
	}
	return s.deal, nil
}

func (s *crmStore) UpdateDealStage(ctx context.Context, id, stage, actorID string) (app.SalesDeal, error) {
	s.movedStage = stage
	s.movedActor = actorID
	updated := s.deal
	updated.Stage = stage
	return updated, nil
}

func (s *crmStore) DeleteDeal(ctx context.Context, id, actorID string) error {
	s.deletedID = id
	s.deletedByID = actorID
	return nil
}

func (s *crmStore) LogActivity(ctx context.Context, userID, action, entityType, entityID, summary string) error {
	return nil
}

func newCRMServer(store Store) *Server {
	cfg := config.Config{SessionSecret: "test-secret"}
	return New(cfg, store, auth.NewService(nil, cfg.SessionSecret, false, time.Hour), nil, nil, nil)
}

// crmRequest builds a POST carrying a CSRF cookie and matching token, with the
// given user attached the way requireAuth attaches it.
func crmRequest(s *Server, target string, form url.Values, user app.User) *http.Request {
	seed := httptest.NewRequest(http.MethodPost, target, nil)
	probe := httptest.NewRecorder()
	token := s.auth.EnsureCSRF(probe, seed)
	cookie := probe.Result().Cookies()[0]

	form.Set("csrf_token", token)
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "fetch")
	req.AddCookie(cookie)
	return httpx.WithUser(req, user)
}

func TestCRMMoveDealAllowsSalesUserToMoveAnotherAgentsDeal(t *testing.T) {
	store := &crmStore{deal: app.SalesDeal{
		ID:                   "deal-1",
		SchoolName:           "Salam school",
		Stage:                "new_lead",
		AssignedSalesAgentID: "other-agent",
		CreatedBy:            "other-agent",
	}}
	server := newCRMServer(store)
	user := app.User{ID: "manzura", Role: app.RoleSales}

	rec := httptest.NewRecorder()
	req := crmRequest(server, "/admin/crm/deals/deal-1/stage", url.Values{"stage": {"contacted"}}, user)
	server.crmMoveDeal(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body %q)", rec.Code, http.StatusOK, rec.Body.String())
	}
	if store.movedStage != "contacted" {
		t.Errorf("stage = %q, want %q", store.movedStage, "contacted")
	}
	if store.movedActor != "manzura" {
		t.Errorf("actor = %q, want %q", store.movedActor, "manzura")
	}
}

func TestCRMDeleteDealAllowsSalesUserToDeleteAnotherAgentsDeal(t *testing.T) {
	store := &crmStore{deal: app.SalesDeal{
		ID:                   "deal-1",
		Stage:                "new_lead",
		AssignedSalesAgentID: "other-agent",
		CreatedBy:            "other-agent",
	}}
	server := newCRMServer(store)
	user := app.User{ID: "manzura", Role: app.RoleSales}

	rec := httptest.NewRecorder()
	req := crmRequest(server, "/admin/crm/deals/deal-1/delete", url.Values{}, user)
	server.crmDeleteDeal(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body %q)", rec.Code, http.StatusOK, rec.Body.String())
	}
	if store.deletedByID != "manzura" {
		t.Errorf("actor = %q, want %q", store.deletedByID, "manzura")
	}
}

func TestCRMMoveDealStillRejectsUnknownDeal(t *testing.T) {
	store := &crmStore{dealErr: errors.New("deal not found")}
	server := newCRMServer(store)
	user := app.User{ID: "manzura", Role: app.RoleSales}

	rec := httptest.NewRecorder()
	req := crmRequest(server, "/admin/crm/deals/missing/stage", url.Values{"stage": {"contacted"}}, user)
	server.crmMoveDeal(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestCRMMoveDealRejectsInvalidStage(t *testing.T) {
	store := &crmStore{deal: app.SalesDeal{ID: "deal-1", Stage: "new_lead"}}
	server := newCRMServer(store)
	user := app.User{ID: "manzura", Role: app.RoleSales}

	rec := httptest.NewRecorder()
	req := crmRequest(server, "/admin/crm/deals/deal-1/stage", url.Values{"stage": {"not_a_stage"}}, user)
	server.crmMoveDeal(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if store.movedStage != "" {
		t.Errorf("stage was updated to %q, want no update", store.movedStage)
	}
}
