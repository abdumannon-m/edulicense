package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"edu-license/pkg/app"
	"edu-license/pkg/auth"
	"edu-license/pkg/config"
	"edu-license/pkg/httpx"
	"edu-license/pkg/notify"
	"edu-license/pkg/storage"
	"github.com/go-chi/chi/v5"
)

type Store interface {
	ListUsers(ctx context.Context) ([]app.User, error)
	ListUsersByRoles(ctx context.Context, roles ...app.Role) ([]app.User, error)
	ListApplications(ctx context.Context, stage, adminID, location string) ([]app.TestCenterApplication, error)
	ApplicationByID(ctx context.Context, id string) (app.TestCenterApplication, error)
	CreateApplication(ctx context.Context, input app.ApplicationInput, actorID string) (app.TestCenterApplication, error)
	UpdateApplication(ctx context.Context, id string, input app.ApplicationInput, actorID string) (app.TestCenterApplication, error)
	DeleteApplication(ctx context.Context, id string, actorID string) error
	CreateApplicationDocument(ctx context.Context, input app.DocumentInput) error
	DealsByStage(ctx context.Context) (map[string][]app.SalesDeal, error)
	DealByID(ctx context.Context, id string) (app.SalesDeal, error)
	CreateDeal(ctx context.Context, input app.DealInput, actorID string) (app.SalesDeal, error)
	UpdateDealStage(ctx context.Context, id, stage, actorID string) (app.SalesDeal, error)
	CreateReminder(ctx context.Context, input app.ReminderInput, actorID string) (app.Reminder, error)
	ListOpenReminders(ctx context.Context, limit int) ([]app.Reminder, error)
	ListDueReminders(ctx context.Context, now time.Time, limit int) ([]app.Reminder, error)
	MarkReminderSent(ctx context.Context, id string, sentAt time.Time) error
	DashboardStats(ctx context.Context, now time.Time) (app.DashboardStats, error)
	LogActivity(ctx context.Context, userID, action, entityType, entityID, summary string) error
	CreateUser(ctx context.Context, input auth.CreateUserInput) (app.User, error)
	CertificateBySlug(ctx context.Context, slug string) (app.Certificate, error)
	CertificatesBySlugs(ctx context.Context, slugs []string) (map[string]app.Certificate, error)
	UpsertCertificate(ctx context.Context, input app.CertificateInput) (app.Certificate, error)
}

type Server struct {
	cfg      config.Config
	store    Store
	auth     *auth.Service
	renderer *app.Renderer
	uploader storage.Uploader
	telegram *notify.Telegram
}

func New(cfg config.Config, store Store, authSvc *auth.Service, renderer *app.Renderer, uploader storage.Uploader, telegram *notify.Telegram) *Server {
	return &Server{
		cfg:      cfg,
		store:    store,
		auth:     authSvc,
		renderer: renderer,
		uploader: uploader,
		telegram: telegram,
	}
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.NotFound(s.notFound)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	r.Get("/", s.home("en"))
	r.Get("/uz", s.home("uz"))
	r.Get("/privacy", s.privacy("en"))
	r.Get("/uz/privacy", s.privacy("uz"))
	r.Get("/verify/{slug}", s.verify("en"))
	r.Get("/uz/verify/{slug}", s.verify("uz"))
	r.Get("/verify/{slug}/qr.png", s.certificateQR)
	r.Get("/robots.txt", s.robots)
	r.Get("/sitemap.xml", s.sitemap)

	r.Get("/admin/login", s.loginPage)
	r.Post("/admin/login", s.loginPost)

	r.Group(func(admin chi.Router) {
		admin.Use(s.requireAuth)
		admin.Get("/admin", s.adminIndex)
		admin.Post("/admin/logout", s.logoutPost)
		admin.Get("/admin/overview", s.requireArea("overview", s.overview))
		admin.Post("/admin/users", s.requireArea("overview", s.userCreate))
		admin.Get("/admin/applications", s.requireArea("applications", s.applicationsIndex))
		admin.Get("/admin/applications/new", s.requireArea("applications", s.applicationNew))
		admin.Post("/admin/applications", s.requireArea("applications", s.applicationCreate))
		admin.Get("/admin/applications/{id}", s.requireArea("applications", s.applicationEdit))
		admin.Post("/admin/applications/{id}", s.requireArea("applications", s.applicationUpdate))
		admin.Post("/admin/applications/{id}/certificate", s.requireArea("applications", s.applicationGenerateCertificate))
		admin.Post("/admin/applications/{id}/delete", s.requireArea("applications", s.applicationDelete))
		admin.Post("/admin/applications/{id}/documents", s.requireArea("applications", s.applicationUploadDocument))
		admin.Post("/admin/reminders", s.requireArea("reminders", s.reminderCreate))
		admin.Get("/admin/crm", s.requireArea("crm", s.crmIndex))
		admin.Post("/admin/crm/deals", s.requireArea("crm", s.crmCreateDeal))
		admin.Post("/admin/crm/deals/{id}/stage", s.requireArea("crm", s.crmMoveDeal))
	})

	return r
}

func (s *Server) adminIndex(w http.ResponseWriter, r *http.Request) {
	user, _ := httpx.CurrentUser(r)
	switch user.Role {
	case app.RoleSuperAdmin:
		http.Redirect(w, r, "/admin/overview", http.StatusSeeOther)
	case app.RoleAdmin:
		http.Redirect(w, r, "/admin/applications", http.StatusSeeOther)
	case app.RoleSales:
		http.Redirect(w, r, "/admin/crm", http.StatusSeeOther)
	default:
		http.Error(w, "forbidden", http.StatusForbidden)
	}
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := s.auth.UserFromRequest(r)
		if !ok {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, httpx.WithUser(r, user))
	})
}

func (s *Server) requireArea(area string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := httpx.CurrentUser(r)
		if user.Role != app.RoleSuperAdmin && area == "overview" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if area != "overview" && !app.RoleCanAccess(user.Role, area) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		handler(w, r)
	}
}

func (s *Server) csrf(w http.ResponseWriter, r *http.Request) string {
	return s.auth.EnsureCSRF(w, r)
}

func (s *Server) validateCSRF(w http.ResponseWriter, r *http.Request) bool {
	if s.auth.ValidateCSRF(r) {
		return true
	}
	http.Error(w, "invalid CSRF token", http.StatusForbidden)
	return false
}

func (s *Server) adminData(w http.ResponseWriter, r *http.Request, title string) app.AdminPageData {
	user, _ := httpx.CurrentUser(r)
	return app.AdminPageData{
		BaseURL:              s.cfg.AppBaseURL,
		Title:                title,
		CurrentUser:          user,
		CSRFToken:            s.csrf(w, r),
		SalesStages:          app.SalesStages,
		ApplicationStages:    app.ApplicationStages,
		ProfitSharingOptions: app.ProfitSharingOptions,
		PaymentOptions:       app.PaymentOptions,
		ChecklistItems:       app.ChecklistItems,
		DocumentTypes:        app.DocumentTypes,
	}
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}

func redirectWithError(w http.ResponseWriter, r *http.Request, target string, err error) {
	if target == "" {
		target = "/admin"
	}
	http.Redirect(w, r, fmt.Sprintf("%s?error=%s", target, urlQuery(err.Error())), http.StatusSeeOther)
}

func redirectWithSuccess(w http.ResponseWriter, r *http.Request, target, message string) {
	if target == "" {
		target = "/admin"
	}
	http.Redirect(w, r, fmt.Sprintf("%s?success=%s", target, urlQuery(message)), http.StatusSeeOther)
}
