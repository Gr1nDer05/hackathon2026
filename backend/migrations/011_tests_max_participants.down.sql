ALTER TABLE tests
DROP CONSTRAINT IF EXISTS tests_max_participants_check;

ALTER TABLE tests
DROP COLUMN IF EXISTS max_participants;
