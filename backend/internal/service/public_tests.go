package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/repository"
)

var ErrPublicTestNotFound = errors.New("public test not found")
var ErrInvalidPublicTestRespondent = errors.New("invalid public test respondent")
var ErrInvalidPublicTestSubmission = errors.New("invalid public test submission")
var ErrPublicTestAlreadyTaken = errors.New("public test already taken")
var ErrPublicTestLimitReached = errors.New("public test limit reached")

const PublicTestSessionTTL = time.Hour

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

	if err := s.repo.DeleteExpiredPublicTestSessions(ctx); err != nil {
		return domain.PublicTest{}, err
	}

	accessInfo, err := s.repo.GetPublicTestAccessInfoBySlug(ctx, slug)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.PublicTest{}, ErrPublicTestNotFound
		}
		return domain.PublicTest{}, err
	}
	if !accessInfo.IsPublic || accessInfo.Status != domain.TestStatusPublished {
		return domain.PublicTest{}, ErrPublicTestNotFound
	}

	test, err := s.repo.GetPublicTestBySlug(ctx, slug)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.PublicTest{}, ErrPublicTestNotFound
		}
		return domain.PublicTest{}, err
	}
	applyDerivedFieldsToPublicTest(&test)

	return test, nil
}

func (s *AppService) StartPublicTest(ctx context.Context, slug string, input domain.StartPublicTestInput) (domain.StartPublicTestResponse, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return domain.StartPublicTestResponse{}, ErrPublicTestNotFound
	}

	if err := s.repo.DeleteExpiredPublicTestSessions(ctx); err != nil {
		return domain.StartPublicTestResponse{}, err
	}

	test, err := s.repo.GetPublicTestBySlug(ctx, slug)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.StartPublicTestResponse{}, ErrPublicTestNotFound
		}
		return domain.StartPublicTestResponse{}, err
	}
	applyDerivedFieldsToPublicTest(&test)

	input.RespondentName = strings.TrimSpace(input.RespondentName)
	normalizedInput, err := normalizePublicRespondentInput(test, input)
	if err != nil {
		return domain.StartPublicTestResponse{}, err
	}

	existingSession, err := s.repo.GetPublicTestSessionByPhone(ctx, slug, normalizedInput.RespondentPhone)
	if err == nil {
		if existingSession.Status == "completed" {
			return domain.StartPublicTestResponse{}, ErrPublicTestAlreadyTaken
		}
		if err := s.repo.ExtendPublicTestSessionExpiry(ctx, existingSession.ID, time.Now().Add(PublicTestSessionTTL)); err != nil {
			return domain.StartPublicTestResponse{}, err
		}
		existingSession, err = s.repo.GetPublicTestSessionByAccessToken(ctx, slug, existingSession.AccessToken)
		if err != nil {
			return domain.StartPublicTestResponse{}, err
		}
		answers, err := s.repo.ListPublicTestAnswersBySessionID(ctx, existingSession.ID, existingSession.TestID)
		if err != nil {
			return domain.StartPublicTestResponse{}, err
		}

		return domain.StartPublicTestResponse{
			Session: existingSession,
			Test:    test,
			Answers: answers,
			Resumed: true,
		}, nil
	}
	if !repository.IsNotFound(err) {
		return domain.StartPublicTestResponse{}, err
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

	const maxAttempts = 5
	for i := 0; i < maxAttempts; i++ {
		accessToken, tokenErr := generateRandomHex(16)
		if tokenErr != nil {
			return domain.StartPublicTestResponse{}, tokenErr
		}

		session, sessionErr := s.repo.StartPublicTestSession(ctx, slug, accessToken, normalizedInput, time.Now().Add(PublicTestSessionTTL))
		if sessionErr != nil {
			if repository.IsLimitReached(sessionErr) {
				return domain.StartPublicTestResponse{}, ErrPublicTestLimitReached
			}
			if repository.IsNotFound(sessionErr) {
				return domain.StartPublicTestResponse{}, ErrPublicTestNotFound
			}
			if repository.IsDuplicatePublicTestPhone(sessionErr) {
				return domain.StartPublicTestResponse{}, ErrPublicTestAlreadyTaken
			}
			if repository.IsUniqueViolation(sessionErr) {
				continue
			}
			return domain.StartPublicTestResponse{}, sessionErr
		}

		return domain.StartPublicTestResponse{
			Session: session,
			Test:    test,
			Answers: []domain.PublicTestAnswer{},
			Resumed: false,
		}, nil
	}

	return domain.StartPublicTestResponse{}, errors.New("failed to generate public session token")
}

func (s *AppService) SavePublicTestProgress(ctx context.Context, slug string, input domain.SubmitPublicTestInput) (domain.SubmitPublicTestResponse, error) {
	slug = strings.TrimSpace(slug)
	input.AccessToken = strings.TrimSpace(input.AccessToken)
	if slug == "" || input.AccessToken == "" {
		return domain.SubmitPublicTestResponse{}, ErrPublicTestNotFound
	}

	if err := s.repo.DeleteExpiredPublicTestSessions(ctx); err != nil {
		return domain.SubmitPublicTestResponse{}, err
	}

	session, err := s.repo.GetPublicTestSessionByAccessToken(ctx, slug, input.AccessToken)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.SubmitPublicTestResponse{}, ErrPublicTestNotFound
		}
		return domain.SubmitPublicTestResponse{}, err
	}
	if session.Status == "completed" {
		return domain.SubmitPublicTestResponse{}, ErrPublicTestAlreadyTaken
	}

	test, err := s.repo.GetPublicTestBySlugAndAccessToken(ctx, slug, input.AccessToken)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.SubmitPublicTestResponse{}, ErrPublicTestNotFound
		}
		return domain.SubmitPublicTestResponse{}, err
	}

	normalizedAnswers, err := normalizePartialPublicAnswers(test, input.Answers)
	if err != nil {
		return domain.SubmitPublicTestResponse{}, err
	}

	response, err := s.repo.SavePublicTestAnswers(ctx, slug, input.AccessToken, normalizedAnswers, time.Now().Add(PublicTestSessionTTL), false)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.SubmitPublicTestResponse{}, ErrPublicTestNotFound
		}
		return domain.SubmitPublicTestResponse{}, err
	}

	return response, nil
}

func (s *AppService) SubmitPublicTest(ctx context.Context, slug string, input domain.SubmitPublicTestInput) (domain.SubmitPublicTestResponse, error) {
	slug = strings.TrimSpace(slug)
	input.AccessToken = strings.TrimSpace(input.AccessToken)
	if slug == "" || input.AccessToken == "" {
		return domain.SubmitPublicTestResponse{}, ErrPublicTestNotFound
	}

	if err := s.repo.DeleteExpiredPublicTestSessions(ctx); err != nil {
		return domain.SubmitPublicTestResponse{}, err
	}

	session, err := s.repo.GetPublicTestSessionByAccessToken(ctx, slug, input.AccessToken)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.SubmitPublicTestResponse{}, ErrPublicTestNotFound
		}
		return domain.SubmitPublicTestResponse{}, err
	}
	if session.Status == "completed" {
		return domain.SubmitPublicTestResponse{}, ErrPublicTestAlreadyTaken
	}

	test, err := s.repo.GetPublicTestBySlugAndAccessToken(ctx, slug, input.AccessToken)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.SubmitPublicTestResponse{}, ErrPublicTestNotFound
		}
		return domain.SubmitPublicTestResponse{}, err
	}

	normalizedAnswers, err := normalizePartialPublicAnswers(test, input.Answers)
	if err != nil {
		return domain.SubmitPublicTestResponse{}, err
	}

	savedAnswers, err := s.repo.ListPublicTestAnswersBySessionID(ctx, session.ID, session.TestID)
	if err != nil {
		return domain.SubmitPublicTestResponse{}, err
	}

	mergedAnswers := mergePublicAnswers(savedAnswersToInputs(savedAnswers), normalizedAnswers)
	if err := validateRequiredPublicAnswers(test, mergedAnswers); err != nil {
		return domain.SubmitPublicTestResponse{}, err
	}

	response, err := s.repo.SavePublicTestAnswers(ctx, slug, input.AccessToken, mergedAnswers, time.Now(), true)
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

func (s *AppService) GetPsychologistTestSubmissionBySessionID(ctx context.Context, userID int64, sessionID int64) (domain.PsychologistTestSubmission, error) {
	submission, err := s.repo.GetPsychologistTestSubmissionBySessionID(ctx, sessionID, userID)
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

func normalizePublicRespondentInput(test domain.PublicTest, input domain.StartPublicTestInput) (domain.StartPublicTestInput, error) {
	name, err := normalizeRespondentName(input.RespondentName)
	if err != nil {
		return domain.StartPublicTestInput{}, ErrInvalidPublicTestRespondent
	}

	phone, err := normalizeRespondentPhone(input.RespondentPhone)
	if err != nil {
		return domain.StartPublicTestInput{}, ErrInvalidPublicTestRespondent
	}

	input.RespondentName = name
	input.RespondentPhone = phone
	input.RespondentEmail = strings.TrimSpace(strings.ToLower(input.RespondentEmail))
	input.RespondentGender = strings.TrimSpace(input.RespondentGender)
	input.RespondentEducation = strings.TrimSpace(input.RespondentEducation)

	if test.CollectRespondentAge {
		if input.RespondentAge < 1 || input.RespondentAge > 120 {
			return domain.StartPublicTestInput{}, ErrInvalidPublicTestRespondent
		}
	} else {
		input.RespondentAge = 0
	}

	if test.CollectRespondentGender {
		if input.RespondentGender == "" {
			return domain.StartPublicTestInput{}, ErrInvalidPublicTestRespondent
		}
	} else {
		input.RespondentGender = ""
	}

	if test.CollectRespondentEducation {
		if input.RespondentEducation == "" {
			return domain.StartPublicTestInput{}, ErrInvalidPublicTestRespondent
		}
	} else {
		input.RespondentEducation = ""
	}

	return input, nil
}

func normalizePublicSubmission(test domain.PublicTest, input domain.SubmitPublicTestInput) ([]domain.PublicAnswerInput, error) {
	if input.AccessToken == "" {
		return nil, ErrInvalidPublicTestSubmission
	}

	normalized, err := normalizePartialPublicAnswers(test, input.Answers)
	if err != nil {
		return nil, err
	}
	if err := validateRequiredPublicAnswers(test, normalized); err != nil {
		return nil, err
	}

	return normalized, nil
}

func normalizePartialPublicAnswers(test domain.PublicTest, answers []domain.PublicAnswerInput) ([]domain.PublicAnswerInput, error) {
	if len(answers) == 0 {
		return nil, ErrInvalidPublicTestSubmission
	}

	questionByID := make(map[int64]domain.PublicQuestion, len(test.Questions))
	for _, question := range test.Questions {
		questionByID[question.ID] = question
	}

	result := make([]domain.PublicAnswerInput, 0, len(answers))
	seenQuestionIDs := make(map[int64]struct{}, len(answers))
	for _, answer := range answers {
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
		result = append(result, normalized)
	}

	return result, nil
}

func validateRequiredPublicAnswers(test domain.PublicTest, answers []domain.PublicAnswerInput) error {
	requiredQuestionIDs := make(map[int64]struct{})
	for _, question := range test.Questions {
		if question.IsRequired {
			requiredQuestionIDs[question.ID] = struct{}{}
		}
	}

	for _, answer := range answers {
		delete(requiredQuestionIDs, answer.QuestionID)
	}

	if len(requiredQuestionIDs) > 0 {
		return ErrInvalidPublicTestSubmission
	}

	return nil
}

func normalizeRespondentName(raw string) (string, error) {
	parts := strings.Fields(raw)
	if len(parts) < 2 {
		return "", ErrInvalidPublicTestRespondent
	}

	return strings.Join(parts, " "), nil
}

func normalizeRespondentPhone(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", ErrInvalidPublicTestRespondent
	}

	var digits strings.Builder
	for i, r := range value {
		switch {
		case r >= '0' && r <= '9':
			digits.WriteRune(r)
		case r == '+' && i == 0:
		case r == ' ' || r == '-' || r == '(' || r == ')':
		default:
			return "", ErrInvalidPublicTestRespondent
		}
	}

	normalized := digits.String()
	switch {
	case len(normalized) == 10:
		normalized = "7" + normalized
	case len(normalized) == 11 && strings.HasPrefix(normalized, "8"):
		normalized = "7" + normalized[1:]
	case len(normalized) < 11 || len(normalized) > 15:
		return "", ErrInvalidPublicTestRespondent
	}

	return "+" + normalized, nil
}

func savedAnswersToInputs(saved []domain.PublicTestAnswer) []domain.PublicAnswerInput {
	result := make([]domain.PublicAnswerInput, 0, len(saved))
	for _, answer := range saved {
		result = append(result, domain.PublicAnswerInput{
			QuestionID:   answer.QuestionID,
			AnswerText:   answer.AnswerText,
			AnswerValue:  answer.AnswerValue,
			AnswerValues: append([]string(nil), answer.AnswerValues...),
		})
	}

	return result
}

func mergePublicAnswers(existing []domain.PublicAnswerInput, updates []domain.PublicAnswerInput) []domain.PublicAnswerInput {
	merged := make(map[int64]domain.PublicAnswerInput, len(existing)+len(updates))
	for _, answer := range existing {
		merged[answer.QuestionID] = answer
	}
	for _, answer := range updates {
		merged[answer.QuestionID] = answer
	}

	result := make([]domain.PublicAnswerInput, 0, len(merged))
	for _, answer := range merged {
		result = append(result, answer)
	}

	return result
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
