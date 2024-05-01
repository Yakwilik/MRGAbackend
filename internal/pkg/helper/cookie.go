package helper

import (
	"context"
	"errors"
	"github.com/Yakwilik/MRGAbackend/internal/core"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"log/slog"
	"net/http"
	"strings"
)

func GetValueCookie(r *http.Request, nameCookie string) (string, error) {
	valueCookie, err := r.Cookie(nameCookie)
	if err != nil {
		return "", err
	}
	return valueCookie.Value, err
}

type userKey struct{}

var contextUserKey userKey = userKey{}

type sessionData struct {
	Email     string
	SessionID string
}

func AuthMiddleware(core core.UseCase, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectRequest := strings.Contains(r.RequestURI, "connect")

		token, err := GetValueCookie(r, sessionKey)
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

		ctx := context.WithValue(r.Context(), contextUserKey, sessionData{
			Email:     email,
			SessionID: token,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetEmailFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(contextUserKey)
	sessionDataValue, ok := v.(sessionData)
	if !ok {
		return "", model.ErrNotFound
	}

	return sessionDataValue.Email, nil
}

func GetSessionIDFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(contextUserKey)
	sessionDataValue, ok := v.(sessionData)
	if !ok {
		return "", model.ErrNotFound
	}

	return sessionDataValue.SessionID, nil
}
