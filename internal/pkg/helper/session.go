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

var secure = func() bool {
	sec, ok := os.LookupEnv("COOKIE_SECURE")
	switch {
	case !ok || sec == "false":
		return false
	default:
		return true
	}
}
var samesiteMode = func() http.SameSite {
	samesite, ok := os.LookupEnv("SAME_SITE_MODE")
	switch {
	case samesite == "LAX":
		return http.SameSiteLaxMode
	case samesite == "STRICT":
		return http.SameSiteStrictMode
	case !ok || samesite == "NONE":
		fallthrough
	default:
		return http.SameSiteNoneMode
	}
}

func init() {
	slog.Info("samesite: ", "mode:", samesiteMode())

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

func SetSessionID(ctx context.Context, domain, sessionID string) {
	c := http.Cookie{
		Name:     sessionKey,
		Value:    sessionID,
		Path:     "/",
		Domain:   domain,
		Expires:  time.Now().Add(time.Hour * 24),
		Secure:   secure(),
		HttpOnly: true,
		SameSite: samesiteMode(),
	}

	_ = grpc.SendHeader(ctx, metadata.New(map[string]string{
		"Set-Cookie": c.String(),
	}))

}
