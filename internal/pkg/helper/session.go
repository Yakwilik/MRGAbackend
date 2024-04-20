package helper

import (
	"context"
	"flag"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
)

var secure = false
var samesiteMode http.SameSite = http.SameSiteLaxMode

func init() {
	withDotEnv := false
	flag.BoolVar(&withDotEnv, "dotenv", false, "используется ли .env файл")
	flag.Parse()

	if withDotEnv {
		err := godotenv.Load()
		log.Println("parsed .env file")
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	sec, ok := os.LookupEnv("COOKIE_SECURE")
	switch {
	case !ok || sec == "false":
		secure = false
	default:
		secure = true
	}

	samesite, ok := os.LookupEnv("SAME_SITE_MODE")
	switch {
	case samesite == "LAX":
		samesiteMode = http.SameSiteLaxMode
	case samesite == "STRICT":
		samesiteMode = http.SameSiteStrictMode
	case !ok || samesite == "NONE":
		fallthrough
	default:
		samesiteMode = http.SameSiteNoneMode
	}
	slog.Info("samesite: ", "mode:", samesite)

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
		Domain:   ".meetme-app.ru",
		Expires:  time.Now().Add(time.Hour * 24),
		Secure:   secure,
		HttpOnly: true,
		SameSite: samesiteMode,
	}

	_ = grpc.SendHeader(ctx, metadata.New(map[string]string{
		"Set-Cookie": c.String(),
	}))

}
