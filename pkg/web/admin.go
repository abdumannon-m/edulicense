package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"edu-license/pkg/app"
	"edu-license/pkg/auth"
	"edu-license/pkg/httpx"
	"edu-license/pkg/storage"
	"edu-license/pkg/store"
	"github.com/go-chi/chi/v5"
)

func (s *Server) loginPage(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.auth.UserFromRequest(r); ok {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	data := app.AdminPageData{
		BaseURL:   s.cfg.AppBaseURL,
		Title:     "Admin login",
		CSRFToken: s.csrf(w, r),
		Error:     queryMessage(r.URL.Query(), "error"),
	}
	s.renderer.Render(w, http.StatusOK, "admin_login", data)
}

func (s *Server) loginPost(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	session, expiresAt, err := s.auth.Login(r.Context(), r.FormValue("email"), r.FormValue("password"))
	if err != nil {
		redirectWithError(w, r, "/admin/login", err)
		return
	}
	s.auth.SetSessionCookie(w, session, expiresAt)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (s *Server) logoutPost(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	if cookie, err := r.Cookie(auth.SessionCookie); err == nil {
		s.auth.Logout(r.Context(), cookie.Value)
	}
	s.auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func (s *Server) overview(w http.ResponseWriter, r *http.Request) {
	data := s.adminData(w, r, "Overview")
	stats, err := s.store.DashboardStats(r.Context(), time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	users, err := s.store.ListUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data.Stats = stats
	data.Users = users
	data.Success = queryMessage(r.URL.Query(), "success")
	data.Error = queryMessage(r.URL.Query(), "error")
	s.renderer.Render(w, http.StatusOK, "admin_overview", data)
}

func (s *Server) userCreate(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	currentUser, _ := httpx.CurrentUser(r)
	role := app.Role(r.FormValue("role"))
	if role != app.RoleSuperAdmin && role != app.RoleAdmin && role != app.RoleSales {
		redirectWithError(w, r, "/admin/overview", errors.New("invalid role"))
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	if name == "" || email == "" {
		redirectWithError(w, r, "/admin/overview", errors.New("name and email are required"))
		return
	}
	password := r.FormValue("password")
	if len(password) < 10 {
		redirectWithError(w, r, "/admin/overview", errors.New("password must be at least 10 characters"))
		return
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		redirectWithError(w, r, "/admin/overview", err)
		return
	}
	created, err := s.store.CreateUser(r.Context(), auth.CreateUserInput{
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		Role:         role,
		Active:       true,
	})
	if err != nil {
		redirectWithError(w, r, "/admin/overview", err)
		return
	}
	_ = s.store.LogActivity(r.Context(), currentUser.ID, "create", "user", created.ID, "Created user "+created.Email)
	redirectWithSuccess(w, r, "/admin/overview", "User created")
}

func (s *Server) applicationsIndex(w http.ResponseWriter, r *http.Request) {
	data := s.adminData(w, r, "SAT test center applications")
	applications, err := s.store.ListApplications(r.Context(), "", "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.attachApplicationCertificates(r.Context(), applications); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	users, err := s.store.ListUsersByRoles(r.Context(), app.RoleAdmin, app.RoleSuperAdmin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data.Applications = applications
	data.Users = users
	data.Success = queryMessage(r.URL.Query(), "success")
	data.Error = queryMessage(r.URL.Query(), "error")
	s.renderer.Render(w, http.StatusOK, "admin_applications", data)
}

func (s *Server) applicationNew(w http.ResponseWriter, r *http.Request) {
	data := s.adminData(w, r, "New application")
	users, err := s.store.ListUsersByRoles(r.Context(), app.RoleAdmin, app.RoleSuperAdmin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data.Users = users
	data.Application = app.TestCenterApplication{
		InstitutionType: "school",
		PaymentOption:   "our_investment",
		PaymentCurrency: "USD",
		Stage:           "start",
		ChecklistStatus: map[string]bool{},
	}
	s.renderer.Render(w, http.StatusOK, "admin_application_form", data)
}

func (s *Server) applicationCreate(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	user, _ := httpx.CurrentUser(r)
	input, err := parseApplicationInput(r)
	if err != nil {
		redirectWithError(w, r, "/admin/applications/new", err)
		return
	}
	created, err := s.store.CreateApplication(r.Context(), input, user.ID)
	if err != nil {
		redirectWithError(w, r, "/admin/applications/new", err)
		return
	}
	redirectWithSuccess(w, r, "/admin/applications/"+created.ID, "Application created")
}

func (s *Server) applicationEdit(w http.ResponseWriter, r *http.Request) {
	data := s.adminData(w, r, "Edit application")
	application, err := s.store.ApplicationByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := s.attachApplicationCertificate(r.Context(), &application); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	users, err := s.store.ListUsersByRoles(r.Context(), app.RoleAdmin, app.RoleSuperAdmin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data.Application = application
	data.Users = users
	data.Success = queryMessage(r.URL.Query(), "success")
	data.Error = queryMessage(r.URL.Query(), "error")
	s.renderer.Render(w, http.StatusOK, "admin_application_form", data)
}

func (s *Server) applicationUpdate(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	id := chi.URLParam(r, "id")
	user, _ := httpx.CurrentUser(r)
	input, err := parseApplicationInput(r)
	if err != nil {
		redirectWithError(w, r, "/admin/applications/"+id, err)
		return
	}
	if _, err := s.store.UpdateApplication(r.Context(), id, input, user.ID); err != nil {
		redirectWithError(w, r, "/admin/applications/"+id, err)
		return
	}
	redirectWithSuccess(w, r, "/admin/applications/"+id, "Application updated")
}

func (s *Server) applicationGenerateCertificate(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	id := chi.URLParam(r, "id")
	user, _ := httpx.CurrentUser(r)
	application, err := s.store.ApplicationByID(r.Context(), id)
	if err != nil {
		redirectWithError(w, r, "/admin/applications/"+id, err)
		return
	}
	if application.Stage != "in_test_center_list" {
		redirectWithError(w, r, "/admin/applications/"+id, errors.New("license can only be generated after the application is in the test center list"))
		return
	}
	if strings.TrimSpace(application.CEEBCode) == "" {
		redirectWithError(w, r, "/admin/applications/"+id, errors.New("test center code is required before generating a license"))
		return
	}
	input := app.CertificateInputForApplication(application, s.nowInAppTimezone())
	if _, err := s.store.CertificateBySlug(r.Context(), input.Slug); err == nil {
		redirectWithSuccess(w, r, "/admin/applications/"+id, "License already generated")
		return
	} else if !store.IsNotFound(err) {
		redirectWithError(w, r, "/admin/applications/"+id, err)
		return
	}
	certificate, err := s.store.UpsertCertificate(r.Context(), input)
	if err != nil {
		redirectWithError(w, r, "/admin/applications/"+id, err)
		return
	}
	_ = s.store.LogActivity(r.Context(), user.ID, "create", "certificate", certificate.ID, "Generated license for "+application.InstitutionName)
	redirectWithSuccess(w, r, "/admin/applications/"+id, "License generated")
}

func (s *Server) applicationDelete(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	user, _ := httpx.CurrentUser(r)
	if err := s.store.DeleteApplication(r.Context(), chi.URLParam(r, "id"), user.ID); err != nil {
		redirectWithError(w, r, "/admin/applications", err)
		return
	}
	redirectWithSuccess(w, r, "/admin/applications", "Application deleted")
}

func (s *Server) applicationUploadDocument(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	if s.uploader == nil || !s.uploader.Configured() {
		redirectWithError(w, r, "/admin/applications/"+chi.URLParam(r, "id"), errors.New("S3/R2 storage is not configured"))
		return
	}
	user, _ := httpx.CurrentUser(r)
	applicationID := chi.URLParam(r, "id")
	file, header, err := r.FormFile("file")
	if err != nil {
		redirectWithError(w, r, "/admin/applications/"+applicationID, err)
		return
	}
	defer file.Close()
	docType := r.FormValue("doc_type")
	key := storage.ApplicationDocumentKey(applicationID, header.Filename)
	contentType := header.Header.Get("Content-Type")
	reader := io.LimitReader(file, 32<<20)
	if err := s.uploader.Upload(r.Context(), key, contentType, reader); err != nil {
		redirectWithError(w, r, "/admin/applications/"+applicationID, err)
		return
	}
	err = s.store.CreateApplicationDocument(r.Context(), app.DocumentInput{
		ApplicationID:    applicationID,
		DocType:          docType,
		OriginalFilename: header.Filename,
		ContentType:      contentType,
		SizeBytes:        header.Size,
		StorageKey:       key,
		UploadedBy:       user.ID,
	})
	if err != nil {
		redirectWithError(w, r, "/admin/applications/"+applicationID, err)
		return
	}
	_ = s.store.LogActivity(r.Context(), user.ID, "upload", "application", applicationID, "Uploaded "+docType)
	redirectWithSuccess(w, r, "/admin/applications/"+applicationID, "Document uploaded")
}

func (s *Server) reminderCreate(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	user, _ := httpx.CurrentUser(r)
	input, err := parseReminderInput(r)
	if err != nil {
		redirectWithError(w, r, r.Referer(), err)
		return
	}
	if _, err := s.store.CreateReminder(r.Context(), input, user.ID); err != nil {
		redirectWithError(w, r, r.Referer(), err)
		return
	}
	redirectWithSuccess(w, r, r.Referer(), "Reminder created")
}

func (s *Server) crmIndex(w http.ResponseWriter, r *http.Request) {
	data := s.adminData(w, r, "Sales CRM")
	deals, err := s.store.DealsByStage(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	users, err := s.store.ListUsersByRoles(r.Context(), app.RoleSales, app.RoleSuperAdmin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data.DealsByStage = deals
	data.Users = users
	data.Success = queryMessage(r.URL.Query(), "success")
	data.Error = queryMessage(r.URL.Query(), "error")
	s.renderer.Render(w, http.StatusOK, "admin_crm", data)
}

func (s *Server) crmCreateDeal(w http.ResponseWriter, r *http.Request) {
	if !s.validateCSRF(w, r) {
		return
	}
	user, _ := httpx.CurrentUser(r)
	input, err := parseDealInput(r)
	if err != nil {
		redirectWithError(w, r, "/admin/crm", err)
		return
	}
	if user.Role == app.RoleSales {
		input.AssignedSalesAgentID = user.ID
	}
	if _, err := s.store.CreateDeal(r.Context(), input, user.ID); err != nil {
		redirectWithError(w, r, "/admin/crm", err)
		return
	}
	redirectWithSuccess(w, r, "/admin/crm", "Deal created")
}

func (s *Server) crmMoveDeal(w http.ResponseWriter, r *http.Request) {
	wantsJSON := wantsJSONResponse(r)
	if !s.auth.ValidateCSRF(r) {
		if wantsJSON {
			writeJSONError(w, http.StatusForbidden, "invalid CSRF token")
			return
		}
		http.Error(w, "invalid CSRF token", http.StatusForbidden)
		return
	}
	user, _ := httpx.CurrentUser(r)
	id := chi.URLParam(r, "id")
	stage := r.FormValue("stage")
	if !app.ValidSalesStage(stage) {
		if wantsJSON {
			writeJSONError(w, http.StatusBadRequest, "invalid CRM stage")
			return
		}
		redirectWithError(w, r, "/admin/crm", errors.New("invalid CRM stage"))
		return
	}
	deal, err := s.store.DealByID(r.Context(), id)
	if err != nil {
		if wantsJSON {
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		redirectWithError(w, r, "/admin/crm", err)
		return
	}
	if user.Role == app.RoleSales && deal.AssignedSalesAgentID != user.ID && deal.CreatedBy != user.ID {
		if wantsJSON {
			writeJSONError(w, http.StatusForbidden, "forbidden")
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	updated, err := s.store.UpdateDealStage(r.Context(), id, stage, user.ID)
	if err != nil {
		if wantsJSON {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		redirectWithError(w, r, "/admin/crm", err)
		return
	}
	if wantsJSON {
		writeJSON(w, http.StatusOK, map[string]string{
			"id":                     updated.ID,
			"stage":                  updated.Stage,
			"stageLabel":             app.StageLabel(updated.Stage),
			"convertedApplicationID": updated.ConvertedApplicationID,
		})
		return
	}
	redirectWithSuccess(w, r, "/admin/crm", "Deal moved")
}

func wantsJSONResponse(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "application/json") ||
		strings.EqualFold(r.Header.Get("X-Requested-With"), "fetch")
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func parseApplicationInput(r *http.Request) (app.ApplicationInput, error) {
	input := app.ApplicationInput{
		InstitutionName:   strings.TrimSpace(r.FormValue("institution_name")),
		InstitutionType:   defaultString(r.FormValue("institution_type"), "school"),
		Location:          strings.TrimSpace(r.FormValue("location")),
		Website:           strings.TrimSpace(r.FormValue("website")),
		ResponsiblePerson: strings.TrimSpace(r.FormValue("responsible_person")),
		Phone:             strings.TrimSpace(r.FormValue("phone")),
		DirectorPhone:     strings.TrimSpace(r.FormValue("director_phone")),
		AssignedAdminID:   r.FormValue("assigned_admin_id"),
		SalesReferralID:   r.FormValue("sales_referral_id"),
		PaymentOption:     defaultString(r.FormValue("payment_option"), "our_investment"),
		PaymentAmount:     strings.TrimSpace(r.FormValue("payment_amount")),
		PaymentCurrency:   defaultString(r.FormValue("payment_currency"), "USD"),
		ProfitSharing:     r.FormValue("profit_sharing"),
		CEEBCode:          strings.TrimSpace(r.FormValue("ceeb_code")),
		Stage:             defaultString(r.FormValue("stage"), "start"),
		ChecklistStatus:   map[string]bool{},
	}
	if input.InstitutionName == "" {
		return input, errors.New("institution name is required")
	}
	if !app.ValidApplicationStage(input.Stage) {
		return input, errors.New("invalid application stage")
	}
	if err := app.ValidateProfitSharing(input.ProfitSharing); err != nil {
		return input, err
	}
	for _, item := range app.ChecklistItems {
		input.ChecklistStatus[item.Key] = r.FormValue("checklist_"+item.Key) == "on"
	}
	return input, nil
}

func parseDealInput(r *http.Request) (app.DealInput, error) {
	capacity, _ := strconv.Atoi(r.FormValue("capacity"))
	input := app.DealInput{
		SchoolName:           strings.TrimSpace(r.FormValue("school_name")),
		Capacity:             capacity,
		AssignedSalesAgentID: r.FormValue("assigned_sales_agent_id"),
		Stage:                defaultString(r.FormValue("stage"), "new_lead"),
		NegotiationPrice:     strings.TrimSpace(r.FormValue("negotiation_price")),
		Currency:             defaultString(r.FormValue("currency"), "USD"),
		ProfitSharing:        r.FormValue("profit_sharing"),
		Notes:                strings.TrimSpace(r.FormValue("notes")),
	}
	if input.SchoolName == "" {
		return input, errors.New("school name is required")
	}
	if !app.ValidSalesStage(input.Stage) {
		return input, errors.New("invalid CRM stage")
	}
	if err := app.ValidateProfitSharing(input.ProfitSharing); err != nil {
		return input, err
	}
	return input, nil
}

func parseReminderInput(r *http.Request) (app.ReminderInput, error) {
	entityType := r.FormValue("entity_type")
	if entityType != "application" && entityType != "deal" {
		return app.ReminderInput{}, errors.New("invalid reminder entity")
	}
	dueAt, err := parseDueAt(r)
	if err != nil {
		return app.ReminderInput{}, err
	}
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		return app.ReminderInput{}, errors.New("reminder title is required")
	}
	return app.ReminderInput{
		EntityType: entityType,
		EntityID:   r.FormValue("entity_id"),
		Title:      title,
		Note:       strings.TrimSpace(r.FormValue("note")),
		DueAt:      dueAt,
	}, nil
}

func parseDueAt(r *http.Request) (time.Time, error) {
	if daysRaw := strings.TrimSpace(r.FormValue("due_days")); daysRaw != "" {
		days, err := strconv.Atoi(daysRaw)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid due days")
		}
		return time.Now().AddDate(0, 0, days), nil
	}
	dateRaw := strings.TrimSpace(r.FormValue("due_date"))
	if dateRaw == "" {
		return time.Time{}, errors.New("due date or due days is required")
	}
	return time.Parse("2006-01-02", dateRaw)
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func (s *Server) nowInAppTimezone() time.Time {
	loc, err := time.LoadLocation(s.cfg.Timezone)
	if err != nil {
		return time.Now()
	}
	return time.Now().In(loc)
}

func (s *Server) attachApplicationCertificate(ctx context.Context, application *app.TestCenterApplication) error {
	if application == nil || application.ID == "" || application.Stage != "in_test_center_list" || strings.TrimSpace(application.CEEBCode) == "" {
		return nil
	}
	input := app.CertificateInputForApplication(*application, time.Now())
	certificate, err := s.store.CertificateBySlug(ctx, input.Slug)
	if err != nil {
		if store.IsNotFound(err) {
			return nil
		}
		return err
	}
	application.CertificateID = certificate.ID
	application.CertificateSlug = certificate.Slug
	application.VerificationID = certificate.VerificationID
	return nil
}

func (s *Server) attachApplicationCertificates(ctx context.Context, applications []app.TestCenterApplication) error {
	slugIndexes := map[string][]int{}
	slugs := []string{}
	seen := map[string]bool{}
	for i := range applications {
		application := &applications[i]
		if application.ID == "" || application.Stage != "in_test_center_list" || strings.TrimSpace(application.CEEBCode) == "" {
			continue
		}
		input := app.CertificateInputForApplication(*application, time.Now())
		if !seen[input.Slug] {
			slugs = append(slugs, input.Slug)
			seen[input.Slug] = true
		}
		slugIndexes[input.Slug] = append(slugIndexes[input.Slug], i)
	}
	certificates, err := s.store.CertificatesBySlugs(ctx, slugs)
	if err != nil {
		return err
	}
	for slug, certificate := range certificates {
		for _, index := range slugIndexes[slug] {
			applications[index].CertificateID = certificate.ID
			applications[index].CertificateSlug = certificate.Slug
			applications[index].VerificationID = certificate.VerificationID
		}
	}
	return nil
}
