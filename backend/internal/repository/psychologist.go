package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

var ErrEmailAlreadyExists = errors.New("email already exists")

func (r *AppRepository) CreatePsychologist(ctx context.Context, input domain.PsychologistRegistrationInput, passwordHash string) (domain.PsychologistWorkspace, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.PsychologistWorkspace{}, err
	}
	defer tx.Rollback()

	var user domain.User
	var createdAt time.Time
	var updatedAt time.Time

	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO users (email, full_name, role, password_hash, is_active)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, email, full_name, role, is_active, created_at, updated_at`,
		input.Email,
		input.FullName,
		domain.RolePsychologist,
		passwordHash,
		input.IsActive,
	).Scan(&user.ID, &user.Email, &user.FullName, &user.Role, &user.IsActive, &createdAt, &updatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return domain.PsychologistWorkspace{}, ErrEmailAlreadyExists
		}
		return domain.PsychologistWorkspace{}, err
	}

	user.CreatedAt = createdAt.Format(time.RFC3339)
	user.UpdatedAt = updatedAt.Format(time.RFC3339)

	if _, err = tx.ExecContext(
		ctx,
		`INSERT INTO psychologist_profiles (user_id) VALUES ($1)
		 ON CONFLICT (user_id) DO NOTHING`,
		user.ID,
	); err != nil {
		return domain.PsychologistWorkspace{}, err
	}

	if _, err = tx.ExecContext(
		ctx,
		`INSERT INTO psychologist_cards (user_id, contact_email) VALUES ($1, $2)
		 ON CONFLICT (user_id) DO NOTHING`,
		user.ID,
		input.Email,
	); err != nil {
		return domain.PsychologistWorkspace{}, err
	}

	if err = tx.Commit(); err != nil {
		return domain.PsychologistWorkspace{}, err
	}

	return r.GetPsychologistWorkspaceByID(ctx, user.ID)
}

func (r *AppRepository) GetPsychologistCredentialsByEmail(ctx context.Context, email string) (domain.UserCredentials, error) {
	var credentials domain.UserCredentials
	var portalAccessUntil sql.NullTime
	var blockedUntil sql.NullTime
	var createdAt time.Time
	var updatedAt time.Time

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, email, full_name, role, is_active, portal_access_until, blocked_until, password_hash, created_at, updated_at
		 FROM users
		 WHERE email = $1 AND role = $2`,
		email,
		domain.RolePsychologist,
	).Scan(
		&credentials.User.ID,
		&credentials.User.Email,
		&credentials.User.FullName,
		&credentials.User.Role,
		&credentials.User.IsActive,
		&portalAccessUntil,
		&blockedUntil,
		&credentials.PasswordHash,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.UserCredentials{}, err
	}

	credentials.User.PortalAccessUntil = formatNullTime(portalAccessUntil)
	credentials.User.BlockedUntil = formatNullTime(blockedUntil)
	applyPsychologistUserStatuses(&credentials.User)
	credentials.User.CreatedAt = createdAt.Format(time.RFC3339)
	credentials.User.UpdatedAt = updatedAt.Format(time.RFC3339)

	return credentials, nil
}

func (r *AppRepository) CreateSession(ctx context.Context, userID int64, sessionHash string, expiresAt time.Time) error {
	if err := r.DeleteExpiredSessions(ctx); err != nil {
		return err
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO user_sessions (session_hash, user_id, expires_at)
		 VALUES ($1, $2, $3)`,
		sessionHash,
		userID,
		expiresAt,
	)
	return err
}

func (r *AppRepository) GetAuthenticatedUserBySession(ctx context.Context, sessionHash string) (domain.AuthenticatedUser, error) {
	if err := r.DeleteExpiredSessions(ctx); err != nil {
		return domain.AuthenticatedUser{}, err
	}

	var user domain.AuthenticatedUser
	var portalAccessUntil sql.NullTime
	var blockedUntil sql.NullTime

	err := r.db.QueryRowContext(
		ctx,
		`SELECT u.id, u.email, u.full_name, u.role, u.is_active, u.portal_access_until, u.blocked_until
		 FROM user_sessions s
		 JOIN users u ON u.id = s.user_id
		 WHERE s.session_hash = $1
		   AND s.expires_at > NOW()`,
		sessionHash,
	).Scan(&user.ID, &user.Email, &user.FullName, &user.Role, &user.IsActive, &portalAccessUntil, &blockedUntil)
	if err != nil {
		return domain.AuthenticatedUser{}, err
	}

	user.PortalAccessUntil = formatNullTime(portalAccessUntil)
	user.BlockedUntil = formatNullTime(blockedUntil)
	return user, nil
}

func (r *AppRepository) DeleteSession(ctx context.Context, sessionHash string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE session_hash = $1`, sessionHash)
	return err
}

func (r *AppRepository) DeleteSessionsByUserID(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE user_id = $1`, userID)
	return err
}

func (r *AppRepository) DeleteExpiredSessions(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE expires_at <= NOW()`)
	return err
}

func (r *AppRepository) GetPsychologistWorkspaceByID(ctx context.Context, userID int64) (domain.PsychologistWorkspace, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT
			u.id, u.email, u.full_name, u.role, u.is_active, u.portal_access_until, u.blocked_until, u.created_at, u.updated_at,
			p.user_id, p.about, p.specialization, p.experience_years, p.education, p.methods, p.city, p.timezone, p.is_public, p.created_at, p.updated_at,
			c.user_id, c.headline, c.short_bio, c.contact_email, c.contact_phone, c.website, c.telegram, c.price_from, c.online_available, c.offline_available, c.created_at, c.updated_at
		 FROM users u
		 JOIN psychologist_profiles p ON p.user_id = u.id
		 JOIN psychologist_cards c ON c.user_id = u.id
		 WHERE u.id = $1 AND u.role = $2`,
		userID,
		domain.RolePsychologist,
	)

	return scanPsychologistWorkspace(row)
}

func (r *AppRepository) GetPsychologistByID(ctx context.Context, userID int64) (domain.User, error) {
	var user domain.User
	var portalAccessUntil sql.NullTime
	var blockedUntil sql.NullTime
	var createdAt time.Time
	var updatedAt time.Time

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, COALESCE(login, ''), email, full_name, role, is_active, portal_access_until, blocked_until, created_at, updated_at
		 FROM users
		 WHERE id = $1 AND role = $2`,
		userID,
		domain.RolePsychologist,
	).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&portalAccessUntil,
		&blockedUntil,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}

	user.PortalAccessUntil = formatNullTime(portalAccessUntil)
	user.BlockedUntil = formatNullTime(blockedUntil)
	applyPsychologistUserStatuses(&user)
	user.CreatedAt = createdAt.Format(time.RFC3339)
	user.UpdatedAt = updatedAt.Format(time.RFC3339)

	return user, nil
}

func (r *AppRepository) UpsertPsychologistProfile(ctx context.Context, userID int64, input domain.UpdatePsychologistProfileInput) (domain.PsychologistProfile, error) {
	row := r.db.QueryRowContext(
		ctx,
		`INSERT INTO psychologist_profiles (
			user_id, about, specialization, experience_years, education, methods, city, timezone, is_public
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id) DO UPDATE SET
			about = EXCLUDED.about,
			specialization = EXCLUDED.specialization,
			experience_years = EXCLUDED.experience_years,
			education = EXCLUDED.education,
			methods = EXCLUDED.methods,
			city = EXCLUDED.city,
			timezone = EXCLUDED.timezone,
			is_public = EXCLUDED.is_public,
			updated_at = NOW()
		RETURNING user_id, about, specialization, experience_years, education, methods, city, timezone, is_public, created_at, updated_at`,
		userID,
		input.About,
		input.Specialization,
		input.ExperienceYears,
		input.Education,
		input.Methods,
		input.City,
		input.Timezone,
		input.IsPublic,
	)

	return scanPsychologistProfile(row)
}

func (r *AppRepository) UpsertPsychologistCard(ctx context.Context, userID int64, input domain.UpdatePsychologistCardInput) (domain.PsychologistCard, error) {
	row := r.db.QueryRowContext(
		ctx,
		`INSERT INTO psychologist_cards (
			user_id, headline, short_bio, contact_email, contact_phone, website, telegram, price_from, online_available, offline_available
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id) DO UPDATE SET
			headline = EXCLUDED.headline,
			short_bio = EXCLUDED.short_bio,
			contact_email = EXCLUDED.contact_email,
			contact_phone = EXCLUDED.contact_phone,
			website = EXCLUDED.website,
			telegram = EXCLUDED.telegram,
			price_from = EXCLUDED.price_from,
			online_available = EXCLUDED.online_available,
			offline_available = EXCLUDED.offline_available,
			updated_at = NOW()
		RETURNING user_id, headline, short_bio, contact_email, contact_phone, website, telegram, price_from, online_available, offline_available, created_at, updated_at`,
		userID,
		input.Headline,
		input.ShortBio,
		input.ContactEmail,
		input.ContactPhone,
		input.Website,
		input.Telegram,
		input.PriceFrom,
		input.OnlineAvailable,
		input.OfflineAvailable,
	)

	return scanPsychologistCard(row)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPsychologistWorkspace(scanner rowScanner) (domain.PsychologistWorkspace, error) {
	var workspace domain.PsychologistWorkspace
	var portalAccessUntil sql.NullTime
	var blockedUntil sql.NullTime
	var userCreatedAt time.Time
	var userUpdatedAt time.Time
	var profileCreatedAt time.Time
	var profileUpdatedAt time.Time
	var cardCreatedAt time.Time
	var cardUpdatedAt time.Time

	err := scanner.Scan(
		&workspace.User.ID,
		&workspace.User.Email,
		&workspace.User.FullName,
		&workspace.User.Role,
		&workspace.User.IsActive,
		&portalAccessUntil,
		&blockedUntil,
		&userCreatedAt,
		&userUpdatedAt,
		&workspace.Profile.UserID,
		&workspace.Profile.About,
		&workspace.Profile.Specialization,
		&workspace.Profile.ExperienceYears,
		&workspace.Profile.Education,
		&workspace.Profile.Methods,
		&workspace.Profile.City,
		&workspace.Profile.Timezone,
		&workspace.Profile.IsPublic,
		&profileCreatedAt,
		&profileUpdatedAt,
		&workspace.Card.UserID,
		&workspace.Card.Headline,
		&workspace.Card.ShortBio,
		&workspace.Card.ContactEmail,
		&workspace.Card.ContactPhone,
		&workspace.Card.Website,
		&workspace.Card.Telegram,
		&workspace.Card.PriceFrom,
		&workspace.Card.OnlineAvailable,
		&workspace.Card.OfflineAvailable,
		&cardCreatedAt,
		&cardUpdatedAt,
	)
	if err != nil {
		return domain.PsychologistWorkspace{}, err
	}

	workspace.User.PortalAccessUntil = formatNullTime(portalAccessUntil)
	workspace.User.BlockedUntil = formatNullTime(blockedUntil)
	applyPsychologistUserStatuses(&workspace.User)
	workspace.User.CreatedAt = userCreatedAt.Format(time.RFC3339)
	workspace.User.UpdatedAt = userUpdatedAt.Format(time.RFC3339)
	workspace.Profile.CreatedAt = profileCreatedAt.Format(time.RFC3339)
	workspace.Profile.UpdatedAt = profileUpdatedAt.Format(time.RFC3339)
	workspace.Card.CreatedAt = cardCreatedAt.Format(time.RFC3339)
	workspace.Card.UpdatedAt = cardUpdatedAt.Format(time.RFC3339)

	return workspace, nil
}

func (r *AppRepository) ListPsychologists(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, COALESCE(login, ''), email, full_name, role, is_active, portal_access_until, blocked_until, created_at, updated_at
		 FROM users
		 WHERE role = $1
		 ORDER BY id`,
		domain.RolePsychologist,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var user domain.User
		var portalAccessUntil sql.NullTime
		var blockedUntil sql.NullTime
		var createdAt time.Time
		var updatedAt time.Time

		if err := rows.Scan(
			&user.ID,
			&user.Login,
			&user.Email,
			&user.FullName,
			&user.Role,
			&user.IsActive,
			&portalAccessUntil,
			&blockedUntil,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}

		user.PortalAccessUntil = formatNullTime(portalAccessUntil)
		user.BlockedUntil = formatNullTime(blockedUntil)
		applyPsychologistUserStatuses(&user)
		user.CreatedAt = createdAt.Format(time.RFC3339)
		user.UpdatedAt = updatedAt.Format(time.RFC3339)
		users = append(users, user)
	}

	return users, rows.Err()
}

func (r *AppRepository) UpdatePsychologistAccount(ctx context.Context, userID int64, input domain.UpdatePsychologistAccountInput) (domain.User, error) {
	var user domain.User
	var portalAccessUntil sql.NullTime
	var blockedUntil sql.NullTime
	var createdAt time.Time
	var updatedAt time.Time

	err := r.db.QueryRowContext(
		ctx,
		`UPDATE users
		 SET email = $2, full_name = $3, is_active = $4, updated_at = NOW()
		 WHERE id = $1 AND role = $5
		 RETURNING id, COALESCE(login, ''), email, full_name, role, is_active, portal_access_until, blocked_until, created_at, updated_at`,
		userID,
		input.Email,
		input.FullName,
		input.IsActive,
		domain.RolePsychologist,
	).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&portalAccessUntil,
		&blockedUntil,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}

	user.PortalAccessUntil = formatNullTime(portalAccessUntil)
	user.BlockedUntil = formatNullTime(blockedUntil)
	applyPsychologistUserStatuses(&user)
	user.CreatedAt = createdAt.Format(time.RFC3339)
	user.UpdatedAt = updatedAt.Format(time.RFC3339)

	return user, nil
}

func scanPsychologistProfile(scanner rowScanner) (domain.PsychologistProfile, error) {
	var profile domain.PsychologistProfile
	var createdAt time.Time
	var updatedAt time.Time

	err := scanner.Scan(
		&profile.UserID,
		&profile.About,
		&profile.Specialization,
		&profile.ExperienceYears,
		&profile.Education,
		&profile.Methods,
		&profile.City,
		&profile.Timezone,
		&profile.IsPublic,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.PsychologistProfile{}, err
	}

	profile.CreatedAt = createdAt.Format(time.RFC3339)
	profile.UpdatedAt = updatedAt.Format(time.RFC3339)

	return profile, nil
}

func scanPsychologistCard(scanner rowScanner) (domain.PsychologistCard, error) {
	var card domain.PsychologistCard
	var createdAt time.Time
	var updatedAt time.Time

	err := scanner.Scan(
		&card.UserID,
		&card.Headline,
		&card.ShortBio,
		&card.ContactEmail,
		&card.ContactPhone,
		&card.Website,
		&card.Telegram,
		&card.PriceFrom,
		&card.OnlineAvailable,
		&card.OfflineAvailable,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.PsychologistCard{}, err
	}

	card.CreatedAt = createdAt.Format(time.RFC3339)
	card.UpdatedAt = updatedAt.Format(time.RFC3339)

	return card, nil
}
