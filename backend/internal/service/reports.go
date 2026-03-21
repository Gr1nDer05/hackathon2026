package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/repository"
)

var ErrReportNotReady = errors.New("report not ready")
var ErrInvalidReportFormat = errors.New("invalid report format")

const ReportContentTypeHTML = "text/html; charset=utf-8"
const ReportContentTypeDOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

type GeneratedReport struct {
	Filename    string
	ContentType string
	Content     []byte
}

type reportFormat string

const (
	reportFormatHTML reportFormat = "html"
	reportFormatDOCX reportFormat = "docx"
)

type reportDocument struct {
	Title    string
	Subtitle string
	Meta     []reportMetaItem
	Sections []reportSection
}

type reportMetaItem struct {
	Key   string
	Label string
	Value string
}

type reportSection struct {
	Key               string
	Title             string
	Paragraphs        []string
	Bullets           []string
	TableHeader       []string
	TableRows         [][]string
	ChartBars         []reportChartBar
	ChartImageDataURI template.URL
	ChartAltText      string
	ChartCaption      string
}

type reportChartBar struct {
	Label string
	Value float64
}

const (
	reportMetaRespondent     = "respondent"
	reportMetaSession        = "session"
	reportMetaStatus         = "status"
	reportMetaTopScales      = "top_scales"
	reportMetaTopProfessions = "top_professions"
)

const (
	reportSectionIntro           = "intro"
	reportSectionSummary         = "summary"
	reportSectionChartData       = "chart_data"
	reportSectionScalesList      = "scales_list"
	reportSectionInterpretation  = "interpretation"
	reportSectionRecommendations = "recommendations"
	reportSectionRawScores       = "raw_scores"
	reportSectionAnswersTable    = "answers_table"
	reportSectionClosing         = "closing"
)

func (s *AppService) GenerateClientReportBySessionID(ctx context.Context, userID int64, sessionID int64, rawFormat string) (GeneratedReport, error) {
	format, err := parseReportFormat(rawFormat)
	if err != nil {
		return GeneratedReport{}, err
	}

	submission, err := s.repo.GetPsychologistTestSubmissionBySessionID(ctx, sessionID, userID)
	if err != nil {
		if repository.IsNotFound(err) {
			return GeneratedReport{}, ErrTestNotFound
		}
		return GeneratedReport{}, err
	}
	if submission.Status != "completed" {
		return GeneratedReport{}, ErrReportNotReady
	}

	test, err := s.GetPsychologistTestByID(ctx, userID, submission.TestID)
	if err != nil {
		return GeneratedReport{}, err
	}
	return s.generateClientReportForSubmission(ctx, userID, test, submission, format)
}

func (s *AppService) GeneratePsychologistReportBySessionID(ctx context.Context, userID int64, sessionID int64, rawFormat string) (GeneratedReport, error) {
	format, err := parseReportFormat(rawFormat)
	if err != nil {
		return GeneratedReport{}, err
	}

	submission, err := s.repo.GetPsychologistTestSubmissionBySessionID(ctx, sessionID, userID)
	if err != nil {
		if repository.IsNotFound(err) {
			return GeneratedReport{}, ErrTestNotFound
		}
		return GeneratedReport{}, err
	}
	if submission.Status != "completed" {
		return GeneratedReport{}, ErrReportNotReady
	}

	test, err := s.GetPsychologistTestByID(ctx, userID, submission.TestID)
	if err != nil {
		return GeneratedReport{}, err
	}
	questions, err := s.repo.ListPsychologistQuestions(ctx, submission.TestID, userID)
	if err != nil {
		return GeneratedReport{}, err
	}
	if err := s.attachResultContractToSubmission(ctx, userID, &submission); err != nil {
		return GeneratedReport{}, err
	}
	if submission.CareerResult == nil && len(submission.Metrics) == 0 && len(submission.TopMetrics) == 0 {
		return GeneratedReport{}, ErrReportNotReady
	}

	document := buildPsychologistReportDocument(test, submission, questions)
	document, err = s.applyStoredReportTemplate(ctx, userID, test.ReportTemplateID, document, reportAudiencePsychologist)
	if err != nil {
		return GeneratedReport{}, err
	}
	return renderReport(document, format, fmt.Sprintf("psychologist-report-%d", submission.SessionID))
}

func (s *AppService) GeneratePublicClientReport(ctx context.Context, slug string, accessToken string, rawFormat string) (GeneratedReport, error) {
	format, err := parseReportFormat(rawFormat)
	if err != nil {
		return GeneratedReport{}, err
	}

	slug = strings.TrimSpace(slug)
	accessToken = strings.TrimSpace(accessToken)
	if slug == "" || accessToken == "" {
		return GeneratedReport{}, ErrPublicTestNotFound
	}

	if err := s.repo.DeleteExpiredPublicTestSessions(ctx); err != nil {
		return GeneratedReport{}, err
	}

	publicTest, err := s.repo.GetPublicTestBySlugAndAccessToken(ctx, slug, accessToken)
	if err != nil {
		if repository.IsNotFound(err) {
			return GeneratedReport{}, ErrPublicTestNotFound
		}
		return GeneratedReport{}, err
	}
	if !publicTest.ShowClientReportImmediately {
		return GeneratedReport{}, ErrPublicClientReportUnavailable
	}

	session, err := s.repo.GetPublicTestSessionByAccessToken(ctx, slug, accessToken)
	if err != nil {
		if repository.IsNotFound(err) {
			return GeneratedReport{}, ErrPublicTestNotFound
		}
		return GeneratedReport{}, err
	}
	if session.Status != "completed" {
		return GeneratedReport{}, ErrReportNotReady
	}

	ownerUserID := publicTest.Psychologist.User.ID
	submission, err := s.repo.GetPsychologistTestSubmissionBySessionID(ctx, session.ID, ownerUserID)
	if err != nil {
		if repository.IsNotFound(err) {
			return GeneratedReport{}, ErrPublicTestNotFound
		}
		return GeneratedReport{}, err
	}

	test, err := s.GetPsychologistTestByID(ctx, ownerUserID, submission.TestID)
	if err != nil {
		if errors.Is(err, ErrTestNotFound) {
			return GeneratedReport{}, ErrPublicTestNotFound
		}
		return GeneratedReport{}, err
	}

	return s.generateClientReportForSubmission(ctx, ownerUserID, test, submission, format)
}

func (s *AppService) applyStoredReportTemplate(ctx context.Context, userID int64, templateID int64, document reportDocument, audience reportAudience) (reportDocument, error) {
	if templateID == 0 {
		return document, nil
	}

	template, err := s.repo.GetReportTemplateByID(ctx, templateID, userID)
	if err != nil {
		if repository.IsNotFound(err) {
			return document, nil
		}
		return reportDocument{}, err
	}

	return applyReportTemplateSafely(document, template, audience), nil
}

func parseReportFormat(raw string) (reportFormat, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "", string(reportFormatHTML):
		return reportFormatHTML, nil
	case string(reportFormatDOCX):
		return reportFormatDOCX, nil
	default:
		return "", ErrInvalidReportFormat
	}
}

func renderReport(document reportDocument, format reportFormat, baseFilename string) (GeneratedReport, error) {
	htmlContent, err := renderReportHTML(document)
	if err != nil {
		return GeneratedReport{}, err
	}

	switch format {
	case reportFormatHTML:
		return GeneratedReport{
			Filename:    baseFilename + ".html",
			ContentType: ReportContentTypeHTML,
			Content:     htmlContent,
		}, nil
	case reportFormatDOCX:
		content, err := renderReportDOCX(htmlContent)
		if err != nil {
			return GeneratedReport{}, err
		}
		return GeneratedReport{
			Filename:    baseFilename + ".docx",
			ContentType: ReportContentTypeDOCX,
			Content:     content,
		}, nil
	default:
		return GeneratedReport{}, ErrInvalidReportFormat
	}
}

func applyReportTemplateSafely(document reportDocument, template domain.ReportTemplate, audience reportAudience) reportDocument {
	customized, err := applyReportTemplate(document, template, audience)
	if err != nil {
		return document
	}

	return customized
}

func (s *AppService) generateClientReportForSubmission(ctx context.Context, userID int64, test domain.Test, submission domain.PsychologistTestSubmission, format reportFormat) (GeneratedReport, error) {
	if submission.CareerResult == nil && len(submission.Metrics) == 0 && len(submission.TopMetrics) == 0 {
		if err := s.attachResultContractToSubmission(ctx, userID, &submission); err != nil {
			return GeneratedReport{}, err
		}
	}
	if submission.CareerResult == nil && len(submission.Metrics) == 0 && len(submission.TopMetrics) == 0 {
		return GeneratedReport{}, ErrReportNotReady
	}

	document := buildClientReportDocument(
		domain.PublicTest{
			ID:          test.ID,
			Title:       test.Title,
			Description: test.Description,
		},
		domain.PublicTestSession{
			ID:             submission.SessionID,
			RespondentName: submission.RespondentName,
			Status:         submission.Status,
		},
		submission.CareerResult,
		submission.Metrics,
		submission.TopMetrics,
	)
	document, err := s.applyStoredReportTemplate(ctx, userID, test.ReportTemplateID, document, reportAudienceClient)
	if err != nil {
		return GeneratedReport{}, err
	}
	return renderReport(document, format, fmt.Sprintf("client-report-%d", submission.SessionID))
}

func buildClientReportDocument(test domain.PublicTest, session domain.PublicTestSession, careerResult *domain.CareerResult, metrics map[string]float64, topMetrics []domain.ResultMetric) reportDocument {
	if careerResult == nil {
		return buildGenericClientReportDocument(test, session, metrics, topMetrics)
	}

	return buildCareerClientReportDocument(test, session, careerResult)
}

func buildCareerClientReportDocument(test domain.PublicTest, session domain.PublicTestSession, careerResult *domain.CareerResult) reportDocument {
	scales := sortedScaleResults(careerResult.Scales)
	topScaleNames := joinScaleNames(careerResult.TopScales)
	topProfessionNames := joinProfessionNames(careerResult.TopProfessions)

	summary := fmt.Sprintf(
		"По результатам теста наиболее выражены шкалы %s. Наиболее близкие профессиональные направления: %s.",
		topScaleNames,
		topProfessionNames,
	)

	interpretation := []string{
		fmt.Sprintf("Профиль клиента опирается прежде всего на сочетание %s.", topScaleNames),
		buildScaleInterpretation(careerResult.TopScales),
		fmt.Sprintf("Лучше всего текущему профилю соответствуют направления %s.", topProfessionNames),
	}

	recommendations := []string{
		fmt.Sprintf("Сравнить между собой профессии %s и выбрать 1-2 направления для более глубокого изучения.", topProfessionNames),
		fmt.Sprintf("Проверить сильные стороны по шкалам %s на практике через мини-проекты, стажировки или профпробы.", topScaleNames),
		"Сопоставить результаты теста с текущими интересами, учебными предметами и желаемым форматом работы.",
		"После знакомства с направлениями вернуться к обсуждению карьерного плана вместе с психологом или наставником.",
	}

	return reportDocument{
		Title:    "Клиентский отчет по профориентации",
		Subtitle: test.Title,
		Meta: []reportMetaItem{
			{Key: reportMetaRespondent, Label: "Респондент", Value: session.RespondentName},
			{Key: reportMetaSession, Label: "Сессия", Value: strconv.FormatInt(session.ID, 10)},
			{Key: reportMetaStatus, Label: "Статус", Value: session.Status},
		},
		Sections: []reportSection{
			{
				Key:        reportSectionSummary,
				Title:      "Краткий вывод",
				Paragraphs: []string{summary},
			},
			{
				Key:       reportSectionChartData,
				Title:     "Профиль результата",
				ChartBars: buildChartBars(scales),
			},
			{
				Key:         reportSectionScalesList,
				Title:       "Шкалы",
				TableHeader: []string{"Шкала", "Процент"},
				TableRows:   buildClientScaleRows(scales),
			},
			{
				Key:        reportSectionInterpretation,
				Title:      "Интерпретация",
				Paragraphs: interpretation,
			},
			{
				Key:     reportSectionRecommendations,
				Title:   "Рекомендации",
				Bullets: recommendations,
			},
		},
	}
}

func buildPsychologistReportDocument(test domain.Test, submission domain.PsychologistTestSubmission, questions []domain.Question) reportDocument {
	if submission.CareerResult == nil {
		return buildGenericPsychologistReportDocument(test, submission, questions)
	}

	return buildCareerPsychologistReportDocument(test, submission, questions)
}

func buildCareerPsychologistReportDocument(test domain.Test, submission domain.PsychologistTestSubmission, questions []domain.Question) reportDocument {
	scales := sortedScaleResults(submission.CareerResult.Scales)
	topScaleNames := joinScaleNames(submission.CareerResult.TopScales)
	topProfessionNames := joinProfessionNames(submission.CareerResult.TopProfessions)

	return reportDocument{
		Title:    "Отчет психолога по профориентации",
		Subtitle: test.Title,
		Meta: []reportMetaItem{
			{Key: reportMetaRespondent, Label: "Респондент", Value: submission.RespondentName},
			{Key: reportMetaSession, Label: "Сессия", Value: strconv.FormatInt(submission.SessionID, 10)},
			{Key: reportMetaStatus, Label: "Статус", Value: submission.Status},
			{Key: reportMetaTopScales, Label: "Топ-шкалы", Value: topScaleNames},
			{Key: reportMetaTopProfessions, Label: "Топ-профессии", Value: topProfessionNames},
		},
		Sections: []reportSection{
			{
				Key:         reportSectionScalesList,
				Title:       "Шкалы",
				TableHeader: []string{"Шкала", "Процент"},
				TableRows:   buildClientScaleRows(scales),
			},
			{
				Key:         reportSectionRawScores,
				Title:       "Детализация шкал",
				TableHeader: []string{"Шкала", "Raw", "Max", "%"},
				TableRows:   buildPsychologistRawScoreRows(scales),
			},
			{
				Key:         reportSectionAnswersTable,
				Title:       "Ответы респондента",
				TableHeader: []string{"#", "Вопрос", "Ответ", "Тип"},
				TableRows:   buildAnswerTableRows(questions, submission.Answers),
			},
		},
	}
}

func buildGenericClientReportDocument(test domain.PublicTest, session domain.PublicTestSession, metrics map[string]float64, topMetrics []domain.ResultMetric) reportDocument {
	sortedMetrics := sortedResultMetrics(metrics)
	strongestMetrics := topMetrics
	if len(strongestMetrics) == 0 {
		strongestMetrics = sortedMetrics
	}

	return reportDocument{
		Title:    "Клиентский отчет по результатам теста",
		Subtitle: test.Title,
		Meta: []reportMetaItem{
			{Key: reportMetaRespondent, Label: "Респондент", Value: session.RespondentName},
			{Key: reportMetaSession, Label: "Сессия", Value: strconv.FormatInt(session.ID, 10)},
			{Key: reportMetaStatus, Label: "Статус", Value: session.Status},
		},
		Sections: []reportSection{
			{
				Key:        reportSectionSummary,
				Title:      "Краткий вывод",
				Paragraphs: []string{buildMetricSummary(strongestMetrics)},
			},
			{
				Key:          reportSectionChartData,
				Title:        "Профиль результата",
				ChartBars:    buildMetricChartBars(strongestMetrics),
				ChartCaption: "Сравнение наиболее выраженных итоговых метрик по текущему прохождению. Чем выше столбец, тем заметнее проявлен показатель.",
				Paragraphs: []string{
					"Для этого теста итог формируется по рассчитанным метрикам. Они помогают увидеть выраженные особенности текущего профиля без жесткой привязки к карьерным шкалам.",
				},
			},
			{
				Key:         reportSectionScalesList,
				Title:       "Метрики",
				TableHeader: []string{"Метрика", "Значение"},
				TableRows:   buildMetricRows(sortedMetrics),
			},
			{
				Key:        reportSectionInterpretation,
				Title:      "Интерпретация",
				Paragraphs: buildMetricInterpretation(strongestMetrics),
			},
			{
				Key:     reportSectionRecommendations,
				Title:   "Рекомендации",
				Bullets: buildMetricRecommendations(strongestMetrics),
			},
		},
	}
}

func buildGenericPsychologistReportDocument(test domain.Test, submission domain.PsychologistTestSubmission, questions []domain.Question) reportDocument {
	sortedMetrics := sortedResultMetrics(submission.Metrics)
	strongestMetrics := submission.TopMetrics
	if len(strongestMetrics) == 0 {
		strongestMetrics = sortedMetrics
	}

	return reportDocument{
		Title:    "Отчет психолога по результатам теста",
		Subtitle: test.Title,
		Meta: []reportMetaItem{
			{Key: reportMetaRespondent, Label: "Респондент", Value: submission.RespondentName},
			{Key: reportMetaSession, Label: "Сессия", Value: strconv.FormatInt(submission.SessionID, 10)},
			{Key: reportMetaStatus, Label: "Статус", Value: submission.Status},
		},
		Sections: []reportSection{
			{
				Key:         reportSectionScalesList,
				Title:       "Ключевые метрики",
				TableHeader: []string{"Метрика", "Значение"},
				TableRows:   buildMetricRows(strongestMetrics),
			},
			{
				Key:         reportSectionRawScores,
				Title:       "Все метрики",
				TableHeader: []string{"Метрика", "Значение"},
				TableRows:   buildMetricRows(sortedMetrics),
			},
			{
				Key:         reportSectionAnswersTable,
				Title:       "Ответы респондента",
				TableHeader: []string{"#", "Вопрос", "Ответ", "Тип"},
				TableRows:   buildAnswerTableRows(questions, submission.Answers),
			},
		},
	}
}

func buildChartBars(scales []domain.CareerScaleResult) []reportChartBar {
	bars := make([]reportChartBar, 0, len(scales))
	for _, scale := range scales {
		bars = append(bars, reportChartBar{
			Label: scaleDisplayName(scale.Scale),
			Value: scale.Percentage,
		})
	}
	return bars
}

func buildClientScaleRows(scales []domain.CareerScaleResult) [][]string {
	rows := make([][]string, 0, len(scales))
	for _, scale := range scales {
		rows = append(rows, []string{
			scaleDisplayName(scale.Scale),
			formatScore(scale.Percentage) + "%",
		})
	}
	return rows
}

func buildPsychologistRawScoreRows(scales []domain.CareerScaleResult) [][]string {
	rows := make([][]string, 0, len(scales))
	for _, scale := range scales {
		rows = append(rows, []string{
			scaleDisplayName(scale.Scale),
			formatScore(scale.RawScore),
			formatScore(scale.MaxScore),
			formatScore(scale.Percentage) + "%",
		})
	}
	return rows
}

func buildMetricRows(metrics []domain.ResultMetric) [][]string {
	rows := make([][]string, 0, len(metrics))
	for _, metric := range metrics {
		rows = append(rows, []string{
			metricDisplayName(metric.Key),
			formatScore(metric.Value),
		})
	}
	return rows
}

func buildMetricChartBars(metrics []domain.ResultMetric) []reportChartBar {
	bars := make([]reportChartBar, 0, len(metrics))
	for _, metric := range metrics {
		if strings.TrimSpace(metric.Key) == "" {
			continue
		}
		bars = append(bars, reportChartBar{
			Label: metricDisplayName(metric.Key),
			Value: metric.Value,
		})
	}
	return bars
}

func buildAnswerTableRows(questions []domain.Question, answers []domain.PublicTestAnswer) [][]string {
	answerByQuestionID := make(map[int64]domain.PublicTestAnswer, len(answers))
	for _, answer := range answers {
		answerByQuestionID[answer.QuestionID] = answer
	}

	rows := make([][]string, 0, len(questions))
	for _, question := range questions {
		answer, ok := answerByQuestionID[question.ID]
		if !ok {
			continue
		}

		rows = append(rows, []string{
			strconv.Itoa(question.OrderNumber),
			question.Text,
			displayAnswer(question, answer),
			question.QuestionType,
		})
	}

	return rows
}

func displayAnswer(question domain.Question, answer domain.PublicTestAnswer) string {
	optionLabelByValue := make(map[string]string, len(question.Options))
	for _, option := range question.Options {
		optionLabelByValue[option.Value] = option.Label
	}

	switch question.QuestionType {
	case domain.QuestionTypeText:
		return answer.AnswerText
	case domain.QuestionTypeNumber:
		if answer.AnswerValue != "" {
			return answer.AnswerValue
		}
		return answer.AnswerText
	case domain.QuestionTypeSingleChoice, domain.QuestionTypeScale:
		if label := optionLabelByValue[answer.AnswerValue]; label != "" {
			return label
		}
		return answer.AnswerValue
	case domain.QuestionTypeMultiple:
		values := make([]string, 0, len(answer.AnswerValues))
		for _, value := range answer.AnswerValues {
			if label := optionLabelByValue[value]; label != "" {
				values = append(values, label)
				continue
			}
			values = append(values, value)
		}
		return strings.Join(values, ", ")
	default:
		if answer.AnswerText != "" {
			return answer.AnswerText
		}
		if answer.AnswerValue != "" {
			return answer.AnswerValue
		}
		return strings.Join(answer.AnswerValues, ", ")
	}
}

func sortedScaleResults(scales []domain.CareerScaleResult) []domain.CareerScaleResult {
	result := append([]domain.CareerScaleResult(nil), scales...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Percentage != result[j].Percentage {
			return result[i].Percentage > result[j].Percentage
		}
		if result[i].RawScore != result[j].RawScore {
			return result[i].RawScore > result[j].RawScore
		}
		return scaleDisplayName(result[i].Scale) < scaleDisplayName(result[j].Scale)
	})
	return result
}

func joinScaleNames(scales []domain.CareerScaleResult) string {
	if len(scales) == 0 {
		return "без выраженных шкал"
	}

	names := make([]string, 0, len(scales))
	for _, scale := range scales {
		names = append(names, scaleDisplayName(scale.Scale))
	}
	return strings.Join(names, ", ")
}

func joinProfessionNames(professions []domain.CareerProfessionResult) string {
	if len(professions) == 0 {
		return "без явно выраженных направлений"
	}

	names := make([]string, 0, len(professions))
	for _, profession := range professions {
		names = append(names, profession.Profession)
	}
	return strings.Join(names, ", ")
}

func sortedResultMetrics(metrics map[string]float64) []domain.ResultMetric {
	items := make([]domain.ResultMetric, 0, len(metrics))
	for key, value := range metrics {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		items = append(items, domain.ResultMetric{
			Key:   trimmedKey,
			Value: roundToTwo(value),
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Value != items[j].Value {
			return items[i].Value > items[j].Value
		}
		return metricDisplayName(items[i].Key) < metricDisplayName(items[j].Key)
	})

	return items
}

func buildMetricSummary(metrics []domain.ResultMetric) string {
	if len(metrics) == 0 {
		return "По этому прохождению рассчитан итоговый результат, но выраженные метрики не выделены."
	}

	return fmt.Sprintf(
		"По результатам теста наиболее заметно проявились показатели %s. Это основной ориентир для чтения результата по текущему прохождению.",
		joinMetricNamesWithValues(metrics),
	)
}

func buildMetricInterpretation(metrics []domain.ResultMetric) []string {
	if len(metrics) == 0 {
		return []string{
			"Итог теста сформирован автоматически. Для точной интерпретации его важно обсуждать вместе с психологом и сопоставлять с контекстом клиента.",
		}
	}

	return []string{
		fmt.Sprintf("Автоматический расчет показывает, что сейчас сильнее всего выражены показатели %s.", joinMetricNamesWithValues(metrics)),
		"Высокие значения сами по себе не являются диагнозом или окончательным выводом. Они помогают выделить темы, которые стоит обсудить подробнее на консультации.",
		"Наиболее полезно читать этот результат вместе с ответами респондента, наблюдениями специалиста и данными следующих встреч.",
	}
}

func buildMetricRecommendations(metrics []domain.ResultMetric) []string {
	if len(metrics) == 0 {
		return []string{
			"Обсудить итог теста вместе с психологом и сопоставить автоматический расчет с реальными наблюдениями по клиенту.",
		}
	}

	return []string{
		fmt.Sprintf("Разобрать вместе с психологом, как показатели %s проявляются в повседневных решениях, поведении и рабочих или учебных задачах клиента.", joinMetricNames(metrics)),
		"Использовать выделенные метрики как основу для уточняющих вопросов и последующего интервью, а не как окончательный вердикт.",
		"Сопоставить автоматический результат с ответами респондента, наблюдениями специалиста и дополнительными данными следующих сессий.",
	}
}

func joinMetricNames(metrics []domain.ResultMetric) string {
	if len(metrics) == 0 {
		return "без выраженных показателей"
	}

	names := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		names = append(names, metricDisplayName(metric.Key))
	}
	return strings.Join(names, ", ")
}

func joinMetricNamesWithValues(metrics []domain.ResultMetric) string {
	if len(metrics) == 0 {
		return "без выраженных показателей"
	}

	items := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		items = append(items, fmt.Sprintf("%s (%s)", metricDisplayName(metric.Key), formatScore(metric.Value)))
	}
	return strings.Join(items, ", ")
}

func buildScaleInterpretation(scales []domain.CareerScaleResult) string {
	if len(scales) == 0 {
		return "Выраженные карьерные шкалы по текущему прохождению не определились."
	}

	parts := make([]string, 0, len(scales))
	for _, scale := range scales {
		description := scaleInterpretation(scale.Scale)
		if description == "" {
			continue
		}
		parts = append(parts, description)
	}
	if len(parts) == 0 {
		return "Полученный профиль стоит обсуждать в связке с интересами и реальным опытом клиента."
	}
	return strings.Join(parts, " ")
}

func scaleDisplayName(scale string) string {
	switch scale {
	case domain.CareerScaleAnalytic:
		return "Аналитичность"
	case domain.CareerScaleCreative:
		return "Креативность"
	case domain.CareerScaleSocial:
		return "Социальность"
	case domain.CareerScaleOrganizer:
		return "Организованность"
	case domain.CareerScalePractical:
		return "Практичность"
	default:
		return scale
	}
}

func metricDisplayName(key string) string {
	switch key {
	case domain.CareerScaleAnalytic, domain.CareerScaleCreative, domain.CareerScaleSocial, domain.CareerScaleOrganizer, domain.CareerScalePractical:
		return scaleDisplayName(key)
	case "total":
		return "Итоговый балл"
	}

	normalized := strings.TrimSpace(key)
	if normalized == "" {
		return key
	}

	normalized = strings.ReplaceAll(normalized, "_", " ")
	normalized = strings.ReplaceAll(normalized, "-", " ")
	return uppercaseFirstRune(normalized)
}

func uppercaseFirstRune(value string) string {
	if value == "" {
		return value
	}

	r, size := utf8.DecodeRuneInString(value)
	if r == utf8.RuneError && size == 0 {
		return value
	}

	return string(unicode.ToUpper(r)) + value[size:]
}

func scaleInterpretation(scale string) string {
	switch scale {
	case domain.CareerScaleAnalytic:
		return "Высокая аналитичность обычно связана с интересом к структурам, логике, данным и поиску причинно-следственных связей."
	case domain.CareerScaleCreative:
		return "Высокая креативность показывает склонность к созданию новых решений, визуальному мышлению и нестандартным подходам."
	case domain.CareerScaleSocial:
		return "Высокая социальность указывает на выраженный интерес к взаимодействию с людьми, поддержке, исследованию потребностей и коммуникации."
	case domain.CareerScaleOrganizer:
		return "Высокая организованность отражает способность держать рамку процесса, координировать задачи и выстраивать приоритеты."
	case domain.CareerScalePractical:
		return "Высокая практичность чаще всего означает ориентацию на применимость результата, работающие инструменты и доведение задач до рабочего состояния."
	default:
		return ""
	}
}

func formatScore(value float64) string {
	text := strconv.FormatFloat(value, 'f', 2, 64)
	text = strings.TrimSuffix(text, "00")
	text = strings.TrimSuffix(text, "0")
	text = strings.TrimSuffix(text, ".")
	if text == "" {
		return "0"
	}
	return text
}

var reportHTMLTemplate = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ .Title }}</title>
  <style>
    :root { color-scheme: light; --ink:#1f2937; --muted:#6b7280; --line:#d7dde8; --panel:#f8fafc; --accent:#116466; --accent-soft:#d8efef; }
    * { box-sizing:border-box; }
    body { margin:0; padding:32px; font-family: "Segoe UI", "Helvetica Neue", Arial, sans-serif; color:var(--ink); background:linear-gradient(180deg,#fbfdff 0%,#eef5f7 100%); }
    .page { max-width:980px; margin:0 auto; background:#fff; border:1px solid var(--line); border-radius:20px; padding:32px; box-shadow:0 20px 45px rgba(17,100,102,.08); }
    h1 { margin:0 0 8px; font-size:30px; line-height:1.2; }
    .subtitle { margin:0 0 24px; color:var(--muted); font-size:16px; }
    .meta { display:grid; grid-template-columns:repeat(auto-fit, minmax(180px, 1fr)); gap:12px; margin-bottom:28px; }
    .meta-item { padding:14px 16px; border-radius:14px; background:var(--panel); border:1px solid var(--line); }
    .meta-label { display:block; font-size:12px; text-transform:uppercase; letter-spacing:.06em; color:var(--muted); margin-bottom:6px; }
    .meta-value { font-size:15px; font-weight:600; }
    section { margin-top:28px; }
    h2 { margin:0 0 14px; font-size:20px; }
    p { margin:0 0 12px; line-height:1.65; }
    ul { margin:0; padding-left:20px; }
    li { margin-bottom:8px; line-height:1.55; }
    table { width:100%; border-collapse:collapse; background:#fff; border:1px solid var(--line); border-radius:14px; overflow:hidden; }
    th, td { text-align:left; padding:12px 14px; border-bottom:1px solid var(--line); vertical-align:top; }
    th { background:var(--panel); font-size:13px; }
    tr:last-child td { border-bottom:none; }
    .chart-card { margin:0; padding:18px; border:1px solid var(--line); border-radius:18px; background:linear-gradient(180deg,#fcffff 0%,#f4fbfb 100%); }
    .chart-image { display:block; width:100%; height:auto; border-radius:14px; background:#fff; }
    .chart-note { margin:14px 0 0; color:var(--muted); font-size:14px; }
  </style>
</head>
<body>
  <main class="page">
    <h1>{{ .Title }}</h1>
    {{ if .Subtitle }}<p class="subtitle">{{ .Subtitle }}</p>{{ end }}
    {{ if .Meta }}
    <div class="meta">
      {{ range .Meta }}
      <div class="meta-item">
        <span class="meta-label">{{ .Label }}</span>
        <span class="meta-value">{{ .Value }}</span>
      </div>
      {{ end }}
    </div>
    {{ end }}
    {{ range .Sections }}
    <section>
      <h2>{{ .Title }}</h2>
      {{ range .Paragraphs }}<p>{{ . }}</p>{{ end }}
      {{ if .ChartImageDataURI }}
      <figure class="chart-card">
        <img class="chart-image" src="{{ .ChartImageDataURI }}" alt="{{ .ChartAltText }}">
        {{ if .ChartCaption }}<p class="chart-note">{{ .ChartCaption }}</p>{{ end }}
      </figure>
      {{ end }}
      {{ if .TableHeader }}
      <table>
        <thead>
          <tr>{{ range .TableHeader }}<th>{{ . }}</th>{{ end }}</tr>
        </thead>
        <tbody>
          {{ range .TableRows }}
          <tr>{{ range . }}<td>{{ . }}</td>{{ end }}</tr>
          {{ end }}
        </tbody>
      </table>
      {{ end }}
      {{ if .Bullets }}
      <ul>{{ range .Bullets }}<li>{{ . }}</li>{{ end }}</ul>
      {{ end }}
    </section>
    {{ end }}
  </main>
</body>
</html>`))

func renderReportHTML(document reportDocument) ([]byte, error) {
	document, err := buildHTMLReportDocument(document)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := reportHTMLTemplate.Execute(&buf, document); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
