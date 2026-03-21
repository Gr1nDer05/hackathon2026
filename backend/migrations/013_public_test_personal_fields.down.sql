DROP INDEX IF EXISTS idx_public_test_sessions_test_phone_unique;

ALTER TABLE public_test_sessions
DROP CONSTRAINT IF EXISTS public_test_sessions_respondent_age_check;

ALTER TABLE public_test_sessions
DROP COLUMN IF EXISTS respondent_education,
DROP COLUMN IF EXISTS respondent_gender,
DROP COLUMN IF EXISTS respondent_age,
DROP COLUMN IF EXISTS respondent_phone;

ALTER TABLE tests
DROP COLUMN IF EXISTS collect_respondent_education,
DROP COLUMN IF EXISTS collect_respondent_gender,
DROP COLUMN IF EXISTS collect_respondent_age;
