package grpc

import (
	"context"
	"errors"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	pb "github.com/Yakwilik/MRGAbackend/internal/pb/ailawyer"
	"google.golang.org/grpc/codes"
)

func (a *Implementation) GetDocumentCategories(ctx context.Context, request *pb.GetCategoriesRequest) (*pb.GetCategoriesResponse, error) {
	categories, err := a.useCase.GetDocumentCategories(ctx)
	if err != nil {
		return nil, err
	}
	logger.Info(ctx, "GetDocumentCategories", "categories", categories)

	return &pb.GetCategoriesResponse{Categories: encodeCategories(categories)}, nil
}

func encodeCategories(cats []model.DocumentCategory) []*pb.DocumentCategory {
	pbCategories := make([]*pb.DocumentCategory, 0, len(cats))

	for _, cat := range cats {
		pbCat := &pb.DocumentCategory{
			CategoryName:  cat.CategoryName,
			DocumentTypes: encodeDocumentTypes(cat.DocumentTypes),
		}
		pbCategories = append(pbCategories, pbCat)
	}

	return pbCategories
}

func encodeDocumentTypes(types []model.DocumentType) []*pb.DocumentType {
	pbTypes := make([]*pb.DocumentType, 0, len(types))

	for _, t := range types {
		pbType := &pb.DocumentType{
			Type:     t.TypeName,
			Variants: encodeDocumentVariants(t.Variants),
		}
		pbTypes = append(pbTypes, pbType)
	}

	return pbTypes
}

func encodeDocumentVariants(variants []model.DocumentVariant) []*pb.DocumentVariant {
	pbVariants := make([]*pb.DocumentVariant, 0, len(variants))

	for _, v := range variants {
		pbVariant := &pb.DocumentVariant{
			Key:  v.Key,
			Name: v.Name,
		}
		pbVariants = append(pbVariants, pbVariant)
	}

	return pbVariants
}

func (a *Implementation) SendRedirectSuggest(ctx context.Context, request *pb.SendRedirectSuggestRequest) (*pb.SendRedirectSuggestResponse, error) {
	return &pb.SendRedirectSuggestResponse{}, nil
}

func (a *Implementation) AddDocumentCategory(ctx context.Context, request *pb.AddDocumentCategoryRequest) (*pb.AddDocumentCategoryResponse, error) {
	if err := a.useCase.AddDocumentCategory(ctx, request.GetCategoryName()); err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			return nil, errValidation.WithDetails(codes.InvalidArgument)
		}
		return nil, err
	}
	return &pb.AddDocumentCategoryResponse{}, nil
}

func (a *Implementation) AddDocumentTypes(ctx context.Context, request *pb.AddDocumentTypesRequest) (*pb.AddDocumentTypesResponse, error) {
	if err := a.useCase.AddDocumentTypes(ctx, model.CreateTypesRequest{
		CategoryName:  request.GetCategoryName(),
		DocumentTypes: request.GetTypeNames(),
	}); err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			return nil, errValidation.WithDetails(codes.InvalidArgument)
		}
		return nil, err
	}
	return &pb.AddDocumentTypesResponse{}, nil
}

func (a *Implementation) AddDocumentVariants(ctx context.Context, documentType *pb.DocumentType) (*pb.AddDocumentVariantsResponse, error) {
	if err := a.useCase.AddDocumentVariants(ctx, model.CreateVariantsRequest{
		TypeName:         documentType.GetType(),
		DocumentVariants: decodeVariants(documentType.GetVariants()),
	}); err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			return nil, errValidation.WithDetails(codes.InvalidArgument)
		}
		return nil, err
	}
	return &pb.AddDocumentVariantsResponse{}, nil
}

func decodeVariants(pbVariants []*pb.DocumentVariant) []model.DocumentVariant {
	result := make([]model.DocumentVariant, 0, len(pbVariants))

	for _, pbVariant := range pbVariants {
		result = append(result, model.DocumentVariant{
			Key:  pbVariant.GetKey(),
			Name: pbVariant.GetName(),
		})
	}

	return result
}

func (a *Implementation) GetDocumentInfoByKey(ctx context.Context, request *pb.GetDocumentInfoByKeyRequest) (*pb.GetDocumentInfoByKeyResponse, error) {
	info, err := a.useCase.GetDocumentInfoByKey(ctx, request.GetKey())
	if err != nil {
		if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
			return nil, errValidation.WithDetails(codes.InvalidArgument)
		}
		return nil, err
	}

	return &pb.GetDocumentInfoByKeyResponse{
		Key:      info.Key,
		Category: info.CategoryName,
		Type:     info.TypeName,
		Name:     info.Name,
	}, nil
}
