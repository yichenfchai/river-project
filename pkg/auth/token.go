package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Claims struct {
	jwt.RegisteredClaims
	Role     string `json:"role"`
	DeviceID string `json:"device_id,omitempty"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	rdb        *redis.Client
	blacklist  sync.Map // fallback when Redis is nil
}

func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return NewTokenManagerWithRedis(secret, accessTTL, refreshTTL, nil)
}

func NewTokenManagerWithRedis(secret string, accessTTL, refreshTTL time.Duration, rdb *redis.Client) *TokenManager {
	m := &TokenManager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		rdb:        rdb,
	}
	if rdb == nil {
		go m.memCleanup(5 * time.Minute)
	}
	return m
}

// ─── Token Issuance ───

func (m *TokenManager) IssueTokens(userID, role, deviceID string) (*TokenPair, error) {
	now := time.Now()
	accessJTI := uuid.New().String()

	accessClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessJTI,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
		Role:     role,
		DeviceID: deviceID,
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(m.secret)
	if err != nil {
		return nil, err
	}

	refreshJTI := uuid.New().String()
	refreshClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTTL)),
		},
		Role:     role,
		DeviceID: deviceID,
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(m.secret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(m.accessTTL.Seconds()),
	}, nil
}

// ─── Token Parsing ───

func (m *TokenManager) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{},
		func(t *jwt.Token) (interface{}, error) {
			return m.secret, nil
		},
	)
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

// ─── Blacklist ───

const blacklistPrefix = "auth:blacklist:"

func (m *TokenManager) BlacklistJTI(jti string, expiresAt time.Time) {
	if m.rdb != nil {
		ttl := time.Until(expiresAt)
		if ttl <= 0 {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = m.rdb.Set(ctx, blacklistPrefix+jti, "1", ttl).Err()
		return
	}
	m.blacklist.Store(jti, expiresAt)
}

func (m *TokenManager) IsBlacklisted(jti string) bool {
	if jti == "" {
		return false
	}
	if m.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		n, err := m.rdb.Exists(ctx, blacklistPrefix+jti).Result()
		if err != nil {
			return false // Redis 不可用时放行
		}
		return n > 0
	}

	val, ok := m.blacklist.Load(jti)
	if !ok {
		return false
	}
	exp, ok := val.(time.Time)
	if !ok || time.Now().After(exp) {
		m.blacklist.Delete(jti)
		return false
	}
	return true
}

func (m *TokenManager) memCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		m.blacklist.Range(func(key, val interface{}) bool {
			exp, ok := val.(time.Time)
			if !ok || now.After(exp) {
				m.blacklist.Delete(key)
			}
			return true
		})
	}
}

func init() { _ = fmt.Sprintf("") }
