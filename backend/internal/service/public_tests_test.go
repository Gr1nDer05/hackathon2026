package service

import (
	"testing"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func TestValidatePublicTestAccessRejectsUnpublished(t *testing.T) {
	err := validatePublicTestAccess(domain.PublicTestAccessInfo{
		IsPublic: true,
		Status:   domain.TestStatusDraft,
	})
	if err != ErrPublicTestNotFound {
		t.Fatalf("expected ErrPublicTestNotFound, got %v", err)
	}
}

func TestValidatePublicTestAccessRejectsLimitReached(t *testing.T) {
	err := validatePublicTestAccess(domain.PublicTestAccessInfo{
		IsPublic:        true,
		Status:          domain.TestStatusPublished,
		MaxParticipants: 2,
		CurrentSessions: 2,
	})
	if err != ErrPublicTestLimitReached {
		t.Fatalf("expected ErrPublicTestLimitReached, got %v", err)
	}
}

func TestNormalizePublicSubmissionRequiresAllRequiredAnswers(t *testing.T) {
	_, err := normalizePublicSubmission(domain.PublicTest{
		Questions: []domain.PublicQuestion{
			{
				ID:           1,
				QuestionType: domain.QuestionTypeText,
				IsRequired:   true,
			},
		},
	}, domain.SubmitPublicTestInput{
		AccessToken: "token",
		Answers:     []domain.PublicAnswerInput{},
	})
	if err != ErrInvalidPublicTestSubmission {
		t.Fatalf("expected ErrInvalidPublicTestSubmission, got %v", err)
	}
}

func TestNormalizePublicAnswerRejectsDuplicateMultipleChoiceValues(t *testing.T) {
	_, err := normalizePublicAnswer(domain.PublicQuestion{
		ID:           1,
		QuestionType: domain.QuestionTypeMultiple,
		Options: []domain.PublicQuestionOption{
			{Value: "a"},
			{Value: "b"},
		},
	}, domain.PublicAnswerInput{
		QuestionID:   1,
		AnswerValues: []string{"a", "a"},
	})
	if err != ErrInvalidPublicTestSubmission {
		t.Fatalf("expected ErrInvalidPublicTestSubmission, got %v", err)
	}
}
