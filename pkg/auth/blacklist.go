package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBlacklist 基于 Redis 的 JWT 黑名单
type TokenBlacklist struct {
	rdb *redis.Client
}

func NewTokenBlacklist(rdb *redis.Client) *TokenBlacklist {
	return &TokenBlacklist{rdb: rdb}
}

// Add 将 Token 加入黑名单，TTL = Token 剩余有效期
func (b *TokenBlacklist) Add(ctx context.Context, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return b.rdb.Set(ctx, key(jti), "1", ttl).Err()
}

// IsBlacklisted 检查 Token 是否在黑名单中
func (b *TokenBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	exists, err := b.rdb.Exists(ctx, key(jti)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// BlacklistAllUserTokens 踢出某用户所有设备
func (b *TokenBlacklist) BlacklistAllUserTokens(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("auth:session:%s:*", userID)
	iter := b.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		sessionKey := iter.Val()
		jtis, err := b.rdb.SMembers(ctx, sessionKey).Result()
		if err != nil {
			continue
		}
		for _, jti := range jtis {
			b.rdb.Set(ctx, key(jti), "1", 24*time.Hour)
		}
		b.rdb.Del(ctx, sessionKey)
	}
	return iter.Err()
}

func key(jti string) string {
	return fmt.Sprintf("auth:blacklist:%s", jti)
}
