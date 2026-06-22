package config

import (
	"os"
	"testing"
	"time"

	"github.com/yichenfchai/river-project/pkg/secrets"
)

func setupEnv() func() {
	// Save and clear relevant env vars
	saved := map[string]string{}
	for _, k := range []string{
		"SERVER_HOST", "SERVER_PORT", "SERVER_MODE",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "DB_CONN_MAX_LIFETIME", "DB_CONN_MAX_IDLE_TIME",
		"JWT_SECRET", "JWT_ACCESS_TTL", "JWT_REFRESH_TTL",
		"LLM_PROVIDER", "LLM_BASE_URL", "LLM_API_KEY", "LLM_MODEL", "LLM_TIMEOUT", "LLM_MAX_TOKENS",
	} {
		saved[k] = os.Getenv(k)
		os.Unsetenv(k)
	}
	return func() {
		for k, v := range saved {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}
}

func TestEnvStr_Fallback(t *testing.T) {
	cleanup := setupEnv()
	defer cleanup()

	if v := envStr("NONEXISTENT_KEY", "default"); v != "default" {
		t.Errorf("got %q, want default", v)
	}
}

func TestEnvStr_FromEnv(t *testing.T) {
	cleanup := setupEnv()
	defer cleanup()

	os.Setenv("SERVER_HOST", "127.0.0.1")
	if v := envStr("SERVER_HOST", "0.0.0.0"); v != "127.0.0.1" {
		t.Errorf("got %q, want 127.0.0.1", v)
	}
}

func TestEnvInt_Fallback(t *testing.T) {
	cleanup := setupEnv()
	defer cleanup()

	if v := envInt("NONEXISTENT", 8080); v != 8080 {
		t.Errorf("got %d, want 8080", v)
	}
}

func TestEnvInt_FromEnv(t *testing.T) {
	cleanup := setupEnv()
	defer cleanup()

	os.Setenv("SERVER_PORT", "3000")
	if v := envInt("SERVER_PORT", 8080); v != 3000 {
		t.Errorf("got %d, want 3000", v)
	}
}

func TestEnvInt_InvalidFormat(t *testing.T) {
	cleanup := setupEnv()
	defer cleanup()

	os.Setenv("SERVER_PORT", "not-a-number")
	if v := envInt("SERVER_PORT", 8080); v != 8080 {
		t.Errorf("got %d, want fallback 8080", v)
	}
}

func TestEnvDuration_Fallback(t *testing.T) {
	cleanup := setupEnv()
	defer cleanup()

	if v := envDuration("NONEXISTENT", 5*time.Minute); v != 5*time.Minute {
		t.Errorf("got %v, want 5m", v)
	}
}

func TestEnvDuration_FromEnv(t *testing.T) {
	cleanup := setupEnv()
	defer cleanup()

	os.Setenv("JWT_ACCESS_TTL", "30m")
	if v := envDuration("JWT_ACCESS_TTL", 15*time.Minute); v != 30*time.Minute {
		t.Errorf("got %v, want 30m", v)
	}
}

func TestEnvDuration_InvalidFormat(t *testing.T) {
	cleanup := setupEnv()
	defer cleanup()

	os.Setenv("JWT_ACCESS_TTL", "not-a-duration")
	if v := envDuration("JWT_ACCESS_TTL", 15*time.Minute); v != 15*time.Minute {
		t.Errorf("got %v, want fallback 15m", v)
	}
}

func TestLoad_Defaults(t *testing.T) {
	cleanup := setupEnv()
	defer cleanup()

	sec := secrets.New("")
	cfg := Load(sec)

	// Server defaults
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d", cfg.Server.Port)
	}

	// DB defaults
	if cfg.Database.Host != "localhost" {
		t.Errorf("DB.Host = %q", cfg.Database.Host)
	}
	if cfg.Database.MaxOpenConns != 25 {
		t.Errorf("DB.MaxOpenConns = %d", cfg.Database.MaxOpenConns)
	}

	// JWT defaults
	if cfg.JWT.AccessTTL != 15*time.Minute {
		t.Errorf("JWT.AccessTTL = %v", cfg.JWT.AccessTTL)
	}
	if cfg.JWT.RefreshTTL != 7*24*time.Hour {
		t.Errorf("JWT.RefreshTTL = %v", cfg.JWT.RefreshTTL)
	}

	// LLM defaults
	if cfg.LLM.Provider != "openai" {
		t.Errorf("LLM.Provider = %q", cfg.LLM.Provider)
	}
	if cfg.LLM.Timeout != 30*time.Second {
		t.Errorf("LLM.Timeout = %v", cfg.LLM.Timeout)
	}

	// JWT secret should be empty (no hardcoded default)
	if cfg.JWT.Secret != "" {
		t.Error("JWT.Secret should default to empty string")
	}
}

func TestLoad_FromEnv(t *testing.T) {
	cleanup := setupEnv()
	defer cleanup()

	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("LLM_PROVIDER", "qwen")

	sec := secrets.New("")
	cfg := Load(sec)

	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Database.Host != "db.example.com" {
		t.Errorf("DB.Host = %q", cfg.Database.Host)
	}
	if cfg.Database.Port != 3306 {
		t.Errorf("DB.Port = %d", cfg.Database.Port)
	}
	if cfg.LLM.Provider != "qwen" {
		t.Errorf("LLM.Provider = %q", cfg.LLM.Provider)
	}
}

func TestLoad_SecretsPriority(t *testing.T) {
	cleanup := setupEnv()
	defer cleanup()

	// Set env var
	os.Setenv("DB_PASSWORD", "env-pass")

	sec := secrets.New("")
	cfg := Load(sec)

	// secrets.Get has priority chain: file > env > fallback
	// Since no file exists, should use env value, not fallback
	if cfg.Database.Password != "env-pass" {
		t.Errorf("DB.Password = %q, want env-pass (env should override fallback)",
			cfg.Database.Password)
	}
}
