ALTER TABLE users
ADD COLUMN IF NOT EXISTS subscription_plan TEXT NOT NULL DEFAULT 'basic';

UPDATE users
SET subscription_plan = 'basic'
WHERE TRIM(COALESCE(subscription_plan, '')) = '';

ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_subscription_plan_check;

ALTER TABLE users
ADD CONSTRAINT users_subscription_plan_check
CHECK (subscription_plan IN ('basic', 'pro'));
