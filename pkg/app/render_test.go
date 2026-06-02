package app

import (
	"net/http/httptest"
	"strings"
	"testing"

	apptemplates "edu-license/pkg/templates"
)

func TestTemplatesParse(t *testing.T) {
	if _, err := NewRenderer("../templates/*.html"); err != nil {
		t.Fatalf("templates should parse: %v", err)
	}
}

func TestEmbeddedTemplatesParse(t *testing.T) {
	if _, err := NewRendererFS(apptemplates.FS, "*.html"); err != nil {
		t.Fatalf("embedded templates should parse: %v", err)
	}
}

func TestAdminCRMTemplateRendersDragBoard(t *testing.T) {
	renderer, err := NewRenderer("../templates/*.html")
	if err != nil {
		t.Fatalf("templates should parse: %v", err)
	}
	data := AdminPageData{
		Title:     "Sales CRM",
		CSRFToken: "csrf-token",
		CurrentUser: User{
			Name: "Admin",
			Role: RoleSuperAdmin,
		},
		SalesStages: []string{"new_lead", "contacted"},
		ProfitSharingOptions: []string{
			"70-30",
		},
		DealsByStage: map[string][]SalesDeal{
			"new_lead": {
				{
					ID:                     "deal-1",
					SchoolName:             "Example School",
					Capacity:               120,
					AssignedSalesAgentName: "Sales Agent",
					Stage:                  "new_lead",
					Currency:               "USD",
					ProfitSharing:          "70-30",
				},
			},
			"contacted": {},
		},
	}

	recorder := httptest.NewRecorder()
	renderer.Render(recorder, 200, "admin_crm", data)

	if recorder.Code != 200 {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		`data-crm-board`,
		`data-stage-column data-stage="new_lead"`,
		`data-deal-card data-deal-id="deal-1"`,
		`data-stage-url="/admin/crm/deals/deal-1/stage"`,
		`draggable="true"`,
		`crm-board--dnd-ready`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("rendered CRM template missing %q", want)
		}
	}
}
