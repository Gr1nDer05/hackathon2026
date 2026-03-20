ALTER TABLE tests
ADD COLUMN IF NOT EXISTS public_slug TEXT,
ADD COLUMN IF NOT EXISTS is_public BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_tests_public_slug_unique
ON tests(public_slug)
WHERE public_slug IS NOT NULL;

CREATE TABLE IF NOT EXISTS public_test_sessions (
    id BIGSERIAL PRIMARY KEY,
    test_id BIGINT NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    access_token TEXT NOT NULL UNIQUE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_public_test_sessions_test_id
ON public_test_sessions(test_id);
