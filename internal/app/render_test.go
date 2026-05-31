package app

import (
	"testing"

	apptemplates "edu-license/internal/templates"
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
