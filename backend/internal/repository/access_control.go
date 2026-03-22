package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func (r *AppRepository) UpdatePsychologistAccess(ctx context.Context, userID int64, update domain.PsychologistAccessUpdate) (domain.User, error) {
	var portalAccessUntil any
	if update.PortalAccessUntil != nil {
		portalAccessUntil = *update.PortalAccessUntil
	}

	var blockedUntil any
	if update.BlockedUntil != nil {
		blockedUntil = *update.BlockedUntil
	}

	var user domain.User
	var rowPortalAccessUntil sql.NullTime
	var rowBlockedUntil sql.NullTime
	var createdAt time.Time
	var updatedAt time.Time

	err := r.db.QueryRowContext(
		ctx,
		`UPDATE users
		 SET is_active = CASE WHEN $2 THEN $3 ELSE is_active END,
		 	portal_access_until = CASE
		 		WHEN $4 THEN $5
		 		WHEN $6 THEN CASE
		 			WHEN portal_access_until IS NOT NULL AND portal_access_until > NOW()
		 				THEN portal_access_until + ($7 * INTERVAL '1 day')
		 			ELSE NOW() + ($7 * INTERVAL '1 day')
		 		END
		 		ELSE portal_access_until
		 	END,
		 	subscription_plan = CASE WHEN $8 THEN $9 ELSE subscription_plan END,
		 	blocked_until = CASE WHEN $10 THEN $11 ELSE blocked_until END,
		 	updated_at = NOW()
		 WHERE id = $1 AND role = $12
		 RETURNING id, COALESCE(login, ''), email, full_name, role, is_active, portal_access_until, blocked_until, subscription_plan, created_at, updated_at`,
		userID,
		update.IsActiveSet,
		update.IsActive,
		update.PortalAccessUntilSet,
		portalAccessUntil,
		update.SubscriptionDaysSet,
		update.SubscriptionDays,
		update.SubscriptionPlanSet,
		update.SubscriptionPlan,
		update.BlockedUntilSet,
		blockedUntil,
		domain.RolePsychologist,
	).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&rowPortalAccessUntil,
		&rowBlockedUntil,
		&user.SubscriptionPlan,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}

	user.PortalAccessUntil = formatNullTime(rowPortalAccessUntil)
	user.BlockedUntil = formatNullTime(rowBlockedUntil)
	applyPsychologistUserStatuses(&user)
	user.CreatedAt = createdAt.Format(time.RFC3339)
	user.UpdatedAt = updatedAt.Format(time.RFC3339)

	return user, nil
}

func formatNullTime(value sql.NullTime) domain.NullableString {
	if !value.Valid {
		return domain.NewNullableString("")
	}

	return domain.NewNullableString(value.Time.Format(time.RFC3339))
}

func applyPsychologistUserStatuses(user *domain.User) {
	if user == nil || user.Role != domain.RolePsychologist {
		return
	}

	user.AccountStatus, user.SubscriptionStatus = domain.ResolvePsychologistStatuses(
		user.IsActive,
		user.PortalAccessUntil,
		user.BlockedUntil,
		time.Now(),
	)
}
