package service

import (
	"context"
	"testing"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func TestNormalizePsychologistAccessInputParsesDateOnlyAsEndOfDay(t *testing.T) {
	input := domain.UpdatePsychologistAccessInput{
		PortalAccessUntil: domain.OptionalStringInput{
			Set:   true,
			Value: ptrString("2026-03-21"),
		},
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

func TestNormalizePsychologistAccessInputTreatsNullAsExplicitClear(t *testing.T) {
	input := domain.UpdatePsychologistAccessInput{
		BlockedUntil: domain.OptionalStringInput{
			Set: true,
		},
	}

	update, err := normalizePsychologistAccessInput(input)
	if err != nil {
		t.Fatalf("normalizePsychologistAccessInput returned error: %v", err)
	}

	if !update.BlockedUntilSet {
		t.Fatalf("expected blocked_until to be marked as explicitly set")
	}
	if update.BlockedUntil != nil {
		t.Fatalf("expected blocked_until to be cleared, got %v", update.BlockedUntil)
	}
}

func TestNormalizePsychologistAccessInputAcceptsSubscriptionDays(t *testing.T) {
	days := 30
	input := domain.UpdatePsychologistAccessInput{
		SubscriptionDays: &days,
	}

	update, err := normalizePsychologistAccessInput(input)
	if err != nil {
		t.Fatalf("normalizePsychologistAccessInput returned error: %v", err)
	}

	if !update.SubscriptionDaysSet || update.SubscriptionDays != 30 {
		t.Fatalf("expected subscription days to be set, got %+v", update)
	}
}

func TestNormalizePsychologistAccessInputAcceptsSubscriptionDaysAlias(t *testing.T) {
	days := 14
	input := domain.UpdatePsychologistAccessInput{
		SubscriptionDaysAlias: &days,
	}

	update, err := normalizePsychologistAccessInput(input)
	if err != nil {
		t.Fatalf("normalizePsychologistAccessInput returned error: %v", err)
	}

	if !update.SubscriptionDaysSet || update.SubscriptionDays != 14 {
		t.Fatalf("expected subscription days alias to be set, got %+v", update)
	}
}

func TestNormalizePsychologistAccessInputAcceptsSubscriptionPlan(t *testing.T) {
	input := domain.UpdatePsychologistAccessInput{
		SubscriptionPlan: "pro",
	}

	update, err := normalizePsychologistAccessInput(input)
	if err != nil {
		t.Fatalf("normalizePsychologistAccessInput returned error: %v", err)
	}

	if !update.SubscriptionPlanSet || update.SubscriptionPlan != domain.SubscriptionPlanPro {
		t.Fatalf("expected subscription plan to be set, got %+v", update)
	}
}

func TestNormalizePsychologistAccessInputRejectsUnknownSubscriptionPlan(t *testing.T) {
	input := domain.UpdatePsychologistAccessInput{
		SubscriptionPlan: "enterprise",
	}

	if _, err := normalizePsychologistAccessInput(input); err != ErrInvalidPsychologistAccess {
		t.Fatalf("expected ErrInvalidPsychologistAccess, got %v", err)
	}
}

func TestNormalizePsychologistAccessInputRejectsMixedDateAndSubscriptionDays(t *testing.T) {
	days := 7
	input := domain.UpdatePsychologistAccessInput{
		PortalAccessUntil: domain.OptionalStringInput{
			Set:   true,
			Value: ptrString("2026-03-21"),
		},
		SubscriptionDays: &days,
	}

	if _, err := normalizePsychologistAccessInput(input); err != ErrInvalidPsychologistAccess {
		t.Fatalf("expected ErrInvalidPsychologistAccess, got %v", err)
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

func TestPsychologistSessionExpiresAtUsesSubscriptionDeadline(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)

	expiresAt := psychologistSessionExpiresAt(now, "2026-03-20T18:00:00Z")

	expected := time.Date(2026, 3, 20, 18, 0, 0, 0, time.UTC)
	if !expiresAt.Equal(expected) {
		t.Fatalf("expected session to expire at %s, got %s", expected.Format(time.RFC3339), expiresAt.Format(time.RFC3339))
	}
}

func TestPsychologistSessionExpiresAtFallsBackToDefaultTTL(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)

	expiresAt := psychologistSessionExpiresAt(now, "")
	expected := now.Add(SessionTTL)
	if !expiresAt.Equal(expected) {
		t.Fatalf("expected default expiry %s, got %s", expected.Format(time.RFC3339), expiresAt.Format(time.RFC3339))
	}
}

func TestPsychologistSessionExpiresAtIgnoresExpiredSubscriptionDeadline(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)

	expiresAt := psychologistSessionExpiresAt(now, "2026-03-20T11:59:59Z")
	expected := now.Add(SessionTTL)
	if !expiresAt.Equal(expected) {
		t.Fatalf("expected default expiry for expired subscription %s, got %s", expected.Format(time.RFC3339), expiresAt.Format(time.RFC3339))
	}
}

func TestCalculateSubscriptionExtensionDaysUsesExplicitDays(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	update := domain.PsychologistAccessUpdate{
		SubscriptionDaysSet: true,
		SubscriptionDays:    30,
	}

	days, ok := calculateSubscriptionExtensionDays(
		domain.NewNullableString("2026-03-25T12:00:00Z"),
		domain.NewNullableString("2026-04-24T12:00:00Z"),
		update,
		now,
	)
	if !ok || days != 30 {
		t.Fatalf("expected 30-day extension, got days=%d ok=%v", days, ok)
	}
}

func TestCalculateSubscriptionExtensionDaysUsesCurrentTimeWhenSubscriptionExpired(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	update := domain.PsychologistAccessUpdate{
		PortalAccessUntilSet: true,
	}

	days, ok := calculateSubscriptionExtensionDays(
		domain.NewNullableString("2026-03-10T12:00:00Z"),
		domain.NewNullableString("2026-03-23T08:00:00Z"),
		update,
		now,
	)
	if !ok || days != 3 {
		t.Fatalf("expected rounded 3-day extension, got days=%d ok=%v", days, ok)
	}
}

func TestCreateSubscriptionPurchaseRequestRejectsUnknownPlan(t *testing.T) {
	service := &AppService{}

	_, err := service.CreateSubscriptionPurchaseRequest(context.Background(), domain.AuthenticatedUser{
		ID:   7,
		Role: domain.RolePsychologist,
	}, domain.CreateSubscriptionPurchaseRequestInput{
		SubscriptionPlan: "enterprise",
	})
	if err != ErrInvalidSubscriptionPurchaseRequest {
		t.Fatalf("expected ErrInvalidSubscriptionPurchaseRequest, got %v", err)
	}
}

func TestCreateSubscriptionPurchaseRequestRejectsNonPsychologist(t *testing.T) {
	service := &AppService{}

	_, err := service.CreateSubscriptionPurchaseRequest(context.Background(), domain.AuthenticatedUser{
		ID:   1,
		Role: domain.RoleAdmin,
	}, domain.CreateSubscriptionPurchaseRequestInput{
		SubscriptionPlan: "basic",
	})
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func ptrString(value string) *string {
	return &value
}
