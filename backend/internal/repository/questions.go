package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func (r *AppRepository) CreatePsychologistQuestion(ctx context.Context, testID int64, createdByUserID int64, input domain.CreateQuestionInput) (domain.Question, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Question{}, err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(
		ctx,
		`INSERT INTO questions (test_id, text, question_type, order_number, is_required)
		 SELECT
			$1,
			$3,
			$4,
			COALESCE(NULLIF($5, 0), (SELECT COALESCE(MAX(q.order_number), 0) + 1 FROM questions q WHERE q.test_id = $1)),
			$6
		 FROM tests t
		 WHERE t.id = $1 AND t.created_by_user_id = $2
		 RETURNING id, test_id, text, question_type, order_number, is_required, created_at, updated_at`,
		testID,
		createdByUserID,
		input.Text,
		input.QuestionType,
		input.OrderNumber,
		input.IsRequired,
	)

	question, err := scanQuestionBase(row)
	if err != nil {
		return domain.Question{}, err
	}

	if err := replaceQuestionOptionsTx(ctx, tx, question.ID, input.Options); err != nil {
		return domain.Question{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Question{}, err
	}

	return r.GetPsychologistQuestionByID(ctx, testID, question.ID, createdByUserID)
}

func (r *AppRepository) ListPsychologistQuestions(ctx context.Context, testID int64, createdByUserID int64) ([]domain.Question, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT q.id, q.test_id, q.text, q.question_type, q.order_number, q.is_required, q.created_at, q.updated_at
		 FROM questions q
		 JOIN tests t ON t.id = q.test_id
		 WHERE q.test_id = $1 AND t.created_by_user_id = $2
		 ORDER BY q.order_number, q.id`,
		testID,
		createdByUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions := make([]domain.Question, 0)
	for rows.Next() {
		question, scanErr := scanQuestionBase(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		questions = append(questions, question)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(questions) == 0 {
		return []domain.Question{}, nil
	}

	optionsByQuestionID, err := r.listQuestionOptionsForTest(ctx, testID, createdByUserID)
	if err != nil {
		return nil, err
	}

	for i := range questions {
		questions[i].Options = optionsByQuestionID[questions[i].ID]
	}

	return questions, nil
}

func (r *AppRepository) GetPsychologistQuestionByID(ctx context.Context, testID int64, questionID int64, createdByUserID int64) (domain.Question, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT q.id, q.test_id, q.text, q.question_type, q.order_number, q.is_required, q.created_at, q.updated_at
		 FROM questions q
		 JOIN tests t ON t.id = q.test_id
		 WHERE q.id = $1 AND q.test_id = $2 AND t.created_by_user_id = $3`,
		questionID,
		testID,
		createdByUserID,
	)

	question, err := scanQuestionBase(row)
	if err != nil {
		return domain.Question{}, err
	}

	options, err := r.listQuestionOptionsByQuestionID(ctx, question.ID)
	if err != nil {
		return domain.Question{}, err
	}
	question.Options = options

	return question, nil
}

func (r *AppRepository) UpdatePsychologistQuestion(ctx context.Context, testID int64, questionID int64, createdByUserID int64, input domain.UpdateQuestionInput) (domain.Question, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Question{}, err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(
		ctx,
		`UPDATE questions q
		 SET text = $4,
		 	question_type = $5,
		 	order_number = COALESCE(NULLIF($6, 0), q.order_number),
		 	is_required = $7,
		 	updated_at = NOW()
		 FROM tests t
		 WHERE q.id = $1
		   AND q.test_id = $2
		   AND t.id = q.test_id
		   AND t.created_by_user_id = $3
		 RETURNING q.id, q.test_id, q.text, q.question_type, q.order_number, q.is_required, q.created_at, q.updated_at`,
		questionID,
		testID,
		createdByUserID,
		input.Text,
		input.QuestionType,
		input.OrderNumber,
		input.IsRequired,
	)

	question, err := scanQuestionBase(row)
	if err != nil {
		return domain.Question{}, err
	}

	if err := replaceQuestionOptionsTx(ctx, tx, question.ID, input.Options); err != nil {
		return domain.Question{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Question{}, err
	}

	return r.GetPsychologistQuestionByID(ctx, testID, question.ID, createdByUserID)
}

func (r *AppRepository) DeletePsychologistQuestion(ctx context.Context, testID int64, questionID int64, createdByUserID int64) (bool, error) {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM questions q
		 USING tests t
		 WHERE q.id = $1
		   AND q.test_id = $2
		   AND t.id = q.test_id
		   AND t.created_by_user_id = $3`,
		questionID,
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

func (r *AppRepository) listQuestionOptionsForTest(ctx context.Context, testID int64, createdByUserID int64) (map[int64][]domain.QuestionOption, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT o.id, o.question_id, o.label, o.value, o.order_number, o.score
		 FROM question_options o
		 JOIN questions q ON q.id = o.question_id
		 JOIN tests t ON t.id = q.test_id
		 WHERE q.test_id = $1 AND t.created_by_user_id = $2
		 ORDER BY o.question_id, o.order_number, o.id`,
		testID,
		createdByUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]domain.QuestionOption)
	for rows.Next() {
		var option domain.QuestionOption
		if err := rows.Scan(
			&option.ID,
			&option.QuestionID,
			&option.Label,
			&option.Value,
			&option.OrderNumber,
			&option.Score,
		); err != nil {
			return nil, err
		}
		result[option.QuestionID] = append(result[option.QuestionID], option)
	}

	return result, rows.Err()
}

func (r *AppRepository) listQuestionOptionsByQuestionID(ctx context.Context, questionID int64) ([]domain.QuestionOption, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, question_id, label, value, order_number, score
		 FROM question_options
		 WHERE question_id = $1
		 ORDER BY order_number, id`,
		questionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := make([]domain.QuestionOption, 0)
	for rows.Next() {
		var option domain.QuestionOption
		if err := rows.Scan(
			&option.ID,
			&option.QuestionID,
			&option.Label,
			&option.Value,
			&option.OrderNumber,
			&option.Score,
		); err != nil {
			return nil, err
		}
		options = append(options, option)
	}

	return options, rows.Err()
}

func replaceQuestionOptionsTx(ctx context.Context, tx *sql.Tx, questionID int64, options []domain.QuestionOptionInput) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM question_options WHERE question_id = $1`, questionID); err != nil {
		return err
	}

	for i, option := range options {
		orderNumber := option.OrderNumber
		if orderNumber <= 0 {
			orderNumber = i + 1
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO question_options (question_id, label, value, order_number, score)
			 VALUES ($1, $2, $3, $4, $5)`,
			questionID,
			option.Label,
			option.Value,
			orderNumber,
			option.Score,
		); err != nil {
			return err
		}
	}

	return nil
}

func scanQuestionBase(scanner rowScanner) (domain.Question, error) {
	var question domain.Question
	var createdAt time.Time
	var updatedAt time.Time

	err := scanner.Scan(
		&question.ID,
		&question.TestID,
		&question.Text,
		&question.QuestionType,
		&question.OrderNumber,
		&question.IsRequired,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.Question{}, err
	}

	question.CreatedAt = createdAt.Format(time.RFC3339)
	question.UpdatedAt = updatedAt.Format(time.RFC3339)

	return question, nil
}
