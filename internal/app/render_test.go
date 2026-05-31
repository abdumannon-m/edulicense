package app

import "testing"

func TestTemplatesParse(t *testing.T) {
	if _, err := NewRenderer("../templates/*.html"); err != nil {
		t.Fatalf("templates should parse: %v", err)
	}
}
