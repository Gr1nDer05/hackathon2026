package service

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func TestRenderReportHTMLIncludesKeySections(t *testing.T) {
	document := buildClientReportDocument(
		domain.PublicTest{Title: "Профориентация"},
		domain.PublicTestSession{ID: 42, RespondentName: "Иван Иванов", Status: "completed"},
		&domain.CareerResult{
			Scales: []domain.CareerScaleResult{
				{Scale: domain.CareerScaleAnalytic, Percentage: 82, RawScore: 16.4, MaxScore: 20},
				{Scale: domain.CareerScalePractical, Percentage: 74, RawScore: 14.8, MaxScore: 20},
			},
			TopScales: []domain.CareerScaleResult{
				{Scale: domain.CareerScaleAnalytic, Percentage: 82},
				{Scale: domain.CareerScalePractical, Percentage: 74},
			},
			TopProfessions: []domain.CareerProfessionResult{
				{Profession: "Backend Developer", Score: 81},
				{Profession: "DevOps Engineer", Score: 78},
				{Profession: "Data Engineer", Score: 74},
			},
		},
	)

	content, err := renderReportHTML(document)
	if err != nil {
		t.Fatalf("renderReportHTML returned error: %v", err)
	}

	html := string(content)
	for _, fragment := range []string{
		"Клиентский отчет по профориентации",
		"Summary",
		"Chart Data",
		"Backend Developer",
		"Аналитичность",
		"data:image/png;base64,",
	} {
		if !bytes.Contains(content, []byte(fragment)) {
			t.Fatalf("expected HTML to contain %q, got %s", fragment, html)
		}
	}
}

func TestRenderReportDOCXCreatesArchiveWithWordDocumentAndChart(t *testing.T) {
	htmlContent, err := renderReportHTML(reportDocument{
		Title:    "Отчет",
		Subtitle: "Тест",
		Sections: []reportSection{
			{Title: "Summary", Paragraphs: []string{"Hello"}},
			{
				Title: "Chart Data",
				ChartBars: []reportChartBar{
					{Label: "Аналитичность", Value: 82},
					{Label: "Практичность", Value: 74},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("renderReportHTML returned error: %v", err)
	}

	content, err := renderReportDOCX(htmlContent)
	if err != nil {
		t.Fatalf("renderReportDOCX returned error: %v", err)
	}
	if len(content) < 2 || content[0] != 'P' || content[1] != 'K' {
		t.Fatalf("expected zip signature, got %v", content[:2])
	}

	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("zip.NewReader returned error: %v", err)
	}

	foundDocumentXML := false
	foundChartPNG := false
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			foundDocumentXML = true
		}
		if file.Name == "word/media/chart-1.png" {
			foundChartPNG = true
		}
	}
	if !foundDocumentXML {
		t.Fatalf("expected word/document.xml inside docx archive")
	}
	if !foundChartPNG {
		t.Fatalf("expected word/media/chart-1.png inside docx archive")
	}
}

func TestBuildPsychologistReportDocumentIsTechnical(t *testing.T) {
	document := buildPsychologistReportDocument(
		domain.Test{Title: "Профориентация"},
		domain.PsychologistTestSubmission{
			SessionID:      7,
			RespondentName: "Иван Иванов",
			Status:         "completed",
			CareerResult: &domain.CareerResult{
				Scales: []domain.CareerScaleResult{
					{Scale: domain.CareerScaleAnalytic, Percentage: 82, RawScore: 16.4, MaxScore: 20},
					{Scale: domain.CareerScalePractical, Percentage: 74, RawScore: 14.8, MaxScore: 20},
				},
				TopScales: []domain.CareerScaleResult{
					{Scale: domain.CareerScaleAnalytic, Percentage: 82},
					{Scale: domain.CareerScalePractical, Percentage: 74},
				},
				TopProfessions: []domain.CareerProfessionResult{
					{Profession: "Backend Developer", Score: 81},
					{Profession: "DevOps Engineer", Score: 78},
				},
			},
			Answers: []domain.PublicTestAnswer{
				{QuestionID: 1, AnswerValue: "often"},
			},
		},
		[]domain.Question{
			{
				ID:           1,
				OrderNumber:  1,
				Text:         "Любите ли вы анализировать?",
				QuestionType: domain.QuestionTypeSingleChoice,
				Options: []domain.QuestionOption{
					{Value: "often", Label: "Часто"},
				},
			},
		},
	)

	if len(document.Sections) != 3 {
		t.Fatalf("expected 3 technical sections, got %+v", document.Sections)
	}

	titles := []string{
		document.Sections[0].Title,
		document.Sections[1].Title,
		document.Sections[2].Title,
	}
	expected := []string{"Scales List", "Raw Scores", "Answers Table"}
	for idx, title := range expected {
		if titles[idx] != title {
			t.Fatalf("expected section %d to be %q, got %+v", idx, title, titles)
		}
	}
}
