package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Security SecurityConfig
}

type ServerConfig struct {
	Port         string
	Env          string
	AllowedOrigins []string
	MaxBodySize  int64
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string
}

type SecurityConfig struct {
	RateLimit        int
	RateLimitWindow  time.Duration
	Argon2Memory     uint32
	Argon2Iterations uint32
	Argon2SaltLen    uint8
	Argon2KeyLen     uint32
}

func Load() (*Config, error) {
	// ТРЕБОВАНИЕ 5: Все секреты из переменных окружения
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Env:          getEnv("ENV", "development"),
			AllowedOrigins: getEnvSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			MaxBodySize:  getEnvInt64("MAX_BODY_SIZE", 10<<20), // 10 MB
		},
		Database: DatabaseConfig{
			Host:     getEnvRequired("DB_HOST"),
			Port:     getEnv("DB_PORT", "5432"),
			DBName:   getEnvRequired("DB_NAME"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:     getEnvRequired("JWT_SECRET"),
			AccessTTL:  30 * time.Minute,
			RefreshTTL: 7 * 24 * time.Hour,
			Issuer:     getEnv("JWT_ISSUER", "devbrain-pro"),
		},
		Security: SecurityConfig{
			RateLimit:        getEnvInt("RATE_LIMIT", 10),
			RateLimitWindow:  time.Minute,
			Argon2Memory:     64 * 1024,
			Argon2Iterations: 3,
			Argon2SaltLen:    16,
			Argon2KeyLen:     32,
		},
	}

	// 🔑 Установка User/Password только если не SQLite
	if cfg.Database.Host != "sqlite" {
		cfg.Database.User = getEnvRequired("DB_USER")
		cfg.Database.Password = getEnvRequired("DB_PASSWORD")
	} else {
		cfg.Database.User = ""
		cfg.Database.Password = ""
	}

	// ТРЕБОВАНИЕ 5: Проверка обязательных переменных
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration error: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Env == "production" && len(c.Server.AllowedOrigins) == 0 {
		return fmt.Errorf("ALLOWED_ORIGINS must be set in production")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	return nil
}

// Вспомогательные функции
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvRequired(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Required environment variable %s is not set", key))
	}
	return val
}

func getEnvSlice(key string, defaultVal []string) []string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	// Simple comma-split (you can improve later)
	split := []string{val}
	return split
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

func getEnvInt64(key string, defaultVal int64) int64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defaultVal
	}
	return i
}
