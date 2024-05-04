package helper

import (
	"context"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
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
	logger.Info(context.Background(), "samesite: ", "mode:", samesiteMode())
}

const SessionKey = "session_id"

func SessionIDFromContextMD(ctx context.Context) (string, bool) {
	md, _ := metadata.FromIncomingContext(ctx)
	cookies := md.Get("cookie")

	request := http.Request{Header: http.Header{"Cookie": cookies}}
	sessionCookie, err := request.Cookie(SessionKey)
	if err != nil {
		slog.Error(err.Error())
		return "", false
	}

	return sessionCookie.Value, true
}

func cookie(name string, value string, expires time.Time) http.Cookie {
	return http.Cookie{
		Name:     name,
		Path:     "/",
		Value:    value,
		Expires:  expires,
		Secure:   secure(),
		HttpOnly: true,
		SameSite: samesiteMode(),
	}
}

func GetSessionCookie(domain, sessionID string, expires time.Time) http.Cookie {
	c := cookie(SessionKey, sessionID, expires)
	c.Domain = domain

	return c
}
