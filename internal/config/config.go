package config

import (
	"context"
	"flag"
	"github.com/joho/godotenv"
	"log"
	"log/slog"
	"os"
)

type Config struct {
	DBHost             func(ctx context.Context) string
	DBPort             func(ctx context.Context) string
	DBUsername         func(ctx context.Context) string
	DBPassword         func(ctx context.Context) string
	DBName             func(ctx context.Context) string
	ChatBotAddr        func(ctx context.Context) string
	ChatBotServiceAddr func(ctx context.Context) string
	LoggerAddr         func(ctx context.Context) string
	LogLevel           func(ctx context.Context) slog.Level
	Environment        func(ctx context.Context) string
	AppName            func(ctx context.Context) string
	Domain             func(ctx context.Context) string
}

func ParseConfig() *Config {
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

	environment := func(ctx context.Context) string {
		env, ok := os.LookupEnv("ENV")
		if !ok {
			return "dev"
		}
		return env
	}

	return &Config{
		DBHost: func(ctx context.Context) string {
			dbHost, ok := os.LookupEnv("DBHost")
			if !ok {
				return "0.0.0.0"
			}
			if environment(ctx) == "container" {
				return "db"
			}

			return dbHost
		},
		DBPort: func(ctx context.Context) string {
			dbPort, ok := os.LookupEnv("DBPort")
			if !ok {
				return "5432"
			}
			return dbPort
		},
		DBUsername: func(ctx context.Context) string {
			dbUsername, ok := os.LookupEnv("DBUsername")
			if !ok {
				return "postgres"
			}
			return dbUsername
		},
		DBPassword: func(ctx context.Context) string {
			dbPassword, ok := os.LookupEnv("DBPassword")
			if !ok {
				return "qwerty"
			}

			return dbPassword
		},
		DBName: func(ctx context.Context) string {
			dbName, ok := os.LookupEnv("DBName")
			if !ok {
				return "postgres"
			}
			return dbName
		},
		Environment: environment,
		LoggerAddr:  func(ctx context.Context) string { return os.Getenv("LOGGER_ADDR") },
		LogLevel: func(ctx context.Context) slog.Level {
			level, ok := os.LookupEnv("LOG_LEVEL")
			switch {
			case !ok || level == "debug":
				return slog.LevelDebug
			case level == "info":
				return slog.LevelInfo
			case level == "warn":
				return slog.LevelWarn
			case level == "error":
				return slog.LevelError
			default:
				return slog.LevelDebug
			}
		},
		AppName: func(ctx context.Context) string {
			appName, ok := os.LookupEnv("APP_NAME")
			if !ok {
				return "app"
			}
			return appName
		},
		ChatBotAddr: func(ctx context.Context) string {
			chatBotAddr, ok := os.LookupEnv("CHAT_BOT_ADDR")
			if !ok {
				return "0.0.0.0:8002"
			}
			return chatBotAddr
		},
		Domain: func(ctx context.Context) string {
			domain, ok := os.LookupEnv("DOMAIN")
			if !ok {
				return ".localhost"
			}
			return domain
		},
		ChatBotServiceAddr: func(ctx context.Context) string {
			addr, ok := os.LookupEnv("CHAT_BOT_SERVICE_ADDR")
			if !ok {
				return "localhost:50051"
			}
			return addr
		},
	}
}
