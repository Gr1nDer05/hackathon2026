ALTER TABLE tests
ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'draft';

UPDATE tests
SET status = 'draft'
WHERE status IS NULL;

ALTER TABLE tests
ADD CONSTRAINT tests_status_check CHECK (status IN ('draft', 'published'));
