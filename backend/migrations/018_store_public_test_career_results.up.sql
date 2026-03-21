ALTER TABLE public_test_sessions
ADD COLUMN IF NOT EXISTS career_result_json JSONB;
