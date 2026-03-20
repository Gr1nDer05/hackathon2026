package repository

import (
	"context"
	"strings"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func (r *AppRepository) UpsertAdminAccount(ctx context.Context, input domain.AdminSeedInput, passwordHash string) error {
	login := strings.TrimSpace(strings.ToLower(input.Login))
	email := login + "@admin.local"
	fullName := strings.TrimSpace(input.FullName)

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (login, email, full_name, role, password_hash, is_active)
		 VALUES ($1, $2, $3, $4, $5, TRUE)
		 ON CONFLICT (login) DO UPDATE SET
			email = EXCLUDED.email,
			full_name = EXCLUDED.full_name,
			role = EXCLUDED.role,
			password_hash = EXCLUDED.password_hash,
			is_active = TRUE,
			updated_at = NOW()`,
		login,
		email,
		fullName,
		domain.RoleAdmin,
		passwordHash,
	)
	return err
}

func (r *AppRepository) GetAdminCredentialsByLogin(ctx context.Context, login string) (domain.UserCredentials, error) {
	var credentials domain.UserCredentials
	var createdAt time.Time
	var updatedAt time.Time

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, login, email, full_name, role, is_active, password_hash, created_at, updated_at
		 FROM users
		 WHERE login = $1 AND role = $2`,
		strings.TrimSpace(strings.ToLower(login)),
		domain.RoleAdmin,
	).Scan(
		&credentials.User.ID,
		&credentials.User.Login,
		&credentials.User.Email,
		&credentials.User.FullName,
		&credentials.User.Role,
		&credentials.User.IsActive,
		&credentials.PasswordHash,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.UserCredentials{}, err
	}

	credentials.User.CreatedAt = createdAt.Format(time.RFC3339)
	credentials.User.UpdatedAt = updatedAt.Format(time.RFC3339)

	return credentials, nil
}

func (r *AppRepository) GetAdminByID(ctx context.Context, userID int64) (domain.User, error) {
	var user domain.User
	var createdAt time.Time
	var updatedAt time.Time

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, login, email, full_name, role, is_active, created_at, updated_at
		 FROM users
		 WHERE id = $1 AND role = $2`,
		userID,
		domain.RoleAdmin,
	).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}

	user.CreatedAt = createdAt.Format(time.RFC3339)
	user.UpdatedAt = updatedAt.Format(time.RFC3339)

	return user, nil
}
