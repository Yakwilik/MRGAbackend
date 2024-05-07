-- +goose Up
-- +goose StatementBegin
ALTER TABLE messages
    ADD COLUMN extra_questions JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TYPE role_type ADD VALUE 'extra_questions';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Удаляем столбец role из таблицы messages
ALTER TABLE messages
DROP COLUMN extra_questions;
-- +goose StatementEnd
