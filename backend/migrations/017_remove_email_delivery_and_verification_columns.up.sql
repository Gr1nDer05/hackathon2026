ALTER TABLE users
DROP COLUMN IF EXISTS subscription_psychologist_notice_sent_at,
DROP COLUMN IF EXISTS subscription_admin_notice_sent_at,
DROP COLUMN IF EXISTS email_verification_expires_at,
DROP COLUMN IF EXISTS email_verification_code_hash,
DROP COLUMN IF EXISTS email_verified_at;
