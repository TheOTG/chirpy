-- +goose Up
CREATE TABLE chirps (
    id UUID,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    body TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
    PRIMARY KEY(id)
);

-- +goose Down
DROP TABLE chirps;