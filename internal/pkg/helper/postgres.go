package helper

import (
	"database/sql"
	"fmt"
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
	return db, nil
}
