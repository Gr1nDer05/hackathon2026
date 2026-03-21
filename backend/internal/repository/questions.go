package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func (r *AppRepository) CreatePsychologistQuestion(ctx context.Context, testID int64, createdByUserID int64, input domain.CreateQuestionInput) (domain.Question, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Question{}, err
	}
	defer tx.Rollback()

	if err := lockPsychologistTestForQuestionMutationTx(ctx, tx, testID, createdByUserID); err != nil {
		return domain.Question{}, err
	}

	orderStates, err := listQuestionOrderStatesForUpdateTx(ctx, tx, testID, createdByUserID)
	if err != nil {
		return domain.Question{}, err
	}

	row := tx.QueryRowContext(
		ctx,
		`INSERT INTO questions (test_id, text, question_type, order_number, is_required, scale_weights_json)
		 SELECT
			$1,
			$3,
			$4,
			$5,
			$6,
			$7::jsonb
		 FROM tests t
		 WHERE t.id = $1 AND t.created_by_user_id = $2
		 RETURNING id, test_id, text, question_type, order_number, is_required, scale_weights_json, created_at, updated_at`,
		testID,
		createdByUserID,
		input.Text,
		input.QuestionType,
		nextQuestionOrder(orderStates),
		input.IsRequired,
		mustMarshalScaleWeights(input.ScaleWeights),
	)

	question, err := scanQuestionBase(row)
	if err != nil {
		return domain.Question{}, err
	}

	orderedQuestionIDs := insertQuestionID(questionOrderIDs(orderStates), question.ID, input.OrderNumber)
	if err := resequenceQuestionOrderNumbersTx(ctx, tx, testID, orderedQuestionIDs); err != nil {
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
		`SELECT q.id, q.test_id, q.text, q.question_type, q.order_number, q.is_required, q.scale_weights_json, q.created_at, q.updated_at
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
		`SELECT q.id, q.test_id, q.text, q.question_type, q.order_number, q.is_required, q.scale_weights_json, q.created_at, q.updated_at
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

	if err := lockPsychologistTestForQuestionMutationTx(ctx, tx, testID, createdByUserID); err != nil {
		return domain.Question{}, err
	}

	orderStates, err := listQuestionOrderStatesForUpdateTx(ctx, tx, testID, createdByUserID)
	if err != nil {
		return domain.Question{}, err
	}

	orderedQuestionIDs, found := moveQuestionID(questionOrderIDs(orderStates), questionID, input.OrderNumber)
	if !found {
		return domain.Question{}, sql.ErrNoRows
	}

	result, err := tx.ExecContext(
		ctx,
		`UPDATE questions q
		 SET text = $4,
		 	question_type = $5,
		 	is_required = $6,
		 	scale_weights_json = COALESCE($7::jsonb, q.scale_weights_json),
		 	updated_at = NOW()
		 FROM tests t
		 WHERE q.id = $1
		   AND q.test_id = $2
		   AND t.id = q.test_id
		   AND t.created_by_user_id = $3`,
		questionID,
		testID,
		createdByUserID,
		input.Text,
		input.QuestionType,
		input.IsRequired,
		marshalOptionalScaleWeights(input.ScaleWeights),
	)
	if err != nil {
		return domain.Question{}, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.Question{}, err
	}
	if rowsAffected == 0 {
		return domain.Question{}, sql.ErrNoRows
	}

	if err := resequenceQuestionOrderNumbersTx(ctx, tx, testID, orderedQuestionIDs); err != nil {
		return domain.Question{}, err
	}

	if err := replaceQuestionOptionsTx(ctx, tx, questionID, input.Options); err != nil {
		return domain.Question{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Question{}, err
	}

	return r.GetPsychologistQuestionByID(ctx, testID, questionID, createdByUserID)
}

func (r *AppRepository) DeletePsychologistQuestion(ctx context.Context, testID int64, questionID int64, createdByUserID int64) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	if err := lockPsychologistTestForQuestionMutationTx(ctx, tx, testID, createdByUserID); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	orderStates, err := listQuestionOrderStatesForUpdateTx(ctx, tx, testID, createdByUserID)
	if err != nil {
		return false, err
	}

	remainingQuestionIDs, found := removeQuestionID(questionOrderIDs(orderStates), questionID)
	if !found {
		return false, nil
	}

	result, err := tx.ExecContext(
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

	if rowsAffected == 0 {
		return false, nil
	}
	if err := resequenceQuestionOrderNumbersTx(ctx, tx, testID, remainingQuestionIDs); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
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

type questionOrderState struct {
	ID          int64
	OrderNumber int
}

func lockPsychologistTestForQuestionMutationTx(ctx context.Context, tx *sql.Tx, testID int64, createdByUserID int64) error {
	var lockedTestID int64
	return tx.QueryRowContext(
		ctx,
		`SELECT id
		 FROM tests
		 WHERE id = $1
		   AND created_by_user_id = $2
		 FOR UPDATE`,
		testID,
		createdByUserID,
	).Scan(&lockedTestID)
}

func listQuestionOrderStatesForUpdateTx(ctx context.Context, tx *sql.Tx, testID int64, createdByUserID int64) ([]questionOrderState, error) {
	rows, err := tx.QueryContext(
		ctx,
		`SELECT q.id, q.order_number
		 FROM questions q
		 JOIN tests t ON t.id = q.test_id
		 WHERE q.test_id = $1
		   AND t.created_by_user_id = $2
		 ORDER BY q.order_number, q.id
		 FOR UPDATE`,
		testID,
		createdByUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := make([]questionOrderState, 0)
	for rows.Next() {
		var state questionOrderState
		if err := rows.Scan(&state.ID, &state.OrderNumber); err != nil {
			return nil, err
		}
		states = append(states, state)
	}

	return states, rows.Err()
}

func nextQuestionOrder(states []questionOrderState) int {
	maxOrder := 0
	for _, state := range states {
		if state.OrderNumber > maxOrder {
			maxOrder = state.OrderNumber
		}
	}

	return maxOrder + 1
}

func questionOrderIDs(states []questionOrderState) []int64 {
	ids := make([]int64, 0, len(states))
	for _, state := range states {
		ids = append(ids, state.ID)
	}

	return ids
}

func insertQuestionID(ids []int64, questionID int64, desiredOrder int) []int64 {
	position := normalizeQuestionPosition(desiredOrder, len(ids)+1)
	index := position - 1

	result := make([]int64, 0, len(ids)+1)
	result = append(result, ids[:index]...)
	result = append(result, questionID)
	result = append(result, ids[index:]...)
	return result
}

func moveQuestionID(ids []int64, questionID int64, desiredOrder int) ([]int64, bool) {
	currentIndex := -1
	for idx, id := range ids {
		if id == questionID {
			currentIndex = idx
			break
		}
	}
	if currentIndex == -1 {
		return nil, false
	}
	if desiredOrder <= 0 {
		return append([]int64(nil), ids...), true
	}

	targetIndex := normalizeQuestionPosition(desiredOrder, len(ids)) - 1
	if targetIndex == currentIndex {
		return append([]int64(nil), ids...), true
	}

	result := make([]int64, 0, len(ids))
	for idx, id := range ids {
		if idx != currentIndex {
			result = append(result, id)
		}
	}

	reordered := make([]int64, 0, len(ids))
	reordered = append(reordered, result[:targetIndex]...)
	reordered = append(reordered, questionID)
	reordered = append(reordered, result[targetIndex:]...)
	return reordered, true
}

func removeQuestionID(ids []int64, questionID int64) ([]int64, bool) {
	result := make([]int64, 0, len(ids))
	found := false
	for _, id := range ids {
		if id == questionID {
			found = true
			continue
		}
		result = append(result, id)
	}

	return result, found
}

func normalizeQuestionPosition(orderNumber int, total int) int {
	if total <= 0 {
		return 1
	}
	if orderNumber <= 0 {
		return total
	}
	if orderNumber > total {
		return total
	}

	return orderNumber
}

func resequenceQuestionOrderNumbersTx(ctx context.Context, tx *sql.Tx, testID int64, orderedQuestionIDs []int64) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE questions
		 SET order_number = -order_number
		 WHERE test_id = $1`,
		testID,
	); err != nil {
		return err
	}

	for index, questionID := range orderedQuestionIDs {
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE questions
			 SET order_number = $2,
			 	updated_at = NOW()
			 WHERE id = $1
			   AND test_id = $3`,
			questionID,
			index+1,
			testID,
		); err != nil {
			return err
		}
	}

	return nil
}

func scanQuestionBase(scanner rowScanner) (domain.Question, error) {
	var question domain.Question
	var rawScaleWeights []byte
	var createdAt time.Time
	var updatedAt time.Time

	err := scanner.Scan(
		&question.ID,
		&question.TestID,
		&question.Text,
		&question.QuestionType,
		&question.OrderNumber,
		&question.IsRequired,
		&rawScaleWeights,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.Question{}, err
	}
	if len(rawScaleWeights) > 0 {
		if err := json.Unmarshal(rawScaleWeights, &question.ScaleWeights); err != nil {
			return domain.Question{}, err
		}
	}
	if question.ScaleWeights == nil {
		question.ScaleWeights = map[string]float64{}
	}

	question.CreatedAt = createdAt.Format(time.RFC3339)
	question.UpdatedAt = updatedAt.Format(time.RFC3339)

	return question, nil
}

func mustMarshalScaleWeights(scaleWeights map[string]float64) string {
	if len(scaleWeights) == 0 {
		return "{}"
	}

	content, err := json.Marshal(scaleWeights)
	if err != nil {
		return "{}"
	}

	return string(content)
}

func marshalOptionalScaleWeights(scaleWeights *map[string]float64) any {
	if scaleWeights == nil {
		return nil
	}

	return mustMarshalScaleWeights(*scaleWeights)
}
