CREATE TABLE subscription_purchase_requests (
    id BIGSERIAL PRIMARY KEY,
    psychologist_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_plan TEXT NOT NULL CHECK (subscription_plan IN ('basic', 'pro')),
    duration_days INTEGER NOT NULL CHECK (duration_days = 30),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processed', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX subscription_purchase_requests_pending_user_idx
    ON subscription_purchase_requests (psychologist_user_id)
    WHERE status = 'pending';
