package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func (r *AppRepository) DeleteExpiredPublicTestSessions(ctx context.Context) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM public_test_sessions
		 WHERE status = 'in_progress'
		   AND expires_at <= NOW()`,
	)
	return err
}

func (r *AppRepository) PublishPsychologistTest(ctx context.Context, testID int64, createdByUserID int64, slug string) (domain.Test, error) {
	row := r.db.QueryRowContext(
		ctx,
		`UPDATE tests
		 SET status = 'published',
		 	is_public = TRUE,
		 	public_slug = COALESCE(NULLIF(public_slug, ''), $3),
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
		slug,
	)

	return scanTest(row)
}

func (r *AppRepository) GetPublicTestBySlug(ctx context.Context, slug string) (domain.PublicTest, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT t.id, t.public_slug, t.title, t.description, t.recommended_duration, t.max_participants,
		 	t.collect_respondent_age, t.collect_respondent_gender, t.collect_respondent_education
		 FROM tests t
		 WHERE t.public_slug = $1
		   AND t.is_public = TRUE
		   AND t.status = 'published'`,
		slug,
	)

	var test domain.PublicTest
	if err := row.Scan(
		&test.ID,
		&test.Slug,
		&test.Title,
		&test.Description,
		&test.RecommendedDuration,
		&test.MaxParticipants,
		&test.CollectRespondentAge,
		&test.CollectRespondentGender,
		&test.CollectRespondentEducation,
	); err != nil {
		return domain.PublicTest{}, err
	}

	questions, err := r.listPublicQuestionsByTestID(ctx, test.ID)
	if err != nil {
		return domain.PublicTest{}, err
	}
	test.Questions = questions

	return test, nil
}

func (r *AppRepository) GetPublicTestBySlugAndAccessToken(ctx context.Context, slug string, accessToken string) (domain.PublicTest, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT t.id, t.public_slug, t.title, t.description, t.recommended_duration, t.max_participants,
		 	t.collect_respondent_age, t.collect_respondent_gender, t.collect_respondent_education
		 FROM tests t
		 JOIN public_test_sessions s ON s.test_id = t.id
		 WHERE t.public_slug = $1
		   AND t.is_public = TRUE
		   AND s.access_token = $2`,
		slug,
		accessToken,
	)

	var test domain.PublicTest
	if err := row.Scan(
		&test.ID,
		&test.Slug,
		&test.Title,
		&test.Description,
		&test.RecommendedDuration,
		&test.MaxParticipants,
		&test.CollectRespondentAge,
		&test.CollectRespondentGender,
		&test.CollectRespondentEducation,
	); err != nil {
		return domain.PublicTest{}, err
	}

	questions, err := r.listPublicQuestionsByTestID(ctx, test.ID)
	if err != nil {
		return domain.PublicTest{}, err
	}
	test.Questions = questions

	return test, nil
}

func (r *AppRepository) GetPublicTestAccessInfoBySlug(ctx context.Context, slug string) (domain.PublicTestAccessInfo, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT t.id, t.status, t.is_public, t.max_participants, COUNT(s.id)
		 FROM tests t
		 LEFT JOIN public_test_sessions s ON s.test_id = t.id
		 WHERE t.public_slug = $1
		 GROUP BY t.id, t.status, t.is_public, t.max_participants`,
		slug,
	)

	var info domain.PublicTestAccessInfo
	if err := row.Scan(
		&info.TestID,
		&info.Status,
		&info.IsPublic,
		&info.MaxParticipants,
		&info.CurrentSessions,
	); err != nil {
		return domain.PublicTestAccessInfo{}, err
	}

	return info, nil
}

func (r *AppRepository) GetPublicTestSessionByPhone(ctx context.Context, slug string, phone string) (domain.PublicTestSession, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT s.id, s.test_id, s.access_token, s.respondent_name, s.respondent_phone, s.respondent_email,
		 	COALESCE(s.respondent_age, 0), s.respondent_gender, s.respondent_education, s.status,
		 	s.started_at, s.expires_at, s.completed_at
		 FROM public_test_sessions s
		 JOIN tests t ON t.id = s.test_id
		 WHERE t.public_slug = $1
		   AND t.is_public = TRUE
		   AND s.respondent_phone = $2`,
		slug,
		phone,
	)

	return scanPublicTestSession(row)
}

func (r *AppRepository) GetPublicTestSessionByAccessToken(ctx context.Context, slug string, accessToken string) (domain.PublicTestSession, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT s.id, s.test_id, s.access_token, s.respondent_name, s.respondent_phone, s.respondent_email,
		 	COALESCE(s.respondent_age, 0), s.respondent_gender, s.respondent_education, s.status,
		 	s.started_at, s.expires_at, s.completed_at
		 FROM public_test_sessions s
		 JOIN tests t ON t.id = s.test_id
		 WHERE t.public_slug = $1
		   AND t.is_public = TRUE
		   AND s.access_token = $2`,
		slug,
		accessToken,
	)

	return scanPublicTestSession(row)
}

func (r *AppRepository) ExtendPublicTestSessionExpiry(ctx context.Context, sessionID int64, expiresAt time.Time) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE public_test_sessions
		 SET expires_at = $2
		 WHERE id = $1
		   AND status = 'in_progress'`,
		sessionID,
		expiresAt,
	)
	return err
}

func (r *AppRepository) StartPublicTestSession(ctx context.Context, slug string, accessToken string, input domain.StartPublicTestInput, expiresAt time.Time) (domain.PublicTestSession, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.PublicTestSession{}, err
	}
	defer tx.Rollback()

	testRow := tx.QueryRowContext(
		ctx,
		`SELECT id, max_participants
		 FROM tests
		 WHERE public_slug = $1
		   AND is_public = TRUE
		   AND status = 'published'
		 FOR UPDATE`,
		slug,
	)

	var testID int64
	var maxParticipants int
	if err := testRow.Scan(&testID, &maxParticipants); err != nil {
		return domain.PublicTestSession{}, err
	}

	var currentSessions int
	if err := tx.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM public_test_sessions WHERE test_id = $1`,
		testID,
	).Scan(&currentSessions); err != nil {
		return domain.PublicTestSession{}, err
	}
	if maxParticipants > 0 && currentSessions >= maxParticipants {
		return domain.PublicTestSession{}, errLimitReached
	}

	var duplicatePhone bool
	if err := tx.QueryRowContext(
		ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM public_test_sessions
			WHERE test_id = $1
			  AND respondent_phone = $2
		)`,
		testID,
		input.RespondentPhone,
	).Scan(&duplicatePhone); err != nil {
		return domain.PublicTestSession{}, err
	}
	if duplicatePhone {
		return domain.PublicTestSession{}, errDuplicateRespondentPhone
	}

	row := tx.QueryRowContext(
		ctx,
		`INSERT INTO public_test_sessions (
			test_id, access_token, respondent_name, respondent_phone, respondent_email,
			respondent_age, respondent_gender, respondent_education, expires_at
		)
		 VALUES ($1, $2, $3, $4, $5, NULLIF($6, 0), $7, $8, $9)
		 RETURNING id, test_id, access_token, respondent_name, respondent_phone, respondent_email,
		 	COALESCE(respondent_age, 0), respondent_gender, respondent_education, status, started_at, expires_at, completed_at`,
		testID,
		accessToken,
		input.RespondentName,
		input.RespondentPhone,
		input.RespondentEmail,
		input.RespondentAge,
		input.RespondentGender,
		input.RespondentEducation,
		expiresAt,
	)

	session, err := scanPublicTestSession(row)
	if err != nil {
		return domain.PublicTestSession{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.PublicTestSession{}, err
	}

	return session, nil
}

func (r *AppRepository) SavePublicTestAnswers(ctx context.Context, slug string, accessToken string, answers []domain.PublicAnswerInput, expiresAt time.Time, complete bool, careerResult *domain.CareerResult) (domain.SubmitPublicTestResponse, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.SubmitPublicTestResponse{}, err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(
		ctx,
		`SELECT s.id, s.test_id
		 FROM public_test_sessions s
		 JOIN tests t ON t.id = s.test_id
		 WHERE t.public_slug = $1
		   AND t.is_public = TRUE
		   AND s.access_token = $2`,
		slug,
		accessToken,
	)

	var sessionID int64
	var testID int64
	if err := row.Scan(&sessionID, &testID); err != nil {
		return domain.SubmitPublicTestResponse{}, err
	}

	for _, answer := range answers {
		answerValuesJSON, err := json.Marshal(answer.AnswerValues)
		if err != nil {
			return domain.SubmitPublicTestResponse{}, err
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO public_test_answers (session_id, question_id, answer_text, answer_value, answer_values_json)
			 VALUES ($1, $2, $3, $4, $5::jsonb)
			 ON CONFLICT (session_id, question_id)
			 DO UPDATE SET
			 	answer_text = EXCLUDED.answer_text,
			 	answer_value = EXCLUDED.answer_value,
			 	answer_values_json = EXCLUDED.answer_values_json,
			 	updated_at = NOW()`,
			sessionID,
			answer.QuestionID,
			answer.AnswerText,
			answer.AnswerValue,
			string(answerValuesJSON),
		); err != nil {
			return domain.SubmitPublicTestResponse{}, err
		}
	}

	if complete {
		careerResultJSON, err := marshalCareerResult(careerResult)
		if err != nil {
			return domain.SubmitPublicTestResponse{}, err
		}
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE public_test_sessions
			 SET status = 'completed',
			 	completed_at = NOW(),
			 	expires_at = NOW(),
			 	career_result_json = $2::jsonb
			 WHERE id = $1`,
			sessionID,
			careerResultJSON,
		); err != nil {
			return domain.SubmitPublicTestResponse{}, err
		}
	} else {
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE public_test_sessions
			 SET expires_at = $2
			 WHERE id = $1
			   AND status = 'in_progress'`,
			sessionID,
			expiresAt,
		); err != nil {
			return domain.SubmitPublicTestResponse{}, err
		}
	}

	savedAnswers, err := listPublicTestAnswersTx(ctx, tx, sessionID, testID)
	if err != nil {
		return domain.SubmitPublicTestResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.SubmitPublicTestResponse{}, err
	}

	return domain.SubmitPublicTestResponse{
		SessionID: sessionID,
		Status:    statusForPublicTestSave(complete),
		Answers:   savedAnswers,
	}, nil
}

func (r *AppRepository) ListPublicTestAnswersBySessionID(ctx context.Context, sessionID int64, testID int64) ([]domain.PublicTestAnswer, error) {
	return listPublicTestAnswersTx(ctx, r.db, sessionID, testID)
}

func (r *AppRepository) ListPsychologistTestSubmissions(ctx context.Context, testID int64, createdByUserID int64) ([]domain.PsychologistTestSubmission, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT s.id, s.test_id, s.respondent_name, s.respondent_phone, s.respondent_email, COALESCE(s.respondent_age, 0),
		 	s.respondent_gender, s.respondent_education, s.status, s.started_at, s.completed_at, COUNT(a.id) AS answers_count
		 FROM public_test_sessions s
		 JOIN tests t ON t.id = s.test_id
		 LEFT JOIN public_test_answers a ON a.session_id = s.id
		 WHERE s.test_id = $1
		   AND t.created_by_user_id = $2
		 GROUP BY s.id, s.test_id, s.respondent_name, s.respondent_phone, s.respondent_email, s.respondent_age, s.respondent_gender, s.respondent_education, s.status, s.started_at, s.completed_at
		 ORDER BY s.started_at DESC, s.id DESC`,
		testID,
		createdByUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	submissions := make([]domain.PsychologistTestSubmission, 0)
	for rows.Next() {
		var submission domain.PsychologistTestSubmission
		var startedAt time.Time
		var completedAt sql.NullTime
		if err := rows.Scan(
			&submission.SessionID,
			&submission.TestID,
			&submission.RespondentName,
			&submission.RespondentPhone,
			&submission.RespondentEmail,
			&submission.RespondentAge,
			&submission.RespondentGender,
			&submission.RespondentEducation,
			&submission.Status,
			&startedAt,
			&completedAt,
			&submission.AnswersCount,
		); err != nil {
			return nil, err
		}
		submission.StartedAt = startedAt.Format(time.RFC3339)
		if completedAt.Valid {
			submission.CompletedAt = completedAt.Time.Format(time.RFC3339)
		}
		submissions = append(submissions, submission)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(submissions) == 0 {
		return []domain.PsychologistTestSubmission{}, nil
	}

	return submissions, nil
}

func (r *AppRepository) GetPsychologistTestSubmissionByID(ctx context.Context, testID int64, sessionID int64, createdByUserID int64) (domain.PsychologistTestSubmission, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT s.id, s.test_id, s.respondent_name, s.respondent_phone, s.respondent_email, COALESCE(s.respondent_age, 0),
		 	s.respondent_gender, s.respondent_education, s.status, s.started_at, s.completed_at, COUNT(a.id) AS answers_count, s.career_result_json
		 FROM public_test_sessions s
		 JOIN tests t ON t.id = s.test_id
		 LEFT JOIN public_test_answers a ON a.session_id = s.id
		 WHERE s.test_id = $1
		   AND s.id = $2
		   AND t.created_by_user_id = $3
		 GROUP BY s.id, s.test_id, s.respondent_name, s.respondent_phone, s.respondent_email, s.respondent_age, s.respondent_gender, s.respondent_education, s.status, s.started_at, s.completed_at, s.career_result_json`,
		testID,
		sessionID,
		createdByUserID,
	)

	var submission domain.PsychologistTestSubmission
	var startedAt time.Time
	var completedAt sql.NullTime
	var rawCareerResult []byte
	if err := row.Scan(
		&submission.SessionID,
		&submission.TestID,
		&submission.RespondentName,
		&submission.RespondentPhone,
		&submission.RespondentEmail,
		&submission.RespondentAge,
		&submission.RespondentGender,
		&submission.RespondentEducation,
		&submission.Status,
		&startedAt,
		&completedAt,
		&submission.AnswersCount,
		&rawCareerResult,
	); err != nil {
		return domain.PsychologistTestSubmission{}, err
	}

	submission.StartedAt = startedAt.Format(time.RFC3339)
	if completedAt.Valid {
		submission.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}
	decodedCareerResult, err := unmarshalCareerResult(rawCareerResult)
	if err != nil {
		return domain.PsychologistTestSubmission{}, err
	}
	submission.CareerResult = decodedCareerResult

	answers, err := listPublicTestAnswersTx(ctx, r.db, sessionID, testID)
	if err != nil {
		return domain.PsychologistTestSubmission{}, err
	}
	submission.Answers = answers

	return submission, nil
}

func (r *AppRepository) GetPsychologistTestSubmissionBySessionID(ctx context.Context, sessionID int64, createdByUserID int64) (domain.PsychologistTestSubmission, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT s.id, s.test_id, s.respondent_name, s.respondent_phone, s.respondent_email, COALESCE(s.respondent_age, 0),
		 	s.respondent_gender, s.respondent_education, s.status, s.started_at, s.completed_at, COUNT(a.id) AS answers_count, s.career_result_json
		 FROM public_test_sessions s
		 JOIN tests t ON t.id = s.test_id
		 LEFT JOIN public_test_answers a ON a.session_id = s.id
		 WHERE s.id = $1
		   AND t.created_by_user_id = $2
		 GROUP BY s.id, s.test_id, s.respondent_name, s.respondent_phone, s.respondent_email, s.respondent_age, s.respondent_gender, s.respondent_education, s.status, s.started_at, s.completed_at, s.career_result_json`,
		sessionID,
		createdByUserID,
	)

	var submission domain.PsychologistTestSubmission
	var startedAt time.Time
	var completedAt sql.NullTime
	var rawCareerResult []byte
	if err := row.Scan(
		&submission.SessionID,
		&submission.TestID,
		&submission.RespondentName,
		&submission.RespondentPhone,
		&submission.RespondentEmail,
		&submission.RespondentAge,
		&submission.RespondentGender,
		&submission.RespondentEducation,
		&submission.Status,
		&startedAt,
		&completedAt,
		&submission.AnswersCount,
		&rawCareerResult,
	); err != nil {
		return domain.PsychologistTestSubmission{}, err
	}

	submission.StartedAt = startedAt.Format(time.RFC3339)
	if completedAt.Valid {
		submission.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}
	decodedCareerResult, err := unmarshalCareerResult(rawCareerResult)
	if err != nil {
		return domain.PsychologistTestSubmission{}, err
	}
	submission.CareerResult = decodedCareerResult

	answers, err := listPublicTestAnswersTx(ctx, r.db, sessionID, submission.TestID)
	if err != nil {
		return domain.PsychologistTestSubmission{}, err
	}
	submission.Answers = answers

	return submission, nil
}

func (r *AppRepository) listPublicQuestionsByTestID(ctx context.Context, testID int64) ([]domain.PublicQuestion, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, text, question_type, order_number, is_required, scale_weights_json
		 FROM questions
		 WHERE test_id = $1
		 ORDER BY order_number, id`,
		testID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions := make([]domain.PublicQuestion, 0)
	for rows.Next() {
		var question domain.PublicQuestion
		var rawScaleWeights []byte
		if err := rows.Scan(
			&question.ID,
			&question.Text,
			&question.QuestionType,
			&question.OrderNumber,
			&question.IsRequired,
			&rawScaleWeights,
		); err != nil {
			return nil, err
		}
		if len(rawScaleWeights) > 0 {
			if err := json.Unmarshal(rawScaleWeights, &question.ScaleWeights); err != nil {
				return nil, err
			}
		}
		if question.ScaleWeights == nil {
			question.ScaleWeights = map[string]float64{}
		}
		questions = append(questions, question)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(questions) == 0 {
		return []domain.PublicQuestion{}, nil
	}

	optionsByQuestionID, err := r.listPublicQuestionOptionsForTest(ctx, testID)
	if err != nil {
		return nil, err
	}

	for i := range questions {
		questions[i].Options = optionsByQuestionID[questions[i].ID]
	}

	return questions, nil
}

func (r *AppRepository) listPublicQuestionOptionsForTest(ctx context.Context, testID int64) (map[int64][]domain.PublicQuestionOption, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT o.question_id, o.id, o.label, o.value, o.order_number, o.score
		 FROM question_options o
		 JOIN questions q ON q.id = o.question_id
		 WHERE q.test_id = $1
		 ORDER BY o.question_id, o.order_number, o.id`,
		testID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]domain.PublicQuestionOption)
	for rows.Next() {
		var questionID int64
		var option domain.PublicQuestionOption
		if err := rows.Scan(
			&questionID,
			&option.ID,
			&option.Label,
			&option.Value,
			&option.OrderNumber,
			&option.Score,
		); err != nil {
			return nil, err
		}
		result[questionID] = append(result[questionID], option)
	}

	return result, rows.Err()
}

func marshalCareerResult(result *domain.CareerResult) (any, error) {
	if result == nil {
		return nil, nil
	}

	content, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return string(content), nil
}

func unmarshalCareerResult(raw []byte) (*domain.CareerResult, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var result domain.CareerResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

type queryContext interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func listPublicTestAnswersTx(ctx context.Context, db queryContext, sessionID int64, testID int64) ([]domain.PublicTestAnswer, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT a.id, a.session_id, a.question_id, a.answer_text, a.answer_value, a.answer_values_json, a.created_at, a.updated_at
		 FROM public_test_answers a
		 JOIN questions q ON q.id = a.question_id
		 WHERE a.session_id = $1
		   AND q.test_id = $2
		 ORDER BY q.order_number, q.id`,
		sessionID,
		testID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	answers := make([]domain.PublicTestAnswer, 0)
	for rows.Next() {
		var answer domain.PublicTestAnswer
		var rawValues []byte
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(
			&answer.ID,
			&answer.SessionID,
			&answer.QuestionID,
			&answer.AnswerText,
			&answer.AnswerValue,
			&rawValues,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		if len(rawValues) > 0 {
			if err := json.Unmarshal(rawValues, &answer.AnswerValues); err != nil {
				return nil, err
			}
		}
		answer.CreatedAt = createdAt.Format(time.RFC3339)
		answer.UpdatedAt = updatedAt.Format(time.RFC3339)
		answers = append(answers, answer)
	}

	return answers, rows.Err()
}

func scanPublicTestSession(scanner rowScanner) (domain.PublicTestSession, error) {
	var session domain.PublicTestSession
	var startedAt time.Time
	var expiresAt time.Time
	var completedAt sql.NullTime

	if err := scanner.Scan(
		&session.ID,
		&session.TestID,
		&session.AccessToken,
		&session.RespondentName,
		&session.RespondentPhone,
		&session.RespondentEmail,
		&session.RespondentAge,
		&session.RespondentGender,
		&session.RespondentEducation,
		&session.Status,
		&startedAt,
		&expiresAt,
		&completedAt,
	); err != nil {
		return domain.PublicTestSession{}, err
	}

	session.StartedAt = startedAt.Format(time.RFC3339)
	session.ExpiresAt = expiresAt.Format(time.RFC3339)
	if completedAt.Valid {
		session.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}

	return session, nil
}

func statusForPublicTestSave(complete bool) string {
	if complete {
		return "completed"
	}

	return "in_progress"
}
