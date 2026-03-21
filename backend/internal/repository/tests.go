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
var errDuplicateRespondentPhone = errors.New("duplicate respondent phone")

func (r *AppRepository) CreateTest(ctx context.Context, createdByUserID int64, input domain.CreateTestInput, publicSlug string) (domain.Test, error) {
	row := r.db.QueryRowContext(
		ctx,
		`INSERT INTO tests (
			title, description, created_by_user_id, report_template_id, recommended_duration, max_participants,
			collect_respondent_age, collect_respondent_gender, collect_respondent_education, status, public_slug, is_public
		)
		 VALUES ($1, $2, $3, NULLIF($4, 0), $5, $6, $7, $8, $9, $10, $11, TRUE)
		 RETURNING id, title, description, created_by_user_id, COALESCE(report_template_id, 0), recommended_duration, max_participants,
		 	collect_respondent_age, collect_respondent_gender, collect_respondent_education, status, COALESCE(public_slug, ''), is_public,
		 	0 AS started_sessions_count, 0 AS in_progress_sessions_count, 0 AS completed_sessions_count,
		 	NULL::timestamptz AS last_started_at, NULL::timestamptz AS last_completed_at, NULL::timestamptz AS last_activity_at,
		 	created_at, updated_at`,
		input.Title,
		input.Description,
		createdByUserID,
		input.ReportTemplateID,
		input.RecommendedDuration,
		input.MaxParticipants,
		input.CollectRespondentAge,
		input.CollectRespondentGender,
		input.CollectRespondentEducation,
		input.Status,
		publicSlug,
	)

	return scanTest(row)
}

func (r *AppRepository) ListPsychologistTests(ctx context.Context, createdByUserID int64) ([]domain.Test, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT t.id, t.title, t.description, t.created_by_user_id, COALESCE(t.report_template_id, 0), t.recommended_duration, t.max_participants,
		 	t.collect_respondent_age, t.collect_respondent_gender, t.collect_respondent_education, t.status, COALESCE(t.public_slug, ''), t.is_public,
		 	(
		 		SELECT COUNT(*)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = t.id
		 	) AS started_sessions_count,
		 	(
		 		SELECT COUNT(*)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = t.id
		 		  AND s.status = 'in_progress'
		 	) AS in_progress_sessions_count,
		 	(
		 		SELECT COUNT(*)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = t.id
		 		  AND s.status = 'completed'
		 	) AS completed_sessions_count,
		 	(
		 		SELECT MAX(s.started_at)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = t.id
		 	) AS last_started_at,
		 	(
		 		SELECT MAX(s.completed_at)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = t.id
		 		  AND s.completed_at IS NOT NULL
		 	) AS last_completed_at,
		 	(
		 		SELECT MAX(COALESCE(s.completed_at, s.started_at))
		 		FROM public_test_sessions s
		 		WHERE s.test_id = t.id
		 	) AS last_activity_at,
		 	t.created_at, t.updated_at
		 FROM tests t
		 WHERE created_by_user_id = $1
		 ORDER BY t.id DESC`,
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
		`SELECT t.id, t.title, t.description, t.created_by_user_id, COALESCE(t.report_template_id, 0), t.recommended_duration, t.max_participants,
		 	t.collect_respondent_age, t.collect_respondent_gender, t.collect_respondent_education, t.status, COALESCE(t.public_slug, ''), t.is_public,
		 	(
		 		SELECT COUNT(*)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = t.id
		 	) AS started_sessions_count,
		 	(
		 		SELECT COUNT(*)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = t.id
		 		  AND s.status = 'in_progress'
		 	) AS in_progress_sessions_count,
		 	(
		 		SELECT COUNT(*)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = t.id
		 		  AND s.status = 'completed'
		 	) AS completed_sessions_count,
		 	(
		 		SELECT MAX(s.started_at)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = t.id
		 	) AS last_started_at,
		 	(
		 		SELECT MAX(s.completed_at)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = t.id
		 		  AND s.completed_at IS NOT NULL
		 	) AS last_completed_at,
		 	(
		 		SELECT MAX(COALESCE(s.completed_at, s.started_at))
		 		FROM public_test_sessions s
		 		WHERE s.test_id = t.id
		 	) AS last_activity_at,
		 	t.created_at, t.updated_at
		 FROM tests t
		 WHERE t.id = $1 AND t.created_by_user_id = $2`,
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
		 	collect_respondent_age = $8,
		 	collect_respondent_gender = $9,
		 	collect_respondent_education = $10,
		 	status = $11,
		 	updated_at = NOW()
		 WHERE id = $1 AND created_by_user_id = $2
		 RETURNING id, title, description, created_by_user_id, COALESCE(report_template_id, 0), recommended_duration, max_participants,
		 	collect_respondent_age, collect_respondent_gender, collect_respondent_education, status, COALESCE(public_slug, ''), is_public,
		 	(
		 		SELECT COUNT(*)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = tests.id
		 	) AS started_sessions_count,
		 	(
		 		SELECT COUNT(*)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = tests.id
		 		  AND s.status = 'in_progress'
		 	) AS in_progress_sessions_count,
		 	(
		 		SELECT COUNT(*)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = tests.id
		 		  AND s.status = 'completed'
		 	) AS completed_sessions_count,
		 	(
		 		SELECT MAX(s.started_at)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = tests.id
		 	) AS last_started_at,
		 	(
		 		SELECT MAX(s.completed_at)
		 		FROM public_test_sessions s
		 		WHERE s.test_id = tests.id
		 		  AND s.completed_at IS NOT NULL
		 	) AS last_completed_at,
		 	(
		 		SELECT MAX(COALESCE(s.completed_at, s.started_at))
		 		FROM public_test_sessions s
		 		WHERE s.test_id = tests.id
		 	) AS last_activity_at,
		 	created_at, updated_at`,
		testID,
		createdByUserID,
		input.Title,
		input.Description,
		input.ReportTemplateID,
		input.RecommendedDuration,
		input.MaxParticipants,
		input.CollectRespondentAge,
		input.CollectRespondentGender,
		input.CollectRespondentEducation,
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
	var lastStartedAt sql.NullTime
	var lastCompletedAt sql.NullTime
	var lastActivityAt sql.NullTime
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
		&test.CollectRespondentAge,
		&test.CollectRespondentGender,
		&test.CollectRespondentEducation,
		&test.Status,
		&test.PublicSlug,
		&test.IsPublic,
		&test.StartedSessionsCount,
		&test.InProgressSessionsCount,
		&test.CompletedSessionsCount,
		&lastStartedAt,
		&lastCompletedAt,
		&lastActivityAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.Test{}, err
	}

	test.LastStartedAt = formatNullTime(lastStartedAt)
	test.LastCompletedAt = formatNullTime(lastCompletedAt)
	test.LastActivityAt = formatNullTime(lastActivityAt)
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

func HasConstraintViolation(err error, name string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == name
}

func IsDuplicatePublicTestPhone(err error) bool {
	return errors.Is(err, errDuplicateRespondentPhone) || HasConstraintViolation(err, "idx_public_test_sessions_test_phone_unique")
}

func IsLimitReached(err error) bool {
	return errors.Is(err, errLimitReached)
}
