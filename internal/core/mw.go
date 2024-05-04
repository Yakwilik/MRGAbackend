package core

import (
	"context"
	"errors"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"log/slog"
	"net/http"
	"strings"
)

func AuthMiddleware(core UseCase, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectRequest := strings.Contains(r.RequestURI, "connect")

		token, err := helper.GetValueCookie(r, helper.SessionKey)
		if err != nil {
			slog.Error(err.Error())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		email, err := core.CheckLogin(r.Context(), token)
		if err != nil {
			slog.Error(err.Error())
			statusCode := http.StatusInternalServerError
			if errValidation := new(model.ValidationError); errors.As(err, &errValidation) {
				statusCode = http.StatusUnauthorized
			}
			if connectRequest {
				statusCode = http.StatusOK
			}
			w.WriteHeader(statusCode)
			if connectRequest {
				w.Write([]byte("{\n  \"disconnect\": {\n    \"code\": 4501,\n    \"reason\": \"unauthorized\"\n  }\n}"))
			}
			return
		}

		ctx := context.WithValue(r.Context(), helper.ContextUserKey, helper.SessionData{
			Email:     email,
			SessionID: token,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
