-- +goose Up
-- +goose StatementBegin
-- Создаем перечисление role_type
CREATE TYPE role_type AS ENUM ('user', 'assistant', 'technical', 'document_redirect');
-- Добавляем столбец role в таблицу messages с временным значением по умолчанию
ALTER TABLE messages
    ADD COLUMN role role_type NOT NULL DEFAULT 'user';
-- Обновляем столбец role в зависимости от значения from_bot
UPDATE messages
SET role = 'assistant'
WHERE from_bot = true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Удаляем столбец role из таблицы messages
ALTER TABLE messages
    DROP COLUMN role;
-- Удаляем тип ENUM role_type
DROP TYPE role_type;

-- +goose StatementEnd
