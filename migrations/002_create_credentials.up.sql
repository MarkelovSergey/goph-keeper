-- Таблица учётных данных пользователей
CREATE TABLE IF NOT EXISTS credentials (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       VARCHAR(50)  NOT NULL,
    name       VARCHAR(255) NOT NULL,
    metadata   TEXT         NOT NULL DEFAULT '',
    data       BYTEA        NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_credentials_user_id ON credentials(user_id);
