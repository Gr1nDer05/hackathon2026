package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUnauthorized              = errors.New("unauthorized")
	ErrForbidden                 = errors.New("forbidden")
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrEmailAlreadyExists        = errors.New("email already exists")
	ErrAccountDisabled           = errors.New("account disabled")
	ErrPortalAccessExpired       = errors.New("portal access expired")
	ErrAccountTemporarilyBlocked = errors.New("account temporarily blocked")
	ErrInvalidPsychologistAccess = errors.New("invalid psychologist access settings")
)

const SessionTTL = 24 * time.Hour

func (s *AppService) LoginPsychologist(ctx context.Context, input domain.PsychologistLoginInput) (string, domain.PsychologistAuthResponse, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	credentials, err := s.repo.GetPsychologistCredentialsByEmail(ctx, input.Email)
	if err != nil {
		return "", domain.PsychologistAuthResponse{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(credentials.PasswordHash), []byte(input.Password)); err != nil {
		return "", domain.PsychologistAuthResponse{}, ErrInvalidCredentials
	}
	if err := psychologistAccessError(credentials.User.IsActive, credentials.User.PortalAccessUntil, credentials.User.BlockedUntil, time.Now()); err != nil {
		return "", domain.PsychologistAuthResponse{}, err
	}

	workspace, err := s.GetPsychologistWorkspace(ctx, credentials.User.ID)
	if err != nil {
		return "", domain.PsychologistAuthResponse{}, err
	}

	return s.createSession(ctx, workspace)
}

func (s *AppService) AuthenticatePsychologist(ctx context.Context, sessionID string) (domain.AuthenticatedUser, error) {
	if strings.TrimSpace(sessionID) == "" {
		return domain.AuthenticatedUser{}, ErrUnauthorized
	}

	sessionHash := hashToken(sessionID)
	user, err := s.repo.GetAuthenticatedUserBySession(ctx, sessionHash)
	if err != nil {
		return domain.AuthenticatedUser{}, ErrUnauthorized
	}

	if user.Role != domain.RolePsychologist {
		return domain.AuthenticatedUser{}, ErrForbidden
	}
	if err := psychologistAccessError(user.IsActive, user.PortalAccessUntil, user.BlockedUntil, time.Now()); err != nil {
		if deleteErr := s.repo.DeleteSession(ctx, sessionHash); deleteErr != nil {
			return domain.AuthenticatedUser{}, deleteErr
		}
		return domain.AuthenticatedUser{}, err
	}

	return user, nil
}

func (s *AppService) GetPsychologistWorkspace(ctx context.Context, userID int64) (domain.PsychologistWorkspace, error) {
	workspace, err := s.repo.GetPsychologistWorkspaceByID(ctx, userID)
	if err != nil {
		return domain.PsychologistWorkspace{}, err
	}

	tests, err := s.repo.ListPsychologistTests(ctx, userID)
	if err != nil {
		return domain.PsychologistWorkspace{}, err
	}
	applyPublicURLsToTests(tests)
	workspace.Tests = tests

	return workspace, nil
}

func (s *AppService) UpdatePsychologistProfile(ctx context.Context, userID int64, input domain.UpdatePsychologistProfileInput) (domain.PsychologistProfile, error) {
	input.Specialization = strings.TrimSpace(input.Specialization)
	input.City = strings.TrimSpace(input.City)
	input.Timezone = strings.TrimSpace(input.Timezone)

	return s.repo.UpsertPsychologistProfile(ctx, userID, input)
}

func (s *AppService) UpdatePsychologistCard(ctx context.Context, userID int64, input domain.UpdatePsychologistCardInput) (domain.PsychologistCard, error) {
	input.Headline = strings.TrimSpace(input.Headline)
	input.ShortBio = strings.TrimSpace(input.ShortBio)
	input.ContactEmail = strings.TrimSpace(strings.ToLower(input.ContactEmail))
	input.ContactPhone = strings.TrimSpace(input.ContactPhone)
	input.Website = strings.TrimSpace(input.Website)
	input.Telegram = strings.TrimSpace(input.Telegram)

	return s.repo.UpsertPsychologistCard(ctx, userID, input)
}

func (s *AppService) LogoutPsychologist(ctx context.Context, sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return nil
	}

	return s.repo.DeleteSession(ctx, hashToken(sessionID))
}

func (s *AppService) CreatePsychologistByAdmin(ctx context.Context, input domain.CreatePsychologistInput) (domain.PsychologistWorkspace, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.FullName = strings.TrimSpace(input.FullName)

	if input.Email == "" || input.FullName == "" || len(input.Password) < 8 {
		return domain.PsychologistWorkspace{}, ErrInvalidCredentials
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.PsychologistWorkspace{}, err
	}

	workspace, err := s.repo.CreatePsychologist(ctx, domain.PsychologistRegistrationInput{
		Email:    input.Email,
		Password: input.Password,
		FullName: input.FullName,
		IsActive: input.IsActive,
	}, string(passwordHash))
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return domain.PsychologistWorkspace{}, ErrEmailAlreadyExists
		}
		return domain.PsychologistWorkspace{}, err
	}

	workspace.Tests = []domain.Test{}
	if s.mailer != nil {
		if err := s.mailer.SendPsychologistCredentials(ctx, workspace.User.Email, workspace.User.FullName, input.Password); err != nil {
			log.Printf("failed to send psychologist credentials email to %s: %v", workspace.User.Email, err)
		}
	}

	return workspace, nil
}

func (s *AppService) ListPsychologists(ctx context.Context) ([]domain.User, error) {
	return s.repo.ListPsychologists(ctx)
}

func (s *AppService) UpdatePsychologistAccount(ctx context.Context, userID int64, input domain.UpdatePsychologistAccountInput) (domain.User, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.FullName = strings.TrimSpace(input.FullName)
	if input.Email == "" || input.FullName == "" {
		return domain.User{}, ErrInvalidCredentials
	}

	return s.repo.UpdatePsychologistAccount(ctx, userID, input)
}

func (s *AppService) UpdatePsychologistAccess(ctx context.Context, userID int64, input domain.UpdatePsychologistAccessInput) (domain.User, error) {
	previousUser, err := s.repo.GetPsychologistByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}

	update, err := normalizePsychologistAccessInput(input)
	if err != nil {
		return domain.User{}, err
	}

	user, err := s.repo.UpdatePsychologistAccess(ctx, userID, update)
	if err != nil {
		return domain.User{}, err
	}

	if psychologistAccessError(user.IsActive, user.PortalAccessUntil, user.BlockedUntil, time.Now()) != nil {
		if err := s.repo.DeleteSessionsByUserID(ctx, userID); err != nil {
			return domain.User{}, err
		}
	}

	s.notifyPsychologistAboutSubscriptionExtension(ctx, previousUser, user, update)

	return user, nil
}

func (s *AppService) createSession(ctx context.Context, workspace domain.PsychologistWorkspace) (string, domain.PsychologistAuthResponse, error) {
	expiresAt := psychologistSessionExpiresAt(time.Now(), workspace.User.PortalAccessUntil)
	sessionID, expiresAt, err := s.createUserSessionWithExpiry(ctx, workspace.User.ID, expiresAt)
	if err != nil {
		return "", domain.PsychologistAuthResponse{}, err
	}

	return sessionID, domain.PsychologistAuthResponse{
		ExpiresAt: expiresAt,
		Workspace: workspace,
	}, nil
}

func (s *AppService) createUserSession(ctx context.Context, userID int64) (string, time.Time, error) {
	return s.createUserSessionWithExpiry(ctx, userID, time.Now().Add(SessionTTL))
}

func (s *AppService) createUserSessionWithExpiry(ctx context.Context, userID int64, expiresAt time.Time) (string, time.Time, error) {
	sessionID, err := generateToken()
	if err != nil {
		return "", time.Time{}, err
	}

	if err := s.repo.CreateSession(ctx, userID, hashToken(sessionID), expiresAt); err != nil {
		return "", time.Time{}, err
	}

	return sessionID, expiresAt, nil
}

func generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func psychologistAccessError(isActive bool, portalAccessUntil domain.NullableString, blockedUntil domain.NullableString, now time.Time) error {
	if !isActive {
		return ErrAccountDisabled
	}

	if until, ok := parseOptionalAccessTime(blockedUntil); ok && until.After(now) {
		return ErrAccountTemporarilyBlocked
	}

	if until, ok := parseOptionalAccessTime(portalAccessUntil); ok && !until.After(now) {
		return ErrPortalAccessExpired
	}

	return nil
}

func psychologistSessionExpiresAt(now time.Time, portalAccessUntil domain.NullableString) time.Time {
	expiresAt := now.Add(SessionTTL)
	if until, ok := parseOptionalAccessTime(portalAccessUntil); ok && until.Before(expiresAt) {
		return until
	}

	return expiresAt
}

func normalizePsychologistAccessInput(input domain.UpdatePsychologistAccessInput) (domain.PsychologistAccessUpdate, error) {
	update := domain.PsychologistAccessUpdate{}

	if input.IsActive != nil {
		update.IsActiveSet = true
		update.IsActive = *input.IsActive
	}

	if input.PortalAccessUntil.Set {
		update.PortalAccessUntilSet = true
		if input.PortalAccessUntil.Value != nil {
			parsed, err := parseAdminAccessDeadline(*input.PortalAccessUntil.Value)
			if err != nil {
				return domain.PsychologistAccessUpdate{}, ErrInvalidPsychologistAccess
			}
			update.PortalAccessUntil = parsed
		}
	}

	days, hasSubscriptionDays, err := normalizeSubscriptionDaysInput(input.SubscriptionDays, input.SubscriptionDaysAlias)
	if err != nil {
		return domain.PsychologistAccessUpdate{}, ErrInvalidPsychologistAccess
	}
	if hasSubscriptionDays {
		if update.PortalAccessUntilSet {
			return domain.PsychologistAccessUpdate{}, ErrInvalidPsychologistAccess
		}
		update.SubscriptionDaysSet = true
		update.SubscriptionDays = days
	}

	if input.BlockedUntil.Set {
		update.BlockedUntilSet = true
		if input.BlockedUntil.Value != nil {
			parsed, err := parseAdminAccessDeadline(*input.BlockedUntil.Value)
			if err != nil {
				return domain.PsychologistAccessUpdate{}, ErrInvalidPsychologistAccess
			}
			update.BlockedUntil = parsed
		}
	}

	if !update.IsActiveSet && !update.PortalAccessUntilSet && !update.SubscriptionDaysSet && !update.BlockedUntilSet {
		return domain.PsychologistAccessUpdate{}, ErrInvalidPsychologistAccess
	}

	return update, nil
}

func normalizeSubscriptionDaysInput(primary *int, alias *int) (int, bool, error) {
	if primary == nil && alias == nil {
		return 0, false, nil
	}

	if primary != nil && alias != nil && *primary != *alias {
		return 0, false, ErrInvalidPsychologistAccess
	}

	daysPtr := primary
	if daysPtr == nil {
		daysPtr = alias
	}
	if daysPtr == nil {
		return 0, false, nil
	}

	days := *daysPtr
	if days < 1 || days > 365 {
		return 0, false, ErrInvalidPsychologistAccess
	}

	return days, true, nil
}

func parseAdminAccessDeadline(raw string) (*time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return &parsed, nil
	}

	parsedDate, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return nil, err
	}

	endOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 23, 59, 59, 0, parsedDate.Location())
	return &endOfDay, nil
}

func parseOptionalAccessTime(raw domain.NullableString) (time.Time, bool) {
	value := strings.TrimSpace(raw.String())
	if value == "" {
		return time.Time{}, false
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, false
	}

	return parsed, true
}

func (s *AppService) notifyPsychologistAboutSubscriptionExtension(ctx context.Context, previous domain.User, current domain.User, update domain.PsychologistAccessUpdate) {
	if s == nil || s.mailer == nil || strings.TrimSpace(current.Email) == "" {
		return
	}

	days, ok := calculateSubscriptionExtensionDays(previous.PortalAccessUntil, current.PortalAccessUntil, update, time.Now())
	if !ok {
		return
	}

	if err := s.mailer.SendPsychologistSubscriptionExtended(
		ctx,
		current.Email,
		current.FullName,
		current.PortalAccessUntil.String(),
		days,
	); err != nil {
		log.Printf("failed to send subscription extension email to %s: %v", current.Email, err)
	}
}

func calculateSubscriptionExtensionDays(previous domain.NullableString, current domain.NullableString, update domain.PsychologistAccessUpdate, now time.Time) (int, bool) {
	currentUntil, ok := parseOptionalAccessTime(current)
	if !ok {
		return 0, false
	}

	if update.SubscriptionDaysSet {
		reference := now
		if previousUntil, previousOK := parseOptionalAccessTime(previous); previousOK && previousUntil.After(now) {
			reference = previousUntil
		}
		if !currentUntil.After(reference) {
			return 0, false
		}

		return update.SubscriptionDays, true
	}

	if !update.PortalAccessUntilSet {
		return 0, false
	}

	reference := now
	if previousUntil, previousOK := parseOptionalAccessTime(previous); previousOK && previousUntil.After(now) {
		reference = previousUntil
	}
	if !currentUntil.After(reference) {
		return 0, false
	}

	return durationDaysRoundedUp(currentUntil.Sub(reference)), true
}

func durationDaysRoundedUp(duration time.Duration) int {
	if duration <= 0 {
		return 0
	}

	days := int(duration / (24 * time.Hour))
	if duration%(24*time.Hour) != 0 {
		days++
	}
	if days == 0 {
		return 1
	}

	return days
}
