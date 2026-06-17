# 大运河平台 — Go 工程化规范

> **Phase 1 已实现**: pkg/errors, pkg/response, pkg/auth, pkg/middleware, pkg/app
> **Phase 2 规划**: pkg/ratelimit (Redis Lua令牌桶), pkg/database (连接池监控), pkg/log (Zap+轮转+脱敏)

> **目标**: 零重复、低耦合、统一错误/响应、完整中间件链、自动文档生成、优雅关闭
>
> **补充**: 01-架构设计.md 和 03-开发者文档.md 的工程化缺口

---

## 目录

1. [新增项目结构 — pkg/ 公共库](#1-新增项目结构)
2. [pkg/errors — 统一错误封装](#2-pkgerrors)
3. [pkg/response — 统一响应格式](#3-pkgresponse)
4. [pkg/middleware — 完整中间件链](#4-pkgmiddleware)
5. [pkg/validator — 参数校验](#5-pkgvalidator)
6. [pkg/app — Gin 启动器](#6-pkgapp)
7. [pkg/transaction — 事务封装](#7-pkgtransaction)
8. [pkg/grpc — 跨服务错误传播](#8-pkggrpc)
9. [swaggo — OpenAPI 自动生成](#9-swaggo)
10. [完整服务示例](#10-完整服务示例)

---

## 1. 新增项目结构

```
grand-canal-guardian/
├── pkg/                           # ★ 公共库 — 所有 Go 服务共享
│   ├── errors/                    # 统一错误
│   │   ├── codes.go               #   错误码枚举
│   │   ├── app_error.go           #   AppError 结构体
│   │   └── i18n.go                #   中文错误消息
│   ├── response/                  # 统一响应
│   │   └── response.go            #   {code, message, data} 包装器
│   ├── auth/                      # ★ 鉴权 (生产级) → 详见 05-production-grade-infra.md
│   │   ├── middleware.go           #   Auth 中间件
│   │   ├── token.go               #   TokenManager (双 Token + 黑名单)
│   │   ├── blacklist.go           #   Redis Token 黑名单
│   │   ├── session.go             #   多设备 Session 管理
│   │   ├── refresh.go             #   Refresh Token 旋转
│   │   └── rbac.go                #   RBAC 位掩码权限矩阵
│   ├── ratelimit/                 # ★ 令牌桶限流 (生产级)
│   │   ├── limiter.go             #   Redis Lua 令牌桶
│   │   ├── token_bucket.lua       #   Lua 脚本 (原子操作)
│   │   ├── middleware.go           #   三级限流中间件 (IP/User/API)
│   │   └── config.go              #   分级限流配置
│   ├── database/                  # ★ 连接池 (生产级)
│   │   ├── pool.go                #   连接池管理 + Prometheus 指标
│   │   ├── health.go              #   健康检查注册
│   │   └── config.go              #   默认配置
│   ├── log/                       # ★ 日志系统 (生产级)
│   │   ├── logger.go              #   Zap 工厂 + 轮转 + 脱敏
│   │   ├── middleware.go           #   Gin 日志中间件
│   │   └── sanitize.go            #   敏感信息脱敏规则
│   ├── middleware/                 # 通用中间件
│   │   ├── chain.go               #   中间件注册顺序
│   │   ├── request_id.go          #   X-Request-ID
│   │   ├── cors.go                #   CORS
│   │   └── trace.go               #   OpenTelemetry
│   ├── validator/                 # 参数校验
│   │   └── validator.go           #   go-playground + 中文翻译
│   ├── app/                       # 启动器
│   │   ├── app.go                 #   Gin 引擎工厂 + 中间件注册
│   │   ├── shutdown.go            #   优雅关闭 (SIGINT/SIGTERM)
│   │   ├── health.go              #   /health + /ready (K8s 探针)
│   │   └── logger.go              #   Zap 工厂
│   ├── transaction/               # 事务封装
│   │   └── tx.go                  #   WithTx / WithTxResult 装饰器
│   ├── grpc/                      # gRPC 工具
│   │   ├── error_convert.go       #   AppError ↔ gRPC Status
│   │   └── interceptors.go        #   Unary 拦截器
│   └── go.mod
├── services/                      # 各业务服务
│   ├── user-service/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── model/
│   │   │   ├── config/
│   │   │   └── wire/              #   ★ 依赖注入
│   │   │       ├── wire.go        #      wire 定义
│   │   │       └── wire_gen.go    #      wire 生成
│   │   ├── docs/                  #   ★ swaggo 生成
│   │   │   ├── docs.go
│   │   │   ├── swagger.json
│   │   │   └── swagger.yaml
│   │   └── go.mod
│   └── ...
└── go.work                        # ★ Go workspace (多模块)
```

**`go.work`** — 解决 pkg/ 与各服务的模块引用:

```
go 1.22

use (
    ./pkg
    ./services/api-gateway
    ./services/user-service
    ./services/content-service
    ./services/map-service
    ./services/quiz-service
    ./services/notification-service
)
```

**`pkg/go.mod`**:
```
module github.com/your-org/grand-canal-guardian/pkg

go 1.22
```

各服务 `go.mod`:
```
require github.com/your-org/grand-canal-guardian/pkg v0.0.0
replace github.com/your-org/grand-canal-guardian/pkg => ../../pkg
```

---

## 2. pkg/errors — 统一错误封装

### 2.1 错误码枚举 (`pkg/errors/codes.go`)

```go
package errors

// ErrorCode 统一错误码 (5 位: 模块(2) + 分类(1) + 序号(2))
// 模块: 00=通用 01=用户 02=内容 03=地图 04=问答 05=LLM 06=视觉
// 分类: 0=系统 1=参数 2=业务 3=权限 4=外部依赖
type ErrorCode int

const (
    // ---- 通用 (00) ----
    ErrInternal       ErrorCode = 10001 // 服务器内部错误
    ErrBadRequest     ErrorCode = 10002 // 请求参数错误
    ErrNotFound       ErrorCode = 10003 // 资源不存在
    ErrConflict       ErrorCode = 10004 // 资源冲突
    ErrTooManyRequest ErrorCode = 10005 // 请求过于频繁

    // ---- 认证 (01) ----
    ErrUnauthorized     ErrorCode = 10101 // 未登录
    ErrTokenExpired     ErrorCode = 10102 // Token 已过期
    ErrTokenInvalid     ErrorCode = 10103 // Token 无效
    ErrForbidden        ErrorCode = 10104 // 无权限
    ErrPasswordWrong    ErrorCode = 10105 // 密码错误
    ErrUserBanned       ErrorCode = 10106 // 账号已封禁

    // ---- 用户 (02) ----
    ErrUserNotFound       ErrorCode = 10201 // 用户不存在
    ErrUsernameExists     ErrorCode = 10202 // 用户名已存在
    ErrEmailExists        ErrorCode = 10203 // 邮箱已注册
    ErrInvalidRole        ErrorCode = 10204 // 无效角色
    ErrCannotBanSelf      ErrorCode = 10205 // 不能封禁自己

    // ---- 内容 (03) ----
    ErrPostNotFound     ErrorCode = 10301 // 帖子不存在
    ErrPostNotOwner     ErrorCode = 10302 // 非帖子作者
    ErrCommentNotFound  ErrorCode = 10303 // 评论不存在
    ErrContentSensitive ErrorCode = 10304 // 内容违规

    // ---- 地图 (04) ----
    ErrPOINotFound   ErrorCode = 10401 // POI 不存在
    ErrRouteNotFound ErrorCode = 10402 // 路线不存在

    // ---- 问答 (05) ----
    ErrQuestionNotFound  ErrorCode = 10501 // 题目不存在
    ErrDuplicateAnswer   ErrorCode = 10502 // 重复提交
    ErrSessionExpired    ErrorCode = 10503 // 答题会话过期

    // ---- LLM (06) ----
    ErrLLMTimeout      ErrorCode = 10601 // LLM 超时
    ErrLLMRateLimit    ErrorCode = 10602 // LLM 限流
    ErrLLMUnavailable  ErrorCode = 10603 // LLM 服务不可用

    // ---- 视觉 (07) ----
    ErrImageTooLarge   ErrorCode = 10701 // 图片过大
    ErrImageFormat     ErrorCode = 10702 // 图片格式不支持
    ErrDetectionFailed ErrorCode = 10703 // 识别失败

    // ---- 外部依赖 (08) ----
    ErrDatabaseError   ErrorCode = 10801 // 数据库错误
    ErrRedisError      ErrorCode = 10802 // Redis 错误
    ErrRabbitMQError   ErrorCode = 10803 // 消息队列错误
    ErrMinIOError      ErrorCode = 10804 // 对象存储错误
)

// HTTPStatus 返回对应的 HTTP 状态码
func (c ErrorCode) HTTPStatus() int {
    switch {
    case c >= 10101 && c <= 10106:
        if c == ErrForbidden { return 403 }
        return 401
    case c >= 10201 && c <= 10205:
        return 400
    case c >= 10301 && c <= 10304:
        if c == ErrPostNotFound || c == ErrCommentNotFound {
            return 404
        }
        if c == ErrPostNotOwner { return 403 }
        return 400
    case c >= 10601 && c <= 10603:
        return 503
    case c >= 10801 && c <= 10804:
        return 500
    default:
        return 500
    }
}
```

### 2.2 AppError 结构体 (`pkg/errors/app_error.go`)

```go
package errors

import (
    "fmt"
)

// AppError 统一应用错误
type AppError struct {
    Code    ErrorCode `json:"code"`              // 错误码
    Message string    `json:"message"`           // 用户可读消息
    Detail  string    `json:"detail,omitempty"`  // 调试详情 (仅 dev 环境返回)
    Cause   error     `json:"-"`                 // 原始错误 (不序列化)
}

func (e *AppError) Error() string {
    if e.Detail != "" {
        return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Detail)
    }
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Cause }

// ---- 构造器 ----

func New(code ErrorCode, message string) *AppError {
    return &AppError{Code: code, Message: message}
}

func Newf(code ErrorCode, format string, args ...any) *AppError {
    return &AppError{Code: code, Message: fmt.Sprintf(format, args...)}
}

func Wrap(code ErrorCode, message string, cause error) *AppError {
    return &AppError{Code: code, Message: message, Cause: cause, Detail: cause.Error()}
}

func Wrapf(code ErrorCode, cause error, format string, args ...any) *AppError {
    return &AppError{
        Code:    code,
        Message: fmt.Sprintf(format, args...),
        Cause:   cause,
        Detail:  cause.Error(),
    }
}

// ---- 快捷构造 (使用默认中文消息) ----

func BadRequest(msg string) *AppError {
    return New(ErrBadRequest, msg)
}

func NotFound(resource string) *AppError {
    return Newf(ErrNotFound, "%s 不存在", resource)
}

func Unauthorized(msg string) *AppError {
    return New(ErrUnauthorized, msg)
}

func Forbidden(msg string) *AppError {
    return New(ErrForbidden, msg)
}

func Internal(cause error) *AppError {
    return Wrap(ErrInternal, "服务器内部错误，请稍后重试", cause)
}

func Conflict(msg string) *AppError {
    return New(ErrConflict, msg)
}
```

### 2.3 中文错误消息 (`pkg/errors/i18n.go`)

```go
package errors

// DefaultMessages 错误码 → 默认中文消息
var DefaultMessages = map[ErrorCode]string{
    ErrInternal:       "服务器内部错误",
    ErrBadRequest:     "请求参数错误",
    ErrNotFound:       "资源不存在",
    ErrConflict:       "资源冲突，请稍后重试",
    ErrTooManyRequest: "请求过于频繁，请稍后再试",

    ErrUnauthorized:  "请先登录",
    ErrTokenExpired:  "登录已过期，请重新登录",
    ErrTokenInvalid:  "Token 无效",
    ErrForbidden:     "无权执行此操作",
    ErrPasswordWrong: "用户名或密码错误",
    ErrUserBanned:    "账号已被封禁",

    ErrUserNotFound:   "用户不存在",
    ErrUsernameExists: "用户名已被注册",
    ErrEmailExists:    "邮箱已被注册",

    ErrPostNotFound:     "帖子不存在",
    ErrPostNotOwner:     "只能操作自己的帖子",
    ErrCommentNotFound:  "评论不存在",
    ErrContentSensitive: "内容包含违规信息，请修改后重试",

    ErrLLMTimeout:      "AI 响应超时，请稍后重试",
    ErrLLMRateLimit:    "AI 服务繁忙，请稍后再试",
    ErrLLMUnavailable:  "AI 服务暂时不可用",

    ErrImageTooLarge:   "图片文件过大，请压缩后重试",
    ErrImageFormat:     "不支持的图片格式",
    ErrDetectionFailed: "垃圾识别失败，请重新拍摄",
}

// GetMessage 获取错误码对应的默认消息
func (c ErrorCode) GetMessage() string {
    if msg, ok := DefaultMessages[c]; ok {
        return msg
    }
    return "未知错误"
}

// NewDefault 使用默认消息创建错误
func NewDefault(code ErrorCode) *AppError {
    return New(code, code.GetMessage())
}

// WrapDefault 使用默认消息包裹错误
func WrapDefault(code ErrorCode, cause error) *AppError {
    return &AppError{
        Code:    code,
        Message: code.GetMessage(),
        Cause:   cause,
        Detail:  cause.Error(),
    }
}
```

### 2.4 Recovery 中间件 (`pkg/middleware/recovery.go`)

```go
package middleware

import (
    "net/http"
    "runtime/debug"

    "github.com/gin-gonic/gin"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
    "github.com/your-org/grand-canal-guardian/pkg/response"
    "go.uber.org/zap"
)

// Recovery 捕获 panic，返回统一错误
// 注意: 必须放在中间件链最内层 (第一个注册)
func Recovery(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                stack := string(debug.Stack())

                logger.Error("panic recovered",
                    zap.String("request_id", GetRequestID(c)),
                    zap.String("method", c.Request.Method),
                    zap.String("path", c.Request.URL.Path),
                    zap.Any("panic", r),
                    zap.String("stack", stack),
                )

                // 区分 AppError panic 和未知 panic
                if appErr, ok := r.(*errors.AppError); ok {
                    response.Error(c, appErr)
                } else {
                    response.Error(c, errors.Internal(
                        fmt.Errorf("panic: %v", r),
                    ))
                }
                c.Abort()
            }
        }()
        c.Next()
    }
}
```

---

## 3. pkg/response — 统一响应格式

(`pkg/response/response.go`)

```go
package response

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
)

// R 统一响应体
type R struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// PageData 分页数据
type PageData struct {
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    Total      int64 `json:"total"`
    TotalPages int   `json:"total_pages"`
}

// List 带分页的列表响应
type List struct {
    Items      interface{} `json:"items"`
    Pagination PageData    `json:"pagination"`
}

// ---- 成功响应 ----

// OK 成功 (data 可选)
func OK(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, R{
        Code:    0,
        Message: "ok",
        Data:    data,
    })
}

// Created 创建成功 (201)
func Created(c *gin.Context, data interface{}) {
    c.JSON(http.StatusCreated, R{
        Code:    0,
        Message: "创建成功",
        Data:    data,
    })
}

// NoContent 成功但无返回体 (204)
func NoContent(c *gin.Context) {
    c.Status(http.StatusNoContent)
}

// OKList 列表+分页
func OKList(c *gin.Context, items interface{}, page, pageSize int, total int64) {
    totalPages := 0
    if pageSize > 0 {
        totalPages = int(total) / pageSize
        if int(total)%pageSize > 0 {
            totalPages++
        }
    }
    c.JSON(http.StatusOK, R{
        Code:    0,
        Message: "ok",
        Data: List{
            Items: items,
            Pagination: PageData{
                Page:       page,
                PageSize:   pageSize,
                Total:      total,
                TotalPages: totalPages,
            },
        },
    })
}

// OKMessage 仅消息
func OKMessage(c *gin.Context, message string) {
    c.JSON(http.StatusOK, R{
        Code:    0,
        Message: message,
    })
}

// ---- 错误响应 ----

// Error 统一错误响应 (所有 handler 唯一的错误出口)
func Error(c *gin.Context, err *errors.AppError) {
    status := err.Code.HTTPStatus()
    c.JSON(status, R{
        Code:    int(err.Code),
        Message: err.Message,
        // Detail 仅在 dev 环境返回 (通过中间件控制)
    })
    c.Abort()
}

// AbortWithError 在中间件中使用
func AbortWithError(c *gin.Context, err *errors.AppError) {
    c.AbortWithStatusJSON(err.Code.HTTPStatus(), R{
        Code:    int(err.Code),
        Message: err.Message,
    })
}
```

**Handler 使用示例（对比）**:

```go
// ❌ 之前 — 错误格式不统一
func GetUser(c *gin.Context) {
    id := c.Param("id")
    user, err := svc.GetUser(c, id)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()}) // 格式乱
        return
    }
    c.JSON(200, user) // 无 code/message 包装
}

// ✓ 现在 — 统一出口
func GetUser(c *gin.Context) {
    id := c.Param("id")
    user, err := svc.GetUser(c.Request.Context(), id)
    if err != nil {
        response.Error(c, err.(*errors.AppError))
        return
    }
    response.OK(c, user)
}

func CreatePost(c *gin.Context) {
    var req CreatePostReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, errors.BadRequest("参数格式错误")) // validator 翻译见第5节
        return
    }
    post, err := svc.CreatePost(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err.(*errors.AppError))
        return
    }
    response.Created(c, post)
}

func ListPosts(c *gin.Context) {
    page := c.GetInt("page")
    pageSize := c.GetInt("page_size")
    posts, total, err := svc.ListPosts(c.Request.Context(), page, pageSize)
    if err != nil {
        response.Error(c, err.(*errors.AppError))
        return
    }
    response.OKList(c, posts, page, pageSize, total)
}
```

---

## 4. pkg/middleware — 完整中间件链

### 4.1 中间件链注册 (`pkg/middleware/chain.go`)

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// Apply 按正确顺序注册所有中间件
// 顺序: Recovery → RequestID → Logger → CORS → Trace → Auth → RateLimit
// Recovery 最内层 (捕获所有后续中间件和 handler 的 panic)
// Logger 紧随其后 (记录 request_id)
// Auth/RateLimit 在最外层 (尽早拦截无效请求)
func Apply(r *gin.Engine, logger *zap.Logger, cfg MiddlewareConfig) {
    // 1. Recovery — 捕获所有 panic (最内层)
    r.Use(Recovery(logger))

    // 2. RequestID — 注入 trace id
    r.Use(RequestID())

    // 3. Logger — 结构化请求日志 (需要 request_id)
    r.Use(Logger(logger))

    // 4. CORS — 跨域
    if cfg.CORS.Enabled {
        r.Use(CORS(cfg.CORS))
    }

    // 5. Trace — OpenTelemetry
    if cfg.Trace.Enabled {
        r.Use(Trace(cfg.Trace.ServiceName))
    }

    // 6. Auth — JWT 鉴权 (白名单路径跳过)
    r.Use(Auth(cfg.Auth))

    // 7. RateLimit — 令牌桶限流
    if cfg.RateLimit.Enabled {
        r.Use(RateLimit(cfg.RateLimit))
    }
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
    CORS      CORSConfig
    Auth      AuthConfig
    RateLimit RateLimitConfig
    Trace     TraceConfig
}

// AuthConfig 鉴权配置
type AuthConfig struct {
    JWTSecret      string
    SkipPaths       []string // 白名单路径, 如 /api/v1/auth/login
    SkipPrefixes    []string // 白名单前缀, 如 /api/v1/public/
}
```

### 4.2 RequestID 中间件 (`pkg/middleware/request_id.go`)

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

const RequestIDKey = "X-Request-ID"
const requestIDCtxKey = "request_id"

// RequestID 注入或透传 X-Request-ID
func RequestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        rid := c.GetHeader(RequestIDKey)
        if rid == "" {
            rid = uuid.New().String()
        }
        c.Set(requestIDCtxKey, rid)
        c.Header(RequestIDKey, rid)
        c.Next()
    }
}

// GetRequestID 从 context 获取 request_id
func GetRequestID(c *gin.Context) string {
    if rid, exists := c.Get(requestIDCtxKey); exists {
        return rid.(string)
    }
    return ""
}
```

### 4.3 Logger 中间件 (`pkg/middleware/logger.go`)

```go
package middleware

import (
    "time"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// Logger 结构化请求日志
func Logger(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()

        fields := []zap.Field{
            zap.String("request_id", GetRequestID(c)),
            zap.Int("status", status),
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.String("query", query),
            zap.String("ip", c.ClientIP()),
            zap.Duration("latency", latency),
            zap.Int("body_size", c.Writer.Size()),
        }

        // 错误请求记录详细信息
        if status >= 500 {
            logger.Error("request completed", fields...)
        } else if status >= 400 {
            logger.Warn("request completed", fields...)
        } else {
            logger.Info("request completed", fields...)
        }
    }
}
```

### 4.4 CORS 中间件 (`pkg/middleware/cors.go`)

```go
package middleware

import (
    "time"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

type CORSConfig struct {
    Enabled        bool
    AllowOrigins   []string
    AllowMethods   []string
    AllowHeaders   []string
    MaxAge         time.Duration
}

func DefaultCORSConfig() CORSConfig {
    return CORSConfig{
        Enabled:      true,
        AllowOrigins: []string{"http://localhost:3000"},
        AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders: []string{
            "Origin", "Content-Type", "Authorization",
            "X-Request-ID", "X-CSRF-Token",
        },
        MaxAge: 12 * time.Hour,
    }
}

func CORS(cfg CORSConfig) gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     cfg.AllowOrigins,
        AllowMethods:     cfg.AllowMethods,
        AllowHeaders:     cfg.AllowHeaders,
        ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
        AllowCredentials: true,
        MaxAge:           cfg.MaxAge,
    })
}
```

### 4.5 Trace 中间件 (`pkg/middleware/trace.go`)

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
    "go.opentelemetry.io/otel/trace"
)

type TraceConfig struct {
    Enabled     bool
    ServiceName string
}

func Trace(serviceName string) gin.HandlerFunc {
    tracer := otel.Tracer(serviceName)
    propagator := otel.GetTextMapPropagator()

    return func(c *gin.Context) {
        // 从请求头提取 trace context
        ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

        spanName := c.FullPath()
        if spanName == "" {
            spanName = c.Request.URL.Path
        }

        ctx, span := tracer.Start(ctx, spanName,
            trace.WithAttributes(
                semconv.HTTPMethod(c.Request.Method),
                semconv.HTTPRoute(c.FullPath()),
                semconv.HTTPURL(c.Request.URL.String()),
            ),
            trace.WithSpanKind(trace.SpanKindServer),
        )
        defer span.End()

        // 注入到 request context
        c.Request = c.Request.WithContext(ctx)

        // 传递 trace headers 到下游
        headers := make(propagation.HeaderCarrier)
        propagator.Inject(ctx, headers)
        for k, v := range headers {
            c.Request.Header.Set(k, v[0])
        }

        c.Next()

        span.SetAttributes(semconv.HTTPStatusCode(c.Writer.Status()))
    }
}
```

### 4.6 Auth 中间件 (`pkg/middleware/auth.go`)

```go
package middleware

import (
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
    "github.com/your-org/grand-canal-guardian/pkg/response"
)

type AuthConfig struct {
    JWTSecret    string
    SkipPaths    []string
    SkipPrefixes []string
}

// ClaimsKey context key
const (
    ClaimsKey = "user_claims"
    UserIDKey = "user_id"
    RoleKey   = "role"
)

type JWTClaims struct {
    jwt.RegisteredClaims
    UserID string `json:"sub"`
    Role   string `json:"role"`
}

func Auth(cfg AuthConfig) gin.HandlerFunc {
    skipPaths := make(map[string]bool, len(cfg.SkipPaths))
    for _, p := range cfg.SkipPaths {
        skipPaths[p] = true
    }

    return func(c *gin.Context) {
        // 白名单跳过
        path := c.Request.URL.Path
        if skipPaths[path] {
            c.Next()
            return
        }
        for _, prefix := range cfg.SkipPrefixes {
            if strings.HasPrefix(path, prefix) {
                c.Next()
                return
            }
        }

        // 提取 Token
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            response.AbortWithError(c, errors.NewDefault(errors.ErrTokenInvalid))
            return
        }
        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

        // 解析 & 验证
        claims := &JWTClaims{}
        token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
            return []byte(cfg.JWTSecret), nil
        })
        if err != nil || !token.Valid {
            response.AbortWithError(c, errors.NewDefault(errors.ErrTokenExpired))
            return
        }

        // 注入 Claims 到 Context
        c.Set(ClaimsKey, claims)
        c.Set(UserIDKey, claims.UserID)
        c.Set(RoleKey, claims.Role)

        c.Next()
    }
}

// ---- 辅助函数，handler 中使用 ----

func GetUserID(c *gin.Context) string {
    id, _ := c.Get(UserIDKey)
    return id.(string)
}

func GetRole(c *gin.Context) string {
    role, _ := c.Get(RoleKey)
    return role.(string)
}

// RequireRole 角色校验中间件工厂
func RequireRole(roles ...string) gin.HandlerFunc {
    roleSet := make(map[string]bool, len(roles))
    for _, r := range roles {
        roleSet[r] = true
    }

    return func(c *gin.Context) {
        currentRole := GetRole(c)
        if !roleSet[currentRole] {
            response.AbortWithError(c, errors.NewDefault(errors.ErrForbidden))
            return
        }
        c.Next()
    }
}
```

### 4.7 RateLimit 中间件 (`pkg/middleware/rate_limit.go`)

```go
package middleware

import (
    "context"
    "fmt"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
    "github.com/your-org/grand-canal-guardian/pkg/response"
)

type RateLimitConfig struct {
    Enabled   bool
    Redis     *redis.Client
    Limit     int           // 窗口内最大请求数
    Window    time.Duration // 时间窗口
    KeyPrefix string        // Redis key 前缀
}

func DefaultRateLimitConfig() RateLimitConfig {
    return RateLimitConfig{
        Limit:     100,
        Window:    time.Minute,
        KeyPrefix: "rate_limit",
    }
}

// RateLimit 令牌桶 / 固定窗口限流
func RateLimit(cfg RateLimitConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := fmt.Sprintf("%s:%s:%s",
            cfg.KeyPrefix,
            c.ClientIP(),
            c.FullPath(),
        )

        // INCR + EXPIRE
        pipe := cfg.Redis.Pipeline()
        incr := pipe.Incr(c.Request.Context(), key)
        pipe.Expire(c.Request.Context(), key, cfg.Window)
        _, err := pipe.Exec(c.Request.Context())

        if err != nil {
            // Redis 不可用 → 放行 (避免限流器本身成为单点)
            c.Next()
            return
        }

        count := incr.Val()
        remaining := cfg.Limit - int(count)

        // 设置响应头
        c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Limit))
        c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, remaining)))
        c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(cfg.Window).Unix()))

        if count > int64(cfg.Limit) {
            response.AbortWithError(c,
                errors.NewDefault(errors.ErrTooManyRequest),
            )
            return
        }

        c.Next()
    }
}

// 防止 import cycle 的工具函数
func max(a, b int) int {
    if a > b { return a }
    return b
}
```

---

## 5. pkg/validator — 参数校验

(`pkg/validator/validator.go`)

```go
package validator

import (
    "reflect"
    "strings"
    "github.com/gin-gonic/gin/binding"
    "github.com/go-playground/locales/zh"
    ut "github.com/go-playground/universal-translator"
    "github.com/go-playground/validator/v10"
    zhTranslations "github.com/go-playground/validator/v10/translations/zh"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
)

var (
    Validate *validator.Validate
    Trans    ut.Translator
)

func init() {
    // 注册中文翻译器
    zh := zh.New()
    uni := ut.New(zh, zh)
    Trans, _ = uni.GetTranslator("zh")

    // 替换 Gin 默认 validator
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        Validate = v
        zhTranslations.RegisterDefaultTranslations(Validate, Trans)

        // 注册 JSON tag 名称提取 (让错误消息用 json tag 名而非字段名)
        Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
            name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
            if name == "-" || name == "" {
                return fld.Name
            }
            return name
        })
    }
}

// TranslateError 将 validator 错误翻译为中文 AppError
func TranslateError(err error) *errors.AppError {
    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        var msgs []string
        for _, e := range validationErrors {
            msgs = append(msgs, e.Translate(Trans))
        }
        return errors.BadRequest(strings.Join(msgs, "; "))
    }
    return errors.BadRequest("参数校验失败")
}

// ---- 自定义校验器 (示例) ----

// RegisterCustomValidators 注册业务自定义校验规则
func RegisterCustomValidators() {
    // 手机号
    Validate.RegisterValidation("mobile", func(fl validator.FieldLevel) bool {
        // 简单实现，生产用正则
        return len(fl.Field().String()) == 11
    })

    // 安全密码 (8-128 位，含大小写+数字)
    Validate.RegisterValidation("secure_password", func(fl validator.FieldLevel) bool {
        s := fl.Field().String()
        hasUpper := strings.ContainsAny(s, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
        hasLower := strings.ContainsAny(s, "abcdefghijklmnopqrstuvwxyz")
        hasDigit := strings.ContainsAny(s, "0123456789")
        return len(s) >= 8 && len(s) <= 128 && hasUpper && hasLower && hasDigit
    })

    // 文件类型白名单
    Validate.RegisterValidation("image_ext", func(fl validator.FieldLevel) bool {
        allowed := map[string]bool{
            "jpg": true, "jpeg": true, "png": true, "gif": true, "webp": true,
        }
        s := strings.ToLower(fl.Field().String())
        // 取扩展名
        parts := strings.Split(s, ".")
        return allowed[parts[len(parts)-1]]
    })
}

// BindAndValidate 通用绑定+校验 (返回翻译后的错误)
func BindAndValidate(c *gin.Context, obj interface{}) error {
    if err := c.ShouldBindJSON(obj); err != nil {
        return TranslateError(err)
    }
    if err := Validate.Struct(obj); err != nil {
        return TranslateError(err)
    }
    return nil
}
```

**Request DTO 示例**:

```go
type CreatePostRequest struct {
    Title   string   `json:"title"   binding:"required,min=1,max=128"`
    Content string   `json:"content" binding:"required,min=1,max=10000"`
    Topic   string   `json:"topic"   binding:"required,oneof=share ecology culture question other"`
    Tags    []string `json:"tags"    binding:"max=5,dive,min=1,max=32"`
}

// Handler 中使用:
func CreatePost(c *gin.Context) {
    var req CreatePostRequest
    if err := validator.BindAndValidate(c, &req); err != nil {
        response.Error(c, err.(*errors.AppError))
        return
    }
    // ... 业务逻辑
}
```

---

## 6. pkg/app — Gin 启动器

### 6.1 引擎工厂 (`pkg/app/app.go`)

```go
package app

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "github.com/your-org/grand-canal-guardian/pkg/middleware"
)

type App struct {
    Engine *gin.Engine
    Server *http.Server
    Logger *zap.Logger
    Config Config
}

type Config struct {
    Name string // 服务名
    Host string // 监听地址
    Port int    // 监听端口
    Mode string // gin mode: debug | release | test

    // 中间件
    EnableCORS      bool
    EnableRateLimit bool
    JWTSecret       string
    AuthSkipPaths   []string // 鉴权白名单
}

// New 创建 Gin 应用实例
func New(cfg Config) (*App, error) {
    // 设置 Gin 模式
    gin.SetMode(cfg.Mode)

    // 创建 Logger
    logger, err := NewLogger(cfg.Name, cfg.Mode)
    if err != nil {
        return nil, fmt.Errorf("创建 logger 失败: %w", err)
    }

    // 创建 Engine
    engine := gin.New()

    // ★ 注册中间件链
    middleware.Apply(engine, logger, middleware.MiddlewareConfig{
        CORS: middleware.CORSConfig{
            Enabled: cfg.EnableCORS,
            AllowOrigins: []string{"http://localhost:3000"},
        },
        Auth: middleware.AuthConfig{
            JWTSecret:   cfg.JWTSecret,
            SkipPaths:   cfg.AuthSkipPaths,
        },
        RateLimit: middleware.RateLimitConfig{
            Enabled: cfg.EnableRateLimit,
        },
        Trace: middleware.TraceConfig{
            Enabled:     true,
            ServiceName: cfg.Name,
        },
    })

    // 健康检查 (无需鉴权)
    engine.GET("/health", HealthHandler)
    engine.GET("/ready",  ReadyHandler)

    return &App{
        Engine: engine,
        Logger: logger,
        Config: cfg,
    }, nil
}

// MountRoutes 挂载业务路由 (由各服务调用)
func (a *App) MountRoutes(register func(r *gin.RouterGroup)) {
    v1 := a.Engine.Group("/api/v1")
    register(v1)
}
```

### 6.2 优雅关闭 (`pkg/app/shutdown.go`)

```go
package app

import (
    "context"
    "fmt"
    "net/http"
    "time"
    "go.uber.org/zap"
)

// Run 启动服务并等待优雅关闭
func (a *App) Run() error {
    a.Server = &http.Server{
        Addr:         fmt.Sprintf("%s:%d", a.Config.Host, a.Config.Port),
        Handler:      a.Engine,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // 启动 (非阻塞)
    errCh := make(chan error, 1)
    go func() {
        a.Logger.Info("服务启动",
            zap.String("name", a.Config.Name),
            zap.String("addr", a.Server.Addr),
        )
        if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            errCh <- err
        }
    }()

    // 等待信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    select {
    case err := <-errCh:
        return fmt.Errorf("服务启动失败: %w", err)
    case sig := <-quit:
        a.Logger.Info("收到关闭信号", zap.String("signal", sig.String()))
    }

    // 优雅关闭
    a.Logger.Info("正在优雅关闭...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // 停止接收新请求
    if err := a.Server.Shutdown(ctx); err != nil {
        a.Logger.Error("强制关闭", zap.Error(err))
        return err
    }

    // 等待现有请求处理完毕
    <-ctx.Done()
    a.Logger.Info("服务已关闭")
    return nil
}
```

### 6.3 健康检查 (`pkg/app/health.go`)

```go
package app

import (
    "github.com/gin-gonic/gin"
)

// 全局健康检查器注册表
var readinessChecks []func() error

// RegisterReadinessCheck 注册就绪检查 (各服务 init 时调用)
func RegisterReadinessCheck(fn func() error) {
    readinessChecks = append(readinessChecks, fn)
}

// HealthHandler liveness probe — 仅检查进程存活
func HealthHandler(c *gin.Context) {
    c.JSON(200, gin.H{"status": "alive"})
}

// ReadyHandler readiness probe — 检查所有依赖
func ReadyHandler(c *gin.Context) {
    for _, check := range readinessChecks {
        if err := check(); err != nil {
            c.JSON(503, gin.H{
                "status": "not_ready",
                "reason": err.Error(),
            })
            return
        }
    }
    c.JSON(200, gin.H{"status": "ready"})
}
```

### 6.4 日志工厂 (`pkg/app/logger.go`)

```go
package app

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func NewLogger(serviceName, mode string) (*zap.Logger, error) {
    var cfg zap.Config

    if mode == "release" {
        cfg = zap.NewProductionConfig()
        cfg.EncoderConfig.TimeKey = "timestamp"
        cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    } else {
        cfg = zap.NewDevelopmentConfig()
        cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    }

    cfg.InitialFields = map[string]interface{}{
        "service": serviceName,
    }

    return cfg.Build()
}
```

### 6.5 服务 main.go 示例

```go
// services/user-service/cmd/main.go
package main

import (
    "github.com/your-org/grand-canal-guardian/pkg/app"
    "github.com/your-org/grand-canal-guardian/pkg/validator"
    "github.com/your-org/grand-canal-guardian/services/user-service/internal/handler"
    "github.com/your-org/grand-canal-guardian/services/user-service/internal/wire"
)

func main() {
    // 1. 注册自定义校验器
    validator.RegisterCustomValidators()

    // 2. 创建 App
    application, err := app.New(app.Config{
        Name: "user-service",
        Host: "0.0.0.0",
        Port: 8001,
        Mode: ginMode(),
        JWTSecret:   os.Getenv("JWT_SECRET"),
        AuthSkipPaths: []string{
            "/api/v1/auth/register",
            "/api/v1/auth/login",
            "/api/v1/auth/refresh",
        },
    })
    if err != nil {
        log.Fatalf("创建应用失败: %v", err)
    }

    // 3. 依赖注入 (见第6.6节)
    deps, err := wire.InitializeApp()
    if err != nil {
        log.Fatalf("初始化依赖失败: %v", err)
    }

    // 4. 注册就绪检查
    app.RegisterReadinessCheck(deps.DB.Ping)
    app.RegisterReadinessCheck(deps.Redis.Ping)

    // 5. 挂载路由
    application.MountRoutes(func(v1 *gin.RouterGroup) {
        h := handler.NewUserHandler(deps.UserService)
        h.RegisterRoutes(v1)
    })

    // 6. 启动 → 等待信号 → 优雅关闭
    if err := application.Run(); err != nil {
        log.Fatalf("服务异常退出: %v", err)
    }
}

func ginMode() string {
    if mode := os.Getenv("GIN_MODE"); mode != "" {
        return mode
    }
    return "debug"
}
```

---

## 7. pkg/transaction — 事务封装

### 7.1 事务装饰器 (`pkg/transaction/tx.go`)

```go
package transaction

import (
    "context"
    "gorm.io/gorm"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
)

// TxFunc 事务函数签名
type TxFunc func(ctx context.Context) error

// WithTx 在事务中执行业务逻辑 (自动 commit/rollback)
// 用法: err := transaction.WithTx(ctx, db, func(ctx context.Context) error { ... })
func WithTx(ctx context.Context, db *gorm.DB, fn TxFunc) error {
    tx := db.WithContext(ctx).Begin()
    if tx.Error != nil {
        return errors.WrapDefault(errors.ErrDatabaseError, tx.Error)
    }

    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r) // 让 Recovery 中间件处理
        }
    }()

    if err := fn(ctx); err != nil {
        tx.Rollback()
        return err
    }

    if err := tx.Commit().Error; err != nil {
        tx.Rollback()
        return errors.WrapDefault(errors.ErrDatabaseError, err)
    }

    return nil
}

// WithTxResult 带返回值的事务
type TxResultFunc[T any] func(ctx context.Context) (T, error)

func WithTxResult[T any](ctx context.Context, db *gorm.DB, fn TxResultFunc[T]) (T, error) {
    var result T
    tx := db.WithContext(ctx).Begin()
    if tx.Error != nil {
        return result, errors.WrapDefault(errors.ErrDatabaseError, tx.Error)
    }

    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()

    result, err := fn(ctx)
    if err != nil {
        tx.Rollback()
        return result, err
    }

    if err := tx.Commit().Error; err != nil {
        tx.Rollback()
        return result, errors.WrapDefault(errors.ErrDatabaseError, err)
    }

    return result, nil
}
```

**Service 层使用示例**:

```go
func (s *ContentService) CreatePostWithReview(ctx context.Context, req *CreatePostReq) (*Post, error) {
    // ★ 事务保证原子性: 创建帖子 + 插入审核记录 同时成功或同时失败
    return transaction.WithTxResult(ctx, s.db, func(ctx context.Context) (*Post, error) {
        post := &Post{
            AuthorID: req.AuthorID,
            Title:    req.Title,
            Content:  req.Content,
            Status:   "pending",
        }
        if err := s.db.Create(post).Error; err != nil {
            return nil, errors.WrapDefault(errors.ErrDatabaseError, err)
        }

        review := &Review{AuditID: post.ID, Status: "pending"}
        if err := s.db.Create(review).Error; err != nil {
            return nil, errors.WrapDefault(errors.ErrDatabaseError, err)
        }

        return post, nil
    })
}
```

---

## 8. pkg/grpc — 跨服务错误传播

### 8.1 AppError ↔ gRPC Status (`pkg/grpc/error_convert.go`)

```go
package grpc

import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
)

// ToGRPCStatus 将 AppError 转换为 gRPC Status
func ToGRPCStatus(err error) error {
    if err == nil {
        return nil
    }
    if appErr, ok := err.(*errors.AppError); ok {
        st := status.New(httpToGRPCCode(appErr.Code.HTTPStatus()), appErr.Message)
        ds, _ := st.WithDetails(/* protobuf 错误详情 */)
        return ds.Err()
    }
    return status.Errorf(codes.Internal, "内部错误: %v", err)
}

// FromGRPCStatus 将 gRPC Status 转换为 AppError
func FromGRPCStatus(err error) *errors.AppError {
    if err == nil {
        return nil
    }
    st, ok := status.FromError(err)
    if !ok {
        return errors.Internal(err)
    }
    return errors.New(
        gRPCCodeToErrorCode(st.Code()),
        st.Message(),
    )
}

func httpToGRPCCode(httpStatus int) codes.Code {
    switch {
    case httpStatus == 400:
        return codes.InvalidArgument
    case httpStatus == 401:
        return codes.Unauthenticated
    case httpStatus == 403:
        return codes.PermissionDenied
    case httpStatus == 404:
        return codes.NotFound
    case httpStatus == 429:
        return codes.ResourceExhausted
    case httpStatus == 503:
        return codes.Unavailable
    case httpStatus >= 500:
        return codes.Internal
    default:
        return codes.Unknown
    }
}

func gRPCCodeToErrorCode(code codes.Code) errors.ErrorCode {
    switch code {
    case codes.InvalidArgument:
        return errors.ErrBadRequest
    case codes.Unauthenticated:
        return errors.ErrUnauthorized
    case codes.PermissionDenied:
        return errors.ErrForbidden
    case codes.NotFound:
        return errors.ErrNotFound
    case codes.ResourceExhausted:
        return errors.ErrTooManyRequest
    case codes.Unavailable:
        return errors.ErrLLMUnavailable
    default:
        return errors.ErrInternal
    }
}
```

### 8.2 Unary 拦截器 (`pkg/grpc/interceptors.go`)

```go
package grpc

import (
    "context"
    "time"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/status"
)

// UnaryServerInterceptor 服务端拦截器: 日志 + 错误转换
func UnaryServerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()

        resp, err := handler(ctx, req)

        latency := time.Since(start)

        logger.Info("gRPC 调用完成",
            zap.String("method", info.FullMethod),
            zap.Duration("latency", latency),
            zap.Error(err),
        )

        // 将业务错误转换为 gRPC Status
        if err != nil {
            return resp, ToGRPCStatus(err)
        }
        return resp, nil
    }
}

// UnaryClientInterceptor 客户端拦截器: gRPC Status → AppError
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        err := invoker(ctx, method, req, reply, cc, opts...)
        if err != nil {
            // 转换 gRPC 错误为 AppError
            st, _ := status.FromError(err)
            return FromGRPCStatus(st.Err())
        }
        return nil
    }
}
```

---

## 9. swaggo — OpenAPI 自动生成

### 9.1 注解示例

```go
// services/user-service/internal/handler/user_handler.go
package handler

// @title           User Service API
// @version         1.0
// @description     大运河平台 - 用户与认证服务

// @host      localhost:8001
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer Token (格式: "Bearer <token>")

// Register godoc
// @Summary      用户注册
// @Description  创建新用户账号，返回 JWT Token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterRequest  true  "注册信息"
// @Success      201   {object}  response.R{data=UserProfile}
// @Failure      409   {object}  response.R  "用户名已存在"
// @Router       /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
    // ...
}

// GetProfile godoc
// @Summary      获取当前用户信息
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.R{data=UserProfile}
// @Failure      401  {object}  response.R  "未登录"
// @Router       /users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
    // ...
}

// ListPosts godoc
// @Summary      帖子列表
// @Tags         Posts
// @Param        page       query     int     false  "页码"        default(1)
// @Param        page_size  query     int     false  "每页数量"    default(20)
// @Param        topic      query     string  false  "话题分类"
// @Success      200  {object}  response.R{data=response.List}
// @Router       /posts [get]
func (h *PostHandler) ListPosts(c *gin.Context) {
    // ...
}
```

### 9.2 Makefile 自动生成

```makefile
# 根目录 Makefile
.PHONY: swagger swagger-all

SWAGGER_SERVICES = user-service content-service map-service quiz-service

# 为指定服务生成 swagger
swagger-%:
	cd services/$* && swag init -g cmd/main.go -o docs --parseDependency --parseInternal

# 生成所有服务 swagger
swagger-all:
	@for svc in $(SWAGGER_SERVICES); do \
		echo ">>> 生成 $$svc swagger 文档..."; \
		cd services/$$svc && swag init -g cmd/main.go -o docs --parseDependency --parseInternal; \
		cd ../..; \
	done
	@echo "Swagger 文档全部生成完毕"

# 合并所有 swagger.json → 统一 openapi.yaml (可选)
swagger-merge:
	python scripts/merge_swagger.py \
		services/user-service/docs/swagger.json \
		services/content-service/docs/swagger.json \
		services/map-service/docs/swagger.json \
		services/quiz-service/docs/swagger.json \
		-o docs/openapi.yaml
```

### 9.3 安装

```bash
go install github.com/swaggo/swag/cmd/swag@latest

# 代码中引入
import _ "github.com/your-org/grand-canal-guardian/services/user-service/docs" // swagger
```

---

## 10. 完整服务示例 — user-service

### 10.1 Handler 层

```go
// services/user-service/internal/handler/user_handler.go
package handler

import (
    "github.com/gin-gonic/gin"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
    "github.com/your-org/grand-canal-guardian/pkg/middleware"
    "github.com/your-org/grand-canal-guardian/pkg/response"
    "github.com/your-org/grand-canal-guardian/pkg/validator"
    "github.com/your-org/grand-canal-guardian/services/user-service/internal/model"
    "github.com/your-org/grand-canal-guardian/services/user-service/internal/service"
)

type UserHandler struct {
    svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
    return &UserHandler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
    auth := r.Group("/auth")
    {
        auth.POST("/register", h.Register)
        auth.POST("/login", h.Login)
        auth.POST("/refresh", h.RefreshToken)
    }

    users := r.Group("/users")
    users.Use(middleware.Auth(...)) // 需要登录
    {
        users.GET("/me", h.GetProfile)
        users.PUT("/me", h.UpdateProfile)
        users.GET("/:id", h.GetUserByID)
    }

    admin := r.Group("/admin/users")
    admin.Use(middleware.RequireRole("admin"))
    {
        admin.GET("", h.ListUsers)
        admin.PUT("/:id/role", h.ChangeRole)
        admin.POST("/:id/ban", h.BanUser)
    }
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
    var req model.RegisterRequest
    if err := validator.BindAndValidate(c, &req); err != nil {
        response.Error(c, err.(*errors.AppError))
        return
    }

    user, tokenPair, err := h.svc.Register(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err.(*errors.AppError))
        return
    }

    response.Created(c, gin.H{
        "user":          user,
        "access_token":  tokenPair.AccessToken,
        "refresh_token": tokenPair.RefreshToken,
        "token_type":    "Bearer",
        "expires_in":    900,
    })
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
    var req model.LoginRequest
    if err := validator.BindAndValidate(c, &req); err != nil {
        response.Error(c, err.(*errors.AppError))
        return
    }

    user, tokenPair, err := h.svc.Login(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err.(*errors.AppError))
        return
    }

    response.OK(c, gin.H{
        "user":          user,
        "access_token":  tokenPair.AccessToken,
        "refresh_token": tokenPair.RefreshToken,
        "token_type":    "Bearer",
        "expires_in":    900,
    })
}

// GetProfile 获取个人信息
func (h *UserHandler) GetProfile(c *gin.Context) {
    userID := middleware.GetUserID(c)
    profile, err := h.svc.GetProfile(c.Request.Context(), userID)
    if err != nil {
        response.Error(c, err.(*errors.AppError))
        return
    }
    response.OK(c, profile)
}

// ListUsers 管理员: 用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
    page := queryInt(c, "page", 1)
    pageSize := queryInt(c, "page_size", 20)
    role := c.Query("role")

    users, total, err := h.svc.ListUsers(c.Request.Context(), page, pageSize, role)
    if err != nil {
        response.Error(c, err.(*errors.AppError))
        return
    }

    response.OKList(c, users, page, pageSize, total)
}

func queryInt(c *gin.Context, key string, defaultVal int) int {
    val := c.Query(key)
    if val == "" {
        return defaultVal
    }
    // 简单处理，生产用 strconv.Atoi
    return defaultVal
}
```

### 10.2 Service 层

```go
// services/user-service/internal/service/user_service.go
package service

import (
    "context"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
    "github.com/your-org/grand-canal-guardian/services/user-service/internal/model"
    "github.com/your-org/grand-canal-guardian/services/user-service/internal/repository"
)

type UserService struct {
    repo       *repository.UserRepo
    jwtSecret  []byte
    accessTTL  time.Duration
    refreshTTL time.Duration
}

func NewUserService(repo *repository.UserRepo, jwtSecret string) *UserService {
    return &UserService{
        repo:       repo,
        jwtSecret:  []byte(jwtSecret),
        accessTTL:  15 * time.Minute,
        refreshTTL: 7 * 24 * time.Hour,
    }
}

func (s *UserService) Register(ctx context.Context, req *model.RegisterRequest) (*model.UserProfile, *model.TokenPair, error) {
    // 检查唯一性
    exists, err := s.repo.ExistsByUsername(ctx, req.Username)
    if err != nil {
        return nil, nil, errors.WrapDefault(errors.ErrDatabaseError, err)
    }
    if exists {
        return nil, nil, errors.NewDefault(errors.ErrUsernameExists)
    }

    // 加密密码
    hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
    if err != nil {
        return nil, nil, errors.Internal(err)
    }

    // 创建用户
    user := &model.User{
        Username: req.Username,
        Password: string(hashed),
        Email:    req.Email,
        Nickname: req.Nickname,
        Role:     "user",
    }
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, nil, errors.WrapDefault(errors.ErrDatabaseError, err)
    }

    // 生成 Token
    tokens, err := s.generateTokens(user)
    if err != nil {
        return nil, nil, err
    }

    return user.ToProfile(), tokens, nil
}

func (s *UserService) Login(ctx context.Context, req *model.LoginRequest) (*model.UserProfile, *model.TokenPair, error) {
    user, err := s.repo.FindByUsername(ctx, req.Username)
    if err != nil {
        return nil, nil, errors.NewDefault(errors.ErrUserNotFound)
    }

    if user.Banned {
        return nil, nil, errors.NewDefault(errors.ErrUserBanned)
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        return nil, nil, errors.NewDefault(errors.ErrPasswordWrong)
    }

    tokens, err := s.generateTokens(user)
    if err != nil {
        return nil, nil, err
    }

    return user.ToProfile(), tokens, nil
}

func (s *UserService) generateTokens(user *model.User) (*model.TokenPair, error) {
    now := time.Now()

    // Access Token
    accessClaims := &middleware.JWTClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   user.ID,
            ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
            IssuedAt:  jwt.NewNumericDate(now),
        },
        UserID: user.ID,
        Role:   user.Role,
    }
    accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
        SignedString(s.jwtSecret)
    if err != nil {
        return nil, errors.Internal(err)
    }

    // Refresh Token
    refreshClaims := &middleware.JWTClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   user.ID,
            ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
            IssuedAt:  jwt.NewNumericDate(now),
        },
        UserID: user.ID,
        Role:   user.Role,
    }
    refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).
        SignedString(s.jwtSecret)
    if err != nil {
        return nil, errors.Internal(err)
    }

    return &model.TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
    }, nil
}
```

### 10.3 error 传播链路图

```
Handler                         Service                      Repository
   │                               │                             │
   │  调用 svc.Register()          │                             │
   │──────────────────────────────►│                             │
   │                               │  repo.Create(user)         │
   │                               │───────────────────────────►│
   │                               │                             │ DB 报错
   │                               │◄──── errors.WrapDefault() ──┤
   │                               │  (ErrDatabaseError)         │
   │                               │                             │
   │◄──── *errors.AppError ────────┤                             │
   │  {Code:10801,                │                             │
   │   Message:"数据库错误",       │                             │
   │   Cause: sql.ErrNoRows}      │                             │
   │                               │                             │
   │  response.Error(c, err)       │                             │
   │  → HTTP 500 + {code:10801,   │                             │
   │    message:"数据库错误"}      │                             │
```

**关键点**:
- Handler 只管 `response.Error()` 或 `response.OK()`
- Service 只返回 `*errors.AppError`，不感知 HTTP
- Repository 返回原始 error，Service 用 `WrapDefault` 包装
- 每一层只加一层语义，不吞错误

---

## 附录: go.mod 依赖清单

```go
// pkg/go.mod
module github.com/your-org/grand-canal-guardian/pkg

go 1.22

require (
    github.com/gin-contrib/cors      v1.7.2
    github.com/gin-gonic/gin         v1.10.0
    github.com/go-playground/locales v0.14.1
    github.com/go-playground/universal-translator v0.18.1
    github.com/go-playground/validator/v10        v10.22.0
    github.com/golang-jwt/jwt/v5     v5.2.1
    github.com/google/uuid           v1.6.0
    github.com/redis/go-redis/v9     v9.5.4
    go.opentelemetry.io/otel         v1.28.0
    go.opentelemetry.io/otel/trace   v1.28.0
    go.uber.org/zap                  v1.27.0
    google.golang.org/grpc           v1.65.0
    gorm.io/gorm                     v1.25.11
)
```

---

> **下一步**: 将此文档的第 1-9 节合并到 `03-developer-guide.md` 的 Go 后端章节，替换原有的工程化空白部分。项目结构树也需同步更新。
