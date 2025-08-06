package config

import (
	"log"
	"os"
)

type Config struct {
	DBUrl string
	Port  string
}

func LoadConfig() Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback
	}

	return Config{
		DBUrl: dbURL,
		Port:  port,
	}
}
