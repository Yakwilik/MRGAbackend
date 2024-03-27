package helper

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"net/http"
	"time"
)

const sessionKey = "session_id"

func SessionIDFromContextMD(ctx context.Context) (string, bool) {
	md, _ := metadata.FromIncomingContext(ctx)
	cookie := md.Get("Cookie")

	request := http.Request{Header: http.Header{"Cookie": cookie}}
	sessionCookie, err := request.Cookie(sessionKey)
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Println(err)
		}
		return "", false
	}

	return sessionCookie.Value, true
}

func SetSessionID(ctx context.Context, sessionID string) {
	c := http.Cookie{
		Name:     sessionKey,
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	}

	_ = grpc.SendHeader(ctx, metadata.New(map[string]string{
		"Set-Cookie": c.String(),
	}))
}
