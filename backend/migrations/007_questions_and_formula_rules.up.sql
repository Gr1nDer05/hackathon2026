CREATE TABLE IF NOT EXISTS question_options (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    label TEXT NOT NULL,
    value TEXT NOT NULL,
    order_number INTEGER NOT NULL DEFAULT 0,
    score DOUBLE PRECISION NOT NULL DEFAULT 0,
    UNIQUE (question_id, order_number),
    UNIQUE (question_id, value)
);

CREATE INDEX IF NOT EXISTS idx_question_options_question_id ON question_options(question_id);

CREATE TABLE IF NOT EXISTS formula_rules (
    id BIGSERIAL PRIMARY KEY,
    test_id BIGINT NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    question_id BIGINT REFERENCES questions(id) ON DELETE SET NULL,
    condition_type TEXT NOT NULL CHECK (
        condition_type IN ('always', 'answer_equals', 'answer_in', 'answer_numeric_gte', 'answer_numeric_lte')
    ),
    expected_value TEXT NOT NULL DEFAULT '',
    score_delta DOUBLE PRECISION NOT NULL DEFAULT 0,
    result_key TEXT NOT NULL DEFAULT 'total',
    priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_formula_rules_test_id ON formula_rules(test_id);
CREATE INDEX IF NOT EXISTS idx_formula_rules_question_id ON formula_rules(question_id);
