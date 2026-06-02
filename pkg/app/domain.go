package app

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"
)

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleSales      Role = "sales"
)

type User struct {
	ID        string
	Name      string
	Email     string
	Role      Role
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Certificate struct {
	ID                     string
	Slug                   string
	Institution            string
	Designation            string
	SATAdministrationDate  string
	CollegeBoardScreenshot string
	VerificationID         string
	IssueDate              string
	TestCenterCode         string
	CountryRegion          string
}

type SalesDeal struct {
	ID                     string
	SchoolName             string
	Capacity               int
	AssignedSalesAgentID   string
	AssignedSalesAgentName string
	Stage                  string
	NegotiationPrice       *string
	Currency               string
	ProfitSharing          string
	Notes                  string
	ConvertedApplicationID string
	CreatedBy              string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type TestCenterApplication struct {
	ID                string
	InstitutionName   string
	InstitutionType   string
	Location          string
	Website           string
	ResponsiblePerson string
	Phone             string
	DirectorPhone     string
	AssignedAdminID   string
	AssignedAdminName string
	SalesReferralID   string
	SalesReferralName string
	PaymentOption     string
	PaymentAmount     *string
	PaymentCurrency   string
	ProfitSharing     string
	CEEBCode          string
	Stage             string
	ChecklistStatus   map[string]bool
	CreatedBy         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CertificateID     string
	CertificateSlug   string
	VerificationID    string
	Documents         []ApplicationDocument
	Reminders         []Reminder
}

type ApplicationDocument struct {
	ID               string
	ApplicationID    string
	DocType          string
	OriginalFilename string
	ContentType      string
	SizeBytes        int64
	StorageKey       string
	UploadedBy       string
	CreatedAt        time.Time
}

type Reminder struct {
	ID           string
	EntityType   string
	EntityID     string
	EntityName   string
	OwnerName    string
	CurrentStage string
	Title        string
	Note         string
	DueAt        time.Time
	CreatedBy    string
	SentAt       *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ActivityLog struct {
	ID         string
	UserID     string
	UserName   string
	Action     string
	EntityType string
	EntityID   string
	Summary    string
	CreatedAt  time.Time
}

type ApplicationInput struct {
	InstitutionName   string
	InstitutionType   string
	Location          string
	Website           string
	ResponsiblePerson string
	Phone             string
	DirectorPhone     string
	AssignedAdminID   string
	SalesReferralID   string
	PaymentOption     string
	PaymentAmount     string
	PaymentCurrency   string
	ProfitSharing     string
	CEEBCode          string
	Stage             string
	ChecklistStatus   map[string]bool
}

type DealInput struct {
	SchoolName           string
	Capacity             int
	AssignedSalesAgentID string
	Stage                string
	NegotiationPrice     string
	Currency             string
	ProfitSharing        string
	Notes                string
}

type ReminderInput struct {
	EntityType string
	EntityID   string
	Title      string
	Note       string
	DueAt      time.Time
}

type DocumentInput struct {
	ApplicationID    string
	DocType          string
	OriginalFilename string
	ContentType      string
	SizeBytes        int64
	StorageKey       string
	UploadedBy       string
}

type CertificateInput struct {
	Slug                   string
	Institution            string
	Designation            string
	SATAdministrationDate  string
	CollegeBoardScreenshot string
	VerificationID         string
	IssueDate              string
	TestCenterCode         string
	CountryRegion          string
}

type DashboardStats struct {
	ApplicationCounts map[string]int
	DealCounts        map[string]int
	OverdueReminders  int
	ActiveUsers       int
	WonDeals          int
	ConvertedDeals    int
	RecentActivity    []ActivityLog
	DueReminders      []Reminder
}

var SalesStages = []string{
	"new_lead",
	"contacted",
	"meeting_scheduled",
	"proposal_sent",
	"negotiating",
	"won",
	"lost",
}

var ApplicationStages = []string{
	"start",
	"document_gather",
	"applied_to_ceeb",
	"got_ceeb",
	"applied_to_test_center",
	"got_test_center",
	"in_test_center_list",
}

var ProfitSharingOptions = []string{
	"80-20",
	"70-30",
	"60-40",
	"50-50",
	"40-60",
	"30-70",
}

var PaymentOptions = []string{
	"investor",
	"our_investment",
	"fixed_amount",
}

var ChecklistItems = []ChecklistItem{
	{Key: "license_english", Label: "Litsenziya Ingliz tilida"},
	{Key: "domain_email", Label: "Domenlik email"},
	{Key: "website_english", Label: "Website ingliz tilida"},
	{Key: "responsible_staff", Label: "1 ta ma'sul xodimning ism familiyasi"},
	{Key: "website_director", Label: "Founder/director name with position on website"},
	{Key: "website_timetable", Label: "Dars jadvali on website"},
	{Key: "website_address_matches_license", Label: "Website address matches license"},
	{Key: "website_phone", Label: "Phone number on website"},
	{Key: "website_email", Label: "Email address on website"},
}

var DocumentTypes = []ChecklistItem{
	{Key: "license_english", Label: "Litsenziya Ingliz tilida"},
	{Key: "domain_email", Label: "Domenlik email"},
	{Key: "website_english", Label: "Website ingliz tilida"},
	{Key: "responsible_staff", Label: "1 ta ma'sul xodimning ism familiyasi"},
}

type ChecklistItem struct {
	Key   string
	Label string
}

func RoleCanAccess(role Role, area string) bool {
	if role == RoleSuperAdmin {
		return true
	}
	switch area {
	case "applications":
		return role == RoleAdmin
	case "crm":
		return role == RoleSales
	case "reminders":
		return role == RoleAdmin
	default:
		return false
	}
}

func ValidSalesStage(stage string) bool {
	return slices.Contains(SalesStages, stage)
}

func ValidApplicationStage(stage string) bool {
	return slices.Contains(ApplicationStages, stage)
}

func ChecklistComplete(status map[string]bool) bool {
	for _, item := range ChecklistItems {
		if !status[item.Key] {
			return false
		}
	}
	return true
}

func StageLabel(stage string) string {
	labels := map[string]string{
		"new_lead":               "New lead",
		"contacted":              "Contacted",
		"meeting_scheduled":      "Meeting scheduled",
		"proposal_sent":          "Proposal sent",
		"negotiating":            "Negotiating",
		"won":                    "Won",
		"lost":                   "Lost",
		"start":                  "Start",
		"document_gather":        "Document gather",
		"applied_to_ceeb":        "Applied to CEEB",
		"got_ceeb":               "Got CEEB",
		"applied_to_test_center": "Applied to test center",
		"got_test_center":        "Got test center",
		"in_test_center_list":    "In the test center list",
	}
	if label, ok := labels[stage]; ok {
		return label
	}
	return stage
}

func PaymentLabel(value string) string {
	switch value {
	case "investor":
		return "Investor"
	case "our_investment":
		return "Our own investment"
	case "fixed_amount":
		return "Fixed amount"
	default:
		return value
	}
}

func CertificateInputForApplication(application TestCenterApplication, issuedAt time.Time) CertificateInput {
	code := strings.TrimSpace(application.CEEBCode)
	year := issuedAt.Format("2006")
	shortID := shortIdentifier(application.ID)
	verificationID := fmt.Sprintf("EL-SAT-%s-%s", code, year)
	if code == "" {
		verificationID = fmt.Sprintf("EL-SAT-%s-%s", strings.ToUpper(shortID), year)
	}
	return CertificateInput{
		Slug:                   certificateSlug(application.InstitutionName, code, shortID),
		Institution:            application.InstitutionName,
		Designation:            "SAT Test Center Listing Verification",
		SATAdministrationDate:  "upcoming SAT administration period",
		CollegeBoardScreenshot: "",
		VerificationID:         verificationID,
		IssueDate:              issuedAt.Format("02/01/2006"),
		TestCenterCode:         code,
		CountryRegion:          "Uzbekistan",
	}
}

func TestCenterCodeFromVerificationID(verificationID string) string {
	parts := strings.Split(verificationID, "-")
	if len(parts) >= 3 && parts[0] == "EL" && parts[1] == "SAT" {
		return parts[2]
	}
	return ""
}

func certificateSlug(institution, code, shortID string) string {
	parts := []string{slugify(institution), "sat-center"}
	if code != "" {
		parts = append(parts, strings.ToLower(code))
	}
	if shortID != "" {
		parts = append(parts, strings.ToLower(shortID))
	}
	return strings.Join(compactStrings(parts), "-")
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		case unicode.IsSpace(r) || r == '-' || r == '_' || r == '/':
			if builder.Len() > 0 && !lastDash {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(builder.String(), "-")
	if out == "" {
		return "certificate"
	}
	return out
}

func shortIdentifier(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	if idx := strings.Index(id, "-"); idx > 0 {
		return id[:idx]
	}
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func compactStrings(values []string) []string {
	out := values[:0]
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return out
}

func ValidateProfitSharing(value string) error {
	if value == "" {
		return nil
	}
	if slices.Contains(ProfitSharingOptions, value) {
		return nil
	}
	return fmt.Errorf("invalid profit sharing option")
}
