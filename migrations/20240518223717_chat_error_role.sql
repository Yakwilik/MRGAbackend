-- +goose Up
-- +goose StatementBegin
ALTER TYPE role_type ADD VALUE 'error';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
