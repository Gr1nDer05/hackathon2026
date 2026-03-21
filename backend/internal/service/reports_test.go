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
		nil,
		nil,
	)

	content, err := renderReportHTML(document)
	if err != nil {
		t.Fatalf("renderReportHTML returned error: %v", err)
	}

	html := string(content)
	for _, fragment := range []string{
		"Клиентский отчет по профориентации",
		"Краткий вывод",
		"Профиль результата",
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
	expected := []string{"Шкалы", "Детализация шкал", "Ответы респондента"}
	for idx, title := range expected {
		if titles[idx] != title {
			t.Fatalf("expected section %d to be %q, got %+v", idx, title, titles)
		}
	}
}

func TestBuildClientReportDocumentSupportsGenericMetricsWithoutCareerResult(t *testing.T) {
	document := buildClientReportDocument(
		domain.PublicTest{Title: "Тест на стрессоустойчивость"},
		domain.PublicTestSession{ID: 15, RespondentName: "Иван Иванов", Status: "completed"},
		nil,
		map[string]float64{
			"stress_resistance": 18,
			"leadership":        11,
			"total":             29,
		},
		[]domain.ResultMetric{
			{Key: "stress_resistance", Value: 18},
			{Key: "leadership", Value: 11},
		},
	)

	if document.Title != "Клиентский отчет по результатам теста" {
		t.Fatalf("expected generic client report title, got %q", document.Title)
	}
	if len(document.Sections) != 5 {
		t.Fatalf("expected 5 generic client sections, got %+v", document.Sections)
	}
	if document.Sections[0].Title != "Краткий вывод" || document.Sections[1].Title != "Профиль результата" || document.Sections[2].Title != "Метрики" {
		t.Fatalf("expected localized generic section titles, got %+v", document.Sections)
	}
	if document.Sections[1].Key != reportSectionChartData || len(document.Sections[1].Paragraphs) == 0 {
		t.Fatalf("expected chart section to contain fallback explanation, got %+v", document.Sections[1])
	}
	if len(document.Sections[1].ChartBars) != 2 {
		t.Fatalf("expected metric chart bars, got %+v", document.Sections[1].ChartBars)
	}
	if len(document.Sections[2].TableRows) < 2 {
		t.Fatalf("expected metrics table rows, got %+v", document.Sections[2].TableRows)
	}
}

func TestGenerateClientReportForSubmissionSupportsGenericMetrics(t *testing.T) {
	service := &AppService{}

	report, err := service.generateClientReportForSubmission(
		t.Context(),
		77,
		domain.Test{
			ID:    5,
			Title: "Стрессоустойчивость",
		},
		domain.PsychologistTestSubmission{
			SessionID:      101,
			TestID:         5,
			RespondentName: "Иван Иванов",
			Status:         "completed",
			Metrics: map[string]float64{
				"stress_resistance": 18,
				"leadership":        11,
				"total":             29,
			},
			TopMetrics: []domain.ResultMetric{
				{Key: "stress_resistance", Value: 18},
				{Key: "leadership", Value: 11},
			},
		},
		reportFormatHTML,
	)
	if err != nil {
		t.Fatalf("expected generic client report to render, got error: %v", err)
	}
	if report.Filename != "client-report-101.html" {
		t.Fatalf("unexpected report filename %q", report.Filename)
	}
	if !bytes.Contains(report.Content, []byte("Stress resistance")) {
		t.Fatalf("expected rendered report to contain generic metric label, got %s", string(report.Content))
	}
	if !bytes.Contains(report.Content, []byte("Итоговый балл")) {
		t.Fatalf("expected rendered report to contain total metric label, got %s", string(report.Content))
	}
}
