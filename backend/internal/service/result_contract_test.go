package service

import (
	"testing"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func TestBuildMetricsFromCareerResultAndFormulaRules(t *testing.T) {
	metrics := buildMetricsFromCareerResult(&domain.CareerResult{
		Scales: []domain.CareerScaleResult{
			{Scale: domain.CareerScaleAnalytic, RawScore: 12},
			{Scale: domain.CareerScaleCreative, RawScore: 4},
		},
	})

	applyFormulaRuleMetrics(metrics, []domain.FormulaRule{
		{
			ConditionType: domain.FormulaConditionAlways,
			ResultKey:     "stress_resistance",
			ScoreDelta:    7,
		},
		{
			ConditionType: domain.FormulaConditionAlways,
			ResultKey:     "analytic",
			ScoreDelta:    3,
		},
	}, []domain.PublicTestAnswer{})

	if metrics[domain.CareerScaleAnalytic] != 15 {
		t.Fatalf("expected analytic metric to include formula score, got %#v", metrics)
	}
	if metrics["stress_resistance"] != 7 {
		t.Fatalf("expected custom metric to be added, got %#v", metrics)
	}
}

func TestBuildTopMetricsSkipsTotalWhenNamedMetricsExist(t *testing.T) {
	topMetrics := buildTopMetrics(map[string]float64{
		"total":             25,
		"stress_resistance": 12,
		"leadership":        7,
		"creativity":        4,
		"communication":     0,
	})

	if len(topMetrics) != 3 {
		t.Fatalf("expected 3 top metrics, got %+v", topMetrics)
	}
	if topMetrics[0].Key != "stress_resistance" || topMetrics[0].Value != 12 {
		t.Fatalf("expected strongest metric first, got %+v", topMetrics)
	}
	if topMetrics[1].Key != "leadership" || topMetrics[2].Key != "creativity" {
		t.Fatalf("unexpected metric order: %+v", topMetrics)
	}
}

func TestCalculateTotalScoreForPublicTestIncludesAnswersAndTriggeredRules(t *testing.T) {
	total := calculateTotalScoreForPublicTest(domain.PublicTest{
		Questions: []domain.PublicQuestion{
			{
				ID:           1,
				QuestionType: domain.QuestionTypeSingleChoice,
				Options: []domain.PublicQuestionOption{
					{Value: "high", Score: 5},
				},
			},
			{
				ID:           2,
				QuestionType: domain.QuestionTypeMultiple,
				Options: []domain.PublicQuestionOption{
					{Value: "a", Score: 2},
					{Value: "b", Score: 3},
				},
			},
		},
	}, []domain.PublicTestAnswer{
		{QuestionID: 1, AnswerValue: "high"},
		{QuestionID: 2, AnswerValues: []string{"a", "b"}},
	}, []domain.FormulaRule{
		{
			QuestionID:    1,
			ConditionType: domain.FormulaConditionAnswerEquals,
			ExpectedValue: "high",
			ResultKey:     "total",
			ScoreDelta:    4,
		},
	})

	if total != 14 {
		t.Fatalf("expected total score 14, got %v", total)
	}
}
