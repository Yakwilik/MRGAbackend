package helper

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log/slog"
	"net/http"
	"os"
	"time"
)

var secure = false
var samesiteMode http.SameSite = http.SameSiteLaxMode

func init() {
	sec, ok := os.LookupEnv("COOKIE_SECURE")
	switch {
	case !ok || sec == "false":
		secure = false
	default:
		secure = true
	}

	samesite, ok := os.LookupEnv("SAME_SITE_NONE")
	switch {
	case !ok || samesite == "false":
		samesiteMode = http.SameSiteLaxMode
	default:
		samesiteMode = http.SameSiteNoneMode
	}

}

const sessionKey = "session_id"

func SessionIDFromContextMD(ctx context.Context) (string, bool) {
	md, _ := metadata.FromIncomingContext(ctx)
	cookie := md.Get("Cookie")

	request := http.Request{Header: http.Header{"Cookie": cookie}}
	sessionCookie, err := request.Cookie(sessionKey)
	if err != nil {
		slog.Error(err.Error())
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
		Secure:   secure,
		HttpOnly: true,
		SameSite: samesiteMode,
	}

	_ = grpc.SendHeader(ctx, metadata.New(map[string]string{
		"Set-Cookie": c.String(),
	}))
}
