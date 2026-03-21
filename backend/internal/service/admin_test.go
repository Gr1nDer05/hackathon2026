package service

import "testing"

func TestIsAdminEmailBoundRejectsPlaceholderEmail(t *testing.T) {
	if IsAdminEmailBound("admin@admin.local") {
		t.Fatalf("expected placeholder admin email to be treated as unbound")
	}
}

func TestIsAdminEmailBoundAcceptsRealEmail(t *testing.T) {
	if !IsAdminEmailBound("admin@example.com") {
		t.Fatalf("expected real admin email to be treated as bound")
	}
}
