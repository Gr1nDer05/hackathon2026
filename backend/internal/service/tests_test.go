package service

import (
	"testing"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

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
