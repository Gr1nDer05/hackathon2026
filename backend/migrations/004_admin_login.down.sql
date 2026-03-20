DROP INDEX IF EXISTS idx_users_login_unique;
ALTER TABLE users DROP COLUMN IF EXISTS login;
