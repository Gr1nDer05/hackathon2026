ALTER TABLE public_test_sessions
ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '1 hour');

CREATE INDEX IF NOT EXISTS idx_public_test_sessions_expires_at
ON public_test_sessions(expires_at)
WHERE status = 'in_progress';
