package core

import (
	"context"
	"errors"
	"github.com/Yakwilik/MRGAbackend/internal/model"
)

func (a *usecase) GetDocumentCategories(ctx context.Context) ([]model.DocumentCategory, error) {
	docInfo, err := a.storage.GetDocumentsInfo(ctx)
	if err != nil {
		return nil, err
	}
	//logger.Info(ctx, "GetDocumentCategories", "docInfo", docInfo)

	return encodeDocumentCategory(docInfo), nil
}

func encodeDocumentCategory(docs []model.DocumentInfo) []model.DocumentCategory {
	categoryMap := make(map[string]model.DocumentCategory)

	for _, doc := range docs {
		// Check if the category already exists
		if cat, exists := categoryMap[doc.CategoryName]; exists {
			// Check if the document type already exists in this category
			typeFound := false
			for i, dt := range cat.DocumentTypes {
				if dt.TypeName == doc.TypeName {
					// Append the variant to the existing document type
					categoryMap[doc.CategoryName].DocumentTypes[i].Variants = append(categoryMap[doc.CategoryName].DocumentTypes[i].Variants, model.DocumentVariant{
						Key:  doc.Key,
						Name: doc.Name,
					})
					typeFound = true
					break
				}
			}
			// If the document type is new for this category
			if !typeFound {
				cat.DocumentTypes = append(cat.DocumentTypes, model.DocumentType{
					TypeName: doc.TypeName,
					Variants: []model.DocumentVariant{{Key: doc.Key, Name: doc.Name}},
				})
				categoryMap[doc.CategoryName] = cat
			}
		} else {
			// Create a new category and document type
			newCat := model.DocumentCategory{
				CategoryName: doc.CategoryName,
				DocumentTypes: []model.DocumentType{{
					TypeName: doc.TypeName,
					Variants: []model.DocumentVariant{{Key: doc.Key, Name: doc.Name}},
				}},
			}
			categoryMap[doc.CategoryName] = newCat
		}
	}

	// Convert map to slice
	var categories []model.DocumentCategory
	for _, c := range categoryMap {
		categories = append(categories, c)
	}

	return categories
}

func (a *usecase) AddDocumentCategory(ctx context.Context, categoryName string) error {
	err := a.storage.AddDocumentCategory(ctx, categoryName)

	if err != nil {
		if errors.Is(err, model.ErrAlreadyExists) {
			return model.NewValidationError("category_name", "category_name already exists", "Категория с таким именем уже существует")
		}

		return err
	}
	return nil
}

func (a *usecase) AddDocumentTypes(ctx context.Context, category model.CreateTypesRequest) error {
	if err := a.storage.AddDocumentTypes(ctx, category); err != nil {
		if errors.Is(err, model.ErrAlreadyExists) {
			return model.NewValidationError("document_type", "document_type already exists", "Тип документа с таким именем уже существует")
		}
		if errors.Is(err, model.ErrNotFound) {
			return model.NewValidationError("category_name", "category_name not found", "Категория с таким именем не найдена")
		}
		return err
	}

	return nil
}

func (a *usecase) AddDocumentVariants(ctx context.Context, doctype model.CreateVariantsRequest) error {
	if err := a.storage.AddDocumentVariants(ctx, doctype); err != nil {
		if errors.Is(err, model.ErrAlreadyExists) {
			return model.NewValidationError("document_variant", "document_variant already exists", "Тип документа с таким именем уже существует")
		}
		if errors.Is(err, model.ErrNotFound) {
			return model.NewValidationError("type_name", "type_name not found", "Категория с таким именем не найдена")
		}
		return err
	}

	return nil
}
