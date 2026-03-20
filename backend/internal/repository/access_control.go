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
		 	portal_access_until = CASE WHEN $4 THEN $5 ELSE portal_access_until END,
		 	blocked_until = CASE WHEN $6 THEN $7 ELSE blocked_until END,
		 	updated_at = NOW()
		 WHERE id = $1 AND role = $8
		 RETURNING id, COALESCE(login, ''), email, full_name, role, is_active, portal_access_until, blocked_until, created_at, updated_at`,
		userID,
		update.IsActiveSet,
		update.IsActive,
		update.PortalAccessUntilSet,
		portalAccessUntil,
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
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}

	user.PortalAccessUntil = formatNullTime(rowPortalAccessUntil)
	user.BlockedUntil = formatNullTime(rowBlockedUntil)
	user.CreatedAt = createdAt.Format(time.RFC3339)
	user.UpdatedAt = updatedAt.Format(time.RFC3339)

	return user, nil
}

func (r *AppRepository) ListAdminNotifications(ctx context.Context, from time.Time, until time.Time) ([]domain.AdminNotification, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, email, full_name, portal_access_until
		 FROM users
		 WHERE role = $1
		   AND is_active = TRUE
		   AND portal_access_until IS NOT NULL
		   AND portal_access_until > $2
		   AND portal_access_until <= $3
		 ORDER BY portal_access_until ASC, id ASC`,
		domain.RolePsychologist,
		from,
		until,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]domain.AdminNotification, 0)
	for rows.Next() {
		var notification domain.AdminNotification
		var portalAccessUntil time.Time
		if err := rows.Scan(
			&notification.PsychologistID,
			&notification.PsychologistEmail,
			&notification.PsychologistName,
			&portalAccessUntil,
		); err != nil {
			return nil, err
		}

		notification.Type = domain.AdminNotificationTypeSubscriptionExpiring
		notification.PortalAccessUntil = portalAccessUntil.Format(time.RFC3339)
		notification.Severity = "warning"
		notification.Message = "Psychologist portal access expires within the next 24 hours"
		notifications = append(notifications, notification)
	}

	return notifications, rows.Err()
}

func formatNullTime(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}

	return value.Time.Format(time.RFC3339)
}
