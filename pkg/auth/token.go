package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/grand-canal-guardian/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// TokenManager JWT 双 Token 管理器
type TokenManager struct {
	secret      []byte
	blacklist   *TokenBlacklist
	sessionRepo *SessionRepo
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

func NewTokenManager(secret string, rdb *redis.Client, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		secret:      []byte(secret),
		blacklist:   NewTokenBlacklist(rdb),
		sessionRepo: NewSessionRepo(rdb),
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
	}
}

// Claims JWT 载荷
type Claims struct {
	jwt.RegisteredClaims
	UserID   string `json:"sub"`
	Role     string `json:"role"`
	DeviceID string `json:"device_id"`
}

// TokenPair Access + Refresh
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// IssueTokens 签发双 Token
func (m *TokenManager) IssueTokens(ctx context.Context, userID, role, deviceID string) (*TokenPair, error) {
	now := time.Now()

	// Access Token (JWT)
	accessJTI := uuid.New().String()
	accessClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessJTI,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
		UserID:   userID,
		Role:     role,
		DeviceID: deviceID,
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(m.secret)
	if err != nil {
		return nil, errors.Internal(err)
	}

	// Refresh Token (随机字符串，非 JWT — 防伪造)
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, errors.Internal(err)
	}

	// 存储 Refresh Token → Redis
	refreshJTI := uuid.New().String()
	if err := m.sessionRepo.SaveRefreshToken(ctx, userID, deviceID, refreshJTI, refreshToken, now.Add(m.refreshTTL)); err != nil {
		return nil, errors.WrapDefault(errors.ErrRedisError, err)
	}

	// 记录 session (多设备管理)
	if err := m.sessionRepo.AddAccessJTI(ctx, userID, deviceID, accessJTI); err != nil {
		return nil, errors.WrapDefault(errors.ErrRedisError, err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: fmt.Sprintf("%s.%s", refreshJTI, refreshToken),
		TokenType:    "Bearer",
		ExpiresIn:    int(m.accessTTL.Seconds()),
	}, nil
}

// RefreshAccess Refresh Token 旋转 — 旧 Token 立即失效
func (m *TokenManager) RefreshAccess(ctx context.Context, refreshTokenRaw string) (*TokenPair, error) {
	jti, token, err := parseRefreshToken(refreshTokenRaw)
	if err != nil {
		return nil, errors.NewDefault(errors.ErrTokenInvalid)
	}

	session, err := m.sessionRepo.VerifyAndRevokeRefreshToken(ctx, jti, token)
	if err != nil {
		if err == ErrRefreshTokenReused {
			// 重放攻击: 全设备踢出
			go m.blacklist.BlacklistAllUserTokens(context.Background(), session.UserID)
			return nil, errors.NewDefault(errors.ErrTokenInvalid)
		}
		return nil, errors.NewDefault(errors.ErrTokenExpired)
	}

	return m.IssueTokens(ctx, session.UserID, session.Role, session.DeviceID)
}

// ValidateAccess 验证 Access Token (含黑名单检查)
func (m *TokenManager) ValidateAccess(ctx context.Context, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, errors.NewDefault(errors.ErrTokenExpired)
	}
	if !token.Valid {
		return nil, errors.NewDefault(errors.ErrTokenInvalid)
	}

	// 黑名单检查
	blacklisted, err := m.blacklist.IsBlacklisted(ctx, claims.ID)
	if err != nil {
		// Redis 不可用 → 放行 (可用性优先)
		return claims, nil
	}
	if blacklisted {
		return nil, errors.NewDefault(errors.ErrTokenInvalid)
	}

	return claims, nil
}

// Logout 登出 — Access Token 入黑名单
func (m *TokenManager) Logout(ctx context.Context, claims *Claims) error {
	if err := m.blacklist.Add(ctx, claims.ID, claims.ExpiresAt.Time); err != nil {
		return err
	}
	return m.sessionRepo.RemoveAccessJTI(ctx, claims.UserID, claims.DeviceID, claims.ID)
}

// ---- 内部工具 ----

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func parseRefreshToken(raw string) (jti string, token string, err error) {
	for i := len(raw) - 1; i >= 0; i-- {
		if raw[i] == '.' {
			return raw[:i], raw[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("invalid refresh token format")
}
