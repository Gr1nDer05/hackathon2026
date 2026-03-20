DROP INDEX IF EXISTS idx_users_psychologist_blocked_until;
DROP INDEX IF EXISTS idx_users_psychologist_portal_access_until;

ALTER TABLE users
DROP COLUMN IF EXISTS blocked_until,
DROP COLUMN IF EXISTS portal_access_until;
