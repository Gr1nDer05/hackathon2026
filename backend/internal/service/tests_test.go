package service

import (
	"testing"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func boolPtr(v bool) *bool {
	return &v
}

func TestNormalizeCreateTestInputDefaultsToPublished(t *testing.T) {
	input, err := normalizeTestCreateInput(domain.CreateTestInput{
		Title: "Public Test",
	})
	if err != nil {
		t.Fatalf("normalizeTestCreateInput returned error: %v", err)
	}

	if input.Status != domain.TestStatusPublished {
		t.Fatalf("expected default status %q, got %q", domain.TestStatusPublished, input.Status)
	}
}

func TestNormalizeCreateTestInputRejectsNegativeMaxParticipants(t *testing.T) {
	_, err := normalizeTestCreateInput(domain.CreateTestInput{
		Title:           "Invalid Test",
		MaxParticipants: -1,
	})
	if err != ErrInvalidTestInput {
		t.Fatalf("expected ErrInvalidTestInput, got %v", err)
	}
}

func TestNormalizeCreateTestInputTreatsZeroMaxParticipantsAsUnlimitedForLegacyClients(t *testing.T) {
	input, err := normalizeTestCreateInput(domain.CreateTestInput{
		Title:           "Unlimited Test",
		MaxParticipants: 0,
	})
	if err != nil {
		t.Fatalf("normalizeTestCreateInput returned error: %v", err)
	}
	if input.MaxParticipants != 0 {
		t.Fatalf("expected max_participants to stay 0, got %d", input.MaxParticipants)
	}
}

func TestNormalizeCreateTestInputRequiresPositiveMaxParticipantsWhenLimitEnabled(t *testing.T) {
	_, err := normalizeTestCreateInput(domain.CreateTestInput{
		Title:               "Limited Test",
		MaxParticipants:     0,
		HasParticipantLimit: boolPtr(true),
	})
	if err != ErrInvalidTestInput {
		t.Fatalf("expected ErrInvalidTestInput, got %v", err)
	}
}

func TestNormalizeCreateTestInputDropsMaxParticipantsWhenLimitDisabled(t *testing.T) {
	input, err := normalizeTestCreateInput(domain.CreateTestInput{
		Title:               "Public Test",
		MaxParticipants:     15,
		HasParticipantLimit: boolPtr(false),
	})
	if err != nil {
		t.Fatalf("normalizeTestCreateInput returned error: %v", err)
	}
	if input.MaxParticipants != 0 {
		t.Fatalf("expected max_participants to be reset to 0, got %d", input.MaxParticipants)
	}
}

func TestApplyPublicURLsToTestsPreservesCompletedSessionsCount(t *testing.T) {
	t.Setenv("PUBLIC_BASE_URL", "https://example.com")

	tests := []domain.Test{
		{
			ID:                     42,
			Title:                  "Stress Test",
			PublicSlug:             "abc123",
			MaxParticipants:        7,
			CompletedSessionsCount: 7,
		},
	}

	applyPublicURLsToTests(tests)

	if tests[0].PublicURL != "https://example.com/public/tests/abc123" {
		t.Fatalf("expected public url to be set, got %q", tests[0].PublicURL)
	}
	if tests[0].CompletedSessionsCount != 7 {
		t.Fatalf("expected completed sessions count to stay unchanged, got %d", tests[0].CompletedSessionsCount)
	}
	if !tests[0].HasParticipantLimit {
		t.Fatalf("expected has_participant_limit to be true")
	}
}
