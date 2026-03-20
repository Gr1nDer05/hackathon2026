package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/repository"
)

var (
	ErrInvalidFormulaRuleInput = errors.New("invalid formula rule input")
	ErrFormulaRuleNotFound     = errors.New("formula rule not found")
)

func (s *AppService) CreateFormulaRule(ctx context.Context, userID int64, testID int64, input domain.CreateFormulaRuleInput) (domain.FormulaRule, error) {
	normalized, err := normalizeCreateFormulaRuleInput(input)
	if err != nil {
		return domain.FormulaRule{}, err
	}

	rule, err := s.repo.CreateFormulaRule(ctx, testID, userID, normalized)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.FormulaRule{}, ErrTestNotFound
		}
		return domain.FormulaRule{}, err
	}

	return rule, nil
}

func (s *AppService) ListFormulaRules(ctx context.Context, userID int64, testID int64) ([]domain.FormulaRule, error) {
	return s.repo.ListFormulaRules(ctx, testID, userID)
}

func (s *AppService) GetFormulaRuleByID(ctx context.Context, userID int64, testID int64, ruleID int64) (domain.FormulaRule, error) {
	rule, err := s.repo.GetFormulaRuleByID(ctx, testID, ruleID, userID)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.FormulaRule{}, ErrFormulaRuleNotFound
		}
		return domain.FormulaRule{}, err
	}

	return rule, nil
}

func (s *AppService) UpdateFormulaRule(ctx context.Context, userID int64, testID int64, ruleID int64, input domain.UpdateFormulaRuleInput) (domain.FormulaRule, error) {
	normalized, err := normalizeUpdateFormulaRuleInput(input)
	if err != nil {
		return domain.FormulaRule{}, err
	}

	rule, err := s.repo.UpdateFormulaRule(ctx, testID, ruleID, userID, normalized)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.FormulaRule{}, ErrFormulaRuleNotFound
		}
		return domain.FormulaRule{}, err
	}

	return rule, nil
}

func (s *AppService) DeleteFormulaRule(ctx context.Context, userID int64, testID int64, ruleID int64) error {
	deleted, err := s.repo.DeleteFormulaRule(ctx, testID, ruleID, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrFormulaRuleNotFound
	}

	return nil
}

func (s *AppService) CalculateFormulaPreview(ctx context.Context, userID int64, testID int64, input domain.CalculateFormulaInput) (domain.CalculateFormulaResponse, error) {
	questions, err := s.repo.ListPsychologistQuestions(ctx, testID, userID)
	if err != nil {
		return domain.CalculateFormulaResponse{}, err
	}
	rules, err := s.repo.ListFormulaRules(ctx, testID, userID)
	if err != nil {
		return domain.CalculateFormulaResponse{}, err
	}

	answerValuesByQuestionID := normalizeAnswers(input.Answers)
	optionScoreMap := buildOptionScoreMap(questions)

	response := domain.CalculateFormulaResponse{
		Metrics:          map[string]float64{"total": 0},
		TriggeredRuleIDs: make([]int64, 0),
	}

	for questionID, answerValues := range answerValuesByQuestionID {
		for _, answerValue := range answerValues {
			response.TotalScore += optionScoreMap[questionID][answerValue]
		}
	}
	response.Metrics["total"] = response.TotalScore

	for _, rule := range rules {
		if ruleTriggered(rule, answerValuesByQuestionID[rule.QuestionID]) {
			response.TotalScore += rule.ScoreDelta
			resultKey := strings.TrimSpace(rule.ResultKey)
			if resultKey == "" {
				resultKey = "total"
			}
			response.Metrics[resultKey] += rule.ScoreDelta
			response.Metrics["total"] = response.TotalScore
			response.TriggeredRuleIDs = append(response.TriggeredRuleIDs, rule.ID)
		}
	}

	return response, nil
}

func normalizeCreateFormulaRuleInput(input domain.CreateFormulaRuleInput) (domain.CreateFormulaRuleInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.ConditionType = normalizeConditionType(input.ConditionType)
	input.ExpectedValue = strings.TrimSpace(input.ExpectedValue)
	input.ResultKey = strings.TrimSpace(input.ResultKey)

	if input.ResultKey == "" {
		input.ResultKey = "total"
	}

	if input.Name == "" || !isAllowedConditionType(input.ConditionType) {
		return domain.CreateFormulaRuleInput{}, ErrInvalidFormulaRuleInput
	}

	return input, nil
}

func normalizeUpdateFormulaRuleInput(input domain.UpdateFormulaRuleInput) (domain.UpdateFormulaRuleInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.ConditionType = normalizeConditionType(input.ConditionType)
	input.ExpectedValue = strings.TrimSpace(input.ExpectedValue)
	input.ResultKey = strings.TrimSpace(input.ResultKey)

	if input.ResultKey == "" {
		input.ResultKey = "total"
	}

	if input.Name == "" || !isAllowedConditionType(input.ConditionType) {
		return domain.UpdateFormulaRuleInput{}, ErrInvalidFormulaRuleInput
	}

	return input, nil
}

func normalizeConditionType(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

func isAllowedConditionType(conditionType string) bool {
	switch conditionType {
	case domain.FormulaConditionAlways, domain.FormulaConditionAnswerEquals, domain.FormulaConditionAnswerIn, domain.FormulaConditionAnswerNumericG, domain.FormulaConditionAnswerNumericL:
		return true
	default:
		return false
	}
}

func normalizeAnswers(answers []domain.FormulaAnswerInput) map[int64][]string {
	result := make(map[int64][]string, len(answers))

	for _, answer := range answers {
		values := make([]string, 0, 1+len(answer.AnswerValues))
		for _, value := range answer.AnswerValues {
			trimmed := strings.TrimSpace(value)
			if trimmed != "" {
				values = append(values, trimmed)
			}
		}

		single := strings.TrimSpace(answer.AnswerValue)
		if single != "" {
			values = append(values, single)
		}

		if len(values) > 0 {
			result[answer.QuestionID] = values
		}
	}

	return result
}

func buildOptionScoreMap(questions []domain.Question) map[int64]map[string]float64 {
	result := make(map[int64]map[string]float64, len(questions))
	for _, question := range questions {
		if len(question.Options) == 0 {
			continue
		}

		optionScores := make(map[string]float64, len(question.Options))
		for _, option := range question.Options {
			optionScores[option.Value] = option.Score
		}
		result[question.ID] = optionScores
	}

	return result
}

func ruleTriggered(rule domain.FormulaRule, answerValues []string) bool {
	switch rule.ConditionType {
	case domain.FormulaConditionAlways:
		return true
	case domain.FormulaConditionAnswerEquals:
		expected := strings.TrimSpace(rule.ExpectedValue)
		for _, value := range answerValues {
			if value == expected {
				return true
			}
		}
		return false
	case domain.FormulaConditionAnswerIn:
		rawExpected := strings.Split(rule.ExpectedValue, ",")
		expected := make(map[string]struct{}, len(rawExpected))
		for _, item := range rawExpected {
			trimmed := strings.TrimSpace(item)
			if trimmed != "" {
				expected[trimmed] = struct{}{}
			}
		}
		for _, value := range answerValues {
			if _, ok := expected[value]; ok {
				return true
			}
		}
		return false
	case domain.FormulaConditionAnswerNumericG:
		answerNumber, ok := parseFirstNumber(answerValues)
		if !ok {
			return false
		}
		expected, err := strconv.ParseFloat(strings.TrimSpace(rule.ExpectedValue), 64)
		if err != nil {
			return false
		}
		return answerNumber >= expected
	case domain.FormulaConditionAnswerNumericL:
		answerNumber, ok := parseFirstNumber(answerValues)
		if !ok {
			return false
		}
		expected, err := strconv.ParseFloat(strings.TrimSpace(rule.ExpectedValue), 64)
		if err != nil {
			return false
		}
		return answerNumber <= expected
	default:
		return false
	}
}

func parseFirstNumber(values []string) (float64, bool) {
	for _, value := range values {
		number, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err == nil {
			return number, true
		}
	}

	return 0, false
}
