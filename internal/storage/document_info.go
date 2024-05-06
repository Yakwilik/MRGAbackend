package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"github.com/blockloop/scan"
	"github.com/lib/pq"
	"strings"
)

type documentInfo struct {
	CategoryName string `db:"category_name"`
	TypeName     string `db:"type_name"`
	Key          string `db:"key"`
	Name         string `db:"name"`
}

func (s *storage) GetDocumentsInfo(ctx context.Context) ([]model.DocumentInfo, error) {
	rows, err := s.db.Query(`SELECT c.category_name,
       dt.type_name,
       dv.key,
       dv.name
FROM document_categories c
         JOIN
     document_types dt ON c.category_id = dt.category_id
         JOIN
     document_variants dv ON dt.type_id = dv.type_id
WHERE enabled = true;`)
	if err != nil {
		return nil, fmt.Errorf("error executing query [GetDocumentsInfo]: %w", err)
	}

	var result []documentInfo
	if err := scan.Rows(&result, rows); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.DocumentInfo{}, nil
		}
		return nil, fmt.Errorf("error executing query [GetDocumentsInfo]: %w", err)
	}

	return decodeDocumentInfo(result), nil
}

func decodeDocumentInfo(dbMessages []documentInfo) []model.DocumentInfo {
	result := make([]model.DocumentInfo, 0, len(dbMessages))
	for _, dbModel := range dbMessages {
		result = append(result, model.DocumentInfo{
			CategoryName: dbModel.CategoryName,
			TypeName:     dbModel.TypeName,
			Key:          dbModel.Key,
			Name:         dbModel.Name,
		})
	}

	return result
}

func (s *storage) AddDocumentCategory(ctx context.Context, categoryName string) error {
	if _, err := s.db.Exec(`INSERT INTO document_categories (category_name) VALUES ($1)`, categoryName); err != nil {
		logger.Error(ctx, "AddDocumentCategory", "error", err)
		if pqErr := new(pq.Error); errors.As(err, &pqErr) {
			if pqErr.Code.Name() == UNIQUE_VIOLATION {
				return model.ErrAlreadyExists
			}
		}
		return fmt.Errorf("error executing query [AddDocumentCategory]: %w", err)
	}

	return nil
}

func (s *storage) AddDocumentTypes(ctx context.Context, category model.CreateTypesRequest) error {
	if len(category.DocumentTypes) == 0 {
		return nil // Возвращаем nil, так как нет типов для вставки
	}

	var categoryID uint32
	if err := s.db.QueryRow(`SELECT category_id FROM document_categories WHERE category_name = $1;`, category.CategoryName).Scan(&categoryID); err != nil {
		logger.Error(ctx, "AddDocumentTypes", "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("error executing query [AddDocumentTypes]: %w", err)
	}

	// Составление запроса для вставки всех типов документов
	var placeholders strings.Builder
	var values []interface{}
	for i, docType := range category.DocumentTypes {
		if i > 0 {
			placeholders.WriteString(", ")
		}
		// Для каждого типа документа добавляем плейсхолдеры и значения
		placeholders.WriteString(fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		values = append(values, categoryID, docType)
	}

	query := fmt.Sprintf("INSERT INTO document_types (category_id, type_name) VALUES %s", placeholders.String())
	logger.Info(ctx, "AddDocumentTypes", "query", query, "values", values)
	if _, err := s.db.Exec(query, values...); err != nil {
		logger.Error(ctx, "AddDocumentTypes", "error", err)
		if pqErr := new(pq.Error); errors.As(err, &pqErr) {
			if pqErr.Code.Name() == UNIQUE_VIOLATION {
				return model.ErrAlreadyExists
			}
		}
		return fmt.Errorf("error executing insert [AddDocumentTypes]: %w", err)
	}

	return nil
}

func (s *storage) AddDocumentVariants(ctx context.Context, docType model.CreateVariantsRequest) error {
	logger.Info(ctx, "AddDocumentVariants", "docType", docType)
	if len(docType.DocumentVariants) == 0 {
		return nil // Возвращаем nil, так как нет вариантов для вставки
	}

	var typeID uint32
	if err := s.db.QueryRow(`SELECT type_id FROM document_types WHERE type_name = $1;`, docType.TypeName).Scan(&typeID); err != nil {
		logger.Error(ctx, "AddDocumentVariants", "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("error executing query [AddDocumentVariants]: %w", err)
	}

	// Составление запроса для вставки всех типов документов
	var placeholders strings.Builder
	var values []interface{}
	for i, variant := range docType.DocumentVariants {
		if i > 0 {
			placeholders.WriteString(", ")
		}
		// Для каждого типа документа добавляем плейсхолдеры и значения
		placeholders.WriteString(fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		values = append(values, typeID, variant.Key, variant.Name)
	}

	query := fmt.Sprintf("INSERT INTO document_variants (type_id, key, name) VALUES %s", placeholders.String())
	logger.Info(ctx, "AddDocumentVariants", "query", query, "values", values)
	if _, err := s.db.Exec(query, values...); err != nil {
		logger.Error(ctx, "AddDocumentVariants", "error", err)
		if pqErr := new(pq.Error); errors.As(err, &pqErr) {
			if pqErr.Code.Name() == UNIQUE_VIOLATION {
				return model.ErrAlreadyExists
			}
		}
		return fmt.Errorf("error executing insert [AddDocumentTypes]: %w", err)
	}

	return nil
}

func (s *storage) GetDocumentInfoByKey(ctx context.Context, key string) (model.DocumentInfo, error) {
	rows, err := s.db.Query(`SELECT c.category_name,
       dt.type_name,
       dv.key,
       dv.name
FROM document_categories c
         JOIN
     document_types dt ON c.category_id = dt.category_id
         JOIN
     document_variants dv ON dt.type_id = dv.type_id
where key = $1;`, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.DocumentInfo{}, model.ErrNotFound
		}
		return model.DocumentInfo{}, fmt.Errorf("error executing insert [GetDocumentInfoByKey]: %w", err)
	}

	var dbResult documentInfo
	if err := scan.Row(&dbResult, rows); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.DocumentInfo{}, model.ErrNotFound
		}
		return model.DocumentInfo{}, fmt.Errorf("error executing insert [GetDocumentInfoByKey]: %w", err)
	}

	return model.DocumentInfo{
		CategoryName: dbResult.CategoryName,
		TypeName:     dbResult.TypeName,
		Key:          dbResult.Key,
		Name:         dbResult.Name,
	}, nil
}
