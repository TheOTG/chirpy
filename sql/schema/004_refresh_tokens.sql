-- +goose Up
CREATE TABLE refresh_tokens (
    token TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    PRIMARY KEY(token)
);

-- +goose Down
DROP TABLE refresh_tokens;