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
	SMTPEmail          string
	SMTPAppPassword    string
	SMTPHost           string
	SMTPPort           string
	TwilioAccountSID   string
	TwilioAuthToken    string
	TwilioPhoneNumber  string
	VonageAPIKey       string
	VonageAPISecret    string
	VonageFrom         string
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
		SMTPEmail:          mustGet("SMTP_EMAIL"),
		SMTPAppPassword:    mustGet("SMTP_APP_PASSWORD"),
		SMTPHost:           mustGet("SMTP_HOST"),
		SMTPPort:           mustGet("SMTP_PORT"),
		TwilioAccountSID:   mustGet("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:    mustGet("TWILIO_AUTH_TOKEN"),
		TwilioPhoneNumber:  mustGet("TWILIO_PHONE_NUMBER"),
		VonageAPIKey:       mustGet("VONAGE_API_KEY"),
		VonageAPISecret:    mustGet("VONAGE_API_SECRET"),
		VonageFrom:         mustGet("VONAGE_FROM"),
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
