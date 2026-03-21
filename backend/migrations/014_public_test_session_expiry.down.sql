DROP INDEX IF EXISTS idx_public_test_sessions_expires_at;

ALTER TABLE public_test_sessions
DROP COLUMN IF EXISTS expires_at;
