ALTER TABLE tests
ADD COLUMN IF NOT EXISTS max_participants INTEGER NOT NULL DEFAULT 0;

ALTER TABLE tests
DROP CONSTRAINT IF EXISTS tests_max_participants_check;

ALTER TABLE tests
ADD CONSTRAINT tests_max_participants_check CHECK (max_participants >= 0);
