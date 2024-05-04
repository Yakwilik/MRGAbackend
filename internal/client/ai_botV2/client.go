package ai_botV2

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	pb "github.com/Yakwilik/MRGAbackend/internal/pb/chatbot"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"math/big"
	"time"
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

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range result {
		randomIndex, _ := rand.Int(rand.Reader, charsetLen)
		result[i] = charset[randomIndex.Int64()]
	}

	return string(result)
}

func mockResponse(config *helper.MockResponseConfig) (<-chan *model.ChatResponseChunk, error) {
	if config.IterationCount == 0 {
		config.IterationCount = 50
	}
	if config.IterationCount < 0 {
		config.IterationCount = -config.IterationCount
	}
	if config.BatchSize == 0 {
		config.BatchSize = 50
	}
	if config.BatchSize < 0 {
		config.BatchSize = -config.BatchSize
	}
	if config.IterationTimeout < time.Millisecond*10 {
		config.IterationTimeout = time.Millisecond * 10
	}
	if config.IterationTimeout.Milliseconds()*int64(config.IterationCount) > time.Minute.Milliseconds() {
		config.IterationTimeout = time.Minute
	}

	responseChan := make(chan *model.ChatResponseChunk)
	go func() {
		defer close(responseChan)

		ticker := time.NewTicker(config.IterationTimeout)
		defer ticker.Stop()

		for i := 0; i < config.IterationCount; i++ {
			select {
			case <-ticker.C:
				responseChan <- &model.ChatResponseChunk{
					Role:          model.RoleAssistant,
					Chunk:         fmt.Sprintf("%sсообщение%d", GenerateRandomString(config.BatchSize), i),
					MessageStatus: model.StatusOk,
				}
			}

		}
	}()

	return responseChan, nil
}

func (a adapter) RespondToUserQuery(ctx context.Context, request model.ChatRequest) (<-chan *model.ChatResponseChunk, error) {
	responseChan := make(chan *model.ChatResponseChunk)

	mockConfig, ok := helper.GetMockResponseConfig(ctx)
	if ok && mockConfig.Generate {
		return mockResponse(mockConfig)
	}
	data := encodeRespondToUserQuery(request)
	grpcStream, err := a.cli.RespondToUserQuery(ctx, data)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(responseChan)
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
					ChatID:        request.ChatID,
				}
				break
			}

			responseChan <- &model.ChatResponseChunk{
				Role:          model.Role(response.GetRole()),
				Chunk:         response.GetOutput(),
				MessageStatus: model.StatusOk,
				UserID:        response.GetUserId(),
				ChatID:        request.ChatID,
			}
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
