-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    users
(
    email         text NOT NULL UNIQUE,
    password_hash text NOT NULL

);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
