ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_subscription_plan_check;

ALTER TABLE users
DROP COLUMN IF EXISTS subscription_plan;
