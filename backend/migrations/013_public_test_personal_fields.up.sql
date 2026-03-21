ALTER TABLE tests
ADD COLUMN IF NOT EXISTS collect_respondent_age BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS collect_respondent_gender BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS collect_respondent_education BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE public_test_sessions
ADD COLUMN IF NOT EXISTS respondent_phone TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS respondent_age INTEGER,
ADD COLUMN IF NOT EXISTS respondent_gender TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS respondent_education TEXT NOT NULL DEFAULT '';

ALTER TABLE public_test_sessions
DROP CONSTRAINT IF EXISTS public_test_sessions_respondent_age_check;

ALTER TABLE public_test_sessions
ADD CONSTRAINT public_test_sessions_respondent_age_check
CHECK (respondent_age IS NULL OR (respondent_age >= 1 AND respondent_age <= 120));

CREATE UNIQUE INDEX IF NOT EXISTS idx_public_test_sessions_test_phone_unique
ON public_test_sessions(test_id, respondent_phone)
WHERE respondent_phone <> '';
