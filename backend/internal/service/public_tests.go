package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/repository"
)

var ErrPublicTestNotFound = errors.New("public test not found")
var ErrInvalidPublicTestSubmission = errors.New("invalid public test submission")
var ErrPublicTestLimitReached = errors.New("public test limit reached")

func (s *AppService) PublishPsychologistTest(ctx context.Context, userID int64, testID int64) (domain.PublishTestResponse, error) {
	const maxAttempts = 5

	for i := 0; i < maxAttempts; i++ {
		slug, err := generateRandomHex(8)
		if err != nil {
			return domain.PublishTestResponse{}, err
		}

		test, err := s.repo.PublishPsychologistTest(ctx, testID, userID, slug)
		if err != nil {
			if repository.IsNotFound(err) {
				return domain.PublishTestResponse{}, ErrTestNotFound
			}
			if repository.IsUniqueViolation(err) {
				continue
			}
			return domain.PublishTestResponse{}, err
		}

		return domain.PublishTestResponse{
			TestID:     test.ID,
			PublicSlug: test.PublicSlug,
			PublicURL:  publicTestURL(test.PublicSlug),
			IsPublic:   test.IsPublic,
			Status:     test.Status,
		}, nil
	}

	return domain.PublishTestResponse{}, errors.New("failed to generate unique public slug")
}

func (s *AppService) GetPublicTestBySlug(ctx context.Context, slug string) (domain.PublicTest, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return domain.PublicTest{}, ErrPublicTestNotFound
	}

	accessInfo, err := s.repo.GetPublicTestAccessInfoBySlug(ctx, slug)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.PublicTest{}, ErrPublicTestNotFound
		}
		return domain.PublicTest{}, err
	}
	if err := validatePublicTestAccess(accessInfo); err != nil {
		return domain.PublicTest{}, err
	}

	test, err := s.repo.GetPublicTestBySlug(ctx, slug)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.PublicTest{}, ErrPublicTestNotFound
		}
		return domain.PublicTest{}, err
	}

	return test, nil
}

func (s *AppService) StartPublicTest(ctx context.Context, slug string, input domain.StartPublicTestInput) (domain.StartPublicTestResponse, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return domain.StartPublicTestResponse{}, ErrPublicTestNotFound
	}

	accessInfo, err := s.repo.GetPublicTestAccessInfoBySlug(ctx, slug)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.StartPublicTestResponse{}, ErrPublicTestNotFound
		}
		return domain.StartPublicTestResponse{}, err
	}
	if err := validatePublicTestAccess(accessInfo); err != nil {
		return domain.StartPublicTestResponse{}, err
	}

	test, err := s.repo.GetPublicTestBySlug(ctx, slug)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.StartPublicTestResponse{}, ErrPublicTestNotFound
		}
		return domain.StartPublicTestResponse{}, err
	}

	input.RespondentName = strings.TrimSpace(input.RespondentName)
	input.RespondentEmail = strings.TrimSpace(strings.ToLower(input.RespondentEmail))

	const maxAttempts = 5
	for i := 0; i < maxAttempts; i++ {
		accessToken, tokenErr := generateRandomHex(16)
		if tokenErr != nil {
			return domain.StartPublicTestResponse{}, tokenErr
		}

		session, sessionErr := s.repo.StartPublicTestSession(ctx, slug, accessToken, input)
		if sessionErr != nil {
			if repository.IsLimitReached(sessionErr) {
				return domain.StartPublicTestResponse{}, ErrPublicTestLimitReached
			}
			if repository.IsNotFound(sessionErr) {
				return domain.StartPublicTestResponse{}, ErrPublicTestNotFound
			}
			if repository.IsUniqueViolation(sessionErr) {
				continue
			}
			return domain.StartPublicTestResponse{}, sessionErr
		}

		return domain.StartPublicTestResponse{
			Session: session,
			Test:    test,
		}, nil
	}

	return domain.StartPublicTestResponse{}, errors.New("failed to generate public session token")
}

func (s *AppService) SubmitPublicTest(ctx context.Context, slug string, input domain.SubmitPublicTestInput) (domain.SubmitPublicTestResponse, error) {
	slug = strings.TrimSpace(slug)
	input.AccessToken = strings.TrimSpace(input.AccessToken)
	if slug == "" || input.AccessToken == "" {
		return domain.SubmitPublicTestResponse{}, ErrPublicTestNotFound
	}

	test, err := s.repo.GetPublicTestBySlugAndAccessToken(ctx, slug, input.AccessToken)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.SubmitPublicTestResponse{}, ErrPublicTestNotFound
		}
		return domain.SubmitPublicTestResponse{}, err
	}

	normalizedAnswers, err := normalizePublicSubmission(test, input)
	if err != nil {
		return domain.SubmitPublicTestResponse{}, err
	}

	response, err := s.repo.SubmitPublicTestAnswers(ctx, slug, input.AccessToken, normalizedAnswers)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.SubmitPublicTestResponse{}, ErrPublicTestNotFound
		}
		return domain.SubmitPublicTestResponse{}, err
	}

	return response, nil
}

func publicTestURL(slug string) string {
	if strings.TrimSpace(slug) == "" {
		return ""
	}

	baseURL := strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL"))
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("APP_BASE_URL"))
	}
	if baseURL == "" {
		return "/public/tests/" + slug
	}

	return strings.TrimRight(baseURL, "/") + "/public/tests/" + slug
}

func generateRandomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

func (s *AppService) ListPsychologistTestSubmissions(ctx context.Context, userID int64, testID int64) ([]domain.PsychologistTestSubmission, error) {
	submissions, err := s.repo.ListPsychologistTestSubmissions(ctx, testID, userID)
	if err != nil {
		return nil, err
	}

	return submissions, nil
}

func (s *AppService) GetPsychologistTestSubmissionByID(ctx context.Context, userID int64, testID int64, sessionID int64) (domain.PsychologistTestSubmission, error) {
	submission, err := s.repo.GetPsychologistTestSubmissionByID(ctx, testID, sessionID, userID)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.PsychologistTestSubmission{}, ErrTestNotFound
		}
		return domain.PsychologistTestSubmission{}, err
	}

	return submission, nil
}

func validatePublicTestAccess(info domain.PublicTestAccessInfo) error {
	if !info.IsPublic || info.Status != domain.TestStatusPublished {
		return ErrPublicTestNotFound
	}
	if info.MaxParticipants > 0 && info.CurrentSessions >= info.MaxParticipants {
		return ErrPublicTestLimitReached
	}

	return nil
}

func normalizePublicSubmission(test domain.PublicTest, input domain.SubmitPublicTestInput) ([]domain.PublicAnswerInput, error) {
	if input.AccessToken == "" || len(input.Answers) == 0 {
		return nil, ErrInvalidPublicTestSubmission
	}

	questionByID := make(map[int64]domain.PublicQuestion, len(test.Questions))
	requiredQuestionIDs := make(map[int64]struct{})
	for _, question := range test.Questions {
		questionByID[question.ID] = question
		if question.IsRequired {
			requiredQuestionIDs[question.ID] = struct{}{}
		}
	}

	result := make([]domain.PublicAnswerInput, 0, len(input.Answers))
	seenQuestionIDs := make(map[int64]struct{}, len(input.Answers))
	for _, answer := range input.Answers {
		question, ok := questionByID[answer.QuestionID]
		if !ok {
			return nil, ErrInvalidPublicTestSubmission
		}
		if _, exists := seenQuestionIDs[answer.QuestionID]; exists {
			return nil, ErrInvalidPublicTestSubmission
		}

		normalized, err := normalizePublicAnswer(question, answer)
		if err != nil {
			return nil, err
		}

		seenQuestionIDs[answer.QuestionID] = struct{}{}
		delete(requiredQuestionIDs, answer.QuestionID)
		result = append(result, normalized)
	}

	if len(requiredQuestionIDs) > 0 {
		return nil, ErrInvalidPublicTestSubmission
	}

	return result, nil
}

func normalizePublicAnswer(question domain.PublicQuestion, answer domain.PublicAnswerInput) (domain.PublicAnswerInput, error) {
	answer.AnswerText = strings.TrimSpace(answer.AnswerText)
	answer.AnswerValue = strings.TrimSpace(answer.AnswerValue)

	cleanValues := make([]string, 0, len(answer.AnswerValues))
	for _, value := range answer.AnswerValues {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		cleanValues = append(cleanValues, trimmed)
	}
	answer.AnswerValues = cleanValues

	allowedOptionValues := make(map[string]struct{}, len(question.Options))
	for _, option := range question.Options {
		allowedOptionValues[option.Value] = struct{}{}
	}

	switch question.QuestionType {
	case domain.QuestionTypeText:
		if answer.AnswerText == "" {
			return domain.PublicAnswerInput{}, ErrInvalidPublicTestSubmission
		}
		answer.AnswerValue = ""
		answer.AnswerValues = nil
	case domain.QuestionTypeNumber:
		if answer.AnswerValue == "" {
			return domain.PublicAnswerInput{}, ErrInvalidPublicTestSubmission
		}
		if _, parseErr := strconv.ParseFloat(answer.AnswerValue, 64); parseErr != nil {
			return domain.PublicAnswerInput{}, ErrInvalidPublicTestSubmission
		}
		answer.AnswerText = ""
		answer.AnswerValues = nil
	case domain.QuestionTypeSingleChoice, domain.QuestionTypeScale:
		if answer.AnswerValue == "" {
			return domain.PublicAnswerInput{}, ErrInvalidPublicTestSubmission
		}
		if _, ok := allowedOptionValues[answer.AnswerValue]; !ok {
			return domain.PublicAnswerInput{}, ErrInvalidPublicTestSubmission
		}
		answer.AnswerText = ""
		answer.AnswerValues = nil
	case domain.QuestionTypeMultiple:
		if len(answer.AnswerValues) == 0 {
			return domain.PublicAnswerInput{}, ErrInvalidPublicTestSubmission
		}
		seenValues := make(map[string]struct{}, len(answer.AnswerValues))
		for _, value := range answer.AnswerValues {
			if _, ok := allowedOptionValues[value]; !ok {
				return domain.PublicAnswerInput{}, ErrInvalidPublicTestSubmission
			}
			if _, exists := seenValues[value]; exists {
				return domain.PublicAnswerInput{}, ErrInvalidPublicTestSubmission
			}
			seenValues[value] = struct{}{}
		}
		answer.AnswerText = ""
		answer.AnswerValue = ""
	default:
		return domain.PublicAnswerInput{}, ErrInvalidPublicTestSubmission
	}

	return answer, nil
}
