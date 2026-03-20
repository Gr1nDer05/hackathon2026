package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/repository"
)

var (
	ErrInvalidQuestionInput = errors.New("invalid question input")
	ErrQuestionNotFound     = errors.New("question not found")
)

func (s *AppService) CreatePsychologistQuestion(ctx context.Context, userID int64, testID int64, input domain.CreateQuestionInput) (domain.Question, error) {
	normalized, err := normalizeCreateQuestionInput(input)
	if err != nil {
		return domain.Question{}, err
	}

	question, err := s.repo.CreatePsychologistQuestion(ctx, testID, userID, normalized)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.Question{}, ErrTestNotFound
		}
		return domain.Question{}, err
	}

	return question, nil
}

func (s *AppService) ListPsychologistQuestions(ctx context.Context, userID int64, testID int64) ([]domain.Question, error) {
	return s.repo.ListPsychologistQuestions(ctx, testID, userID)
}

func (s *AppService) GetPsychologistQuestionByID(ctx context.Context, userID int64, testID int64, questionID int64) (domain.Question, error) {
	question, err := s.repo.GetPsychologistQuestionByID(ctx, testID, questionID, userID)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.Question{}, ErrQuestionNotFound
		}
		return domain.Question{}, err
	}

	return question, nil
}

func (s *AppService) UpdatePsychologistQuestion(ctx context.Context, userID int64, testID int64, questionID int64, input domain.UpdateQuestionInput) (domain.Question, error) {
	normalized, err := normalizeUpdateQuestionInput(input)
	if err != nil {
		return domain.Question{}, err
	}

	question, err := s.repo.UpdatePsychologistQuestion(ctx, testID, questionID, userID, normalized)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.Question{}, ErrQuestionNotFound
		}
		return domain.Question{}, err
	}

	return question, nil
}

func (s *AppService) DeletePsychologistQuestion(ctx context.Context, userID int64, testID int64, questionID int64) error {
	deleted, err := s.repo.DeletePsychologistQuestion(ctx, testID, questionID, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrQuestionNotFound
	}

	return nil
}

func normalizeCreateQuestionInput(input domain.CreateQuestionInput) (domain.CreateQuestionInput, error) {
	input.Text = strings.TrimSpace(input.Text)
	input.QuestionType = normalizeQuestionType(input.QuestionType)
	options, err := normalizeQuestionOptions(input.Options)
	if err != nil {
		return domain.CreateQuestionInput{}, err
	}
	input.Options = options

	if !isAllowedQuestionType(input.QuestionType) || input.Text == "" || input.OrderNumber < 0 {
		return domain.CreateQuestionInput{}, ErrInvalidQuestionInput
	}
	if requiresOptions(input.QuestionType) && len(input.Options) == 0 {
		return domain.CreateQuestionInput{}, ErrInvalidQuestionInput
	}

	return input, nil
}

func normalizeUpdateQuestionInput(input domain.UpdateQuestionInput) (domain.UpdateQuestionInput, error) {
	input.Text = strings.TrimSpace(input.Text)
	input.QuestionType = normalizeQuestionType(input.QuestionType)
	options, err := normalizeQuestionOptions(input.Options)
	if err != nil {
		return domain.UpdateQuestionInput{}, err
	}
	input.Options = options

	if !isAllowedQuestionType(input.QuestionType) || input.Text == "" || input.OrderNumber < 0 {
		return domain.UpdateQuestionInput{}, ErrInvalidQuestionInput
	}
	if requiresOptions(input.QuestionType) && len(input.Options) == 0 {
		return domain.UpdateQuestionInput{}, ErrInvalidQuestionInput
	}

	return input, nil
}

func normalizeQuestionOptions(options []domain.QuestionOptionInput) ([]domain.QuestionOptionInput, error) {
	result := make([]domain.QuestionOptionInput, 0, len(options))
	seenValues := make(map[string]struct{}, len(options))

	for i, option := range options {
		option.Label = strings.TrimSpace(option.Label)
		option.Value = strings.TrimSpace(option.Value)
		if option.OrderNumber < 0 {
			return nil, ErrInvalidQuestionInput
		}
		if option.OrderNumber == 0 {
			option.OrderNumber = i + 1
		}
		if option.Label == "" || option.Value == "" {
			return nil, ErrInvalidQuestionInput
		}
		if _, exists := seenValues[option.Value]; exists {
			return nil, ErrInvalidQuestionInput
		}
		seenValues[option.Value] = struct{}{}
		result = append(result, option)
	}

	return result, nil
}

func normalizeQuestionType(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

func isAllowedQuestionType(questionType string) bool {
	switch questionType {
	case domain.QuestionTypeSingleChoice, domain.QuestionTypeMultiple, domain.QuestionTypeScale, domain.QuestionTypeText, domain.QuestionTypeNumber:
		return true
	default:
		return false
	}
}

func requiresOptions(questionType string) bool {
	switch questionType {
	case domain.QuestionTypeSingleChoice, domain.QuestionTypeMultiple, domain.QuestionTypeScale:
		return true
	default:
		return false
	}
}
