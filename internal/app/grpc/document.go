package grpc

import (
	"context"
	pb "github.com/Yakwilik/MRGAbackend/internal/pb/ailawyer"
)

func (a *Implementation) GetDocumentCategories(ctx context.Context, request *pb.GetCategoriesRequest) (*pb.GetCategoriesResponse, error) {
	categories := []*pb.DocumentCategory{{
		CategoryName: "Семейное право",
		DocumentTypes: []*pb.DocumentType{{
			Type: "Заключение брака",
			Variants: []*pb.DocumentVariant{{
				Key:  "marryOne",
				Name: "Заявление на заключение брака для подачи без пары",
			}},
		}},
	}}

	return &pb.GetCategoriesResponse{Categories: categories}, nil
}
