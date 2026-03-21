package service

import (
	"testing"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func TestNormalizeUpdateQuestionInputPreservesOmittedScaleWeights(t *testing.T) {
	input := domain.UpdateQuestionInput{
		Text:         "Question",
		QuestionType: domain.QuestionTypeSingleChoice,
		Options: []domain.QuestionOptionInput{
			{Label: "One", Value: "1"},
		},
	}

	normalized, err := normalizeUpdateQuestionInput(input)
	if err != nil {
		t.Fatalf("normalizeUpdateQuestionInput returned error: %v", err)
	}
	if normalized.ScaleWeights != nil {
		t.Fatalf("expected omitted scale_weights to stay nil, got %+v", *normalized.ScaleWeights)
	}
}

func TestNormalizeUpdateQuestionInputAllowsExplicitScaleWeightsClear(t *testing.T) {
	scaleWeights := map[string]float64{}
	input := domain.UpdateQuestionInput{
		Text:         "Question",
		QuestionType: domain.QuestionTypeSingleChoice,
		ScaleWeights: &scaleWeights,
		Options: []domain.QuestionOptionInput{
			{Label: "One", Value: "1"},
		},
	}

	normalized, err := normalizeUpdateQuestionInput(input)
	if err != nil {
		t.Fatalf("normalizeUpdateQuestionInput returned error: %v", err)
	}
	if normalized.ScaleWeights == nil {
		t.Fatalf("expected explicit scale_weights payload to stay non-nil")
	}
	if len(*normalized.ScaleWeights) != 0 {
		t.Fatalf("expected explicit clear to produce empty map, got %+v", *normalized.ScaleWeights)
	}
}
