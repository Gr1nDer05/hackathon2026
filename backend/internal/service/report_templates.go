package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/repository"
)

var (
	ErrInvalidReportTemplateInput = errors.New("invalid report template input")
	ErrReportTemplateNotFound     = errors.New("report template not found")
)

type reportAudience string

const (
	reportAudienceClient       reportAudience = "client"
	reportAudiencePsychologist reportAudience = "psychologist"
)

type reportTemplateConfig struct {
	Client       reportAudienceTemplateConfig `json:"client,omitempty"`
	Psychologist reportAudienceTemplateConfig `json:"psychologist,omitempty"`
}

type reportAudienceTemplateConfig struct {
	Title             string            `json:"title,omitempty"`
	MetaLabels        map[string]string `json:"meta_labels,omitempty"`
	SectionTitles     map[string]string `json:"section_titles,omitempty"`
	IntroParagraphs   []string          `json:"intro_paragraphs,omitempty"`
	ClosingParagraphs []string          `json:"closing_paragraphs,omitempty"`
	ChartCaption      string            `json:"chart_caption,omitempty"`
}

func (s *AppService) CreateReportTemplate(ctx context.Context, userID int64, input domain.CreateReportTemplateInput) (domain.ReportTemplate, error) {
	normalized, err := normalizeCreateReportTemplateInput(input)
	if err != nil {
		return domain.ReportTemplate{}, err
	}

	template, err := s.repo.CreateReportTemplate(ctx, userID, normalized)
	if err != nil {
		return domain.ReportTemplate{}, err
	}

	return template, nil
}

func (s *AppService) ListReportTemplates(ctx context.Context, userID int64) ([]domain.ReportTemplate, error) {
	return s.repo.ListReportTemplates(ctx, userID)
}

func (s *AppService) GetReportTemplateByID(ctx context.Context, userID int64, templateID int64) (domain.ReportTemplate, error) {
	template, err := s.repo.GetReportTemplateByID(ctx, templateID, userID)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.ReportTemplate{}, ErrReportTemplateNotFound
		}
		return domain.ReportTemplate{}, err
	}

	return template, nil
}

func (s *AppService) UpdateReportTemplate(ctx context.Context, userID int64, templateID int64, input domain.UpdateReportTemplateInput) (domain.ReportTemplate, error) {
	normalized, err := normalizeUpdateReportTemplateInput(input)
	if err != nil {
		return domain.ReportTemplate{}, err
	}

	template, err := s.repo.UpdateReportTemplate(ctx, templateID, userID, normalized)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.ReportTemplate{}, ErrReportTemplateNotFound
		}
		return domain.ReportTemplate{}, err
	}

	return template, nil
}

func (s *AppService) DeleteReportTemplate(ctx context.Context, userID int64, templateID int64) error {
	deleted, err := s.repo.DeleteReportTemplate(ctx, templateID, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrReportTemplateNotFound
	}

	return nil
}

func (s *AppService) ensureReportTemplateAccessible(ctx context.Context, userID int64, templateID int64) error {
	if templateID == 0 {
		return nil
	}

	if _, err := s.repo.GetReportTemplateByID(ctx, templateID, userID); err != nil {
		if repository.IsNotFound(err) {
			return ErrReportTemplateNotFound
		}
		return err
	}

	return nil
}

func normalizeCreateReportTemplateInput(input domain.CreateReportTemplateInput) (domain.CreateReportTemplateInput, error) {
	configBody, err := normalizeReportTemplateBody(input.TemplateBody)
	if err != nil {
		return domain.CreateReportTemplateInput{}, err
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.TemplateBody = configBody
	if input.Name == "" {
		return domain.CreateReportTemplateInput{}, ErrInvalidReportTemplateInput
	}

	return input, nil
}

func normalizeUpdateReportTemplateInput(input domain.UpdateReportTemplateInput) (domain.UpdateReportTemplateInput, error) {
	configBody, err := normalizeReportTemplateBody(input.TemplateBody)
	if err != nil {
		return domain.UpdateReportTemplateInput{}, err
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.TemplateBody = configBody
	if input.Name == "" {
		return domain.UpdateReportTemplateInput{}, ErrInvalidReportTemplateInput
	}

	return input, nil
}

func normalizeReportTemplateBody(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidReportTemplateInput
	}

	var config reportTemplateConfig
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return "", ErrInvalidReportTemplateInput
	}

	normalizeReportAudienceTemplateConfig(&config.Client)
	normalizeReportAudienceTemplateConfig(&config.Psychologist)
	if err := validateReportTemplateConfig(config); err != nil {
		return "", err
	}

	normalized, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}

	return string(normalized), nil
}

func normalizeReportAudienceTemplateConfig(config *reportAudienceTemplateConfig) {
	if config == nil {
		return
	}

	config.Title = strings.TrimSpace(config.Title)
	config.ChartCaption = strings.TrimSpace(config.ChartCaption)

	if len(config.MetaLabels) > 0 {
		normalized := make(map[string]string, len(config.MetaLabels))
		for key, value := range config.MetaLabels {
			trimmedKey := strings.TrimSpace(strings.ToLower(key))
			trimmedValue := strings.TrimSpace(value)
			if trimmedKey == "" || trimmedValue == "" {
				continue
			}
			normalized[trimmedKey] = trimmedValue
		}
		config.MetaLabels = normalized
	}

	if len(config.SectionTitles) > 0 {
		normalized := make(map[string]string, len(config.SectionTitles))
		for key, value := range config.SectionTitles {
			trimmedKey := strings.TrimSpace(strings.ToLower(key))
			trimmedValue := strings.TrimSpace(value)
			if trimmedKey == "" || trimmedValue == "" {
				continue
			}
			normalized[trimmedKey] = trimmedValue
		}
		config.SectionTitles = normalized
	}

	config.IntroParagraphs = normalizeTemplateParagraphs(config.IntroParagraphs)
	config.ClosingParagraphs = normalizeTemplateParagraphs(config.ClosingParagraphs)
}

func normalizeTemplateParagraphs(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func validateReportTemplateConfig(config reportTemplateConfig) error {
	if err := validateReportAudienceTemplateConfig(config.Client, reportAudienceClient); err != nil {
		return err
	}
	if err := validateReportAudienceTemplateConfig(config.Psychologist, reportAudiencePsychologist); err != nil {
		return err
	}
	return nil
}

func validateReportAudienceTemplateConfig(config reportAudienceTemplateConfig, audience reportAudience) error {
	allowedSections := allowedReportSectionKeys(audience)
	for key := range config.SectionTitles {
		if _, ok := allowedSections[key]; !ok {
			return ErrInvalidReportTemplateInput
		}
	}

	allowedMeta := allowedReportMetaKeys(audience)
	for key := range config.MetaLabels {
		if _, ok := allowedMeta[key]; !ok {
			return ErrInvalidReportTemplateInput
		}
	}

	return nil
}

func parseReportTemplateBody(raw string) (reportTemplateConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return reportTemplateConfig{}, nil
	}

	var config reportTemplateConfig
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return reportTemplateConfig{}, err
	}
	return config, nil
}

func applyReportTemplate(document reportDocument, template domain.ReportTemplate, audience reportAudience) (reportDocument, error) {
	config, err := parseReportTemplateBody(template.TemplateBody)
	if err != nil {
		return reportDocument{}, err
	}

	audienceConfig := config.forAudience(audience)
	if audienceConfig.Title != "" {
		document.Title = audienceConfig.Title
	}

	if len(audienceConfig.MetaLabels) > 0 {
		for idx := range document.Meta {
			if label, ok := audienceConfig.MetaLabels[strings.ToLower(document.Meta[idx].Key)]; ok {
				document.Meta[idx].Label = label
			}
		}
	}

	if len(audienceConfig.SectionTitles) > 0 {
		for idx := range document.Sections {
			if title, ok := audienceConfig.SectionTitles[strings.ToLower(document.Sections[idx].Key)]; ok {
				document.Sections[idx].Title = title
			}
		}
	}

	if audienceConfig.ChartCaption != "" {
		for idx := range document.Sections {
			if document.Sections[idx].Key == reportSectionChartData {
				document.Sections[idx].ChartCaption = audienceConfig.ChartCaption
			}
		}
	}

	if len(audienceConfig.IntroParagraphs) > 0 {
		introSection := reportSection{
			Key:        reportSectionIntro,
			Title:      "Введение",
			Paragraphs: append([]string(nil), audienceConfig.IntroParagraphs...),
		}
		if title, ok := audienceConfig.SectionTitles[reportSectionIntro]; ok {
			introSection.Title = title
		}
		document.Sections = append([]reportSection{introSection}, document.Sections...)
	}

	if len(audienceConfig.ClosingParagraphs) > 0 {
		closingSection := reportSection{
			Key:        reportSectionClosing,
			Title:      "Заключение",
			Paragraphs: append([]string(nil), audienceConfig.ClosingParagraphs...),
		}
		if title, ok := audienceConfig.SectionTitles[reportSectionClosing]; ok {
			closingSection.Title = title
		}
		document.Sections = append(document.Sections, closingSection)
	}

	return document, nil
}

func (c reportTemplateConfig) forAudience(audience reportAudience) reportAudienceTemplateConfig {
	if audience == reportAudiencePsychologist {
		return c.Psychologist
	}
	return c.Client
}

func allowedReportSectionKeys(audience reportAudience) map[string]struct{} {
	keys := map[string]struct{}{
		reportSectionIntro:   {},
		reportSectionClosing: {},
	}

	if audience == reportAudiencePsychologist {
		keys[reportSectionScalesList] = struct{}{}
		keys[reportSectionRawScores] = struct{}{}
		keys[reportSectionAnswersTable] = struct{}{}
		return keys
	}

	keys[reportSectionSummary] = struct{}{}
	keys[reportSectionChartData] = struct{}{}
	keys[reportSectionScalesList] = struct{}{}
	keys[reportSectionInterpretation] = struct{}{}
	keys[reportSectionRecommendations] = struct{}{}
	return keys
}

func allowedReportMetaKeys(audience reportAudience) map[string]struct{} {
	keys := map[string]struct{}{
		reportMetaRespondent: {},
		reportMetaSession:    {},
		reportMetaStatus:     {},
	}
	if audience == reportAudiencePsychologist {
		keys[reportMetaTopScales] = struct{}{}
		keys[reportMetaTopProfessions] = struct{}{}
	}
	return keys
}
