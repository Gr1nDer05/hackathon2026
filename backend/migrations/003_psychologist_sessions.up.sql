DROP INDEX IF EXISTS idx_auth_tokens_expires_at;
DROP INDEX IF EXISTS idx_auth_tokens_user_id;
DROP TABLE IF EXISTS auth_tokens;

CREATE TABLE IF NOT EXISTS user_sessions (
    session_hash TEXT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
