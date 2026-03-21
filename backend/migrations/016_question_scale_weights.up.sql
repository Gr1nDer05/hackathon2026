ALTER TABLE questions
ADD COLUMN IF NOT EXISTS scale_weights_json JSONB NOT NULL DEFAULT '{}'::jsonb;
