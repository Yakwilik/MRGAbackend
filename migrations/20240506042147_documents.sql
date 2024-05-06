-- +goose Up
-- +goose StatementBegin
CREATE TABLE document_categories
(
    category_id   INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_name TEXT NOT NULL UNIQUE
);

CREATE TABLE document_types
(
    type_id     INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_id INT  NOT NULL REFERENCES document_categories (category_id) ON DELETE CASCADE,
    type_name   TEXT NOT NULL UNIQUE,
    enabled     BOOL NOT NULL DEFAULT true
);

CREATE TABLE document_variants
(
    variant_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type_id    INT  NOT NULL REFERENCES document_types (type_id) ON DELETE CASCADE,
    key        TEXT NOT NULL UNIQUE,
    name       TEXT NOT NULL UNIQUE
);



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Откат создания таблиц для управления документами

-- Удаление таблицы вариантов документов
DROP TABLE document_variants;

-- Удаление таблицы типов документов
DROP TABLE document_types;

-- Удаление таблицы категорий
DROP TABLE document_categories;

-- +goose StatementEnd
