DROP TABLE IF EXISTS public_test_sessions;

DROP INDEX IF EXISTS idx_tests_public_slug_unique;

ALTER TABLE tests
DROP COLUMN IF EXISTS is_public,
DROP COLUMN IF EXISTS public_slug;
