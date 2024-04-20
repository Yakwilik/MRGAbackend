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

var ContextUserKey userKey = userKey{}

func AuthMiddleware(core core.UseCase, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectRequest := strings.Contains(r.RequestURI, "connect")
		//if connectRequest {
		//	ctx := context.WithValue(r.Context(), ContextUserKey, "khas.2015@gmail.com")
		//	next.ServeHTTP(w, r.WithContext(ctx))
		//	return
		//}

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

		ctx := context.WithValue(r.Context(), ContextUserKey, email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetEmailFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(ContextUserKey)
	email, ok := v.(string)
	if !ok {
		return "", model.ErrNotFound
	}

	return email, nil
}
