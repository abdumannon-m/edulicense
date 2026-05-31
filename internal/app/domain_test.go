package app

import "testing"

func TestRoleCanAccess(t *testing.T) {
	tests := []struct {
		role Role
		area string
		want bool
	}{
		{RoleSuperAdmin, "applications", true},
		{RoleSuperAdmin, "crm", true},
		{RoleAdmin, "applications", true},
		{RoleAdmin, "crm", false},
		{RoleSales, "crm", true},
		{RoleSales, "applications", false},
	}
	for _, tt := range tests {
		if got := RoleCanAccess(tt.role, tt.area); got != tt.want {
			t.Fatalf("RoleCanAccess(%q, %q) = %v, want %v", tt.role, tt.area, got, tt.want)
		}
	}
}

func TestStageValidation(t *testing.T) {
	for _, stage := range SalesStages {
		if !ValidSalesStage(stage) {
			t.Fatalf("sales stage %q should be valid", stage)
		}
	}
	if ValidSalesStage("signed_contract") {
		t.Fatal("unexpected sales stage should be invalid")
	}
	for _, stage := range ApplicationStages {
		if !ValidApplicationStage(stage) {
			t.Fatalf("application stage %q should be valid", stage)
		}
	}
	if ValidApplicationStage("done") {
		t.Fatal("unexpected application stage should be invalid")
	}
}

func TestChecklistComplete(t *testing.T) {
	status := map[string]bool{}
	if ChecklistComplete(status) {
		t.Fatal("empty checklist should not be complete")
	}
	for _, item := range ChecklistItems {
		status[item.Key] = true
	}
	if !ChecklistComplete(status) {
		t.Fatal("all checklist items marked true should be complete")
	}
	status[ChecklistItems[0].Key] = false
	if ChecklistComplete(status) {
		t.Fatal("missing one checklist item should be incomplete")
	}
}

func TestValidateProfitSharing(t *testing.T) {
	if err := ValidateProfitSharing(""); err != nil {
		t.Fatalf("empty profit sharing should be accepted: %v", err)
	}
	if err := ValidateProfitSharing("50-50"); err != nil {
		t.Fatalf("known profit sharing should be accepted: %v", err)
	}
	if err := ValidateProfitSharing("90-10"); err == nil {
		t.Fatal("unknown profit sharing should fail")
	}
}
