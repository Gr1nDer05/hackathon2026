DROP TABLE IF EXISTS public_test_answers;

ALTER TABLE public_test_sessions
DROP CONSTRAINT IF EXISTS public_test_sessions_status_check;

ALTER TABLE public_test_sessions
DROP COLUMN IF EXISTS completed_at,
DROP COLUMN IF EXISTS status;
