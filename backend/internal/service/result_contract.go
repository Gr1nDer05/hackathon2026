package service

import (
	"context"
	"sort"
	"strings"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

const maxTopMetrics = 3

func (s *AppService) attachResultContractToSubmitResponse(ctx context.Context, test domain.PublicTest, response *domain.SubmitPublicTestResponse, careerResult *domain.CareerResult) error {
	if response == nil {
		return nil
	}

	response.CareerResult = careerResult
	if response.Status != "completed" {
		return nil
	}

	rules, err := s.listPublicResultFormulaRules(ctx, test)
	if err != nil {
		return err
	}

	metrics := buildMetricsFromCareerResult(careerResult)
	applyFormulaRuleMetrics(metrics, rules, response.Answers)
	if len(metrics) == 0 {
		metrics = map[string]float64{
			"total": calculateTotalScoreForPublicTest(test, response.Answers, rules),
		}
	}

	response.Metrics = metrics
	response.TopMetrics = buildTopMetrics(metrics)
	return nil
}

func (s *AppService) attachResultContractToSubmission(ctx context.Context, userID int64, submission *domain.PsychologistTestSubmission) error {
	if submission == nil || submission.Status != "completed" {
		return nil
	}

	questions, err := s.repo.ListPsychologistQuestions(ctx, submission.TestID, userID)
	if err != nil {
		return err
	}
	if err := s.ensureSubmissionCareerResultSnapshot(ctx, submission, questions); err != nil {
		return err
	}

	rules, err := s.repo.ListFormulaRules(ctx, submission.TestID, userID)
	if err != nil {
		return err
	}

	metrics := buildMetricsFromCareerResult(submission.CareerResult)
	applyFormulaRuleMetrics(metrics, rules, submission.Answers)
	if len(metrics) == 0 {
		metrics = map[string]float64{
			"total": calculateTotalScoreForQuestions(questions, submission.Answers, rules),
		}
	}

	submission.Metrics = metrics
	submission.TopMetrics = buildTopMetrics(metrics)
	return nil
}

func (s *AppService) ensureSubmissionCareerResultSnapshot(ctx context.Context, submission *domain.PsychologistTestSubmission, questions []domain.Question) error {
	if submission == nil || submission.Status != "completed" || submission.CareerResult != nil {
		return nil
	}

	submission.CareerResult = calculateCareerResultForQuestions(questions, submission.Answers)
	if submission.CareerResult == nil {
		return nil
	}

	return s.repo.BackfillPublicTestSessionCareerResult(ctx, submission.SessionID, submission.CareerResult)
}

func (s *AppService) listPublicResultFormulaRules(ctx context.Context, test domain.PublicTest) ([]domain.FormulaRule, error) {
	if test.Psychologist.User.ID == 0 {
		return []domain.FormulaRule{}, nil
	}

	return s.repo.ListFormulaRules(ctx, test.ID, test.Psychologist.User.ID)
}

func buildMetricsFromCareerResult(careerResult *domain.CareerResult) map[string]float64 {
	metrics := make(map[string]float64)
	if careerResult == nil {
		return metrics
	}

	for _, scale := range careerResult.Scales {
		metrics[scale.Scale] = roundToTwo(scale.RawScore)
	}

	return metrics
}

func applyFormulaRuleMetrics(metrics map[string]float64, rules []domain.FormulaRule, answers []domain.PublicTestAnswer) {
	answerValuesByQuestionID := answersToFormulaRuleValues(answers)
	for _, rule := range rules {
		if !ruleTriggered(rule, answerValuesByQuestionID[rule.QuestionID]) {
			continue
		}

		resultKey := strings.TrimSpace(rule.ResultKey)
		if resultKey == "" {
			resultKey = "total"
		}
		metrics[resultKey] = roundToTwo(metrics[resultKey] + rule.ScoreDelta)
	}
}

func buildTopMetrics(metrics map[string]float64) []domain.ResultMetric {
	items := make([]domain.ResultMetric, 0, len(metrics))
	for key, value := range metrics {
		if strings.TrimSpace(key) == "" || key == "total" || value == 0 {
			continue
		}
		items = append(items, domain.ResultMetric{
			Key:   key,
			Value: roundToTwo(value),
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Value != items[j].Value {
			return items[i].Value > items[j].Value
		}
		return items[i].Key < items[j].Key
	})

	if len(items) == 0 {
		total, ok := metrics["total"]
		if !ok {
			return nil
		}
		return []domain.ResultMetric{{Key: "total", Value: roundToTwo(total)}}
	}
	if len(items) > maxTopMetrics {
		items = items[:maxTopMetrics]
	}

	return items
}

func answersToFormulaRuleValues(answers []domain.PublicTestAnswer) map[int64][]string {
	result := make(map[int64][]string, len(answers))
	for _, answer := range answers {
		values := make([]string, 0, 1+len(answer.AnswerValues))
		for _, answerValue := range answer.AnswerValues {
			trimmed := strings.TrimSpace(answerValue)
			if trimmed != "" {
				values = append(values, trimmed)
			}
		}

		singleValue := strings.TrimSpace(answer.AnswerValue)
		if singleValue != "" {
			values = append(values, singleValue)
		}
		result[answer.QuestionID] = values
	}

	return result
}

func calculateTotalScoreForPublicTest(test domain.PublicTest, answers []domain.PublicTestAnswer, rules []domain.FormulaRule) float64 {
	return calculateTotalScore(buildCareerQuestionsFromPublicTest(test), answers, rules)
}

func calculateTotalScoreForQuestions(questions []domain.Question, answers []domain.PublicTestAnswer, rules []domain.FormulaRule) float64 {
	return calculateTotalScore(buildCareerQuestionsFromQuestions(questions), answers, rules)
}

func calculateTotalScore(questions []careerQuestion, answers []domain.PublicTestAnswer, rules []domain.FormulaRule) float64 {
	if len(answers) == 0 {
		return 0
	}

	questionByID := make(map[int64]careerQuestion, len(questions))
	for _, question := range questions {
		questionByID[question.ID] = question
	}

	total := 0.0
	for _, answer := range answers {
		question, ok := questionByID[answer.QuestionID]
		if !ok {
			continue
		}
		total += resolvedAnswerValue(question, answer)
	}

	answerValuesByQuestionID := answersToFormulaRuleValues(answers)
	for _, rule := range rules {
		if ruleTriggered(rule, answerValuesByQuestionID[rule.QuestionID]) {
			total += rule.ScoreDelta
		}
	}

	return roundToTwo(total)
}
