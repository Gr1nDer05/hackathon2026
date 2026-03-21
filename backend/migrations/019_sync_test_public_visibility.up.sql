UPDATE tests
SET is_public = (status = 'published');

ALTER TABLE tests
DROP CONSTRAINT IF EXISTS tests_public_visibility_check;

ALTER TABLE tests
ADD CONSTRAINT tests_public_visibility_check
CHECK (is_public = (status = 'published'));
