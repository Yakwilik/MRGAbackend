-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats
    ADD COLUMN chat_name TEXT NOT NULL DEFAULT ''::TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats
DROP COLUMN chat_name;
-- +goose StatementEnd
