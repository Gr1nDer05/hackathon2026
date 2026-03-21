package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultDemoPsychologistEmail    = "demo.psychologist@profdnk.local"
	defaultDemoPsychologistPassword = "demo12345"
	defaultDemoPsychologistName     = "Demo Psychologist"
	demoReportTemplateName          = "Базовый шаблон ПрофДНК"
	demoTestTitle                   = "ПрофДНК: IT-профориентация"
)

func (s *AppService) SeedDemoData(ctx context.Context) error {
	if !demoDataEnabled() {
		return nil
	}

	userID, err := s.ensureDemoPsychologist(ctx)
	if err != nil {
		return err
	}

	templateID, err := s.ensureDemoReportTemplate(ctx, userID)
	if err != nil {
		return err
	}

	testID, err := s.ensureDemoTest(ctx, userID, templateID)
	if err != nil {
		return err
	}

	return s.ensureDemoQuestions(ctx, userID, testID)
}

func demoDataEnabled() bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv("DEMO_DATA_ENABLED")))
	if raw == "true" {
		return true
	}
	if raw == "false" {
		return false
	}

	return !strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production")
}

func (s *AppService) ensureDemoPsychologist(ctx context.Context) (int64, error) {
	email := strings.TrimSpace(strings.ToLower(getenvDefault("DEMO_PSYCHOLOGIST_EMAIL", defaultDemoPsychologistEmail)))
	password := strings.TrimSpace(getenvDefault("DEMO_PSYCHOLOGIST_PASSWORD", defaultDemoPsychologistPassword))
	fullName := strings.TrimSpace(getenvDefault("DEMO_PSYCHOLOGIST_FULL_NAME", defaultDemoPsychologistName))

	credentials, err := s.repo.GetPsychologistCredentialsByEmail(ctx, email)
	if err == nil {
		return credentials.User.ID, nil
	}
	if !repository.IsNotFound(err) {
		return 0, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	workspace, err := s.repo.CreatePsychologist(ctx, domain.PsychologistRegistrationInput{
		Email:    email,
		Password: password,
		FullName: fullName,
		IsActive: true,
	}, string(passwordHash))
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			credentials, getErr := s.repo.GetPsychologistCredentialsByEmail(ctx, email)
			if getErr != nil {
				return 0, getErr
			}
			return credentials.User.ID, nil
		}
		return 0, err
	}

	return workspace.User.ID, nil
}

func (s *AppService) ensureDemoReportTemplate(ctx context.Context, userID int64) (int64, error) {
	templates, err := s.repo.ListReportTemplates(ctx, userID)
	if err != nil {
		return 0, err
	}

	for _, template := range templates {
		if template.Name == demoReportTemplateName {
			return template.ID, nil
		}
	}

	template, err := s.CreateReportTemplate(ctx, userID, domain.CreateReportTemplateInput{
		Name:         demoReportTemplateName,
		Description:  "Демо-шаблон для клиентского и профессионального отчета",
		TemplateBody: buildDemoReportTemplateBody(),
	})
	if err != nil {
		return 0, err
	}

	return template.ID, nil
}

func (s *AppService) ensureDemoTest(ctx context.Context, userID int64, templateID int64) (int64, error) {
	tests, err := s.repo.ListPsychologistTests(ctx, userID)
	if err != nil {
		return 0, err
	}

	for _, test := range tests {
		if test.Title != demoTestTitle {
			continue
		}
		if test.ReportTemplateID != templateID {
			_, err := s.UpdatePsychologistTest(ctx, userID, test.ID, domain.UpdateTestInput{
				Title:                      test.Title,
				Description:                test.Description,
				ReportTemplateID:           templateID,
				RecommendedDuration:        test.RecommendedDuration,
				MaxParticipants:            test.MaxParticipants,
				HasParticipantLimit:        demoBoolPtr(test.HasParticipantLimit),
				CollectRespondentAge:       test.CollectRespondentAge,
				CollectRespondentGender:    test.CollectRespondentGender,
				CollectRespondentEducation: test.CollectRespondentEducation,
				Status:                     test.Status,
			})
			if err != nil {
				return 0, err
			}
		}
		return test.ID, nil
	}

	test, err := s.CreatePsychologistTest(ctx, userID, domain.CreateTestInput{
		Title:                      demoTestTitle,
		Description:                "Демо-методика для оценки склонности к IT-направлениям.",
		ReportTemplateID:           templateID,
		RecommendedDuration:        15,
		MaxParticipants:            0,
		HasParticipantLimit:        demoBoolPtr(false),
		CollectRespondentAge:       true,
		CollectRespondentGender:    true,
		CollectRespondentEducation: true,
		Status:                     domain.TestStatusPublished,
	})
	if err != nil {
		return 0, err
	}

	return test.ID, nil
}

func (s *AppService) ensureDemoQuestions(ctx context.Context, userID int64, testID int64) error {
	questions, err := s.repo.ListPsychologistQuestions(ctx, testID, userID)
	if err != nil {
		return err
	}
	if len(questions) > 0 {
		return nil
	}

	demoQuestions := []domain.CreateQuestionInput{
		{
			Text:         "Насколько вам интересно искать закономерности в данных и сложных системах?",
			QuestionType: domain.QuestionTypeScale,
			OrderNumber:  1,
			IsRequired:   true,
			ScaleWeights: map[string]float64{domain.CareerScaleAnalytic: 1.0, domain.CareerScalePractical: 0.2},
			Options:      demoScaleOptions(),
		},
		{
			Text:         "Насколько вам нравится придумывать новые визуальные или продуктовые решения?",
			QuestionType: domain.QuestionTypeScale,
			OrderNumber:  2,
			IsRequired:   true,
			ScaleWeights: map[string]float64{domain.CareerScaleCreative: 1.0, domain.CareerScaleSocial: 0.2},
			Options:      demoScaleOptions(),
		},
		{
			Text:         "Насколько вам комфортно много общаться, собирать обратную связь и помогать пользователям?",
			QuestionType: domain.QuestionTypeScale,
			OrderNumber:  3,
			IsRequired:   true,
			ScaleWeights: map[string]float64{domain.CareerScaleSocial: 1.0, domain.CareerScaleOrganizer: 0.2},
			Options:      demoScaleOptions(),
		},
		{
			Text:         "Насколько вам нравится координировать задачи, сроки и работу команды?",
			QuestionType: domain.QuestionTypeScale,
			OrderNumber:  4,
			IsRequired:   true,
			ScaleWeights: map[string]float64{domain.CareerScaleOrganizer: 1.0, domain.CareerScaleSocial: 0.2},
			Options:      demoScaleOptions(),
		},
		{
			Text:         "Насколько вам важно, чтобы результат можно было быстро применить на практике и довести до рабочего состояния?",
			QuestionType: domain.QuestionTypeScale,
			OrderNumber:  5,
			IsRequired:   true,
			ScaleWeights: map[string]float64{domain.CareerScalePractical: 1.0, domain.CareerScaleAnalytic: 0.2},
			Options:      demoScaleOptions(),
		},
	}

	for _, question := range demoQuestions {
		if _, err := s.CreatePsychologistQuestion(ctx, userID, testID, question); err != nil {
			return err
		}
	}

	return nil
}

func demoScaleOptions() []domain.QuestionOptionInput {
	return []domain.QuestionOptionInput{
		{Label: "Совсем не похоже на меня", Value: "1", OrderNumber: 1, Score: 1},
		{Label: "Скорее не похоже", Value: "2", OrderNumber: 2, Score: 2},
		{Label: "Иногда похоже", Value: "3", OrderNumber: 3, Score: 3},
		{Label: "Скорее похоже", Value: "4", OrderNumber: 4, Score: 4},
		{Label: "Очень похоже на меня", Value: "5", OrderNumber: 5, Score: 5},
	}
}

func buildDemoReportTemplateBody() string {
	config := reportTemplateConfig{
		Client: reportAudienceTemplateConfig{
			Title: "Индивидуальный отчет ПрофДНК",
			SectionTitles: map[string]string{
				reportSectionIntro:           "Как читать этот отчет",
				reportSectionSummary:         "Короткий вывод",
				reportSectionChartData:       "График профиля",
				reportSectionScalesList:      "Результаты по шкалам",
				reportSectionInterpretation:  "Что это означает",
				reportSectionRecommendations: "Рекомендации по следующим шагам",
				reportSectionClosing:         "Что делать дальше",
			},
			MetaLabels: map[string]string{
				reportMetaRespondent: "Клиент",
				reportMetaSession:    "Номер сессии",
				reportMetaStatus:     "Статус прохождения",
			},
			IntroParagraphs: []string{
				"Отчет помогает увидеть ваши сильные стороны и возможные направления развития в IT-сфере.",
				"Он не ставит жесткий ярлык, а дает основу для обсуждения с профориентологом и для дальнейших проб.",
			},
			ClosingParagraphs: []string{
				"Лучший эффект дает не только чтение отчета, но и проверка гипотез через реальные мини-проекты, стажировки и разговор с наставником.",
			},
			ChartCaption: "Профиль по пяти шкалам ПрофДНК. Чем выше столбец, тем сильнее выражено направление интересов и способов работы.",
		},
		Psychologist: reportAudienceTemplateConfig{
			Title: "Технический отчет профориентолога",
			SectionTitles: map[string]string{
				reportSectionIntro:        "Контекст",
				reportSectionScalesList:   "Нормализованные шкалы",
				reportSectionRawScores:    "Сырые показатели",
				reportSectionAnswersTable: "Таблица ответов",
			},
			MetaLabels: map[string]string{
				reportMetaRespondent:     "Клиент",
				reportMetaSession:        "Сессия",
				reportMetaStatus:         "Статус",
				reportMetaTopScales:      "Ведущие шкалы",
				reportMetaTopProfessions: "Релевантные профессии",
			},
			IntroParagraphs: []string{
				"Отчет предназначен для внутренней работы специалиста и содержит только технические данные интерпретации.",
			},
		},
	}

	content, _ := json.MarshalIndent(config, "", "  ")
	return string(content)
}

func demoBoolPtr(value bool) *bool {
	return &value
}

func getenvDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
