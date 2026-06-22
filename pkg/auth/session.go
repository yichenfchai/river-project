//go:build redis
// +build redis

package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrRefreshTokenReused = fmt.Errorf("refresh token reused")

// SessionRepo Redis session 存储
type SessionRepo struct {
	rdb *redis.Client
}

// RefreshSession Refresh Token 对应的 session 信息
type RefreshSession struct {
	UserID   string
	Role     string
	DeviceID string
}

func NewSessionRepo(rdb *redis.Client) *SessionRepo {
	return &SessionRepo{rdb: rdb}
}

// SaveRefreshToken 存储 Refresh Token
func (r *SessionRepo) SaveRefreshToken(ctx context.Context, userID, deviceID, jti, token string, expiresAt time.Time) error {
	key := fmt.Sprintf("auth:refresh:%s", jti)
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}

	if err := r.rdb.HSet(ctx, key, map[string]interface{}{
		"user_id":    userID,
		"device_id":  deviceID,
		"token_hash": hashToken(token),
	}).Err(); err != nil {
		return err
	}
	return r.rdb.Expire(ctx, key, ttl).Err()
}

// VerifyAndRevokeRefreshToken 验证并原子删除 (一次性使用)
// 返回 session 信息; 如已被使用 → ErrRefreshTokenReused
func (r *SessionRepo) VerifyAndRevokeRefreshToken(ctx context.Context, jti, token string) (*RefreshSession, error) {
	key := fmt.Sprintf("auth:refresh:%s", jti)

	script := `
		local data = redis.call('HGETALL', KEYS[1])
		if #data == 0 then
			return nil
		end
		redis.call('DEL', KEYS[1])
		return data
	`
	result, err := r.rdb.Eval(ctx, script, []string{key}).Result()
	if err != nil {
		return nil, err
	}

	// 未找到 → Token 已被使用或过期
	if result == nil {
		return nil, ErrRefreshTokenReused
	}

	data, ok := result.([]interface{})
	if !ok || len(data) < 4 {
		return nil, ErrRefreshTokenReused
	}

	fields := make(map[string]string, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		k, _ := data[i].(string)
		v, _ := data[i+1].(string)
		fields[k] = v
	}

	// 常量时间比较 (防时序攻击)
	expectedHash := fields["token_hash"]
	actualHash := hashToken(token)
	if subtle.ConstantTimeCompare([]byte(expectedHash), []byte(actualHash)) != 1 {
		return nil, ErrRefreshTokenReused
	}

	return &RefreshSession{
		UserID:   fields["user_id"],
		DeviceID: fields["device_id"],
	}, nil
}

// AddAccessJTI 记录设备会话 (多设备管理)
func (r *SessionRepo) AddAccessJTI(ctx context.Context, userID, deviceID, jti string) error {
	key := fmt.Sprintf("auth:session:%s:%s", userID, deviceID)
	return r.rdb.SAdd(ctx, key, jti).Err()
}

// RemoveAccessJTI 登出时移除
func (r *SessionRepo) RemoveAccessJTI(ctx context.Context, userID, deviceID, jti string) error {
	key := fmt.Sprintf("auth:session:%s:%s", userID, deviceID)
	return r.rdb.SRem(ctx, key, jti).Err()
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
