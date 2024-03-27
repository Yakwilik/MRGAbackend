-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    session (
              session_id text NOT NULL PRIMARY KEY,
              email text NOT NULL references users(email)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE session;
-- +goose StatementEnd
