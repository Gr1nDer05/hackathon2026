ALTER TABLE users
ADD COLUMN IF NOT EXISTS login TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_login_unique
ON users(login);
