package service

import (
	"strings"
	"testing"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func TestNormalizeReportTemplateBodyRejectsUnknownSection(t *testing.T) {
	_, err := normalizeReportTemplateBody(`{
  "client": {
    "section_titles": {
      "raw_scores": "Should not exist in client report"
    }
  }
}`)
	if err != ErrInvalidReportTemplateInput {
		t.Fatalf("expected ErrInvalidReportTemplateInput, got %v", err)
	}
}

func TestApplyReportTemplateOverridesLabelsAndAddsIntro(t *testing.T) {
	normalized, err := normalizeReportTemplateBody(`{
  "client": {
    "title": "Индивидуальный отчет",
    "meta_labels": {
      "respondent": "Клиент"
    },
    "section_titles": {
      "summary": "Итог",
      "intro": "Как читать отчет"
    },
    "intro_paragraphs": [
      "Первый вводный абзац."
    ],
    "chart_caption": "Новый заголовок графика"
  }
}`)
	if err != nil {
		t.Fatalf("normalizeReportTemplateBody returned error: %v", err)
	}

	document, err := applyReportTemplate(reportDocument{
		Title: "Базовый отчет",
		Meta: []reportMetaItem{
			{Key: reportMetaRespondent, Label: "Респондент", Value: "Иван Иванов"},
		},
		Sections: []reportSection{
			{Key: reportSectionSummary, Title: "Summary", Paragraphs: []string{"Короткий итог"}},
			{Key: reportSectionChartData, Title: "Chart Data", ChartCaption: "Старый заголовок"},
		},
	}, domain.ReportTemplate{TemplateBody: normalized}, reportAudienceClient)
	if err != nil {
		t.Fatalf("applyReportTemplate returned error: %v", err)
	}

	if document.Title != "Индивидуальный отчет" {
		t.Fatalf("expected custom title, got %q", document.Title)
	}
	if document.Meta[0].Label != "Клиент" {
		t.Fatalf("expected custom meta label, got %q", document.Meta[0].Label)
	}
	if len(document.Sections) != 3 {
		t.Fatalf("expected intro section to be added, got %d sections", len(document.Sections))
	}
	if document.Sections[0].Key != reportSectionIntro {
		t.Fatalf("expected intro section first, got %q", document.Sections[0].Key)
	}
	if !strings.Contains(document.Sections[0].Paragraphs[0], "вводный") {
		t.Fatalf("expected intro paragraphs to be preserved, got %#v", document.Sections[0].Paragraphs)
	}
	if document.Sections[1].Title != "Итог" {
		t.Fatalf("expected summary title override, got %q", document.Sections[1].Title)
	}
	if document.Sections[2].ChartCaption != "Новый заголовок графика" {
		t.Fatalf("expected chart caption override, got %q", document.Sections[2].ChartCaption)
	}
}
