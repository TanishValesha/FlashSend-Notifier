package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BindAddr           string
	JwtSecret          string
	APIKeyHmacSecret   string
	JwtExpirationHours string
	DB                 string
}

var Cfg Config

func Load() {
	_ = godotenv.Load()

	Cfg = Config{
		BindAddr:           getEnv("BIND_ADDR", ":8080"),
		JwtSecret:          mustGet("JWT_SECRET"),
		APIKeyHmacSecret:   mustGet("APIKEY_HMAC_SECRET"),
		JwtExpirationHours: getEnv("JWT_EXPIRATION_HOURS", "24"),
		DB:                 mustGet("DATABASE_URL"),
	}

}

func mustGet(key string) string {
	env := os.Getenv(key)
	if env == "" {
		log.Fatalf("Missing required env: %s", key)
	}
	return env
}

func getEnv(key string, fallback string) string {
	env := os.Getenv(key)

	if env == "" {
		return fallback
	}

	return env
}
