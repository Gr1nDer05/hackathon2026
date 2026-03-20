package service

import (
	"testing"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func TestNormalizePsychologistAccessInputParsesDateOnlyAsEndOfDay(t *testing.T) {
	input := domain.UpdatePsychologistAccessInput{
		PortalAccessUntil: ptrString("2026-03-21"),
	}

	update, err := normalizePsychologistAccessInput(input)
	if err != nil {
		t.Fatalf("normalizePsychologistAccessInput returned error: %v", err)
	}

	if !update.PortalAccessUntilSet || update.PortalAccessUntil == nil {
		t.Fatalf("expected portal access date to be set")
	}
	if update.PortalAccessUntil.Hour() != 23 || update.PortalAccessUntil.Minute() != 59 || update.PortalAccessUntil.Second() != 59 {
		t.Fatalf("expected end of day timestamp, got %s", update.PortalAccessUntil.Format(time.RFC3339))
	}
}

func TestPsychologistAccessErrorRejectsExpiredSubscription(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	err := psychologistAccessError(true, "2026-03-20T11:59:59Z", "", now)
	if err != ErrPortalAccessExpired {
		t.Fatalf("expected ErrPortalAccessExpired, got %v", err)
	}
}

func TestPsychologistAccessErrorRejectsTemporaryBlock(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	err := psychologistAccessError(true, "", "2026-03-21T12:00:00Z", now)
	if err != ErrAccountTemporarilyBlocked {
		t.Fatalf("expected ErrAccountTemporarilyBlocked, got %v", err)
	}
}

func ptrString(value string) *string {
	return &value
}
