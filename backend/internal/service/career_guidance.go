package service

import (
	"sort"
	"strconv"
	"strings"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

type careerQuestion struct {
	ID           int64
	QuestionType string
	ScaleWeights map[string]float64
	Options      map[string]careerOption
}

type careerOption struct {
	Value string
	Score float64
}

type careerProfessionProfile struct {
	Name    string
	Weights map[string]float64
}

var careerProfessionCatalog = []careerProfessionProfile{
	{Name: "Backend Developer", Weights: map[string]float64{"analytic": 0.95, "practical": 0.8, "organizer": 0.35}},
	{Name: "Frontend Developer", Weights: map[string]float64{"creative": 0.9, "analytic": 0.75, "social": 0.35}},
	{Name: "Fullstack Developer", Weights: map[string]float64{"analytic": 0.9, "practical": 0.8, "creative": 0.45}},
	{Name: "Mobile Developer", Weights: map[string]float64{"analytic": 0.82, "practical": 0.78, "creative": 0.4}},
	{Name: "Game Developer", Weights: map[string]float64{"creative": 0.88, "analytic": 0.72, "practical": 0.62}},
	{Name: "Embedded Engineer", Weights: map[string]float64{"practical": 0.95, "analytic": 0.82, "organizer": 0.22}},
	{Name: "Robotics Engineer", Weights: map[string]float64{"practical": 0.92, "analytic": 0.84, "creative": 0.36}},
	{Name: "DevOps Engineer", Weights: map[string]float64{"practical": 0.92, "analytic": 0.88, "organizer": 0.42}},
	{Name: "Site Reliability Engineer", Weights: map[string]float64{"practical": 0.95, "analytic": 0.85, "organizer": 0.4}},
	{Name: "Cloud Engineer", Weights: map[string]float64{"practical": 0.9, "analytic": 0.82, "organizer": 0.38}},
	{Name: "Solutions Architect", Weights: map[string]float64{"analytic": 0.9, "organizer": 0.72, "social": 0.52}},
	{Name: "QA Engineer", Weights: map[string]float64{"analytic": 0.86, "practical": 0.68, "organizer": 0.4}},
	{Name: "Automation QA Engineer", Weights: map[string]float64{"analytic": 0.88, "practical": 0.78, "organizer": 0.33}},
	{Name: "Cybersecurity Analyst", Weights: map[string]float64{"analytic": 0.94, "practical": 0.73, "organizer": 0.35}},
	{Name: "Penetration Tester", Weights: map[string]float64{"analytic": 0.92, "practical": 0.82, "creative": 0.34}},
	{Name: "Security Engineer", Weights: map[string]float64{"analytic": 0.9, "practical": 0.8, "organizer": 0.32}},
	{Name: "Data Analyst", Weights: map[string]float64{"analytic": 0.96, "organizer": 0.42, "social": 0.28}},
	{Name: "BI Developer", Weights: map[string]float64{"analytic": 0.92, "practical": 0.62, "organizer": 0.44}},
	{Name: "Data Engineer", Weights: map[string]float64{"analytic": 0.91, "practical": 0.85, "organizer": 0.31}},
	{Name: "Data Scientist", Weights: map[string]float64{"analytic": 0.98, "creative": 0.55, "practical": 0.42}},
	{Name: "Machine Learning Engineer", Weights: map[string]float64{"analytic": 0.95, "practical": 0.75, "creative": 0.44}},
	{Name: "AI Engineer", Weights: map[string]float64{"analytic": 0.94, "creative": 0.56, "practical": 0.52}},
	{Name: "MLOps Engineer", Weights: map[string]float64{"analytic": 0.87, "practical": 0.9, "organizer": 0.4}},
	{Name: "Product Manager", Weights: map[string]float64{"organizer": 0.94, "social": 0.82, "analytic": 0.55}},
	{Name: "Project Manager", Weights: map[string]float64{"organizer": 0.96, "social": 0.76, "practical": 0.33}},
	{Name: "Scrum Master", Weights: map[string]float64{"organizer": 0.88, "social": 0.84, "analytic": 0.26}},
	{Name: "Business Analyst", Weights: map[string]float64{"analytic": 0.82, "social": 0.66, "organizer": 0.58}},
	{Name: "System Analyst", Weights: map[string]float64{"analytic": 0.93, "organizer": 0.5, "social": 0.3}},
	{Name: "UX Researcher", Weights: map[string]float64{"social": 0.88, "analytic": 0.62, "creative": 0.55}},
	{Name: "UI/UX Designer", Weights: map[string]float64{"creative": 0.96, "social": 0.54, "analytic": 0.3}},
	{Name: "Product Designer", Weights: map[string]float64{"creative": 0.92, "social": 0.56, "analytic": 0.42}},
	{Name: "Graphic Designer", Weights: map[string]float64{"creative": 0.95, "social": 0.35, "practical": 0.24}},
	{Name: "DevRel Engineer", Weights: map[string]float64{"social": 0.9, "creative": 0.62, "analytic": 0.48}},
	{Name: "Technical Writer", Weights: map[string]float64{"analytic": 0.65, "creative": 0.58, "social": 0.45}},
	{Name: "Database Administrator", Weights: map[string]float64{"analytic": 0.9, "practical": 0.79, "organizer": 0.36}},
	{Name: "ETL Developer", Weights: map[string]float64{"analytic": 0.9, "practical": 0.7, "organizer": 0.34}},
	{Name: "Network Engineer", Weights: map[string]float64{"practical": 0.88, "analytic": 0.75, "organizer": 0.28}},
	{Name: "Blockchain Developer", Weights: map[string]float64{"analytic": 0.91, "practical": 0.76, "creative": 0.32}},
	{Name: "AR/VR Developer", Weights: map[string]float64{"creative": 0.88, "practical": 0.67, "analytic": 0.52}},
	{Name: "Support Engineer", Weights: map[string]float64{"social": 0.72, "practical": 0.62, "analytic": 0.48}},
}

func isAllowedCareerScale(scale string) bool {
	switch strings.TrimSpace(strings.ToLower(scale)) {
	case domain.CareerScaleAnalytic, domain.CareerScaleCreative, domain.CareerScaleSocial, domain.CareerScaleOrganizer, domain.CareerScalePractical:
		return true
	default:
		return false
	}
}

func calculateCareerResultForPublicTest(test domain.PublicTest, answers []domain.PublicTestAnswer) *domain.CareerResult {
	return calculateCareerResult(buildCareerQuestionsFromPublicTest(test), answers)
}

func calculateCareerResultForQuestions(questions []domain.Question, answers []domain.PublicTestAnswer) *domain.CareerResult {
	return calculateCareerResult(buildCareerQuestionsFromQuestions(questions), answers)
}

func calculateCareerResult(questions []careerQuestion, answers []domain.PublicTestAnswer) *domain.CareerResult {
	if len(questions) == 0 || len(answers) == 0 {
		return nil
	}

	questionByID := make(map[int64]careerQuestion, len(questions))
	maxAnswerValues := make(map[int64]float64, len(questions))
	rawScores := make(map[string]float64, len(domain.CareerScales))
	maxScores := make(map[string]float64, len(domain.CareerScales))
	hasWeightedQuestions := false

	for _, scale := range domain.CareerScales {
		rawScores[scale] = 0
		maxScores[scale] = 0
	}

	for _, question := range questions {
		questionByID[question.ID] = question
		if len(question.ScaleWeights) == 0 {
			continue
		}
		maxAnswerValue := maxQuestionAnswerValue(question)
		if maxAnswerValue <= 0 {
			continue
		}
		maxAnswerValues[question.ID] = maxAnswerValue
		hasWeightedQuestions = true
		for scale, weight := range question.ScaleWeights {
			maxScores[scale] += maxAnswerValue * weight
		}
	}

	if !hasWeightedQuestions {
		return nil
	}

	for _, answer := range answers {
		question, ok := questionByID[answer.QuestionID]
		if !ok || len(question.ScaleWeights) == 0 {
			continue
		}
		if maxAnswerValues[question.ID] <= 0 {
			continue
		}

		answerValue := resolvedAnswerValue(question, answer)
		for scale, weight := range question.ScaleWeights {
			rawScores[scale] += answerValue * weight
		}
	}

	scaleResults := make([]domain.CareerScaleResult, 0, len(domain.CareerScales))
	for _, scale := range domain.CareerScales {
		percentage := 0.0
		if maxScores[scale] > 0 {
			percentage = clampPercentage(rawScores[scale] / maxScores[scale] * 100)
		}
		scaleResults = append(scaleResults, domain.CareerScaleResult{
			Scale:      scale,
			RawScore:   roundToTwo(rawScores[scale]),
			MaxScore:   roundToTwo(maxScores[scale]),
			Percentage: roundToTwo(percentage),
		})
	}

	topScales := append([]domain.CareerScaleResult(nil), scaleResults...)
	sort.SliceStable(topScales, func(i, j int) bool {
		if topScales[i].Percentage != topScales[j].Percentage {
			return topScales[i].Percentage > topScales[j].Percentage
		}
		if topScales[i].RawScore != topScales[j].RawScore {
			return topScales[i].RawScore > topScales[j].RawScore
		}
		return careerScaleOrder(topScales[i].Scale) < careerScaleOrder(topScales[j].Scale)
	})
	if len(topScales) > 2 {
		topScales = topScales[:2]
	}

	topProfessions := calculateTopProfessions(scaleResults)

	return &domain.CareerResult{
		Scales:         scaleResults,
		TopScales:      topScales,
		TopProfessions: topProfessions,
	}
}

func buildCareerQuestionsFromQuestions(questions []domain.Question) []careerQuestion {
	result := make([]careerQuestion, 0, len(questions))
	for _, question := range questions {
		options := make(map[string]careerOption, len(question.Options))
		for _, option := range question.Options {
			options[option.Value] = careerOption{
				Value: option.Value,
				Score: option.Score,
			}
		}
		result = append(result, careerQuestion{
			ID:           question.ID,
			QuestionType: question.QuestionType,
			ScaleWeights: question.ScaleWeights,
			Options:      options,
		})
	}

	return result
}

func buildCareerQuestionsFromPublicTest(test domain.PublicTest) []careerQuestion {
	result := make([]careerQuestion, 0, len(test.Questions))
	for _, question := range test.Questions {
		options := make(map[string]careerOption, len(question.Options))
		for _, option := range question.Options {
			options[option.Value] = careerOption{
				Value: option.Value,
				Score: option.Score,
			}
		}
		result = append(result, careerQuestion{
			ID:           question.ID,
			QuestionType: question.QuestionType,
			ScaleWeights: question.ScaleWeights,
			Options:      options,
		})
	}

	return result
}

func maxQuestionAnswerValue(question careerQuestion) float64 {
	if len(question.Options) == 0 {
		return 0
	}

	switch question.QuestionType {
	case domain.QuestionTypeMultiple:
		total := 0.0
		for _, option := range question.Options {
			value := optionNumericValue(option)
			if value > 0 {
				total += value
			}
		}
		return total
	case domain.QuestionTypeSingleChoice, domain.QuestionTypeScale:
		maxValue := 0.0
		for _, option := range question.Options {
			value := optionNumericValue(option)
			if value > maxValue {
				maxValue = value
			}
		}
		return maxValue
	default:
		return 0
	}
}

func resolvedAnswerValue(question careerQuestion, answer domain.PublicTestAnswer) float64 {
	switch question.QuestionType {
	case domain.QuestionTypeMultiple:
		total := 0.0
		for _, value := range answer.AnswerValues {
			total += tokenNumericValue(strings.TrimSpace(value), question.Options)
		}
		return total
	case domain.QuestionTypeSingleChoice, domain.QuestionTypeScale, domain.QuestionTypeNumber:
		return tokenNumericValue(strings.TrimSpace(answer.AnswerValue), question.Options)
	default:
		return 0
	}
}

func tokenNumericValue(token string, options map[string]careerOption) float64 {
	if token == "" {
		return 0
	}

	if parsed, err := strconv.ParseFloat(token, 64); err == nil {
		return parsed
	}

	option, ok := options[token]
	if !ok {
		return 0
	}

	return optionNumericValue(option)
}

func optionNumericValue(option careerOption) float64 {
	if parsed, err := strconv.ParseFloat(strings.TrimSpace(option.Value), 64); err == nil {
		return parsed
	}

	return option.Score
}

func calculateTopProfessions(scaleResults []domain.CareerScaleResult) []domain.CareerProfessionResult {
	scalePercentages := make(map[string]float64, len(scaleResults))
	for _, result := range scaleResults {
		scalePercentages[result.Scale] = result.Percentage
	}

	professions := make([]domain.CareerProfessionResult, 0, len(careerProfessionCatalog))
	for _, profession := range careerProfessionCatalog {
		score := 0.0
		for scale, weight := range profession.Weights {
			score += scalePercentages[scale] * weight
		}
		professions = append(professions, domain.CareerProfessionResult{
			Profession: profession.Name,
			Score:      roundToTwo(score),
		})
	}

	sort.SliceStable(professions, func(i, j int) bool {
		if professions[i].Score != professions[j].Score {
			return professions[i].Score > professions[j].Score
		}
		return professions[i].Profession < professions[j].Profession
	})
	if len(professions) > 3 {
		professions = professions[:3]
	}

	return professions
}

func careerScaleOrder(scale string) int {
	for idx, candidate := range domain.CareerScales {
		if scale == candidate {
			return idx
		}
	}

	return len(domain.CareerScales)
}

func clampPercentage(value float64) float64 {
	switch {
	case value < 0:
		return 0
	case value > 100:
		return 100
	default:
		return value
	}
}

func roundToTwo(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
