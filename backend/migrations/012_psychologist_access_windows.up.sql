ALTER TABLE users
ADD COLUMN IF NOT EXISTS portal_access_until TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS blocked_until TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_users_psychologist_portal_access_until
ON users(portal_access_until)
WHERE role = 'psychologist';

CREATE INDEX IF NOT EXISTS idx_users_psychologist_blocked_until
ON users(blocked_until)
WHERE role = 'psychologist';
