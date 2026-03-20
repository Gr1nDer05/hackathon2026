package repository

import (
	"context"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func (r *AppRepository) CreateFormulaRule(ctx context.Context, testID int64, createdByUserID int64, input domain.CreateFormulaRuleInput) (domain.FormulaRule, error) {
	row := r.db.QueryRowContext(
		ctx,
		`INSERT INTO formula_rules (test_id, name, question_id, condition_type, expected_value, score_delta, result_key, priority)
		 SELECT
			$1,
			$3,
			NULLIF($4, 0),
			$5,
			$6,
			$7,
			$8,
			$9
		 FROM tests t
		 WHERE t.id = $1
		   AND t.created_by_user_id = $2
		   AND ($4 = 0 OR EXISTS (SELECT 1 FROM questions q WHERE q.id = $4 AND q.test_id = t.id))
		 RETURNING id, test_id, COALESCE(question_id, 0), name, condition_type, expected_value, score_delta, result_key, priority, created_at, updated_at`,
		testID,
		createdByUserID,
		input.Name,
		input.QuestionID,
		input.ConditionType,
		input.ExpectedValue,
		input.ScoreDelta,
		input.ResultKey,
		input.Priority,
	)

	return scanFormulaRule(row)
}

func (r *AppRepository) ListFormulaRules(ctx context.Context, testID int64, createdByUserID int64) ([]domain.FormulaRule, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT fr.id, fr.test_id, COALESCE(fr.question_id, 0), fr.name, fr.condition_type, fr.expected_value, fr.score_delta, fr.result_key, fr.priority, fr.created_at, fr.updated_at
		 FROM formula_rules fr
		 JOIN tests t ON t.id = fr.test_id
		 WHERE fr.test_id = $1 AND t.created_by_user_id = $2
		 ORDER BY fr.priority, fr.id`,
		testID,
		createdByUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]domain.FormulaRule, 0)
	for rows.Next() {
		rule, scanErr := scanFormulaRule(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

func (r *AppRepository) GetFormulaRuleByID(ctx context.Context, testID int64, ruleID int64, createdByUserID int64) (domain.FormulaRule, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT fr.id, fr.test_id, COALESCE(fr.question_id, 0), fr.name, fr.condition_type, fr.expected_value, fr.score_delta, fr.result_key, fr.priority, fr.created_at, fr.updated_at
		 FROM formula_rules fr
		 JOIN tests t ON t.id = fr.test_id
		 WHERE fr.id = $1 AND fr.test_id = $2 AND t.created_by_user_id = $3`,
		ruleID,
		testID,
		createdByUserID,
	)

	return scanFormulaRule(row)
}

func (r *AppRepository) UpdateFormulaRule(ctx context.Context, testID int64, ruleID int64, createdByUserID int64, input domain.UpdateFormulaRuleInput) (domain.FormulaRule, error) {
	row := r.db.QueryRowContext(
		ctx,
		`UPDATE formula_rules fr
		 SET name = $4,
		 	question_id = NULLIF($5, 0),
		 	condition_type = $6,
		 	expected_value = $7,
		 	score_delta = $8,
		 	result_key = $9,
		 	priority = $10,
		 	updated_at = NOW()
		 FROM tests t
		 WHERE fr.id = $1
		   AND fr.test_id = $2
		   AND t.id = fr.test_id
		   AND t.created_by_user_id = $3
		   AND ($5 = 0 OR EXISTS (SELECT 1 FROM questions q WHERE q.id = $5 AND q.test_id = t.id))
		 RETURNING fr.id, fr.test_id, COALESCE(fr.question_id, 0), fr.name, fr.condition_type, fr.expected_value, fr.score_delta, fr.result_key, fr.priority, fr.created_at, fr.updated_at`,
		ruleID,
		testID,
		createdByUserID,
		input.Name,
		input.QuestionID,
		input.ConditionType,
		input.ExpectedValue,
		input.ScoreDelta,
		input.ResultKey,
		input.Priority,
	)

	return scanFormulaRule(row)
}

func (r *AppRepository) DeleteFormulaRule(ctx context.Context, testID int64, ruleID int64, createdByUserID int64) (bool, error) {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM formula_rules fr
		 USING tests t
		 WHERE fr.id = $1
		   AND fr.test_id = $2
		   AND t.id = fr.test_id
		   AND t.created_by_user_id = $3`,
		ruleID,
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

func scanFormulaRule(scanner rowScanner) (domain.FormulaRule, error) {
	var rule domain.FormulaRule
	var createdAt time.Time
	var updatedAt time.Time

	err := scanner.Scan(
		&rule.ID,
		&rule.TestID,
		&rule.QuestionID,
		&rule.Name,
		&rule.ConditionType,
		&rule.ExpectedValue,
		&rule.ScoreDelta,
		&rule.ResultKey,
		&rule.Priority,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.FormulaRule{}, err
	}

	rule.CreatedAt = createdAt.Format(time.RFC3339)
	rule.UpdatedAt = updatedAt.Format(time.RFC3339)

	return rule, nil
}
