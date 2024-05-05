package grpc

import (
	"context"
	pb "github.com/Yakwilik/MRGAbackend/internal/pb/ailawyer"
)

func (a *Implementation) GetDocumentCategories(ctx context.Context, request *pb.GetCategoriesRequest) (*pb.GetCategoriesResponse, error) {
	categories := []*pb.DocumentCategory{
		{
			CategoryName: "Семейное право",
			DocumentTypes: []*pb.DocumentType{
				{
					Type: "Заключение брака",
					Variants: []*pb.DocumentVariant{{
						Key:  "marryOne",
						Name: "Заявление на заключение брака для подачи без пары",
					}},
				},
				//{
				//	Type: "Расторжение брака",
				//	Variants: []*pb.DocumentVariant{
				//		{
				//			Key:  "divorce1",
				//			Name: "Исковое заявление о расторжении брака",
				//		},
				//		{
				//			Key:  "divorce2",
				//			Name: "Исковое заявление о расторжении брака с одним ребенком",
				//		},
				//		{
				//			Key:  "divorce3",
				//			Name: "Исковое заявление о расторжении брака с 2 и более детьми",
				//		},
				//		{
				//			Key:  "divorce4",
				//			Name: "Исковое заявление о расторжении брака при согласии супругов",
				//		},
				//		{
				//			Key:  "divorce5",
				//			Name: "Исковое заявление о расторжении брака, когда один супруг не согласен",
				//		},
				//		{
				//			Key:  "divorce6",
				//			Name: "Исковое заявление о расторжении брака и разделе имущества",
				//		},
				//		{
				//			Key:  "divorce7",
				//			Name: "Исковое заявление о расторжении брака и взыскании алиментов",
				//		},
				//		{
				//			Key:  "divorce8",
				//			Name: "Исковое заявление о расторжении брака и определении места жительства ребенка",
				//		},
				//	},
				//},
				//{
				//	Type: "Взыскание алиментов",
				//	Variants: []*pb.DocumentVariant{
				//		{
				//			Key:  "aliments1",
				//			Name: "Заявление на судебный приказ о взыскании алиментов",
				//		},
				//		{
				//			Key:  "aliments2",
				//			Name: "Исковое заявление на взыскание алиментов",
				//		},
				//	},
				//},
			},
		},
		{
			CategoryName: "Трудовое право",
			DocumentTypes: []*pb.DocumentType{
				{
					Type: "Заявление об увольнении",
					Variants: []*pb.DocumentVariant{
						{
							Key:  "voluntaryDismissal",
							Name: "Заявление об увольнении по собственному желанию",
						},
					},
				},
			},
		},
	}

	return &pb.GetCategoriesResponse{Categories: categories}, nil
}

func (a *Implementation) SendRedirectSuggest(ctx context.Context, request *pb.SendRedirectSuggestRequest) (*pb.SendRedirectSuggestResponse, error) {
	return &pb.SendRedirectSuggestResponse{}, nil
}
