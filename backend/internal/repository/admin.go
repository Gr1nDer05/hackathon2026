package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func (r *AppRepository) UpsertAdminAccount(ctx context.Context, input domain.AdminSeedInput, passwordHash string) error {
	login := strings.TrimSpace(strings.ToLower(input.Login))
	email := placeholderAdminEmail(login)
	fullName := strings.TrimSpace(input.FullName)

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (login, email, full_name, role, password_hash, is_active)
		 VALUES ($1, $2, $3, $4, $5, TRUE)
		 ON CONFLICT (login) DO UPDATE SET
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

func (r *AppRepository) UpdateAdminEmail(ctx context.Context, userID int64, email string) (domain.User, error) {
	var user domain.User
	var emailVerifiedAt sql.NullTime
	var createdAt time.Time
	var updatedAt time.Time

	err := r.db.QueryRowContext(
		ctx,
		`UPDATE users
		 SET email = $2,
		 	email_verified_at = CASE WHEN email = $2 THEN email_verified_at ELSE NULL END,
		 	email_verification_code_hash = NULL,
		 	email_verification_expires_at = NULL,
		 	updated_at = NOW()
		 WHERE id = $1
		   AND role = $3
		 RETURNING id, login, email, email_verified_at, full_name, role, is_active, created_at, updated_at`,
		userID,
		email,
		domain.RoleAdmin,
	).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&emailVerifiedAt,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return domain.User{}, ErrEmailAlreadyExists
		}
		return domain.User{}, err
	}

	user.EmailVerifiedAt = formatNullTime(emailVerifiedAt)
	user.CreatedAt = createdAt.Format(time.RFC3339)
	user.UpdatedAt = updatedAt.Format(time.RFC3339)

	return user, nil
}

func (r *AppRepository) StoreAdminEmailVerificationCode(ctx context.Context, userID int64, codeHash string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE users
		 SET email_verification_code_hash = $2,
		 	email_verification_expires_at = $3,
		 	updated_at = NOW()
		 WHERE id = $1
		   AND role = $4`,
		userID,
		codeHash,
		expiresAt,
		domain.RoleAdmin,
	)
	return err
}

func (r *AppRepository) VerifyAdminEmail(ctx context.Context, userID int64, codeHash string) (domain.User, error) {
	var user domain.User
	var emailVerifiedAt sql.NullTime
	var createdAt time.Time
	var updatedAt time.Time

	err := r.db.QueryRowContext(
		ctx,
		`UPDATE users
		 SET email_verified_at = NOW(),
		 	email_verification_code_hash = NULL,
		 	email_verification_expires_at = NULL,
		 	updated_at = NOW()
		 WHERE id = $1
		   AND role = $2
		   AND email_verification_code_hash = $3
		   AND email_verification_expires_at IS NOT NULL
		   AND email_verification_expires_at > NOW()
		 RETURNING id, login, email, email_verified_at, full_name, role, is_active, created_at, updated_at`,
		userID,
		domain.RoleAdmin,
		codeHash,
	).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&emailVerifiedAt,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}

	user.EmailVerifiedAt = formatNullTime(emailVerifiedAt)
	user.CreatedAt = createdAt.Format(time.RFC3339)
	user.UpdatedAt = updatedAt.Format(time.RFC3339)

	return user, nil
}

func (r *AppRepository) ListVerifiedAdmins(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, login, email, email_verified_at, full_name, role, is_active, created_at, updated_at
		 FROM users
		 WHERE role = $1
		   AND is_active = TRUE
		   AND email_verified_at IS NOT NULL
		 ORDER BY id`,
		domain.RoleAdmin,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	admins := make([]domain.User, 0)
	for rows.Next() {
		var user domain.User
		var emailVerifiedAt sql.NullTime
		var createdAt time.Time
		var updatedAt time.Time

		if err := rows.Scan(
			&user.ID,
			&user.Login,
			&user.Email,
			&emailVerifiedAt,
			&user.FullName,
			&user.Role,
			&user.IsActive,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}

		user.EmailVerifiedAt = formatNullTime(emailVerifiedAt)
		user.CreatedAt = createdAt.Format(time.RFC3339)
		user.UpdatedAt = updatedAt.Format(time.RFC3339)
		admins = append(admins, user)
	}

	return admins, rows.Err()
}

func (r *AppRepository) GetAdminCredentialsByLogin(ctx context.Context, login string) (domain.UserCredentials, error) {
	var credentials domain.UserCredentials
	var emailVerifiedAt sql.NullTime
	var createdAt time.Time
	var updatedAt time.Time

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, login, email, email_verified_at, full_name, role, is_active, password_hash, created_at, updated_at
		 FROM users
		 WHERE login = $1 AND role = $2`,
		strings.TrimSpace(strings.ToLower(login)),
		domain.RoleAdmin,
	).Scan(
		&credentials.User.ID,
		&credentials.User.Login,
		&credentials.User.Email,
		&emailVerifiedAt,
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

	credentials.User.EmailVerifiedAt = formatNullTime(emailVerifiedAt)
	credentials.User.CreatedAt = createdAt.Format(time.RFC3339)
	credentials.User.UpdatedAt = updatedAt.Format(time.RFC3339)

	return credentials, nil
}

func (r *AppRepository) GetAdminByID(ctx context.Context, userID int64) (domain.User, error) {
	var user domain.User
	var emailVerifiedAt sql.NullTime
	var createdAt time.Time
	var updatedAt time.Time

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, login, email, email_verified_at, full_name, role, is_active, created_at, updated_at
		 FROM users
		 WHERE id = $1 AND role = $2`,
		userID,
		domain.RoleAdmin,
	).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&emailVerifiedAt,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}

	user.EmailVerifiedAt = formatNullTime(emailVerifiedAt)
	user.CreatedAt = createdAt.Format(time.RFC3339)
	user.UpdatedAt = updatedAt.Format(time.RFC3339)

	return user, nil
}

func placeholderAdminEmail(login string) string {
	return strings.TrimSpace(strings.ToLower(login)) + "@admin.local"
}
