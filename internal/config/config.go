package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Env             string
	HTTPAddr        string
	ShutdownTimeout time.Duration

	LogLevel  string
	LogFormat string

	DatabaseURL string

	JWT JWTConfig

	RBACCacheTTL time.Duration

	SeedAdminEmail string
}

type JWTConfig struct {
	Secret string
	TTL    time.Duration
	Issuer string
}

const minJWTSecretLength = 32

// Load membaca konfigurasi dari environment variable.
func Load() (Config, error) {
	l := &loader{}
	cfg := Config{
		Env:             l.str("APP_ENV", "development"),
		HTTPAddr:        l.str("HTTP_ADDR", ":8080"),
		ShutdownTimeout: l.duration("SHUTDOWN_TIMEOUT", 10*time.Second),
		LogLevel:        l.str("LOG_LEVEL", "info"),
		LogFormat:       l.str("LOG_FORMAT", "json"),
		DatabaseURL:     l.required("DATABASE_URL"),
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
			TTL:    l.duration("JWT_TTL", 24*time.Hour),
			Issuer: l.str("JWT_ISSUER", "backend-core-go"),
		},
		RBACCacheTTL:   l.duration("RBAC_CACHE_TTL", 5*time.Minute),
		SeedAdminEmail: os.Getenv("SEED_ADMIN_EMAIL"),
	}
	if err := errors.Join(l.errs...); err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}
	return cfg, nil
}

// ValidateJWT dipanggil oleh proses yang menerbitkan/memverifikasi token (API).
func (c Config) ValidateJWT() error {
	if len(c.JWT.Secret) < minJWTSecretLength {
		return fmt.Errorf("JWT_SECRET must be at least %d characters", minJWTSecretLength)
	}
	if c.JWT.TTL <= 0 {
		return errors.New("JWT_TTL must be positive")
	}
	return nil
}

func (c Config) IsProduction() bool {
	return c.Env == "production"
}

type loader struct {
	errs []error
}

func (l *loader) str(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func (l *loader) required(key string) string {
	v := os.Getenv(key)
	if v == "" {
		l.errs = append(l.errs, fmt.Errorf("%s is required", key))
	}
	return v
}

func (l *loader) duration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s: %w", key, err))
		return def
	}
	return d
}
