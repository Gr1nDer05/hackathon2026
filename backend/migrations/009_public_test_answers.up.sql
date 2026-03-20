ALTER TABLE public_test_sessions
ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'in_progress',
ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ;

ALTER TABLE public_test_sessions
DROP CONSTRAINT IF EXISTS public_test_sessions_status_check;

ALTER TABLE public_test_sessions
ADD CONSTRAINT public_test_sessions_status_check CHECK (status IN ('in_progress', 'completed'));

CREATE TABLE IF NOT EXISTS public_test_answers (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES public_test_sessions(id) ON DELETE CASCADE,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    answer_text TEXT NOT NULL DEFAULT '',
    answer_value TEXT NOT NULL DEFAULT '',
    answer_values_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (session_id, question_id)
);

CREATE INDEX IF NOT EXISTS idx_public_test_answers_session_id
ON public_test_answers(session_id);
