package helper

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	_ "github.com/lib/pq"
)

type PGConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
}

func NewPostgresDB(cfg PGConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	db.SetMaxOpenConns(1000)
	db.SetMaxIdleConns(100)
	if err != nil {
		return nil, err
	}
	logger.Info(context.Background(), "connected to postgres")
	return db, nil
}
