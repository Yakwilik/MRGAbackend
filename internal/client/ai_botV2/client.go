package ai_botV2

import (
	"context"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	pb "github.com/Yakwilik/MRGAbackend/internal/pb/chatbot"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
)

type client interface {
	RespondToUserQuery(ctx context.Context, in *pb.ChatRequest, opts ...grpc.CallOption) (pb.ChatService_RespondToUserQueryClient, error)
}

type Interface interface {
	RespondToUserQuery(ctx context.Context, request model.ChatRequest) (<-chan *model.ChatResponseChunk, error)
}

type adapter struct {
	cli client
}

type Config struct {
	ServiceAddr string
}

func MustNew(cfg Config) Interface {
	cli, err := grpc.Dial(cfg.ServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect to aibot service: error %v, addr: %s", err, cfg.ServiceAddr)
	}
	return &adapter{
		cli: pb.NewChatServiceClient(cli),
	}
}

func (a adapter) RespondToUserQuery(ctx context.Context, request model.ChatRequest) (<-chan *model.ChatResponseChunk, error) {
	data := encodeRespondToUserQuery(request)
	grpcStream, err := a.cli.RespondToUserQuery(ctx, data)
	if err != nil {
		return nil, err
	}
	//mockResp := []model.ChatResponseChunk{
	//	//{
	//	//	Role:          model.RoleAssistant,
	//	//	Chunk:         "Привет; Спасибо, что пишешь мне",
	//	//	MessageStatus: model.StatusOk,
	//	//	ErrorDetails:  "",
	//	//},
	//	//{
	//	//	Role:          model.RoleAssistant,
	//	//	Chunk:         "Привет2; Спасибо, что пишешь мне",
	//	//	MessageStatus: model.StatusOk,
	//	//	ErrorDetails:  "",
	//	//},
	//	//{
	//	//	Role:          model.RoleAssistant,
	//	//	Chunk:         "сообщение",
	//	//	MessageStatus: model.StatusOk,
	//	//	ErrorDetails:  "",
	//	//},
	//}
	//for i := 0; i < 500; i++ {
	//	//for i, d := range mockResp {
	//	mockResp = append(mockResp, model.ChatResponseChunk{
	//		Role:          model.RoleAssistant,
	//		Chunk:         fmt.Sprintf("сообщение%d", i),
	//		MessageStatus: model.StatusOk,
	//		ErrorDetails:  "",
	//	})
	//	//}
	//}

	responseChan := make(chan *model.ChatResponseChunk)
	go func() {
		defer close(responseChan)
		//for _, response := range mockResp {
		for {
			response, err := grpcStream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				logger.Error(ctx, "RespondToUserQuery", "err", err)

				responseChan <- &model.ChatResponseChunk{
					Role:          model.RoleTechnical,
					Chunk:         "Произошла ошибка, попробуйте повторить запрос",
					MessageStatus: model.StatusFail,
					ErrorDetails:  err.Error(),
				}
				break
			}

			responseChan <- &model.ChatResponseChunk{
				Role:          model.Role(response.GetRole()),
				Chunk:         response.GetOutput(),
				MessageStatus: model.StatusOk,
				UserID:        response.GetUserId(),
				ChatID:        response.GetChatId(),
			}
			//responseChan <- &response
		}
	}()
	return responseChan, nil
}

func encodeRespondToUserQuery(request model.ChatRequest) *pb.ChatRequest {
	return &pb.ChatRequest{
		UserId:      request.UserID,
		ChatId:      request.ChatID,
		UserQuery:   request.UserQuery,
		ChatHistory: encodeChatHistory(request.ChatHistory),
	}
}

func encodeChatHistory(chatHistory []model.HistoryMessage) []*pb.ChatMessage {
	if len(chatHistory) == 0 {
		return []*pb.ChatMessage{}
	}
	lastMessageID := len(chatHistory) - 1
	response := make([]*pb.ChatMessage, 0, len(chatHistory)-1)
	for index, msg := range chatHistory {
		if index == lastMessageID {
			continue
		} else {
			response = append(response, &pb.ChatMessage{
				Role: msg.Role.String(),
				Text: msg.Text,
			})
		}
	}

	return response
}
