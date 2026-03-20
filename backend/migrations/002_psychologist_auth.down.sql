DROP INDEX IF EXISTS idx_auth_tokens_expires_at;
DROP INDEX IF EXISTS idx_auth_tokens_user_id;
DROP TABLE IF EXISTS auth_tokens;
DROP TABLE IF EXISTS psychologist_cards;
DROP TABLE IF EXISTS psychologist_profiles;
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
