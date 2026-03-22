package service

import (
	"context"
	"errors"
	"net/mail"
	"os"
	"strings"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const defaultAdminAccounts = ""

var ErrInvalidAdminEmail = errors.New("invalid admin email")

func LoadAdminSeedInputs() []domain.AdminSeedInput {
	raw := strings.TrimSpace(os.Getenv("ADMIN_ACCOUNTS"))
	if raw == "" {
		if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
			return nil
		}
		raw = defaultAdminAccounts
	}

	parts := strings.Split(raw, ",")
	admins := make([]domain.AdminSeedInput, 0, len(parts))

	for _, part := range parts {
		chunks := strings.SplitN(strings.TrimSpace(part), ":", 3)
		if len(chunks) != 3 {
			continue
		}

		login := strings.TrimSpace(chunks[0])
		password := strings.TrimSpace(chunks[1])
		fullName := strings.TrimSpace(chunks[2])
		if login == "" || password == "" || fullName == "" {
			continue
		}

		admins = append(admins, domain.AdminSeedInput{
			Login:    login,
			Password: password,
			FullName: fullName,
		})
	}

	return admins
}

func (s *AppService) SeedAdminAccounts(ctx context.Context, admins []domain.AdminSeedInput) error {
	for _, admin := range admins {
		if len(admin.Password) < 8 {
			continue
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		if err := s.repo.UpsertAdminAccount(ctx, admin, string(passwordHash)); err != nil {
			return err
		}
	}

	return nil
}

func (s *AppService) LoginAdmin(ctx context.Context, input domain.AdminLoginInput) (string, domain.AdminAuthResponse, error) {
	input.Login = strings.TrimSpace(strings.ToLower(input.Login))

	credentials, err := s.repo.GetAdminCredentialsByLogin(ctx, input.Login)
	if err != nil {
		return "", domain.AdminAuthResponse{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(credentials.PasswordHash), []byte(input.Password)); err != nil {
		return "", domain.AdminAuthResponse{}, ErrInvalidCredentials
	}

	sessionID, expiresAt, err := s.createUserSession(ctx, credentials.User.ID)
	if err != nil {
		return "", domain.AdminAuthResponse{}, err
	}

	user, err := s.repo.GetAdminByID(ctx, credentials.User.ID)
	if err != nil {
		return "", domain.AdminAuthResponse{}, err
	}

	return sessionID, domain.AdminAuthResponse{
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

func (s *AppService) AuthenticateAdmin(ctx context.Context, sessionID string) (domain.AuthenticatedUser, error) {
	if strings.TrimSpace(sessionID) == "" {
		return domain.AuthenticatedUser{}, ErrUnauthorized
	}

	user, err := s.repo.GetAuthenticatedUserBySession(ctx, hashToken(sessionID))
	if err != nil {
		return domain.AuthenticatedUser{}, ErrUnauthorized
	}

	if user.Role != domain.RoleAdmin {
		return domain.AuthenticatedUser{}, ErrForbidden
	}

	return user, nil
}

func (s *AppService) LogoutAdmin(ctx context.Context, sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return nil
	}

	return s.repo.DeleteSession(ctx, hashToken(sessionID))
}

func (s *AppService) GetAdminMe(ctx context.Context, userID int64) (domain.User, error) {
	user, err := s.repo.GetAdminByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}
	if user.Role != domain.RoleAdmin {
		return domain.User{}, errors.New("user is not admin")
	}

	return user, nil
}

func (s *AppService) ListPendingSubscriptionPurchaseRequests(ctx context.Context) ([]domain.SubscriptionPurchaseRequest, error) {
	return s.repo.ListPendingSubscriptionPurchaseRequests(ctx)
}

func (s *AppService) UpdateAdminMe(ctx context.Context, userID int64, input domain.UpdateAdminMeInput) (domain.User, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if !IsAdminEmailBound(email) {
		return domain.User{}, ErrInvalidAdminEmail
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return domain.User{}, ErrInvalidAdminEmail
	}

	user, err := s.repo.UpdateAdminEmail(ctx, userID, email)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return domain.User{}, ErrEmailAlreadyExists
		}
		return domain.User{}, err
	}
	if user.Role != domain.RoleAdmin {
		return domain.User{}, errors.New("user is not admin")
	}

	return user, nil
}

func IsAdminEmailBound(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return false
	}

	if strings.HasSuffix(email, "@admin.local") {
		return false
	}

	return true
}
