package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	Port        string
	DBPath      string
	JWTSecret   string
	CORSOrigins []string
}

func loadConfig() Config {
	_ = godotenv.Load()

	env := getEnv("APP_ENV", "development")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		if env == "production" {
			slog.Error("JWT_SECRET must be set in production")
			os.Exit(1)
		}
		slog.Warn("JWT_SECRET not set, using an insecure default for local development only")
		jwtSecret = "xqjfa5dgWiPajvAZ1aOR/IYju+Vbd2FEkRXkYq+qtpI="
	}

	corsOrigins := strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"), ",")
	for i := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
	}

	return Config{
		Env:         env,
		Port:        getEnv("PORT", "8080"),
		DBPath:      getEnv("DB_PATH", "blog.db"),
		JWTSecret:   jwtSecret,
		CORSOrigins: corsOrigins,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
