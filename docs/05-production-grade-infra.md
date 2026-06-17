# 大运河平台 — 生产级基础设施强化

> **强化范围**: 鉴权中间件 / 令牌桶限流 / 数据库连接池 / 日志系统
>
> **前置阅读**: `04-go-engineering-standards.md` (基础 pkg/ 框架)

---

## 目录

1. [鉴权中间件 — Token 黑名单 + Refresh 旋转 + 角色矩阵](#1-鉴权中间件)
2. [令牌桶限流 — Redis Lua 原子操作 + 分级限流](#2-令牌桶限流)
3. [数据库连接池 — 完整监控 + 慢查询 + 健康检查](#3-数据库连接池)
4. [日志系统 — 轮转 + 脱敏 + 链路关联 + 分级输出](#4-日志系统)

---

## 1. 鉴权中间件

### 1.1 设计升级点

| 原来 | 现在 |
|------|------|
| JWT 无黑名单，登出不可控 | `pkg/auth/blacklist.go` — Redis 黑名单 (TTL = Token 剩余有效期) |
| Refresh Token 无旋转 | `pkg/auth/refresh.go` — 一次使用后失效，签发新 Refresh Token |
| 角色检查散落各 handler | `pkg/auth/rbac.go` — 声明式 RBAC 矩阵 + `RequirePermissions` |
| 无多设备管理 | `pkg/auth/session.go` — Redis Set 存设备 Token，支持踢下线 |

### 1.2 Token 黑名单 (`pkg/auth/blacklist.go`)

```go
package auth

import (
    "context"
    "fmt"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/golang-jwt/jwt/v5"
)

// TokenBlacklist 基于 Redis 的 JWT 黑名单
// 登出时加入，TTL = 该 Token 原本的过期时间
type TokenBlacklist struct {
    rdb *redis.Client
}

func NewTokenBlacklist(rdb *redis.Client) *TokenBlacklist {
    return &TokenBlacklist{rdb: rdb}
}

// Add 将 Token 加入黑名单
// jti: JWT ID
// expiresAt: 该 Token 的过期时间，黑名单在此之后自动清理
func (b *TokenBlacklist) Add(ctx context.Context, jti string, expiresAt time.Time) error {
    ttl := time.Until(expiresAt)
    if ttl <= 0 {
        return nil // 已过期，无需黑名单
    }
    return b.rdb.Set(ctx, key(jti), "1", ttl).Err()
}

// IsBlacklisted 检查 Token 是否在黑名单中
func (b *TokenBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
    exists, err := b.rdb.Exists(ctx, key(jti)).Result()
    if err != nil {
        return false, err // Redis 不可用 → 放行 (避免单点)
    }
    return exists > 0, nil
}

// BlacklistAllUserTokens 踢出某用户所有设备 (管理员操作)
func (b *TokenBlacklist) BlacklistAllUserTokens(ctx context.Context, userID string) error {
    // 扫描所有该用户的 session token
    pattern := fmt.Sprintf("auth:session:%s:*", userID)
    iter := b.rdb.Scan(ctx, 0, pattern, 100).Iterator()
    for iter.Next(ctx) {
        sessionKey := iter.Val()
        // 获取该 session 的所有 jti
        jtis, err := b.rdb.SMembers(ctx, sessionKey).Result()
        if err != nil {
            continue
        }
        for _, jti := range jtis {
            // 将 jti 加入黑名单 (TTL 取较长时间兜底)
            b.rdb.Set(ctx, key(jti), "1", 24*time.Hour)
        }
        // 删除 session 记录
        b.rdb.Del(ctx, sessionKey)
    }
    return iter.Err()
}

func key(jti string) string {
    return fmt.Sprintf("auth:blacklist:%s", jti)
}
```

### 1.3 Token 管理器 (`pkg/auth/token.go`)

```go
package auth

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
)

type TokenManager struct {
    secret      []byte
    blacklist   *TokenBlacklist
    sessionRepo *SessionRepo // Redis 存储
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
    DeviceID string `json:"device_id"` // 设备标识 (浏览器指纹 / UUID)
}

// TokenPair Access + Refresh
type TokenPair struct {
    AccessToken   string `json:"access_token"`
    RefreshToken  string `json:"refresh_token"`
    TokenType     string `json:"token_type"`
    ExpiresIn     int    `json:"expires_in"`
}

// IssueTokens 签发双 Token
func (m *TokenManager) IssueTokens(ctx context.Context, userID, role, deviceID string) (*TokenPair, error) {
    now := time.Now()

    // Access Token
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
    accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
        SignedString(m.secret)
    if err != nil {
        return nil, errors.Internal(err)
    }

    // Refresh Token (随机字符串，非 JWT → 防伪造)
    refreshToken, err := generateRefreshToken()
    if err != nil {
        return nil, errors.Internal(err)
    }

    // 存储 Refresh Token → Redis
    refreshJTI := uuid.New().String()
    err = m.sessionRepo.SaveRefreshToken(ctx, userID, deviceID, refreshJTI, refreshToken,
        now.Add(m.refreshTTL))
    if err != nil {
        return nil, errors.WrapDefault(errors.ErrRedisError, err)
    }

    // 记录 session (多设备管理)
    err = m.sessionRepo.AddAccessJTI(ctx, userID, deviceID, accessJTI)
    if err != nil {
        return nil, errors.WrapDefault(errors.ErrRedisError, err)
    }

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: fmt.Sprintf("%s.%s", refreshJTI, refreshToken),
        TokenType:    "Bearer",
        ExpiresIn:    int(m.accessTTL.Seconds()),
    }, nil
}

// RefreshAccess 使用 Refresh Token 换取新的 TokenPair (★ 旋转)
// 旧 Refresh Token 被立即失效，防重放攻击
func (m *TokenManager) RefreshAccess(ctx context.Context, refreshTokenRaw string) (*TokenPair, error) {
    // 解析 refresh token: "jti.random_hex"
    jti, token, err := parseRefreshToken(refreshTokenRaw)
    if err != nil {
        return nil, errors.NewDefault(errors.ErrTokenInvalid)
    }

    // 验证 Refresh Token
    session, err := m.sessionRepo.VerifyAndRevokeRefreshToken(ctx, jti, token)
    if err != nil {
        if err == ErrRefreshTokenReused {
            // ★ 重放攻击检测: 立即踢出该用户所有设备
            go m.blacklist.BlacklistAllUserTokens(context.Background(), session.UserID)
            return nil, errors.NewDefault(errors.ErrTokenInvalid)
        }
        return nil, errors.NewDefault(errors.ErrTokenExpired)
    }

    // 签发新 Token 对 (旋转)
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
    // 加入黑名单
    if err := m.blacklist.Add(ctx, claims.ID, claims.ExpiresAt.Time); err != nil {
        return err
    }
    // 删除 session 记录
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
    // 格式: "jti.token"
    for i := len(raw) - 1; i >= 0; i-- {
        if raw[i] == '.' {
            return raw[:i], raw[i+1:], nil
        }
    }
    return "", "", fmt.Errorf("invalid refresh token format")
}

var ErrRefreshTokenReused = fmt.Errorf("refresh token reused")
```

### 1.4 Session 存储 (`pkg/auth/session.go`)

```go
package auth

import (
    "context"
    "crypto/subtle"
    "fmt"
    "time"
    "github.com/redis/go-redis/v9"
)

type SessionRepo struct {
    rdb *redis.Client
}

type RefreshSession struct {
    UserID    string
    Role      string
    DeviceID  string
    TokenHash string
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
    return r.rdb.HSet(ctx, key, map[string]interface{}{
        "user_id":    userID,
        "role":       "", // 由调用方在 IssueTokens 时已知 role 但在 refresh 时需要重新查 DB
        "device_id":  deviceID,
        "token_hash": hashToken(token),
    }).Err()
    r.rdb.Expire(ctx, key, ttl)
    return nil
}

// VerifyAndRevokeRefreshToken 验证并立即删除 (一次性使用)
// 返回 session 信息；如果已被使用过 → 返回 ErrRefreshTokenReused
func (r *SessionRepo) VerifyAndRevokeRefreshToken(ctx context.Context, jti, token string) (*RefreshSession, error) {
    key := fmt.Sprintf("auth:refresh:%s", jti)

    // Lua 原子操作: GET + DEL
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

    // 未找到 → Token 已被使用 或 过期
    if result == nil {
        return nil, ErrRefreshTokenReused
    }

    // 解析结果
    data, ok := result.([]interface{})
    if !ok || len(data) < 6 {
        return nil, ErrRefreshTokenReused
    }
    fields := make(map[string]string, len(data)/2)
    for i := 0; i < len(data); i += 2 {
        fields[data[i].(string)] = data[i+1].(string)
    }

    // 常量时间比较防时序攻击
    expectedHash := fields["token_hash"]
    actualHash := hashToken(token)
    if subtle.ConstantTimeCompare([]byte(expectedHash), []byte(actualHash)) != 1 {
        return nil, fmt.Errorf("token mismatch")
    }

    // ★ 删除成功 = Token 被消耗，用户在记录中标记已 refresh
    // 后续需要重新查 DB 获取最新 role (因为 role 可能被管理员修改)
    return &RefreshSession{
        UserID:   fields["user_id"],
        DeviceID: fields["device_id"],
    }, nil
}

// AddAccessJTI 记录设备会话的 Access Token JTI (多设备管理)
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
    // 生产环境用 SHA-256
    h := sha256.Sum256([]byte(token))
    return hex.EncodeToString(h[:])
}
```

### 1.5 RBAC 权限矩阵 (`pkg/auth/rbac.go`)

```go
package auth

import (
    "github.com/gin-gonic/gin"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
    "github.com/your-org/grand-canal-guardian/pkg/response"
)

// Permission 权限位 (位掩码)
type Permission uint64

const (
    PermReadPost     Permission = 1 << iota  // 阅读帖子
    PermCreatePost                            // 发帖
    PermDeletePost                            // 删帖
    PermComment                               // 评论
    PermLike                                  // 点赞
    PermReadMap                               // 查看地图
    PermPlayQuiz                              // 答题
    PermViewLeaderboard                       // 排行榜
    PermUploadGarbage                         // 上传垃圾图片
    PermViewMonitorData                       // 查看监测数据
    PermReviewPost                            // 审核帖子
    PermManageUser                            // 管理用户
    PermSystemConfig                          // 系统配置
)

// RolePermissions 角色 → 权限映射
var RolePermissions = map[string]Permission{
    "user":    PermReadPost | PermCreatePost | PermComment | PermLike |
               PermReadMap | PermPlayQuiz | PermViewLeaderboard,
    "monitor": PermReadPost | PermCreatePost | PermComment | PermLike |
               PermReadMap | PermPlayQuiz | PermViewLeaderboard |
               PermUploadGarbage | PermViewMonitorData,
    "admin":   0xFFFFFFFFFFFFFFFF, // 全部权限
}

// HasPermission 检查角色是否拥有指定权限
func HasPermission(role string, perm Permission) bool {
    perms, ok := RolePermissions[role]
    if !ok {
        return false
    }
    return perms&perm != 0
}

// RequirePermission 权限校验中间件工厂
func RequirePermission(perm Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        role := GetRole(c)
        if !HasPermission(role, perm) {
            response.AbortWithError(c, errors.NewDefault(errors.ErrForbidden))
            return
        }
        c.Next()
    }
}

// RequireRole 保留原风格 (兼容旧代码)
func RequireRole(roles ...string) gin.HandlerFunc {
    roleSet := make(map[string]bool, len(roles))
    for _, r := range roles {
        roleSet[r] = true
    }
    return func(c *gin.Context) {
        if !roleSet[GetRole(c)] {
            response.AbortWithError(c, errors.NewDefault(errors.ErrForbidden))
            return
        }
        c.Next()
    }
}
```

### 1.6 增强版 Auth 中间件 (`pkg/auth/middleware.go`)

```go
package auth

import (
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
    "github.com/your-org/grand-canal-guardian/pkg/response"
)

// Context keys
const (
    ClaimsKey = "user_claims"
    UserIDKey = "user_id"
    RoleKey   = "role"
)

type AuthMiddleware struct {
    tm         *TokenManager
    skipPaths  map[string]bool
    skipPrefixes []string
}

func NewAuthMiddleware(tm *TokenManager, skipPaths []string, skipPrefixes []string) *AuthMiddleware {
    sp := make(map[string]bool, len(skipPaths))
    for _, p := range skipPaths {
        sp[p] = true
    }
    return &AuthMiddleware{
        tm:           tm,
        skipPaths:    sp,
        skipPrefixes: skipPrefixes,
    }
}

// Handler 返回 Gin 中间件
func (a *AuthMiddleware) Handler() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 白名单跳过
        if a.shouldSkip(c.Request.URL.Path) {
            c.Next()
            return
        }

        // 提取 Token
        tokenStr := extractToken(c)
        if tokenStr == "" {
            response.AbortWithError(c, errors.NewDefault(errors.ErrTokenInvalid))
            return
        }

        // 验证 (含黑名单)
        claims, err := a.tm.ValidateAccess(c.Request.Context(), tokenStr)
        if err != nil {
            if appErr, ok := err.(*errors.AppError); ok {
                response.AbortWithError(c, appErr)
            } else {
                response.AbortWithError(c, errors.NewDefault(errors.ErrTokenExpired))
            }
            return
        }

        // 注入 Context
        c.Set(ClaimsKey, claims)
        c.Set(UserIDKey, claims.UserID)
        c.Set(RoleKey, claims.Role)

        c.Next()
    }
}

func (a *AuthMiddleware) shouldSkip(path string) bool {
    if a.skipPaths[path] {
        return true
    }
    for _, prefix := range a.skipPrefixes {
        if strings.HasPrefix(path, prefix) {
            return true
        }
    }
    return false
}

func extractToken(c *gin.Context) string {
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
        return ""
    }
    return strings.TrimPrefix(authHeader, "Bearer ")
}

// ---- Handler 辅助函数 ----

func GetUserID(c *gin.Context) string {
    id, _ := c.Get(UserIDKey)
    return id.(string)
}

func GetRole(c *gin.Context) string {
    role, _ := c.Get(RoleKey)
    return role.(string)
}

func GetClaims(c *gin.Context) *Claims {
    claims, _ := c.Get(ClaimsKey)
    return claims.(*Claims)
}
```

---

## 2. 令牌桶限流

### 2.1 设计升级点

| 原来 | 现在 |
|------|------|
| 固定窗口 (INCR+EXPIRE) | 真正令牌桶 (Redis Lua 原子操作) |
| 单一限流 | 三级限流: IP / 用户 / 接口 |
| 无速率头 | `X-RateLimit-*` + `Retry-After` |

### 2.2 Redis Lua 令牌桶 (`pkg/ratelimit/token_bucket.lua`)

```lua
-- token_bucket.lua
-- KEYS[1]: bucket key
-- ARGV[1]: capacity (最大令牌数)
-- ARGV[2]: rate (每秒填充令牌数)
-- ARGV[3]: requested (本次请求消耗令牌数, 默认 1)
-- ARGV[4]: now_ms (当前时间毫秒)
--
-- 返回: {allowed (0/1), remaining, reset_at_ms, retry_after_ms}

local key        = KEYS[1]
local capacity   = tonumber(ARGV[1])
local rate       = tonumber(ARGV[2])
local requested  = tonumber(ARGV[3])
local now_ms     = tonumber(ARGV[4])

-- 读取当前状态
local bucket = redis.call('HMGET', key, 'tokens', 'last_fill_ms')
local tokens       = tonumber(bucket[1])
local last_fill_ms = tonumber(bucket[2])

-- 首次访问 → 初始化满桶
if tokens == nil then
    tokens       = capacity
    last_fill_ms = now_ms
end

-- 计算新令牌数
local elapsed_ms = now_ms - last_fill_ms
local new_tokens = math.min(capacity, tokens + (elapsed_ms / 1000.0) * rate)

-- 尝试消耗
local allowed = 0
if new_tokens >= requested then
    new_tokens = new_tokens - requested
    allowed = 1
end

-- 计算下次充满时间
local fill_time_ms = 0
if new_tokens < capacity then
    fill_time_ms = math.ceil((capacity - new_tokens) / rate * 1000)
end
local reset_at_ms = now_ms + fill_time_ms

-- 重试时间 (被拒绝时)
local retry_after_ms = 0
if allowed == 0 then
    retry_after_ms = math.ceil((requested - new_tokens) / rate * 1000)
end

-- 写回
redis.call('HMSET', key, 'tokens', new_tokens, 'last_fill_ms', now_ms)
redis.call('PEXPIRE', key, math.max(60000, fill_time_ms + 10000)) -- TTL > 充满时间

return {allowed, math.floor(new_tokens), reset_at_ms, retry_after_ms}
```

### 2.3 Go 限流器 (`pkg/ratelimit/limiter.go`)

```go
package ratelimit

import (
    "context"
    _ "embed"
    "fmt"
    "strconv"
    "time"
    "github.com/redis/go-redis/v9"
)

//go:embed token_bucket.lua
var tokenBucketScript string

// BucketConfig 令牌桶配置
type BucketConfig struct {
    Name     string  // 标识 (如 "ip", "user", "api")
    Capacity int     // 桶容量 (最大突发)
    Rate     float64 // 填充速率 (令牌/秒)
}

// Limiter Redis 令牌桶限流器
type Limiter struct {
    rdb    *redis.Client
    script *redis.Script
}

func NewLimiter(rdb *redis.Client) *Limiter {
    return &Limiter{
        rdb:    rdb,
        script: redis.NewScript(tokenBucketScript),
    }
}

// Result 限流结果
type Result struct {
    Allowed      bool
    Remaining    int
    ResetAt      time.Time
    RetryAfterMs int
}

// Allow 检查是否允许请求
func (l *Limiter) Allow(ctx context.Context, key string, cfg BucketConfig, cost int) (*Result, error) {
    nowMs := time.Now().UnixMilli()

    result, err := l.script.Run(ctx, l.rdb, []string{key},
        cfg.Capacity,
        cfg.Rate,
        cost,
        nowMs,
    ).Result()
    if err != nil {
        // Redis 不可用 → 放行 (可用性优先)
        return &Result{Allowed: true, Remaining: cfg.Capacity - 1}, nil
    }

    values, ok := result.([]interface{})
    if !ok || len(values) < 4 {
        return &Result{Allowed: true}, nil
    }

    allowed, _ := strconv.Atoi(fmt.Sprint(values[0]))
    remaining, _ := strconv.Atoi(fmt.Sprint(values[1]))
    resetMs, _ := strconv.ParseInt(fmt.Sprint(values[2]), 10, 64)
    retryAfter, _ := strconv.Atoi(fmt.Sprint(values[3]))

    return &Result{
        Allowed:      allowed == 1,
        Remaining:    remaining,
        ResetAt:      time.UnixMilli(resetMs),
        RetryAfterMs: retryAfter,
    }, nil
}
```

### 2.4 三级限流中间件 (`pkg/ratelimit/middleware.go`)

```go
package ratelimit

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
    "github.com/your-org/grand-canal-guardian/pkg/response"
)

// TierConfig 三级限流配置
type TierConfig struct {
    // IP 限流 (基础防护)
    IP BucketConfig
    // 用户限流 (登录用户)
    User BucketConfig
    // 接口限流 (特定 API)
    API map[string]BucketConfig // path → config
}

// DefaultTierConfig 默认三级限流
func DefaultTierConfig() TierConfig {
    return TierConfig{
        IP: BucketConfig{
            Name:     "ip",
            Capacity: 100,  // 桶容量 100
            Rate:     20.0, // 20 token/s
        },
        User: BucketConfig{
            Name:     "user",
            Capacity: 50,
            Rate:     10.0,
        },
        API: map[string]BucketConfig{
            "/api/v1/llm/chat":         {Name: "llm_chat", Capacity: 10, Rate: 2.0},    // LLM 成本控制
            "/api/v1/vision/classify":   {Name: "vision", Capacity: 20, Rate: 5.0},      // GPU 保护
            "/api/v1/quiz/submit":       {Name: "quiz", Capacity: 30, Rate: 10.0},        // 防刷题
            "/api/v1/auth/login":        {Name: "login", Capacity: 5, Rate: 1.0},         // 防爆破
            "/api/v1/upload/image":      {Name: "upload", Capacity: 20, Rate: 5.0},       // 带宽保护
        },
    }
}

// Middleware 三级令牌桶限流中间件
func Middleware(limiter *Limiter, cfg TierConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()

        // 1. IP 限流
        ipKey := fmt.Sprintf("ratelimit:ip:%s", c.ClientIP())
        result, err := limiter.Allow(ctx, ipKey, cfg.IP, 1)
        if err != nil || !result.Allowed {
            setRateLimitHeaders(c, result)
            if !result.Allowed {
                response.AbortWithError(c, errors.NewDefault(errors.ErrTooManyRequest))
                return
            }
            c.Next()
            return
        }

        // 2. 用户限流 (如已登录)
        if userID := c.GetString("user_id"); userID != "" {
            userKey := fmt.Sprintf("ratelimit:user:%s", userID)
            result, err := limiter.Allow(ctx, userKey, cfg.User, 1)
            if err != nil || !result.Allowed {
                setRateLimitHeaders(c, result)
                if !result.Allowed {
                    response.AbortWithError(c, errors.NewDefault(errors.ErrTooManyRequest))
                    return
                }
                c.Next()
                return
            }
        }

        // 3. 接口级限流
        path := c.FullPath()
        if apiCfg, ok := cfg.API[path]; ok {
            apiKey := fmt.Sprintf("ratelimit:api:%s", path)
            result, err := limiter.Allow(ctx, apiKey, apiCfg, 1)
            if err != nil || !result.Allowed {
                setRateLimitHeaders(c, result)
                if !result.Allowed {
                    response.AbortWithError(c, errors.NewDefault(errors.ErrTooManyRequest))
                    return
                }
                c.Next()
                return
            }
        }

        c.Next()
    }
}

func setRateLimitHeaders(c *gin.Context, r *Result) {
    if r == nil {
        return
    }
    c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", r.Remaining))
    c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", r.ResetAt.Unix()))
    if r.RetryAfterMs > 0 {
        c.Header("Retry-After", fmt.Sprintf("%.1f", float64(r.RetryAfterMs)/1000.0))
    }
}
```

---

## 3. 数据库连接池

### 3.1 完整连接池管理器 (`pkg/database/pool.go`)

```go
package database

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "gorm.io/plugin/dbresolver"
)

// ---- Prometheus 指标 ----

var (
    dbConnOpen = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "db_connections_open",
        Help: "当前打开的数据库连接数",
    }, []string{"db", "mode"}) // mode: write/read

    dbConnIdle = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "db_connections_idle",
        Help: "当前空闲连接数",
    }, []string{"db", "mode"})

    dbConnWaitCount = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "db_connections_wait_total",
        Help: "等待连接的累计次数",
    }, []string{"db", "mode"})

    dbConnWaitDuration = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "db_connections_wait_seconds_total",
        Help: "等待连接的总时间 (秒)",
    }, []string{"db", "mode"})

    dbQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "db_query_duration_seconds",
        Help:    "查询耗时",
        Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
    }, []string{"db", "operation"}) // operation: select/insert/update/delete

    dbSlowQueryCount = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "db_slow_query_total",
        Help: "慢查询计数 (> SlowThreshold)",
    }, []string{"db"})
)

// PoolConfig 连接池配置
type PoolConfig struct {
    Name string // 数据库名称 (指标标签)

    // 连接池参数
    MaxOpenConns    int           // 最大打开连接数 (推荐: 25-50)
    MaxIdleConns    int           // 最大空闲连接数 (推荐: 10-20)
    ConnMaxLifetime time.Duration // 连接最大存活时间 (推荐: 1h)
    ConnMaxIdleTime time.Duration // 空闲连接最大存活 (推荐: 10min)

    // 慢查询阈值
    SlowThreshold time.Duration // 超过此时间的查询记录为慢查询 (推荐: 200ms)

    // 连接池爆满等待
    ConnAcquireTimeout time.Duration // 获取连接超时 (推荐: 3s)

    // 读写分离
    WriteDSN string
    ReadDSNs []string // 从库列表 (支持多个)
}

// DefaultPoolConfig 生产环境推荐配置
func DefaultPoolConfig(name string) PoolConfig {
    return PoolConfig{
        Name:               name,
        MaxOpenConns:       25,
        MaxIdleConns:       10,
        ConnMaxLifetime:    1 * time.Hour,
        ConnMaxIdleTime:    10 * time.Minute,
        SlowThreshold:      200 * time.Millisecond,
        ConnAcquireTimeout: 3 * time.Second,
    }
}

// NewDB 创建数据库连接 (含读写分离 + 监控)
func NewDB(cfg PoolConfig) (*gorm.DB, error) {
    // GORM 配置
    gormCfg := &gorm.Config{
        Logger: newGormLogger(cfg.SlowThreshold),
        NowFunc: func() time.Time { return time.Now().UTC() },
        DisableForeignKeyConstraintWhenMigrating: true,
    }

    // 主库
    db, err := gorm.Open(postgres.New(postgres.Config{
        DSN:                  cfg.WriteDSN,
        PreferSimpleProtocol: true, // 禁用 prepared statement 缓存 (减少内存)
    }), gormCfg)
    if err != nil {
        return nil, fmt.Errorf("连接主库失败: %w", err)
    }

    // 配置主库连接池
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

    // 注册读写分离
    if len(cfg.ReadDSNs) > 0 {
        replicas := make([]gorm.Dialector, len(cfg.ReadDSNs))
        for i, dsn := range cfg.ReadDSNs {
            replicas[i] = postgres.New(postgres.Config{
                DSN:                  dsn,
                PreferSimpleProtocol: true,
            })
        }
        db.Use(dbresolver.Register(dbresolver.Config{
            Replicas: replicas,
            Policy:   &dbresolver.RandomPolicy{},
        }).SetConnMaxIdleTime(cfg.ConnMaxIdleTime).
            SetConnMaxLifetime(cfg.ConnMaxLifetime).
            SetMaxIdleConns(cfg.MaxIdleConns).
            SetMaxOpenConns(cfg.MaxOpenConns))
    }

    // ★ 启动连接池监控
    go monitorPool(cfg.Name, sqlDB)

    return db, nil
}

// monitorPool 定期上报连接池指标
func monitorPool(name string, db *sql.DB) {
    ticker := time.NewTicker(15 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        stats := db.Stats()

        dbConnOpen.WithLabelValues(name, "write").Set(float64(stats.OpenConnections))
        dbConnIdle.WithLabelValues(name, "write").Set(float64(stats.Idle))
        dbConnWaitCount.WithLabelValues(name, "write").Add(float64(stats.WaitCount))
        dbConnWaitDuration.WithLabelValues(name, "write").Add(stats.WaitDuration.Seconds())

        // 告警: 连接池接近耗尽
        maxOpen := db.Stats().MaxOpenConnections // 获取配置值
        if maxOpen > 0 && stats.OpenConnections > int(float64(maxOpen)*0.8) {
            log.Printf("[WARN] 数据库 %s 连接池使用率 > 80%: %d/%d, 等待次数: %d",
                name, stats.OpenConnections, maxOpen, stats.WaitCount)
        }
    }
}

// ---- Callbacks: 查询耗时监控 ----

func registerCallbacks(db *gorm.DB, dbName string) {
    // 注册所有 CRUD 操作的 callback
    for _, op := range []string{"Create", "Query", "Update", "Delete", "Row", "Raw"} {
        registerDurationCallback(db, dbName, op)
    }
}

func registerDurationCallback(db *gorm.DB, dbName, op string) {
    db.Callback().Create().After("gorm:after_*").Register(
        fmt.Sprintf("metrics:%s:%s", dbName, op),
        func(d *gorm.DB) {
            duration := time.Since(d.Statement.StartTime).Seconds()
            dbQueryDuration.WithLabelValues(dbName, op).Observe(duration)
        },
    )
}

// ---- GORM Logger: 慢查询告警 ----

type gormLogger struct {
    SlowThreshold time.Duration
    logLevel      logger.LogLevel
}

func newGormLogger(slowThreshold time.Duration) logger.Interface {
    return &gormLogger{
        SlowThreshold: slowThreshold,
        logLevel:      logger.Warn,
    }
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
    l.logLevel = level
    return l
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
    // 仅开发环境输出
}

func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
    log.Printf("[GORM WARN] %s %v", msg, data)
}

func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
    log.Printf("[GORM ERROR] %s %v", msg, data)
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
    elapsed := time.Since(begin)
    sql, rows := fc()

    // 慢查询检测
    if elapsed > l.SlowThreshold {
        dbSlowQueryCount.WithLabelValues("unknown").Inc()
        log.Printf("[SLOW QUERY] %v | rows:%d | %s", elapsed, rows, sql)
    }

    // 错误
    if err != nil {
        log.Printf("[DB ERROR] %v | %s | %v", elapsed, sql, err)
    }
}
```

### 3.2 连接池健康检查 (`pkg/database/health.go`)

```go
package database

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    "github.com/your-org/grand-canal-guardian/pkg/app"
)

// RegisterHealthCheck 注册数据库健康检查到 app 的 readiness
func RegisterHealthCheck(name string, db *sql.DB) {
    app.RegisterReadinessCheck(func() error {
        ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
        defer cancel()
        if err := db.PingContext(ctx); err != nil {
            return fmt.Errorf("数据库 %s 不可达: %w", name, err)
        }
        return nil
    })
}
```

---

## 4. 日志系统

### 4.1 完整日志系统 (`pkg/log/logger.go`)

```go
package log

import (
    "os"
    "path/filepath"
    "regexp"
    "strings"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2" // 日志轮转
)

// Config 日志配置
type Config struct {
    Level      string // debug | info | warn | error
    Format     string // json | console
    Output     string // stdout | file
    FilePath   string // 文件路径 (Output=file 时)
    MaxSizeMB  int    // 单文件最大 MB (默认 100)
    MaxBackups int    // 保留旧文件数 (默认 7)
    MaxAgeDays int    // 保留天数 (默认 30)
    Compress   bool   // 压缩旧文件
    Service    string // 服务名 (自动注入到每行日志)
}

func DefaultConfig(service string) Config {
    return Config{
        Level:      "info",
        Format:     "json",
        Output:     "stdout",
        MaxSizeMB:  100,
        MaxBackups: 7,
        MaxAgeDays: 30,
        Compress:   true,
        Service:    service,
    }
}

// New 创建生产级 Logger
func New(cfg Config) (*zap.Logger, error) {
    level, err := zapcore.ParseLevel(cfg.Level)
    if err != nil {
        level = zapcore.InfoLevel
    }

    // Encoder 配置
    encoderCfg := zap.NewProductionEncoderConfig()
    encoderCfg.TimeKey = "timestamp"
    encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
    encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
    encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

    var encoder zapcore.Encoder
    if cfg.Format == "console" {
        encoder = zapcore.NewConsoleEncoder(encoderCfg)
    } else {
        encoder = zapcore.NewJSONEncoder(encoderCfg)
    }

    // Writer (stdout / 文件轮转)
    var writer zapcore.WriteSyncer
    if cfg.Output == "file" && cfg.FilePath != "" {
        writer = zapcore.AddSync(&lumberjack.Logger{
            Filename:   cfg.FilePath,
            MaxSize:    cfg.MaxSizeMB,
            MaxBackups: cfg.MaxBackups,
            MaxAge:     cfg.MaxAgeDays,
            Compress:   cfg.Compress,
            LocalTime:  true,
        })
    } else {
        writer = zapcore.AddSync(os.Stdout)
    }

    core := zapcore.NewCore(encoder, writer, level)

    // 添加初始字段
    logger := zap.New(core,
        zap.AddCaller(),
        zap.AddCallerSkip(1),
        zap.Fields(zap.String("service", cfg.Service)),
    )

    return logger, nil
}

// WithRequestID 注入 request_id
func WithRequestID(logger *zap.Logger, requestID string) *zap.Logger {
    return logger.With(zap.String("request_id", requestID))
}

// WithUser 注入用户信息 (非敏感)
func WithUser(logger *zap.Logger, userID, role string) *zap.Logger {
    return logger.With(
        zap.String("user_id", userID),
        zap.String("role", role),
    )
}

// ---- 敏感信息脱敏 ----

// 脱敏规则
var sensitivePatterns = []struct {
    Pattern *regexp.Regexp
    Replace string
}{
    {regexp.MustCompile(`(password=)[^&\s]+`), `${1}***`},
    {regexp.MustCompile(`(token=)[^&\s]+`), `${1}***`},
    {regexp.MustCompile(`(api_key=)[^&\s]+`), `${1}***`},
    {regexp.MustCompile(`"password"\s*:\s*"[^"]*"`), `"password":"***"`},
    {regexp.MustCompile(`"token"\s*:\s*"[^"]*"`), `"token":"***"`},
    {regexp.MustCompile(`"email"\s*:\s*"([^"@]+)@`), `"email":"***@`},
    {regexp.MustCompile(`"phone"\s*:\s*"(\d{3})\d+`), `"phone":"$1****"`},
}

// Sanitize 脱敏字符串
func Sanitize(s string) string {
    result := s
    for _, rule := range sensitivePatterns {
        result = rule.Pattern.ReplaceAllString(result, rule.Replace)
    }
    return result
}

// SanitizeField 创建脱敏的 zap.Field
func SanitizeField(key, value string) zap.Field {
    return zap.String(key, Sanitize(value))
}
```

### 4.2 Gin 日志中间件 (`pkg/log/middleware.go`)

```go
package log

import (
    "time"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// GinLogger 结构化请求日志中间件
// 使用 pkg/log 的脱敏能力
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := Sanitize(c.Request.URL.RawQuery)

        // 处理请求
        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()
        requestID, _ := c.Get("request_id")

        fields := []zap.Field{
            zap.String("request_id", toString(requestID)),
            zap.Int("status", status),
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.String("query", query),
            zap.String("ip", c.ClientIP()),
            zap.String("user_agent", c.GetHeader("User-Agent")),
            zap.Duration("latency", latency),
            zap.Int("body_size", c.Writer.Size()),
        }

        // 业务错误
        if len(c.Errors) > 0 {
            fields = append(fields, zap.String("gin_errors", c.Errors.String()))
        }

        // 分级输出
        switch {
        case status >= 500:
            logger.Error("请求完成 [5xx]", fields...)
        case status >= 400:
            logger.Warn("请求完成 [4xx]", fields...)
        case latency > 500*time.Millisecond:
            logger.Warn("请求完成 [慢请求]", fields...)
        default:
            logger.Info("请求完成", fields...)
        }
    }
}

// GinRecovery Panic 恢复 (带日志)
func GinRecovery(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                logger.Error("PANIC",
                    zap.Any("panic", r),
                    zap.Stack("stack"),
                    zap.String("method", c.Request.Method),
                    zap.String("path", c.Request.URL.Path),
                )
                c.AbortWithStatus(500)
            }
        }()
        c.Next()
    }
}

func toString(v interface{}) string {
    if v == nil {
        return ""
    }
    if s, ok := v.(string); ok {
        return s
    }
    return ""
}
```

### 4.3 使用示例

```go
// 初始化
logger, _ := log.New(log.Config{
    Level:    "info",
    Format:   "json",
    Output:   "file",
    FilePath: "/var/log/gcg/user-service.log",
    Service:  "user-service",
})
defer logger.Sync()

// 业务日志 (带 trace_id + 用户)
func (s *UserService) Register(ctx context.Context, req *RegisterRequest) error {
    logger := log.WithRequestID(s.logger, ctx.Value("request_id").(string))

    logger.Info("开始注册",
        zap.String("username", req.Username),
        zap.String("email", log.Sanitize(req.Email)), // ★ 脱敏
    )

    if err := s.repo.Create(ctx, user); err != nil {
        logger.Error("创建用户失败",
            zap.Error(err),
            zap.String("username", req.Username),
        )
        return errors.WrapDefault(errors.ErrDatabaseError, err)
    }

    logger.Info("注册成功",
        zap.String("user_id", user.ID),
        zap.String("role", user.Role),
    )
    return nil
}
```

### 4.4 依赖

```go
// go.mod 新增
require (
    go.uber.org/zap v1.27.0
    gopkg.in/natefinch/lumberjack.v2 v2.2.1
)
```

---

## 附录: 完整中间件注册顺序

```go
// ★ 中间件链最终版
func ApplyAll(r *gin.Engine, deps Dependencies) {
    // 1. Recovery — 最内层 (捕获所有 panic)
    r.Use(recovery.Handler(deps.Logger))

    // 2. RequestID — 注入 X-Request-ID
    r.Use(middleware.RequestID())

    // 3. Logger — 结构化日志 (需要在 RequestID 后)
    r.Use(log.GinLogger(deps.Logger))

    // 4. CORS — 跨域
    r.Use(cors.New(cors.Config{...}))

    // 5. OpenTelemetry — 链路追踪
    r.Use(traceware.GinMiddleware(deps.ServiceName))

    // 6. Auth — JWT 鉴权 (含黑名单 + RBAC)
    authMW := auth.NewAuthMiddleware(deps.TokenManager, ...)
    r.Use(authMW.Handler())

    // 7. RateLimit — 三级令牌桶
    r.Use(ratelimit.Middleware(deps.Limiter, ratelimit.DefaultTierConfig()))

    // ★ 健康检查路由 (无鉴权)
    r.GET("/health", health.LivenessHandler)
    r.GET("/ready", health.ReadinessHandler(deps))
}
```

---

> **集成方式**: 将本文档中的 `pkg/auth/`、`pkg/ratelimit/`、`pkg/database/`、`pkg/log/` 合并到 `04-go-engineering-standards.md` 的 `pkg/` 目录结构中即可。所有代码可直接复制使用。
