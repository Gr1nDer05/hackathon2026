package service

import (
	"testing"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func TestNormalizePublicRespondentInputRequiresFullNameAndPhone(t *testing.T) {
	_, err := normalizePublicRespondentInput(domain.PublicTest{}, domain.StartPublicTestInput{
		RespondentName:  "Иван",
		RespondentPhone: "+79991234567",
	})
	if err != ErrInvalidPublicTestRespondent {
		t.Fatalf("expected ErrInvalidPublicTestRespondent, got %v", err)
	}
}

func TestNormalizePublicRespondentInputNormalizesPhone(t *testing.T) {
	input, err := normalizePublicRespondentInput(domain.PublicTest{}, domain.StartPublicTestInput{
		RespondentName:  "Иванов Иван",
		RespondentPhone: "8 (999) 123-45-67",
	})
	if err != nil {
		t.Fatalf("normalizePublicRespondentInput returned error: %v", err)
	}
	if input.RespondentPhone != "+79991234567" {
		t.Fatalf("expected normalized phone, got %q", input.RespondentPhone)
	}
}

func TestNormalizePublicRespondentInputRequiresEnabledPersonalFields(t *testing.T) {
	_, err := normalizePublicRespondentInput(domain.PublicTest{
		CollectRespondentAge:       true,
		CollectRespondentGender:    true,
		CollectRespondentEducation: true,
	}, domain.StartPublicTestInput{
		RespondentName:  "Иванов Иван",
		RespondentPhone: "+79991234567",
	})
	if err != ErrInvalidPublicTestRespondent {
		t.Fatalf("expected ErrInvalidPublicTestRespondent, got %v", err)
	}
}

func TestNormalizePublicRespondentInputClearsDisabledPersonalFields(t *testing.T) {
	input, err := normalizePublicRespondentInput(domain.PublicTest{}, domain.StartPublicTestInput{
		RespondentName:      "Иванов Иван",
		RespondentPhone:     "+79991234567",
		RespondentAge:       34,
		RespondentGender:    "male",
		RespondentEducation: "higher",
	})
	if err != nil {
		t.Fatalf("normalizePublicRespondentInput returned error: %v", err)
	}
	if input.RespondentAge != 0 || input.RespondentGender != "" || input.RespondentEducation != "" {
		t.Fatalf("expected disabled personal fields to be cleared, got %+v", input)
	}
}

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

func TestValidatePublicTestAccessAllowsUnlimitedParticipantLinks(t *testing.T) {
	err := validatePublicTestAccess(domain.PublicTestAccessInfo{
		IsPublic:        true,
		Status:          domain.TestStatusPublished,
		MaxParticipants: 0,
		CurrentSessions: 999,
	})
	if err != nil {
		t.Fatalf("expected unlimited public link to stay available, got %v", err)
	}
}

func TestAttachPublicReportAccessToSubmitResponseIncludesURLWhenEnabled(t *testing.T) {
	t.Setenv("PUBLIC_BASE_URL", "https://example.com")

	response := domain.SubmitPublicTestResponse{Status: "completed"}
	attachPublicReportAccessToSubmitResponse(domain.PublicTest{
		Slug:                        "demo-slug",
		ShowClientReportImmediately: true,
	}, "access-token", &response)

	if !response.ClientReportAvailable {
		t.Fatalf("expected client report to be available")
	}
	expectedURL := "https://example.com/public/tests/demo-slug/report?access_token=access-token"
	if response.ClientReportURL != expectedURL {
		t.Fatalf("expected client report url %q, got %q", expectedURL, response.ClientReportURL)
	}
}

func TestAttachPublicReportAccessToSubmitResponseSkipsDisabledReports(t *testing.T) {
	response := domain.SubmitPublicTestResponse{Status: "completed"}
	attachPublicReportAccessToSubmitResponse(domain.PublicTest{
		Slug:                        "demo-slug",
		ShowClientReportImmediately: false,
	}, "access-token", &response)

	if response.ClientReportAvailable {
		t.Fatalf("expected client report to stay unavailable")
	}
	if response.ClientReportURL != "" {
		t.Fatalf("expected empty client report url, got %q", response.ClientReportURL)
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

func TestMergePublicAnswersOverridesExistingAnswer(t *testing.T) {
	merged := mergePublicAnswers(
		[]domain.PublicAnswerInput{
			{QuestionID: 1, AnswerValue: "old"},
			{QuestionID: 2, AnswerValue: "keep"},
		},
		[]domain.PublicAnswerInput{
			{QuestionID: 1, AnswerValue: "new"},
		},
	)

	answersByQuestionID := make(map[int64]string, len(merged))
	for _, answer := range merged {
		answersByQuestionID[answer.QuestionID] = answer.AnswerValue
	}

	if answersByQuestionID[1] != "new" || answersByQuestionID[2] != "keep" {
		t.Fatalf("unexpected merged answers: %+v", merged)
	}
}

func TestValidateRequiredPublicAnswersAcceptsMergedAnswers(t *testing.T) {
	err := validateRequiredPublicAnswers(domain.PublicTest{
		Questions: []domain.PublicQuestion{
			{ID: 1, QuestionType: domain.QuestionTypeText, IsRequired: true},
			{ID: 2, QuestionType: domain.QuestionTypeText, IsRequired: true},
		},
	}, []domain.PublicAnswerInput{
		{QuestionID: 1, AnswerText: "one"},
		{QuestionID: 2, AnswerText: "two"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
