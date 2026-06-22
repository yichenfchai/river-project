package auth

import (
	"testing"
	"time"
)

func TestNewTokenManager(t *testing.T) {
	tm := NewTokenManager("my-secret-key", 15*time.Minute, 7*24*time.Hour)
	if tm == nil {
		t.Fatal("NewTokenManager returned nil")
	}
	if tm.accessTTL != 15*time.Minute {
		t.Errorf("accessTTL = %v, want %v", tm.accessTTL, 15*time.Minute)
	}
	if tm.refreshTTL != 7*24*time.Hour {
		t.Errorf("refreshTTL = %v, want %v", tm.refreshTTL, 7*24*time.Hour)
	}
}

func TestIssueTokens(t *testing.T) {
	tm := NewTokenManager("test-secret", 15*time.Minute, 7*24*time.Hour)

	pair, err := tm.IssueTokens("user-123", "user", "device-abc")
	if err != nil {
		t.Fatalf("IssueTokens error: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("AccessToken is empty")
	}
	if pair.RefreshToken == "" {
		t.Error("RefreshToken is empty")
	}
	if pair.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", pair.TokenType, "Bearer")
	}
	if pair.ExpiresIn != int(15*time.Minute.Seconds()) {
		t.Errorf("ExpiresIn = %d, want %d", pair.ExpiresIn, int(15*time.Minute.Seconds()))
	}
	if pair.AccessToken == pair.RefreshToken {
		t.Error("AccessToken and RefreshToken should differ")
	}
}

func TestParseToken_Valid(t *testing.T) {
	tm := NewTokenManager("secret", 15*time.Minute, 7*24*time.Hour)
	pair, err := tm.IssueTokens("user-456", "admin", "dev-x")
	if err != nil {
		t.Fatalf("IssueTokens error: %v", err)
	}

	claims, err := tm.ParseToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}

	if claims.Subject != "user-456" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "user-456")
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %q, want %q", claims.Role, "admin")
	}
	if claims.DeviceID != "dev-x" {
		t.Errorf("DeviceID = %q, want %q", claims.DeviceID, "dev-x")
	}
}

func TestParseToken_InvalidSignature(t *testing.T) {
	tm := NewTokenManager("secret-a", 15*time.Minute, 7*24*time.Hour)
	pair, err := tm.IssueTokens("user-1", "user", "")
	if err != nil {
		t.Fatalf("IssueTokens error: %v", err)
	}

	// Parse with different secret
	tm2 := NewTokenManager("secret-b", 15*time.Minute, 7*24*time.Hour)
	_, err = tm2.ParseToken(pair.AccessToken)
	if err == nil {
		t.Error("Expected error for invalid signature, got nil")
	}
}

func TestParseToken_Expired(t *testing.T) {
	// Use very short TTL to force expire
	tm := NewTokenManager("secret", -1*time.Second, -1*time.Second)
	pair, err := tm.IssueTokens("user-1", "user", "")
	if err != nil {
		t.Fatalf("IssueTokens error: %v", err)
	}

	_, err = tm.ParseToken(pair.AccessToken)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestParseToken_Malformed(t *testing.T) {
	tm := NewTokenManager("secret", 15*time.Minute, 7*24*time.Hour)

	_, err := tm.ParseToken("not-a-jwt-token-at-all")
	if err == nil {
		t.Error("Expected error for malformed token, got nil")
	}

	_, err = tm.ParseToken("")
	if err == nil {
		t.Error("Expected error for empty token, got nil")
	}
}

func TestParseToken_RefreshToken(t *testing.T) {
	tm := NewTokenManager("secret", 15*time.Minute, 7*24*time.Hour)
	pair, err := tm.IssueTokens("user-789", "monitor", "mobile")
	if err != nil {
		t.Fatalf("IssueTokens error: %v", err)
	}

	// Refresh token should also be parseable
	claims, err := tm.ParseToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("ParseToken(refresh) error: %v", err)
	}

	if claims.Subject != "user-789" {
		t.Errorf("Refresh Subject = %q, want %q", claims.Subject, "user-789")
	}
	if claims.Role != "monitor" {
		t.Errorf("Refresh Role = %q, want %q", claims.Role, "monitor")
	}
}
