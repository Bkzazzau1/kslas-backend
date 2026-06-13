package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName           string
	AppEnv            string
	HTTPPort          string
	DatabaseURL       string
	AutoMigrate       bool
	SeedRBAC          bool
	DBMaxIdleConns    int
	DBMaxOpenConns    int
	DBConnMaxLifetime time.Duration
	JWTSecret         string
	JWTExpiresHours   int
	StorageDir        string
	MaxUploadBytes    int64
	BootstrapAdmin    BootstrapAdminConfig
}

type BootstrapAdminConfig struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

func Load() (Config, error) {
	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}

	if len(jwtSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must be at least 32 characters")
	}

	cfg := Config{
		AppName:           getEnv("APP_NAME", "kslasbackend"),
		AppEnv:            getEnv("APP_ENV", "development"),
		HTTPPort:          getEnv("APP_PORT", "8080"),
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		AutoMigrate:       getEnvBool("AUTO_MIGRATE", true),
		SeedRBAC:          getEnvBool("SEED_RBAC", true),
		DBMaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		DBMaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 20),
		DBConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
		JWTSecret:         jwtSecret,
		JWTExpiresHours:   getEnvInt("JWT_EXPIRES_HOURS", 24),
		StorageDir:        getEnv("STORAGE_DIR", "storage"),
		MaxUploadBytes:    getEnvInt64("MAX_UPLOAD_BYTES", 1<<30),
		BootstrapAdmin: BootstrapAdminConfig{
			Email:     strings.ToLower(getEnv("BOOTSTRAP_ADMIN_EMAIL", "")),
			Password:  strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_PASSWORD")),
			FirstName: getEnv("BOOTSTRAP_ADMIN_FIRST_NAME", "System"),
			LastName:  getEnv("BOOTSTRAP_ADMIN_LAST_NAME", "Admin"),
		},
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	if cfg.JWTExpiresHours < 1 {
		return Config{}, errors.New("JWT_EXPIRES_HOURS must be greater than zero")
	}

	if strings.TrimSpace(cfg.StorageDir) == "" {
		return Config{}, errors.New("STORAGE_DIR is required")
	}

	if cfg.MaxUploadBytes < 1<<20 {
		return Config{}, errors.New("MAX_UPLOAD_BYTES must be at least 1048576")
	}

	if (cfg.BootstrapAdmin.Email == "") != (cfg.BootstrapAdmin.Password == "") {
		return Config{}, errors.New("BOOTSTRAP_ADMIN_EMAIL and BOOTSTRAP_ADMIN_PASSWORD must be provided together")
	}

	if cfg.BootstrapAdmin.Password != "" && len(cfg.BootstrapAdmin.Password) < 8 {
		return Config{}, errors.New("BOOTSTRAP_ADMIN_PASSWORD must be at least 8 characters")
	}

	if cfg.BootstrapAdmin.Email != "" && !strings.Contains(cfg.BootstrapAdmin.Email, "@") {
		return Config{}, fmt.Errorf("invalid BOOTSTRAP_ADMIN_EMAIL %q", cfg.BootstrapAdmin.Email)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	value, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}

	return value
}

func getEnvInt64(key string, fallback int64) int64 {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return fallback
	}

	return value
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	value, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}

	return value
}
