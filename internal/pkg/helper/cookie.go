package helper

import (
	"context"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"net/http"
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

type SessionData struct {
	Email     string
	SessionID string
}

func GetEmailFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(ContextUserKey)
	sessionDataValue, ok := v.(SessionData)
	if !ok {
		return "", model.ErrNotFound
	}

	return sessionDataValue.Email, nil
}

func GetSessionIDFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(ContextUserKey)
	sessionDataValue, ok := v.(SessionData)
	if !ok {
		return "", model.ErrNotFound
	}

	return sessionDataValue.SessionID, nil
}
