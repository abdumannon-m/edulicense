package app

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// assetVersion changes on every start, so a deploy busts cached stylesheets.
var assetVersion = strconv.FormatInt(time.Now().Unix(), 36)

type Renderer struct {
	templates *template.Template
}

func NewRenderer(pattern string) (*Renderer, error) {
	return newRenderer(func(t *template.Template) (*template.Template, error) {
		return t.ParseGlob(pattern)
	}, filepath.Base(pattern))
}

func NewRendererFS(fsys fs.FS, pattern string) (*Renderer, error) {
	return newRenderer(func(t *template.Template) (*template.Template, error) {
		return t.ParseFS(fsys, pattern)
	}, pattern)
}

func newRenderer(parse func(*template.Template) (*template.Template, error), name string) (*Renderer, error) {
	funcs := template.FuncMap{
		"year":         func() int { return time.Now().Year() },
		"assetVersion": func() string { return assetVersion },
		"roleLabel":    RoleLabel,
		"stageLabel":   StageLabel,
		"paymentLabel": PaymentLabel,
		"whatsapp": func(message string) string {
			return "https://wa.me/998901234567?text=" + url.QueryEscape(message)
		},
		"emailURL": func(subject, body string) string {
			values := url.Values{}
			values.Set("subject", subject)
			if body != "" {
				values.Set("body", body)
			}
			return "mailto:info@edulicense.uz?" + values.Encode()
		},
		"ptrString": func(value *string) string {
			if value == nil {
				return ""
			}
			return *value
		},
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"inc":   func(i int) int { return i + 1 },
		"trimUZ": func(path string) string {
			if path == "/uz" {
				return "/"
			}
			return strings.TrimPrefix(path, "/uz")
		},
		"initial": func(value string) string {
			if value == "" {
				return ""
			}
			return string([]rune(value)[0])
		},
		"eq": func(a, b any) bool { return fmt.Sprint(a) == fmt.Sprint(b) },
		"selected": func(current, value string) template.HTMLAttr {
			if current == value {
				return `selected`
			}
			return ``
		},
		"checked": func(status map[string]bool, key string) template.HTMLAttr {
			if status != nil && status[key] {
				return `checked`
			}
			return ``
		},
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02 15:04")
		},
		"formatDateInput": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02")
		},
		"dict": func(values ...any) map[string]any {
			out := make(map[string]any, len(values)/2)
			for i := 0; i+1 < len(values); i += 2 {
				out[fmt.Sprint(values[i])] = values[i+1]
			}
			return out
		},
	}

	for key, fn := range FinTemplateFuncs() {
		funcs[key] = fn
	}
	t := template.New(filepath.Base(name)).Funcs(funcs)
	parsed, err := parse(t)
	if err != nil {
		return nil, err
	}
	return &Renderer{templates: parsed}, nil
}

func (r *Renderer) Render(w http.ResponseWriter, status int, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := r.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type PublicPageData struct {
	BaseURL      string
	Path         string
	CanonicalURL string
	Messages     Messages
	Title        string
	Description  string
	Certificate  Certificate
}

type AdminPageData struct {
	BaseURL              string
	Title                string
	CurrentUser          User
	CSRFToken            string
	Error                string
	Success              string
	Users                []User
	Stats                DashboardStats
	Applications         []TestCenterApplication
	Application          TestCenterApplication
	DealsByStage         map[string][]SalesDeal
	Deal                 SalesDeal
	Reminders            []Reminder
	SalesStages          []string
	ApplicationStages    []string
	ProfitSharingOptions []string
	PaymentOptions       []string
	ChecklistItems       []ChecklistItem
	DocumentTypes        []ChecklistItem
	Fin                  *FinPage
}

// Section names the admin area a page belongs to, for the navigation.
func (d AdminPageData) Section() string {
	switch {
	case d.Fin != nil:
		return "finance"
	case d.Title == "Overview":
		return "overview"
	case d.Title == "Sales CRM":
		return "crm"
	default:
		return "applications"
	}
}

// RoleLabel is the role name shown beside the signed-in user.
func RoleLabel(role Role) string {
	switch role {
	case "super_admin":
		return "Super admin"
	case "admin":
		return "Admin"
	case "sales":
		return "Sales"
	}
	return string(role)
}
