package service

import (
	"testing"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

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

func TestIsAdminEmailVerifiedRequiresTimestamp(t *testing.T) {
	admin := domain.User{
		Email:           "admin@example.com",
		EmailVerifiedAt: domain.NewNullableString("2026-03-21T10:00:00Z"),
	}

	if !IsAdminEmailVerified(admin) {
		t.Fatalf("expected verified admin email to be treated as verified")
	}

	admin.EmailVerifiedAt = domain.NewNullableString("")
	if IsAdminEmailVerified(admin) {
		t.Fatalf("expected missing verification timestamp to be treated as unverified")
	}
}
