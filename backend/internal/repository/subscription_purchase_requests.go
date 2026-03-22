package repository

import (
	"context"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func (r *AppRepository) CreateOrReplacePendingSubscriptionPurchaseRequest(ctx context.Context, psychologistUserID int64, subscriptionPlan string, durationDays int) (domain.SubscriptionPurchaseRequest, error) {
	row := r.db.QueryRowContext(
		ctx,
		`WITH upserted AS (
			INSERT INTO subscription_purchase_requests (psychologist_user_id, subscription_plan, duration_days, status)
			VALUES ($1, $2, $3, 'pending')
			ON CONFLICT (psychologist_user_id) WHERE status = 'pending'
			DO UPDATE SET
				subscription_plan = EXCLUDED.subscription_plan,
				duration_days = EXCLUDED.duration_days,
				updated_at = NOW()
			RETURNING id, psychologist_user_id, subscription_plan, duration_days, status, created_at, updated_at
		)
		SELECT r.id, r.psychologist_user_id, u.full_name, u.email, r.subscription_plan, r.duration_days, r.status, r.created_at, r.updated_at
		FROM upserted r
		JOIN users u ON u.id = r.psychologist_user_id`,
		psychologistUserID,
		subscriptionPlan,
		durationDays,
	)

	return scanSubscriptionPurchaseRequest(row)
}

func (r *AppRepository) ListPendingSubscriptionPurchaseRequests(ctx context.Context) ([]domain.SubscriptionPurchaseRequest, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT r.id, r.psychologist_user_id, u.full_name, u.email, r.subscription_plan, r.duration_days, r.status, r.created_at, r.updated_at
		 FROM subscription_purchase_requests r
		 JOIN users u ON u.id = r.psychologist_user_id
		 WHERE r.status = 'pending'
		 ORDER BY r.created_at DESC, r.id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := make([]domain.SubscriptionPurchaseRequest, 0)
	for rows.Next() {
		request, err := scanSubscriptionPurchaseRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

type subscriptionPurchaseRequestScanner interface {
	Scan(dest ...any) error
}

func scanSubscriptionPurchaseRequest(scanner subscriptionPurchaseRequestScanner) (domain.SubscriptionPurchaseRequest, error) {
	var request domain.SubscriptionPurchaseRequest
	var createdAt time.Time
	var updatedAt time.Time

	if err := scanner.Scan(
		&request.ID,
		&request.PsychologistID,
		&request.PsychologistName,
		&request.PsychologistEmail,
		&request.SubscriptionPlan,
		&request.DurationDays,
		&request.Status,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.SubscriptionPurchaseRequest{}, err
	}

	request.CreatedAt = createdAt.Format(time.RFC3339)
	request.UpdatedAt = updatedAt.Format(time.RFC3339)
	return request, nil
}
