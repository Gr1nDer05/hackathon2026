package service

import (
	"testing"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func TestNormalizeQuestionScaleWeightsNormalizesAllowedScales(t *testing.T) {
	scaleWeights, err := normalizeQuestionScaleWeights(map[string]float64{
		" Analytic ": 1,
		"creative":   0.5,
		"social":     0,
		"organizer":  2,
		"practical":  0.25,
	})
	if err != nil {
		t.Fatalf("normalizeQuestionScaleWeights returned error: %v", err)
	}

	if len(scaleWeights) != 4 {
		t.Fatalf("expected zero-weight scale to be dropped, got %+v", scaleWeights)
	}
	if scaleWeights[domain.CareerScaleAnalytic] != 1 {
		t.Fatalf("expected analytic weight to be normalized, got %+v", scaleWeights)
	}
}

func TestNormalizeQuestionScaleWeightsRejectsUnknownScale(t *testing.T) {
	if _, err := normalizeQuestionScaleWeights(map[string]float64{"leadership": 1}); err != ErrInvalidQuestionInput {
		t.Fatalf("expected ErrInvalidQuestionInput, got %v", err)
	}
}

func TestCalculateCareerResultUsesOptionScoreFallbackAndRanksResults(t *testing.T) {
	result := calculateCareerResultForPublicTest(domain.PublicTest{
		Questions: []domain.PublicQuestion{
			{
				ID:           1,
				QuestionType: domain.QuestionTypeSingleChoice,
				ScaleWeights: map[string]float64{
					domain.CareerScaleAnalytic: 1,
				},
				Options: []domain.PublicQuestionOption{
					{Value: "low", Score: 1},
					{Value: "high", Score: 5},
				},
			},
			{
				ID:           2,
				QuestionType: domain.QuestionTypeSingleChoice,
				ScaleWeights: map[string]float64{
					domain.CareerScalePractical: 1,
				},
				Options: []domain.PublicQuestionOption{
					{Value: "base", Score: 1},
					{Value: "strong", Score: 5},
				},
			},
		},
	}, []domain.PublicTestAnswer{
		{QuestionID: 1, AnswerValue: "high"},
		{QuestionID: 2, AnswerValue: "strong"},
	})

	if result == nil {
		t.Fatalf("expected career result, got nil")
	}
	if len(result.TopScales) != 2 {
		t.Fatalf("expected top 2 scales, got %+v", result.TopScales)
	}
	if result.TopScales[0].Scale != domain.CareerScaleAnalytic || result.TopScales[1].Scale != domain.CareerScalePractical {
		t.Fatalf("unexpected top scales: %+v", result.TopScales)
	}
	if result.TopScales[0].Percentage != 100 || result.TopScales[1].Percentage != 100 {
		t.Fatalf("expected 100%% scale percentages, got %+v", result.TopScales)
	}
	if len(result.TopProfessions) != 3 {
		t.Fatalf("expected top 3 professions, got %+v", result.TopProfessions)
	}

	expectedProfessions := []string{
		"DevOps Engineer",
		"Site Reliability Engineer",
		"Embedded Engineer",
	}
	for idx, profession := range expectedProfessions {
		if result.TopProfessions[idx].Profession != profession {
			t.Fatalf("expected profession %q at index %d, got %+v", profession, idx, result.TopProfessions)
		}
	}
}

func TestCalculateCareerResultReturnsNilWithoutWeightedQuestions(t *testing.T) {
	result := calculateCareerResultForQuestions([]domain.Question{
		{ID: 1, QuestionType: domain.QuestionTypeText},
	}, []domain.PublicTestAnswer{
		{QuestionID: 1, AnswerText: "hello"},
	})

	if result != nil {
		t.Fatalf("expected nil result for unweighted questions, got %+v", result)
	}
}
