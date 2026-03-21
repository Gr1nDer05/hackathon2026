package repository

import (
	"testing"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func TestCareerResultJSONRoundTrip(t *testing.T) {
	source := &domain.CareerResult{
		Scales: []domain.CareerScaleResult{
			{Scale: domain.CareerScaleAnalytic, RawScore: 12, MaxScore: 20, Percentage: 60},
		},
		TopScales: []domain.CareerScaleResult{
			{Scale: domain.CareerScaleAnalytic, RawScore: 12, MaxScore: 20, Percentage: 60},
		},
		TopProfessions: []domain.CareerProfessionResult{
			{Profession: "Backend Developer", Score: 72.5},
		},
	}

	raw, err := marshalCareerResult(source)
	if err != nil {
		t.Fatalf("marshalCareerResult returned error: %v", err)
	}

	text, ok := raw.(string)
	if !ok || text == "" {
		t.Fatalf("expected JSON string payload, got %#v", raw)
	}

	result, err := unmarshalCareerResult([]byte(text))
	if err != nil {
		t.Fatalf("unmarshalCareerResult returned error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected decoded result, got nil")
	}
	if len(result.TopProfessions) != 1 || result.TopProfessions[0].Profession != "Backend Developer" {
		t.Fatalf("unexpected decoded professions: %+v", result.TopProfessions)
	}
}

func TestCareerResultJSONNilRoundTrip(t *testing.T) {
	raw, err := marshalCareerResult(nil)
	if err != nil {
		t.Fatalf("marshalCareerResult returned error: %v", err)
	}
	if raw != nil {
		t.Fatalf("expected nil raw payload, got %#v", raw)
	}

	result, err := unmarshalCareerResult(nil)
	if err != nil {
		t.Fatalf("unmarshalCareerResult returned error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
}
