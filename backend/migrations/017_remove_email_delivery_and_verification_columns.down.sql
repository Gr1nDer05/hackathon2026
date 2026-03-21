ALTER TABLE users
ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS email_verification_code_hash TEXT,
ADD COLUMN IF NOT EXISTS email_verification_expires_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS subscription_admin_notice_sent_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS subscription_psychologist_notice_sent_at TIMESTAMPTZ;
