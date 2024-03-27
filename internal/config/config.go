package config

import "os"

type Config struct {
	DBHost     string
	DBPort     string
	DBUsername string
	DBPassword string
	DBName     string
}

func ParseConfig() Config {
	dbHost, ok := os.LookupEnv("DBHost")
	if !ok {
		dbHost = "0.0.0.0"
	}

	dbPort, ok := os.LookupEnv("DBPort")
	if !ok {
		dbPort = "5432"
	}

	dbUsername, ok := os.LookupEnv("DBUsername")
	if !ok {
		dbUsername = "postgres"
	}

	dbPassword, ok := os.LookupEnv("DBPassword")
	if !ok {
		dbPassword = "qwerty"
	}

	dbName, ok := os.LookupEnv("DBName")
	if !ok {
		dbName = "postgres"
	}

	return Config{
		DBHost:     dbHost,
		DBPort:     dbPort,
		DBUsername: dbUsername,
		DBPassword: dbPassword,
		DBName:     dbName,
	}
}
