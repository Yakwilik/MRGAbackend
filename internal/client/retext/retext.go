package retext

import (
	"context"
	"errors"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"time"
)

type Interface interface {
	Summarize(ctx context.Context, message string, resultMaxLen uint32) (string, error)
}

type adapter struct {
	cli client
}

func New() Interface {
	return &adapter{
		cli: NewClient(),
	}
}

func (a *adapter) Summarize(ctx context.Context, message string, resultMaxLen uint32) (string, error) {
	postTaskResp, err := a.cli.PostTask(ctx, taskRequest{
		Task:       taskSummarize,
		SourceText: message,
		Additional: additional{
			Lang:      langRU,
			MaxLength: resultMaxLen,
		},
	})
	if err != nil {
		return "", model.WrapErrorWithMethodName(err, "Summarize")
	}

	if postTaskResp.Status != statusOK {
		return "", errors.New("couldn't summarize message")
	}

	getStatusResp, err := a.cli.GetTaskStatus(ctx, postTaskResp.Data.TaskId)
	if err != nil {
		return "", model.WrapErrorWithMethodName(err, "Summarize")
	}

	if getStatusResp.Status != statusOK {
		return "", errors.New("couldn't get task status")
	}

	if getStatusResp.Data.Ready {
		if getStatusResp.Data.Successful {
			if len(getStatusResp.Data.Result) > 0 {
				return getStatusResp.Data.Result[0], nil
			}
			return "", errors.New("no result found")
		}
		return "", errors.New("task failed") // Возвращаем ошибку, если задача завершилась неудачно
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "WHY context was canceled", "reason", context.Cause(ctx).Error())
			return "", ctx.Err() // Возвращает ошибку контекста в случае таймаута
		case <-ticker.C:
			getStatusResp, err := a.cli.GetTaskStatus(ctx, postTaskResp.Data.TaskId)
			if err != nil {
				return "", err // Возвращаем ошибку в случае ошибки при получении статуса задачи
			}

			if getStatusResp.Status != statusOK {
				return "", errors.New("couldn't get task status")
			}

			if getStatusResp.Data.Ready {
				// Если получен ответ о готовности, возвращаем результат
				if getStatusResp.Data.Successful {
					// Если задача выполнена успешно, возвращаем первый результат
					if len(getStatusResp.Data.Result) > 0 {
						return getStatusResp.Data.Result[0], nil
					}
					return "", errors.New("no result found")
				}
				return "", errors.New("task failed") // Возвращаем ошибку, если задача завершилась неудачно
			}
		}
	}
}
