package config

import (
	"os"
	"strconv"
	"time"

	"github.com/yichenfchai/river-project/pkg/secrets"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	LLM      LLMConfig
}

type LLMConfig struct {
	Provider    string
	BaseURL     string
	APIKey      string
	Model       string
	Timeout     time.Duration
	MaxTokens   int
	Temperature float64
}

type ServerConfig struct {
	Host string
	Port int
	Mode string
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func Load(s *secrets.Store) Config {
	return Config{
		Server: ServerConfig{
			Host: envStr("SERVER_HOST", "0.0.0.0"),
			Port: envInt("SERVER_PORT", 8080),
			Mode: envStr("SERVER_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:            envStr("DB_HOST", "localhost"),
			Port:            envInt("DB_PORT", 5432),
			User:            envStr("DB_USER", "postgres"),
			Password:        s.Get("DB_PASSWORD", ""),
			DBName:          envStr("DB_NAME", "grand_canal_db"),
			SSLMode:         envStr("DB_SSLMODE", "disable"),
			MaxOpenConns:    envInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    envInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: envDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: envDuration("DB_CONN_MAX_IDLE_TIME", 1*time.Minute),
		},
		JWT: JWTConfig{
			Secret:     s.Get("JWT_SECRET", ""),
			AccessTTL:  envDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL: envDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
		},
		LLM: LLMConfig{
			Provider:    envStr("LLM_PROVIDER", "openai"),
			BaseURL:     envStr("LLM_BASE_URL", "https://api.openai.com/v1"),
			APIKey:      s.Get("LLM_API_KEY", ""),
			Model:       envStr("LLM_MODEL", "gpt-3.5-turbo"),
			Timeout:     envDuration("LLM_TIMEOUT", 30*time.Second),
			MaxTokens:   envInt("LLM_MAX_TOKENS", 1024),
			Temperature: 0.7,
		},
	}
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
