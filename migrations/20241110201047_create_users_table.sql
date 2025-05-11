-- +goose Up
-- +goose StatementBegin
CREATE TABLE users
(
    id         UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    username   VARCHAR(25)  NOT NULL UNIQUE,
    email      VARCHAR(25)  NOT NULL UNIQUE,
    password   VARCHAR(100) NOT NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE
OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at
= NOW();
RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER set_updated_at
    BEFORE UPDATE
    ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX idx_users_created_at ON users (created_at);
CREATE INDEX idx_users_updated_at ON users (updated_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd