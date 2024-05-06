package helper

import (
	"context"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"net/http"
	"strconv"
	"time"
)

type MockResponseConfig struct {
	Generate            bool
	IterationCount      int
	BatchSize           int
	Redirect            string
	AdditionalQuestions string
	IterationTimeout    time.Duration
}

type mockResponseConfigKey struct{}

var contextMockResponseConfigKey mockResponseConfigKey

func MockResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		genMockResponse := r.Header.Get("X-Generate-Mock-Response")
		mockResponseIterationCount := r.Header.Get("X-Mock-Response-Iteration-Count")
		mockResponseBatchSize := r.Header.Get("X-Mock-Response-Batch-Size")
		mockResponseIterationTimeout := r.Header.Get("X-Mock-Response-Iteration-Timeout")
		mockResponseRedirectKey := r.Header.Get("X-Mock-Response-Redirect-Key")
		mockResponseAdditionalQuestions := r.Header.Get("X-Mock-Response-Additional-Questions")

		// Преобразуем данные в нужный формат
		generate, _ := strconv.ParseBool(genMockResponse)
		iterationCount, _ := strconv.Atoi(mockResponseIterationCount)
		batchSize, _ := strconv.Atoi(mockResponseBatchSize)
		iterationTimeout, _ := time.ParseDuration(mockResponseIterationTimeout)

		// Создаем структуру MockResponseConfig
		config := MockResponseConfig{
			Generate:            generate,
			IterationCount:      iterationCount,
			BatchSize:           batchSize,
			IterationTimeout:    iterationTimeout,
			Redirect:            mockResponseRedirectKey,
			AdditionalQuestions: mockResponseAdditionalQuestions,
		}

		// Положим структуру в контекст
		ctx := context.WithValue(r.Context(), contextMockResponseConfigKey, &config)

		logger.Info(ctx, "Mock Response Config", "config", config)
		// Передаем управление следующему обработчику
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetMockResponseConfig(ctx context.Context) (*MockResponseConfig, bool) {
	config, ok := ctx.Value(contextMockResponseConfigKey).(*MockResponseConfig)
	return config, ok
}
