package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

var errLimitReached = errors.New("limit reached")

func (r *AppRepository) CreateTest(ctx context.Context, createdByUserID int64, input domain.CreateTestInput, publicSlug string) (domain.Test, error) {
	row := r.db.QueryRowContext(
		ctx,
		`INSERT INTO tests (title, description, created_by_user_id, report_template_id, recommended_duration, max_participants, status, public_slug, is_public)
		 VALUES ($1, $2, $3, NULLIF($4, 0), $5, $6, $7, $8, TRUE)
		 RETURNING id, title, description, created_by_user_id, COALESCE(report_template_id, 0), recommended_duration, max_participants, status, COALESCE(public_slug, ''), is_public, created_at, updated_at`,
		input.Title,
		input.Description,
		createdByUserID,
		input.ReportTemplateID,
		input.RecommendedDuration,
		input.MaxParticipants,
		input.Status,
		publicSlug,
	)

	return scanTest(row)
}

func (r *AppRepository) ListPsychologistTests(ctx context.Context, createdByUserID int64) ([]domain.Test, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, title, description, created_by_user_id, COALESCE(report_template_id, 0), recommended_duration, max_participants, status, COALESCE(public_slug, ''), is_public, created_at, updated_at
		 FROM tests
		 WHERE created_by_user_id = $1
		 ORDER BY id DESC`,
		createdByUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tests := make([]domain.Test, 0)
	for rows.Next() {
		test, scanErr := scanTest(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		tests = append(tests, test)
	}

	return tests, rows.Err()
}

func (r *AppRepository) GetPsychologistTestByID(ctx context.Context, testID int64, createdByUserID int64) (domain.Test, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, title, description, created_by_user_id, COALESCE(report_template_id, 0), recommended_duration, max_participants, status, COALESCE(public_slug, ''), is_public, created_at, updated_at
		 FROM tests
		 WHERE id = $1 AND created_by_user_id = $2`,
		testID,
		createdByUserID,
	)

	return scanTest(row)
}

func (r *AppRepository) UpdatePsychologistTest(ctx context.Context, testID int64, createdByUserID int64, input domain.UpdateTestInput) (domain.Test, error) {
	row := r.db.QueryRowContext(
		ctx,
		`UPDATE tests
		 SET title = $3,
		 	description = $4,
		 	report_template_id = NULLIF($5, 0),
		 	recommended_duration = $6,
		 	max_participants = $7,
		 	status = $8,
		 	updated_at = NOW()
		 WHERE id = $1 AND created_by_user_id = $2
		 RETURNING id, title, description, created_by_user_id, COALESCE(report_template_id, 0), recommended_duration, max_participants, status, COALESCE(public_slug, ''), is_public, created_at, updated_at`,
		testID,
		createdByUserID,
		input.Title,
		input.Description,
		input.ReportTemplateID,
		input.RecommendedDuration,
		input.MaxParticipants,
		input.Status,
	)

	return scanTest(row)
}

func (r *AppRepository) DeletePsychologistTest(ctx context.Context, testID int64, createdByUserID int64) (bool, error) {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM tests
		 WHERE id = $1 AND created_by_user_id = $2`,
		testID,
		createdByUserID,
	)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func scanTest(scanner rowScanner) (domain.Test, error) {
	var test domain.Test
	var createdAt time.Time
	var updatedAt time.Time

	err := scanner.Scan(
		&test.ID,
		&test.Title,
		&test.Description,
		&test.CreatedByUserID,
		&test.ReportTemplateID,
		&test.RecommendedDuration,
		&test.MaxParticipants,
		&test.Status,
		&test.PublicSlug,
		&test.IsPublic,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.Test{}, err
	}

	test.CreatedAt = createdAt.Format(time.RFC3339)
	test.UpdatedAt = updatedAt.Format(time.RFC3339)

	return test, nil
}

func IsNotFound(err error) bool {
	return err == sql.ErrNoRows
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func IsLimitReached(err error) bool {
	return errors.Is(err, errLimitReached)
}
