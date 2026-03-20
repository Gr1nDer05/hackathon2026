package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/repository"
)

var (
	ErrInvalidTestInput = errors.New("invalid test input")
	ErrTestNotFound     = errors.New("test not found")
)

func (s *AppService) CreatePsychologistTest(ctx context.Context, userID int64, input domain.CreateTestInput) (domain.Test, error) {
	normalizedInput, err := normalizeTestCreateInput(input)
	if err != nil {
		return domain.Test{}, err
	}

	const maxAttempts = 5
	for i := 0; i < maxAttempts; i++ {
		publicSlug, slugErr := generateRandomHex(8)
		if slugErr != nil {
			return domain.Test{}, slugErr
		}

		test, createErr := s.repo.CreateTest(ctx, userID, normalizedInput, publicSlug)
		if createErr != nil {
			if repository.IsUniqueViolation(createErr) {
				continue
			}
			return domain.Test{}, createErr
		}

		test.PublicURL = publicTestURL(test.PublicSlug)
		return test, nil
	}

	return domain.Test{}, errors.New("failed to generate unique public slug")
}

func (s *AppService) ListPsychologistTests(ctx context.Context, userID int64) ([]domain.Test, error) {
	tests, err := s.repo.ListPsychologistTests(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range tests {
		tests[i].PublicURL = publicTestURL(tests[i].PublicSlug)
	}

	return tests, nil
}

func (s *AppService) GetPsychologistTestByID(ctx context.Context, userID int64, testID int64) (domain.Test, error) {
	test, err := s.repo.GetPsychologistTestByID(ctx, testID, userID)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.Test{}, ErrTestNotFound
		}
		return domain.Test{}, err
	}

	test.PublicURL = publicTestURL(test.PublicSlug)
	return test, nil
}

func (s *AppService) UpdatePsychologistTest(ctx context.Context, userID int64, testID int64, input domain.UpdateTestInput) (domain.Test, error) {
	normalizedInput, err := normalizeTestUpdateInput(input)
	if err != nil {
		return domain.Test{}, err
	}

	test, err := s.repo.UpdatePsychologistTest(ctx, testID, userID, normalizedInput)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.Test{}, ErrTestNotFound
		}
		return domain.Test{}, err
	}

	test.PublicURL = publicTestURL(test.PublicSlug)
	return test, nil
}

func (s *AppService) DeletePsychologistTest(ctx context.Context, userID int64, testID int64) error {
	deleted, err := s.repo.DeletePsychologistTest(ctx, testID, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrTestNotFound
	}

	return nil
}

func normalizeTestCreateInput(input domain.CreateTestInput) (domain.CreateTestInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Status = normalizeCreateTestStatus(input.Status)

	if input.Title == "" || input.RecommendedDuration < 0 || input.MaxParticipants < 0 {
		return domain.CreateTestInput{}, ErrInvalidTestInput
	}
	if !isAllowedTestStatus(input.Status) {
		return domain.CreateTestInput{}, ErrInvalidTestInput
	}

	return input, nil
}

func normalizeTestUpdateInput(input domain.UpdateTestInput) (domain.UpdateTestInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Status = normalizeTestStatus(input.Status)

	if input.Title == "" || input.RecommendedDuration < 0 || input.MaxParticipants < 0 {
		return domain.UpdateTestInput{}, ErrInvalidTestInput
	}
	if !isAllowedTestStatus(input.Status) {
		return domain.UpdateTestInput{}, ErrInvalidTestInput
	}

	return input, nil
}

func normalizeCreateTestStatus(raw string) string {
	status := strings.TrimSpace(strings.ToLower(raw))
	if status == "" {
		return domain.TestStatusPublished
	}

	return status
}

func normalizeTestStatus(raw string) string {
	status := strings.TrimSpace(strings.ToLower(raw))
	if status == "" {
		return domain.TestStatusDraft
	}

	return status
}

func isAllowedTestStatus(status string) bool {
	return status == domain.TestStatusDraft || status == domain.TestStatusPublished
}
