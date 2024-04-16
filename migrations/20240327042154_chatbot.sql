-- +goose Up
-- +goose StatementBegin
CREATE TABLE chats (
    chat_id bigserial primary key,
    email text NOT NULL references users(email)
);

CREATE TABLE messages (
    chat_id bigint NOT NULL references chats(chat_id),
    sent_at timestamptz NOT NULL DEFAULT now(),
    message text NOT NULL DEFAULT ''::text,
    from_bot bool NOT NULL DEFAULT false
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE chats;
-- +goose StatementEnd
