package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSSLMode string
	JWTSecret string
	JWTHours  int
	AppEnv    string
}

func Load() *Config {
	_ = godotenv.Load() // ignore if missing

	jwtHours := 72
	if s := os.Getenv("JWT_EXP_HOURS"); s != "" {
		if h, err := strconv.Atoi(s); err == nil {
			jwtHours = h
		}
	}

	cfg := &Config{
		DBHost:    getenv("DB_HOST", "localhost"),
		DBPort:    getenv("DB_PORT", "5432"),
		DBUser:    getenv("DB_USER", "postgres"),
		DBPass:    getenv("DB_PASS", "postgres"),
		DBName:    getenv("DB_NAME", "rbac_db"),
		DBSSLMode: getenv("DB_SSLMODE", "disable"),
		JWTSecret: getenv("JWT_SECRET", "changemeplease"),
		JWTHours:  jwtHours,
		AppEnv:    getenv("APP_ENV", "development"),
	}

	log.Printf("Loaded config: env=%s db=%s@%s", cfg.AppEnv, cfg.DBUser, cfg.DBHost)
	return cfg
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
