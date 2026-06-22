# 大运河平台 — Go 工程化规范（合并自洽版）

> **版本**: v1.1 | **日期**: 2026-06-18
>
> **本文件整合了**: 原 `04-go-engineering-standards.md`（基础框架）+ `05-production-grade-infra.md`（生产强化）+ `06-error-desensitization-patch.md`（脱敏补丁），并修复了三份草稿在并行期间产生的冲突与 bug。`05`、`06` 文档已废弃，以本文件为准。
>
> **目标**: 零重复、低耦合、统一错误/响应、完整中间件链、自动文档生成、优雅关闭、错误脱敏。

---

## 目录

1. [项目结构与模块布局](#1-项目结构与模块布局)
2. [pkg/errors — 统一错误（含脱敏）](#2-pkgerrors)
3. [pkg/response — 统一响应（双模式）](#3-pkgresponse)
4. [pkg/log — 日志系统（Zap+轮转+脱敏）](#4-pkglog)
5. [pkg/middleware — 基础中间件（RequestID/CORS/Trace/Recovery）](#5-pkgmiddleware)
6. [pkg/validator — 参数校验](#6-pkgvalidator)
7. [pkg/auth — 鉴权（黑名单+Refresh旋转+RBAC）](#7-pkgauth)
8. [pkg/ratelimit — 令牌桶限流（Lua）](#8-pkgratelimit)
9. [pkg/database — 连接池与监控](#9-pkgdatabase)
10. [pkg/app — 启动器与优雅关闭](#10-pkgapp)
11. [pkg/transaction — 事务装饰器](#11-pkgtransaction)
12. [pkg/grpc — 跨服务错误传播](#12-pkggrpc)
13. [swaggo — OpenAPI 自动生成](#13-swaggo)
14. [统一中间件链注册](#14-统一中间件链注册)
15. [完整服务示例 — user-service](#15-完整服务示例--user-service)
16. [附录：go.mod 依赖清单](#16-附录-gomod-依赖清单)

---

## 1. 项目结构与模块布局

```
grand-canal-guardian/
├── pkg/                            # 公共库 — 所有 Go 服务共享
│   ├── errors/
│   │   ├── codes.go                #   错误码枚举 (5 位: 模块+分类+序号)
│   │   ├── app_error.go            #   AppError (Code + Message + Safe + Cause)
│   │   ├── i18n.go                 #   中文错误消息映射
│   │   └── sanitize.go             #   ★ 脱敏引擎 + 错误分类
│   ├── response/
│   │   └── response.go             #   {code, message, data} + 双模式 Error
│   ├── log/
│   │   ├── logger.go               #   ★ Zap 工厂 + 轮转 + 脱敏
│   │   ├── middleware.go           #   Gin 请求日志 + Recovery 中间件
│   │   └── sanitize.go             #   日志字段脱敏
│   ├── middleware/                  # 仅保留无状态通用中间件
│   │   ├── chain.go                #   统一中间件链注册 (Apply)
│   │   ├── request_id.go           #   X-Request-ID
│   │   ├── cors.go                 #   CORS
│   │   └── trace.go                #   OpenTelemetry
│   ├── auth/                       # ★ 鉴权（生产级）
│   │   ├── token.go                #   TokenManager (双Token + 黑名单 + Refresh旋转)
│   │   ├── blacklist.go            #   Redis Token 黑名单
│   │   ├── session.go              #   多设备 Session + 一次性 Refresh
│   │   ├── rbac.go                 #   RBAC 位掩码权限矩阵
│   │   └── middleware.go           #   Auth 中间件 + 辅助函数
│   ├── ratelimit/                  # ★ 令牌桶限流（生产级）
│   │   ├── token_bucket.lua        #   Lua 原子脚本
│   │   ├── limiter.go              #   Redis 令牌桶
│   │   ├── middleware.go           #   三级限流中间件 (IP/User/API)
│   │   └── config.go               #   分级限流配置
│   ├── database/                   # ★ 连接池（生产级）
│   │   ├── pool.go                 #   连接池 + Prometheus 指标 + 慢查询
│   │   ├── callbacks.go            #   GORM Callback 查询耗时监控
│   │   └── health.go               #   健康检查注册
│   ├── validator/
│   │   └── validator.go            #   go-playground + 中文翻译
│   ├── app/
│   │   ├── app.go                  #   Gin 引擎工厂 + 路由挂载
│   │   ├── shutdown.go             #   优雅关闭 (SIGINT/SIGTERM)
│   │   └── health.go               #   /health + /ready (K8s 探针)
│   ├── transaction/
│   │   └── tx.go                   #   WithTx / WithTxResult (闭包接收 *gorm.DB)
│   ├── grpc/
│   │   ├── error_convert.go        #   AppError ↔ gRPC Status
│   │   └── interceptors.go         #   Unary 拦截器
│   └── go.mod
├── services/                       # 各业务服务
│   └── user-service/
│       ├── cmd/main.go
│       └── internal/
│           ├── handler/            #   只调 response.OK/Error
│           ├── service/            #   只返回 *errors.AppError
│           ├── repository/         #   返回原始 error
│           ├── model/
│           ├── config/
│           └── wire/               #   依赖注入
└── go.work                         # Go workspace
```

### 1.1 go.work — 多模块

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

### 1.2 pkg/go.mod

```
module github.com/your-org/grand-canal-guardian/pkg

go 1.22
```

各服务 `go.mod` 通过 `go.work` 自动解析 `pkg/`，**无需 `replace` 指令**（避免与 workspace 冗余）:

```
require github.com/your-org/grand-canal-guardian/pkg v0.0.0
```

---

## 2. pkg/errors

> 统一错误封装 + 错误码 + 中文消息 + **脱敏**（合并自 doc 04 基础与 doc 06 脱敏补丁）。
>
> **核心原则**: `Detail` 字段在生产环境经过脱敏处理，绝不直接暴露原始 error 文本（防数据库结构泄露）。

### 2.1 错误码枚举 (`pkg/errors/codes.go`)

```go
package errors

// ErrorCode 统一错误码 (5 位: 模块(2) + 分类(1) + 序号(2))
// 模块: 00=通用 01=认证 02=用户 03=内容 04=地图 05=问答 06=LLM 07=视觉 08=外部依赖
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
    ErrUnauthorized  ErrorCode = 10101 // 未登录
    ErrTokenExpired  ErrorCode = 10102 // Token 已过期
    ErrTokenInvalid  ErrorCode = 10103 // Token 无效
    ErrForbidden     ErrorCode = 10104 // 无权限
    ErrPasswordWrong ErrorCode = 10105 // 密码错误
    ErrUserBanned    ErrorCode = 10106 // 账号已封禁

    // ---- 用户 (02) ----
    ErrUserNotFound   ErrorCode = 10201
    ErrUsernameExists ErrorCode = 10202
    ErrEmailExists    ErrorCode = 10203
    ErrInvalidRole    ErrorCode = 10204
    ErrCannotBanSelf  ErrorCode = 10205

    // ---- 内容 (03) ----
    ErrPostNotFound     ErrorCode = 10301
    ErrPostNotOwner     ErrorCode = 10302
    ErrCommentNotFound  ErrorCode = 10303
    ErrContentSensitive ErrorCode = 10304

    // ---- 地图 (04) ----
    ErrPOINotFound   ErrorCode = 10401
    ErrRouteNotFound ErrorCode = 10402

    // ---- 问答 (05) ----
    ErrQuestionNotFound ErrorCode = 10501
    ErrDuplicateAnswer  ErrorCode = 10502
    ErrSessionExpired   ErrorCode = 10503

    // ---- LLM (06) ----
    ErrLLMTimeout     ErrorCode = 10601
    ErrLLMRateLimit   ErrorCode = 10602
    ErrLLMUnavailable ErrorCode = 10603

    // ---- 视觉 (07) ----
    ErrImageTooLarge   ErrorCode = 10701
    ErrImageFormat    ErrorCode = 10702
    ErrDetectionFailed ErrorCode = 10703

    // ---- 外部依赖 (08) ----
    ErrDatabaseError ErrorCode = 10801
    ErrRedisError   ErrorCode = 10802
    ErrRabbitMQError ErrorCode = 10803
    ErrMinIOError   ErrorCode = 10804
)

// HTTPStatus 返回对应的 HTTP 状态码
func (c ErrorCode) HTTPStatus() int {
    switch {
    case c == ErrBadRequest:
        return 400
    case c == ErrConflict:
        return 409
    case c >= 10101 && c <= 10103:
        return 401
    case c == ErrForbidden || c == ErrPostNotOwner:
        return 403
    case c == ErrNotFound || c == ErrPostNotFound || c == ErrCommentNotFound ||
        c == ErrUserNotFound || c == ErrPOINotFound || c == ErrRouteNotFound ||
        c == ErrQuestionNotFound:
        return 404
    case c == ErrTooManyRequest || c == ErrDuplicateAnswer || c == ErrSessionExpired:
        return 429
    case c == ErrLLMTimeout || c == ErrLLMRateLimit || c == ErrLLMUnavailable:
        return 503
    case c == ErrImageTooLarge:
        return 413
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

// Mode 控制脱敏级别，由 main.go 在启动时调用 SetMode 设置
//   "release" — 生产: Detail 不序列化，只输出 Safe 分类
//   "debug"   — 开发: Detail 输出脱敏后的摘要，便于排查
var Mode = "release"

// SetMode 由启动时调用
func SetMode(mode string) { Mode = mode }

// SafeDetail 生产环境安全摘要（不含任何原始数据，用于前端错误分类 + 监控聚合）
type SafeDetail struct {
    Type    string `json:"type,omitempty"`    // 错误类型 (如 "unique_violation")
    Summary string `json:"summary,omitempty"` // 安全摘要 (如 "数据重复")
}

// AppError 统一应用错误
type AppError struct {
    Code    ErrorCode   `json:"code"`
    Message string      `json:"message"`
    Detail  string      `json:"-"`           // ★ 不直接序列化；由 response.Error 按模式处理
    Safe    *SafeDetail `json:"-"`           // ★ 同上
    Cause   error       `json:"-"`           // 永远不序列化
}

func (e *AppError) Error() string {
    if e.Detail != "" {
        return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Detail)
    }
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Cause }

// ---- 构造器（基础，不含 cause）----

func New(code ErrorCode, message string) *AppError {
    return &AppError{Code: code, Message: message}
}

func Newf(code ErrorCode, format string, args ...any) *AppError {
    return &AppError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// NewDefault 使用错误码的默认中文消息
func NewDefault(code ErrorCode) *AppError {
    return New(code, code.GetMessage())
}

// ---- 构造器（包裹原始 cause，自动脱敏）----

// Wrap 用自定义消息包裹；Detail 经脱敏后存储，Safe 分类自动生成
func Wrap(code ErrorCode, message string, cause error) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
        Cause:   cause,
        Detail:  sanitizeDetail(code, cause),
        Safe:    classifyError(code, cause),
    }
}

// Wrapf 同 Wrap，格式化消息
func Wrapf(code ErrorCode, cause error, format string, args ...any) *AppError {
    return Wrap(code, fmt.Sprintf(format, args...), cause)
}

// WrapDefault 用错误码默认消息包裹原始错误
func WrapDefault(code ErrorCode, cause error) *AppError {
    return Wrap(code, code.GetMessage(), cause)
}

// ---- 快捷构造 ----

func BadRequest(msg string) *AppError { return New(ErrBadRequest, msg) }

func NotFound(resource string) *AppError {
    return Newf(ErrNotFound, "%s 不存在", resource)
}

func Unauthorized(msg string) *AppError { return New(ErrUnauthorized, msg) }
func Forbidden(msg string) *AppError    { return New(ErrForbidden, msg) }
func Conflict(msg string) *AppError     { return New(ErrConflict, msg) }

// Internal 包裹内部错误 — Detail 固定为 "internal_error"，不泄露任何原始信息
func Internal(cause error) *AppError {
    return &AppError{
        Code:    ErrInternal,
        Message: ErrInternal.GetMessage(),
        Cause:   cause,
        Detail:  "internal_error",
        Safe:    &SafeDetail{Type: "internal", Summary: "internal_error"},
    }
}
```

### 2.3 中文错误消息 (`pkg/errors/i18n.go`)

```go
package errors

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

    ErrLLMTimeout:     "AI 响应超时，请稍后重试",
    ErrLLMRateLimit:   "AI 服务繁忙，请稍后再试",
    ErrLLMUnavailable: "AI 服务暂时不可用",

    ErrImageTooLarge:    "图片文件过大，请压缩后重试",
    ErrImageFormat:     "不支持的图片格式",
    ErrDetectionFailed: "垃圾识别失败，请重新拍摄",

    ErrDatabaseError: "数据库错误",
}

func (c ErrorCode) GetMessage() string {
    if msg, ok := DefaultMessages[c]; ok {
        return msg
    }
    return "未知错误"
}
```

### 2.4 脱敏引擎 (`pkg/errors/sanitize.go`)

```go
package errors

import (
    "regexp"
    "strings"
)

// leakPatterns 匹配数据库结构/连接信息等敏感片段 → 替换
var leakPatterns = []struct {
    Pattern *regexp.Regexp
    Replace string
}{
    // PostgreSQL 约束名
    {regexp.MustCompile(`"[a-z_]+_[a-z_]+_key"`), `"***_key"`},
    {regexp.MustCompile(`relation "([a-z_]+)"`), `relation "***"`},
    {regexp.MustCompile(`column "([a-z_]+)"`), `column "***"`},
    // 重复键值（值脱敏，保留字段名）
    {regexp.MustCompile(`Key \(([^)]+)\)=\([^)]*\)`), `Key ($1)=(***)`)},
    // MySQL 重复
    {regexp.MustCompile(`Duplicate entry '([^']+)' for key`), `Duplicate entry '***' for key`},
    // 连接信息
    {regexp.MustCompile(`host=[^\s]+`), `host=***`},
    {regexp.MustCompile(`dbname=[^\s]+`), `dbname=***`},
    {regexp.MustCompile(`user=[^\s]+`), `user=***`},
    {regexp.MustCompile(`password=[^\s]+`), `password=***`},
    // panic 中的具体值
    {regexp.MustCompile(`index out of range \[\d+:\d+\]`), `index out of range [N:N]`},
    {regexp.MustCompile(`invalid memory address 0x[0-9a-f]+`), `invalid memory address ***`},
    // UUID / 长数字 ID
    {regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`), `<uuid>`},
}

// sanitizeDetail 脱敏错误详情
func sanitizeDetail(code ErrorCode, cause error) string {
    if cause == nil {
        return code.GetMessage()
    }

    // 内部错误 / 数据库错误: 不返回任何原始消息（最敏感）
    if code == ErrInternal || code == ErrDatabaseError {
        return code.GetMessage()
    }

    raw := cause.Error()
    for _, rule := range leakPatterns {
        raw = rule.Pattern.ReplaceAllString(raw, rule.Replace)
    }

    // 截断（防超长注入）
    if len(raw) > 200 {
        raw = raw[:200] + "..."
    }
    return raw
}

// classifyError 生成生产环境安全分类（用于监控聚合，不含具体数据）
func classifyError(code ErrorCode, cause error) *SafeDetail {
    if cause == nil {
        return nil
    }

    errStr := cause.Error()
    sd := &SafeDetail{}

    switch {
    case strings.Contains(errStr, "duplicate key") ||
        strings.Contains(errStr, "Duplicate entry") ||
        strings.Contains(errStr, "unique constraint"):
        sd.Type = "unique_violation"
        sd.Summary = "数据重复"

    case strings.Contains(errStr, "foreign key") ||
        strings.Contains(errStr, "violates foreign"):
        sd.Type = "foreign_key_violation"
        sd.Summary = "外键约束"

    case strings.Contains(errStr, "connection refused") ||
        strings.Contains(errStr, "no such host") ||
        strings.Contains(errStr, "timeout"):
        sd.Type = "connection_error"
        sd.Summary = "数据库连接失败"

    case strings.Contains(errStr, "deadlock") ||
        strings.Contains(errStr, "could not serialize"):
        sd.Type = "deadlock"
        sd.Summary = "死锁"

    case strings.Contains(errStr, "no rows"):
        sd.Type = "not_found"
        sd.Summary = "记录不存在"

    default:
        sd.Type = "database_error"
        sd.Summary = "数据库错误"
    }

    return sd
}
```

**脱敏前后对比**（注册重复用户名）:

```
脱敏前 日志: [10801] 数据库错误: ERROR: duplicate key value
                violates unique constraint "users_username_key" (SQLSTATE 23505)
脱敏后 日志: {"code":10801,"message":"数据库错误","error_type":"unique_violation"}
脱敏后 HTTP: {"code":10801,"message":"数据库错误","data":{"error_type":"unique_violation"}}
```

---

## 3. pkg/response

> 统一响应格式 + **双模式 Error**（debug 返回脱敏 Detail，release 返回 Safe 分类）。

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

// PageData 分页元数据
type PageData struct {
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    Total      int64 `json:"total"`
    TotalPages int   `json:"total_pages"`
}

// List 带分页的列表
type List struct {
    Items      interface{} `json:"items"`
    Pagination PageData    `json:"pagination"`
}

// ---- 成功响应 ----

func OK(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, R{Code: 0, Message: "ok", Data: data})
}

func Created(c *gin.Context, data interface{}) {
    c.JSON(http.StatusCreated, R{Code: 0, Message: "创建成功", Data: data})
}

func NoContent(c *gin.Context) {
    c.Status(http.StatusNoContent)
}

func OKMessage(c *gin.Context, message string) {
    c.JSON(http.StatusOK, R{Code: 0, Message: message})
}

// OKList 列表 + 分页
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
                Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages,
            },
        },
    })
}

// ---- 错误响应（唯一出口）----

// Error 统一错误响应。所有 handler / 中间件通过此函数返回错误。
// 生产模式: 响应体仅含 code + message + data.error_type（安全分类）。
// 调试模式: 额外附带脱敏后的 detail，便于本地排查。
func Error(c *gin.Context, err *errors.AppError) {
    status := err.Code.HTTPStatus()
    body := R{Code: int(err.Code), Message: err.Message}

    switch errors.Mode {
    case "debug":
        if err.Detail != "" {
            body.Data = map[string]interface{}{"detail": err.Detail}
        }
    default: // release
        if err.Safe != nil {
            body.Data = map[string]interface{}{"error_type": err.Safe.Type}
        }
    }

    c.JSON(status, body)
    c.Abort()
}

// AbortWithError 中间件中使用（abort + 错误响应）
func AbortWithError(c *gin.Context, err *errors.AppError) {
    Error(c, err)
}
```

**Handler 唯一错误出口**:

```go
// ❌ 不要这样
func GetUser(c *gin.Context) {
    user, err := svc.GetUser(c, id)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()}) // 格式乱、泄露内部信息
        return
    }
    c.JSON(200, user)
}

// ✓ 统一出口
func GetUser(c *gin.Context) {
    user, err := svc.GetUser(c.Request.Context(), id)
    if err != nil {
        response.Error(c, err) // err 必须是 *errors.AppError
        return
    }
    response.OK(c, user)
}
```

> **错误传播链路**:
> ```
> Handler                         Service                      Repository
>    │                               │                             │
>    │  svc.GetUser(ctx, id)         │                             │
>    │──────────────────────────────►│  repo.FindByID(ctx, id)     │
>    │                               │───────────────────────────►│
>    │                               │       DB 报错（原始 error） │
>    │                               │◄── errors.WrapDefault() ───┤
>    │                               │    (ErrDatabaseError+脱敏)  │
>    │◄── *errors.AppError ─────────┤                             │
>    │  {Code:10801,                 │                             │
>    │   Message:"数据库错误",        │                             │
>    │   Safe:{unique_violation}}    │                             │
>    │                               │                             │
>    │  response.Error(c, err)       │                             │
>    │  → HTTP 500 + {code:10801,    │                             │
>    │    message:"数据库错误",       │                             │
>    │    data:{error_type:...}}     │                             │
> ```
> - Handler 只管 `response.Error()` / `response.OK()`
> - Service 只返回 `*errors.AppError`，不感知 HTTP
> - Repository 返回原始 `error`，由 Service 用 `WrapDefault` 包裹

---

## 4. pkg/log

> 日志系统：Zap + 文件轮转（lumberjack）+ 敏感信息脱敏 + Gin 请求日志 + Recovery。
>
> **注意**: 本包是日志的唯一工厂。`pkg/app` 的启动器通过本包创建 Logger，不再自带实现。

### 4.1 Logger 工厂 (`pkg/log/logger.go`)

```go
package log

import (
    "os"
    "regexp"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
)

// Config 日志配置
type Config struct {
    Level      string // debug | info | warn | error
    Format     string // json | console
    Output     string // stdout | file
    FilePath   string // 文件路径 (Output=file 时)
    MaxSizeMB  int    // 单文件最大 MB
    MaxBackups int    // 保留旧文件数
    MaxAgeDays int    // 保留天数
    Compress   bool   // 压缩旧文件
    Service    string // 服务名（自动注入每行日志）
}

func DefaultConfig(service string) Config {
    return Config{
        Level: "info", Format: "json", Output: "stdout",
        MaxSizeMB: 100, MaxBackups: 7, MaxAgeDays: 30, Compress: true,
        Service: service,
    }
}

// New 创建生产级 Logger
func New(cfg Config) (*zap.Logger, error) {
    level, err := zapcore.ParseLevel(cfg.Level)
    if err != nil {
        level = zapcore.InfoLevel
    }

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

    var writer zapcore.WriteSyncer
    if cfg.Output == "file" && cfg.FilePath != "" {
        writer = zapcore.AddSync(&lumberjack.Logger{
            Filename: cfg.FilePath, MaxSize: cfg.MaxSizeMB,
            MaxBackups: cfg.MaxBackups, MaxAge: cfg.MaxAgeDays,
            Compress: cfg.Compress, LocalTime: true,
        })
    } else {
        writer = zapcore.AddSync(os.Stdout)
    }

    core := zapcore.NewCore(encoder, writer, level)
    return zap.New(core,
        zap.AddCaller(),
        zap.AddCallerSkip(1),
        zap.Fields(zap.String("service", cfg.Service)),
    ), nil
}

// WithRequestID 注入 request_id
func WithRequestID(l *zap.Logger, requestID string) *zap.Logger {
    return l.With(zap.String("request_id", requestID))
}

// WithUser 注入用户信息（非敏感）
func WithUser(l *zap.Logger, userID, role string) *zap.Logger {
    return l.With(zap.String("user_id", userID), zap.String("role", role))
}
```

### 4.2 脱敏规则 (`pkg/log/sanitize.go`)

```go
package log

import (
    "fmt"
    "regexp"

    "github.com/your-org/grand-canal-guardian/pkg/errors"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// 敏感信息脱敏规则
var sensitivePatterns = []struct {
    Pattern *regexp.Regexp
    Replace string
}{
    {regexp.MustCompile(`(password=)[^&\s]+`), `${1}***`},
    {regexp.MustCompile(`(token=)[^&\s]+`), `${1}***`},
    {regexp.MustCompile(`(api_key=)[^&\s]+`), `${1}***`},
    {regexp.MustCompile(`"password"\s*:\s*"[^"]*"`), `"password":"***"`},
    {regexp.MustCompile(`"token"\s*:\s*"[^"]*"`), `"token":"***"`},
    {regexp.MustCompile(`"email"\s*:\s*"[^"@]+"@`), `"email":"***@`},
    {regexp.MustCompile(`"phone"\s*:\s*"\d{3})\d+`), `"phone":"$1****"`},
}

// Sanitize 脱敏字符串
func Sanitize(s string) string {
    for _, rule := range sensitivePatterns {
        s = rule.Pattern.ReplaceAllString(s, rule.Replace)
    }
    return s
}

// SanitizeField 创建脱敏 zap.Field
func SanitizeField(key, value string) zap.Field {
    return zap.String(key, Sanitize(value))
}

// SanitizeError 脱敏 error 字段（用于 zap.Error 替代）
// 日志中只输出 Code + Message + error_type，不含 Detail/Cause
func SanitizeError(err error) zap.Field {
    if appErr, ok := err.(*errors.AppError); ok {
        return zap.Object("error", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
            enc.AddInt("code", int(appErr.Code))
            enc.AddString("message", appErr.Message)
            if appErr.Safe != nil {
                enc.AddString("error_type", appErr.Safe.Type)
            }
            return nil
        }))
    }
    // 非 AppError → 只输出类型，不输出消息（可能是 driver 原始错误）
    return zap.String("error_type", fmt.Sprintf("%T", err))
}
```

### 4.3 Gin 中间件 (`pkg/log/middleware.go`)

```go
package log

import (
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// GinLogger 结构化请求日志（带脱敏）
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := Sanitize(c.Request.URL.RawQuery)

        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()

        // request_id 由 pkg/middleware.RequestID 注入
        requestID, _ := c.Get("X-Request-ID")

        fields := []zap.Field{
            zap.Any("request_id", requestID),
            zap.Int("status", status),
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.String("query", query),
            zap.String("ip", c.ClientIP()),
            zap.String("user_agent", c.GetHeader("User-Agent")),
            zap.Duration("latency", latency),
            zap.Int("body_size", c.Writer.Size()),
        }

        if len(c.Errors) > 0 {
            fields = append(fields, zap.String("gin_errors", c.Errors.String()))
        }

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

// Recovery 捕获 panic，记录完整栈到日志，HTTP 响应不泄露任何原始信息
// 必须放在中间件链最内层（第一个注册）
func Recovery(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                // 日志记录完整信息（服务器内部，受访问控制保护）
                logger.Error("panic recovered",
                    zap.Any("request_id", c.Value("X-Request-ID")),
                    zap.Any("panic", r),
                    zap.Stack("stack"),
                    zap.String("method", c.Request.Method),
                    zap.String("path", c.Request.URL.Path),
                )
                // HTTP 响应: 只返回通用内部错误
                c.AbortWithStatusJSON(500, gin.H{
                    "code":    10001,
                    "message":  "服务器内部错误",
                    "data":    gin.H{"error_type": "panic"},
                })
            }
        }()
        c.Next()
    }
}
```

---

## 5. pkg/middleware

> 仅保留**无状态、无外部依赖**的通用中间件：RequestID、CORS、Trace。
> 有状态的中间件（Auth、RateLimit、Logger、Recovery）分别在各自的包中实现。
> 本包还提供 `Apply`，按正确顺序串联**所有**中间件（跨包）。

### 5.1 RequestID (`pkg/middleware/request_id.go`)

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

// RequestID 注入或透传 X-Request-ID
func RequestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        rid := c.GetHeader(RequestIDHeader)
        if rid == "" {
            rid = uuid.New().String()
        }
        c.Set(RequestIDHeader, rid)
        c.Header(RequestIDHeader, rid)
        c.Next()
    }
}

// GetRequestID 从 context 获取 request_id
func GetRequestID(c *gin.Context) string {
    if rid, exists := c.Get(RequestIDHeader); exists {
        if s, ok := rid.(string); ok {
            return s
        }
    }
    return ""
}
```

### 5.2 CORS (`pkg/middleware/cors.go`)

```go
package middleware

import (
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

type CORSConfig struct {
    Enabled      bool
    AllowOrigins []string
    AllowMethods []string
    AllowHeaders []string
    MaxAge       time.Duration
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

### 5.3 Trace (`pkg/middleware/trace.go`)

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

        c.Request = c.Request.WithContext(ctx)

        // 注入 trace headers 到下游
        carrier := make(propagation.HeaderCarrier)
        propagator.Inject(ctx, carrier)
        for k, v := range carrier {
            c.Request.Header.Set(k, v[0])
        }

        c.Next()

        span.SetAttributes(semconv.HTTPStatusCode(c.Writer.Status()))
    }
}
```

### 5.4 统一中间件链注册 (`pkg/middleware/chain.go`)

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    pkgauth "github.com/your-org/grand-canal-guardian/pkg/auth"
    "github.com/your-org/grand-canal-guardian/pkg/log"
    pkgratelimit "github.com/your-org/grand-canal-guardian/pkg/ratelimit"
)

// Dependencies Apply 所需的所有有状态中间件依赖
type Dependencies struct {
    Logger    *zap.Logger
    Auth      *pkgauth.AuthMiddleware
    RateLimit gin.HandlerFunc // 由 ratelimit.Middleware(...) 生成
    CORS      CORSConfig
    Trace     TraceConfig
}

// Apply 按正确顺序注册所有中间件。
// 顺序: Recovery → RequestID → Logger → CORS → Trace → Auth → RateLimit
//   - Recovery 最内层（捕获所有后续中间件 + handler 的 panic）
//   - Logger 紧随其后（记录 request_id）
//   - Auth/RateLimit 最外层（尽早拦截无效请求）
//
// 注意: /health、/ready 等探针路由应在调用 Apply 后单独注册，
//       并确保它们不在 Auth 白名单外（见 pkg/app/health.go 处理）。
func Apply(r *gin.Engine, deps Dependencies) {
    // 1. Recovery
    r.Use(log.Recovery(deps.Logger))

    // 2. RequestID
    r.Use(RequestID())

    // 3. Logger
    r.Use(log.GinLogger(deps.Logger))

    // 4. CORS
    if deps.CORS.Enabled {
        r.Use(CORS(deps.CORS))
    }

    // 5. Trace
    if deps.Trace.Enabled {
        r.Use(Trace(deps.Trace.ServiceName))
    }

    // 6. Auth（内含白名单跳过逻辑）
    if deps.Auth != nil {
        r.Use(deps.Auth.Handler())
    }

    // 7. RateLimit
    if deps.RateLimit != nil {
        r.Use(deps.RateLimit)
    }
}
```

---

## 6. pkg/validator

> go-playground/validator + 中文翻译 + 业务自定义校验器。

(`pkg/validator/validator.go`)

```go
package validator

import (
    "reflect"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "github.com/go-playground/locales/zh"
    ut "github.com/go-playground/universal-translator"
    "github.com/go-playground/validator/v10"
    zhTranslations "github.com/go-playground/validator/v10/translations/zh"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
)

var (
    validate *validator.Validate
    trans    ut.Translator
)

func init() {
    // ★ 注意: locale 包变量重命名为 zhLocale，避免遮蔽
    zhLocale := zh.New()
    uni := ut.New(zhLocale, zhLocale)
    trans, _ = uni.GetTranslator("zh")

    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        validate = v
        _ = zhTranslations.RegisterDefaultTranslations(validate, trans)

        // 让错误消息用 json tag 名而非字段名
        validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
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
            msgs = append(msgs, e.Translate(trans))
        }
        return errors.BadRequest(strings.Join(msgs, "; "))
    }
    return errors.BadRequest("参数校验失败")
}

// RegisterCustomValidators 注册业务自定义校验规则（main.go 启动时调用）
func RegisterCustomValidators() {
    if validate == nil {
        return
    }
    // 手机号
    _ = validate.RegisterValidation("mobile", func(fl validator.FieldLevel) bool {
        s := fl.Field().String()
        return len(s) == 11
    })
    // 安全密码 (8-128 位，含大小写+数字)
    _ = validate.RegisterValidation("secure_password", func(fl validator.FieldLevel) bool {
        s := fl.Field().String()
        hasUpper := strings.ContainsAny(s, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
        hasLower := strings.ContainsAny(s, "abcdefghijklmnopqrstuvwxyz")
        hasDigit := strings.ContainsAny(s, "0123456789")
        return len(s) >= 8 && len(s) <= 128 && hasUpper && hasLower && hasDigit
    })
}

// BindAndValidate 绑定 + 校验，返回翻译后的 *AppError
func BindAndValidate(c *gin.Context, obj interface{}) error {
    if err := c.ShouldBindJSON(obj); err != nil {
        return TranslateError(err)
    }
    if validate != nil {
        if err := validate.Struct(obj); err != nil {
            return TranslateError(err)
        }
    }
    return nil
}

// QueryInt 解析 query int 参数（带默认值）
func QueryInt(c *gin.Context, key string, defaultVal int) int {
    val := c.Query(key)
    if val == "" {
        return defaultVal
    }
    n, err := strconv.Atoi(val)
    if err != nil || n < 1 {
        return defaultVal
    }
    return n
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
        response.Error(c, err) // err 已是 *AppError
        return
    }
    // ...
}
```

---

## 7. pkg/auth

> 鉴权（生产级）：Token 黑名单 + Refresh 旋转 + 重放检测 + RBAC 位掩码。
>
> **修复的关键 bug**:
> - JWT Claims 不再重复定义 `UserID`（用 `RegisteredClaims.Subject`）
> - RefreshAccess 处理 nil session，重放时通过 jti 反查 userID
> - RefreshAccess 在签发新 token 前查 DB 取最新 role（防 role 丢失）
> - SaveRefreshToken 修复 Expire 不可达代码
> - VerifyAndRevokeRefreshToken 返回 jti 以支持重放溯源

### 7.1 JWT Claims (`pkg/auth/claims.go`)

```go
package auth

import (
    "github.com/golang-jwt/jwt/v5"
)

// Claims JWT 载荷。
// ★ 注意: 不再单独定义 UserID 字段，统一使用 RegisteredClaims.Subject (json:"sub")，
//         避免 JSON key 冲突导致序列化相互覆盖。
type Claims struct {
    jwt.RegisteredClaims
    Role     string `json:"role"`
    DeviceID string `json:"device_id,omitempty"` // 设备标识（浏览器指纹 / UUID）
}

// UserID 是 Claims.Subject 的语义化别名
func (c *Claims) UserID() string {
    if c == nil {
        return ""
    }
    return c.Subject
}
```

### 7.2 Token 黑名单 (`pkg/auth/blacklist.go`)

```go
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

func blacklistKey(jti string) string {
    return fmt.Sprintf("auth:blacklist:%s", jti)
}

// Add 将 Token 的 jti 加入黑名单，TTL = Token 剩余有效期
func (b *TokenBlacklist) Add(ctx context.Context, jti string, expiresAt time.Time) error {
    ttl := time.Until(expiresAt)
    if ttl <= 0 {
        return nil // 已过期，无需黑名单
    }
    return b.rdb.Set(ctx, blacklistKey(jti), "1", ttl).Err()
}

// IsBlacklisted 检查 jti 是否在黑名单中
func (b *TokenBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
    n, err := b.rdb.Exists(ctx, blacklistKey(jti)).Result()
    if err != nil {
        return false, err // Redis 不可用 → 由调用方决定是否放行
    }
    return n > 0, nil
}

// BlacklistAllUserTokens 踢出某用户所有设备（管理员操作 / 检测到重放攻击）
// sessionSetKey 形如 "auth:session:{userID}:{deviceID}"
func (b *TokenBlacklist) BlacklistAllUserTokens(ctx context.Context, userID string) error {
    pattern := fmt.Sprintf("auth:session:%s:*", userID)
    var cursor uint64
    for {
        keys, next, err := b.rdb.Scan(ctx, cursor, pattern, 100).Result()
        if err != nil {
            return err
        }
        for _, sessionKey := range keys {
            jtis, err := b.rdb.SMembers(ctx, sessionKey).Result()
            if err != nil {
                continue
            }
            for _, jti := range jtis {
                // 用较长 TTL 兜底（无法精确知道每个 token 的过期时间）
                _ = b.rdb.Set(ctx, blacklistKey(jti), "1", 24*time.Hour).Err()
            }
            _ = b.rdb.Del(ctx, sessionKey).Err()
        }
        if next == 0 {
            break
        }
        cursor = next
    }
    return nil
}
```

### 7.3 Session 存储 (`pkg/auth/session.go`)

```go
package auth

import (
    "context"
    "crypto/subtle"
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

// ErrRefreshTokenReused Refresh Token 被重放（已使用过或不存在）
var ErrRefreshTokenReused = errors.New("refresh token reused or expired")

// ErrRefreshTokenInvalid Refresh Token 格式错误或哈希不匹配
var ErrRefreshTokenInvalid = errors.New("refresh token invalid")

type SessionRepo struct {
    rdb *redis.Client
}

// RefreshSession Refresh Token 对应的会话信息
type RefreshSession struct {
    JTI      string
    UserID   string
    DeviceID string
}

func NewSessionRepo(rdb *redis.Client) *SessionRepo {
    return &SessionRepo{rdb: rdb}
}

func hashToken(token string) string {
    h := sha256.Sum256([]byte(token))
    return hex.EncodeToString(h[:])
}

func refreshKey(jti string) string {
    return fmt.Sprintf("auth:refresh:%s", jti)
}

// SaveRefreshToken 存储 Refresh Token（哈希后存入，原子设置 TTL）
func (r *SessionRepo) SaveRefreshToken(
    ctx context.Context, userID, role, deviceID, jti, token string, expiresAt time.Time,
) error {
    key := refreshKey(jti)
    ttl := time.Until(expiresAt)
    if ttl <= 0 {
        return nil
    }
    // ★ 用 Pipeline 保证 HSet + Expire 原子提交（修复原来 return 后 Expire 不可达的 bug）
    pipe := r.rdb.TxPipeline()
    pipe.HSet(ctx, key,
        "user_id", userID,
        "role", role,
        "device_id", deviceID,
        "token_hash", hashToken(token),
    )
    pipe.Expire(ctx, key, ttl)
    _, err := pipe.Exec(ctx)
    return err
}

// verifyAndRevokeRefreshToken 验证并立即删除 Refresh Token（一次性使用）。
// 返回 (session, nil) 表示成功；返回 (nil, ErrRefreshTokenReused) 表示重放/过期。
func (r *SessionRepo) verifyAndRevokeRefreshToken(ctx context.Context, jti, token string) (*RefreshSession, error) {
    key := refreshKey(jti)

    // Lua 原子操作: HGETALL + DEL
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
    if result == nil {
        // 不存在 = 已被使用 或 已过期
        return nil, ErrRefreshTokenReused
    }

    data, ok := result.([]interface{})
    if !ok || len(data) < 8 { // 至少 4 个字段 = 8 个元素
        return nil, ErrRefreshTokenReused
    }
    fields := make(map[string]string, len(data)/2)
    for i := 0; i < len(data); i += 2 {
        k, _ := data[i].(string)
        v, _ := data[i+1].(string)
        fields[k] = v
    }

    // 常量时间比较防时序攻击
    expected := fields["token_hash"]
    actual := hashToken(token)
    if subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) != 1 {
        return nil, ErrRefreshTokenInvalid
    }

    return &RefreshSession{
        JTI:      jti,
        UserID:   fields["user_id"],
        DeviceID: fields["device_id"],
    }, nil
}

// addAccessJTI 记录设备会话的 Access Token JTI（多设备管理）
func (r *SessionRepo) addAccessJTI(ctx context.Context, userID, deviceID, jti string) error {
    key := fmt.Sprintf("auth:session:%s:%s", userID, deviceID)
    return r.rdb.SAdd(ctx, key, jti).Err()
}

func (r *SessionRepo) removeAccessJTI(ctx context.Context, userID, deviceID, jti string) error {
    key := fmt.Sprintf("auth:session:%s:%s", userID, deviceID)
    return r.rdb.SRem(ctx, key, jti).Err()
}

// lookupRefreshTokenUserID 通过 jti 反查 userID（用于重放攻击溯源）
func (r *SessionRepo) lookupRefreshTokenUserID(ctx context.Context, jti string) (string, error) {
    // refresh token 已被上面的 Lua DEL，无法反查；改用一个独立的索引:
    // 在 SaveRefreshToken 时额外写 auth:refresh_index:{jti} = userID (短 TTL)
    key := fmt.Sprintf("auth:refresh_index:%s", jti)
    uid, err := r.rdb.Get(ctx, key).Result()
    if err != nil {
        return "", err
    }
    return uid, nil
}
```

### 7.4 Token 管理器 (`pkg/auth/token.go`)

```go
package auth

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "errors"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"

    apperrors "github.com/your-org/grand-canal-guardian/pkg/errors"
)

// RoleFetcher 在 Refresh 时获取用户最新 role（防 role 丢失/被改后仍用旧 role）
type RoleFetcher interface {
    GetRole(ctx context.Context, userID string) (string, error)
}

// TokenManager 统一的 Token 管理器
type TokenManager struct {
    secret      []byte
    accessTTL   time.Duration
    refreshTTL  time.Duration
    blacklist   *TokenBlacklist
    sessions    *SessionRepo
    roleFetcher RoleFetcher // 可选；nil 时使用 refresh session 中缓存的 role
    rdb         *redis.Client
}

func NewTokenManager(
    secret string, rdb *redis.Client,
    accessTTL, refreshTTL time.Duration,
    roleFetcher RoleFetcher,
) *TokenManager {
    return &TokenManager{
        secret:      []byte(secret),
        accessTTL:   accessTTL,
        refreshTTL:  refreshTTL,
        blacklist:   NewTokenBlacklist(rdb),
        sessions:    NewSessionRepo(rdb),
        roleFetcher: roleFetcher,
        rdb:         rdb,
    }
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
    accessJTI := uuid.New().String()

    claims := &Claims{
        RegisteredClaims: jwt.RegisteredClaims{
            ID:        accessJTI,
            Subject:   userID, // ★ 用 Subject，不再重复 UserID 字段
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
        },
        Role:     role,
        DeviceID: deviceID,
    }
    accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
        SignedString(m.secret)
    if err != nil {
        return nil, apperrors.Internal(err)
    }

    // Refresh Token: 随机字符串（非 JWT → 防伪造）
    refreshJTI := uuid.New().String()
    refreshRandom, err := generateRefreshToken()
    if err != nil {
        return nil, apperrors.Internal(err)
    }

    // 存储 refresh + 索引（用于重放溯源）
    expiresAt := now.Add(m.refreshTTL)
    if err := m.sessions.SaveRefreshToken(ctx, userID, role, deviceID, refreshJTI, refreshRandom, expiresAt); err != nil {
        return nil, apperrors.WrapDefault(apperrors.ErrRedisError, err)
    }
    // 索引 key，TTL 略长于 refresh token（重放时 token 已被删，靠索引反查）
    indexKey := fmt.Sprintf("auth:refresh_index:%s", refreshJTI)
    _ = m.rdb.Set(ctx, indexKey, userID, m.refreshTTL+time.Hour).Err()

    // 记录 session（多设备管理）
    if err := m.sessions.addAccessJTI(ctx, userID, deviceID, accessJTI); err != nil {
        return nil, apperrors.WrapDefault(apperrors.ErrRedisError, err)
    }

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: fmt.Sprintf("%s.%s", refreshJTI, refreshRandom),
        TokenType:    "Bearer",
        ExpiresIn:    int(m.accessTTL.Seconds()),
    }, nil
}

// RefreshAccess 用 Refresh Token 换取新的 TokenPair（旋转）。
// 旧 Refresh Token 立即失效，防重放攻击。
func (m *TokenManager) RefreshAccess(ctx context.Context, refreshTokenRaw string) (*TokenPair, error) {
    jti, token, err := parseRefreshToken(refreshTokenRaw)
    if err != nil {
        return nil, apperrors.NewDefault(apperrors.ErrTokenInvalid)
    }

    session, err := m.sessions.verifyAndRevokeRefreshToken(ctx, jti, token)
    if err != nil {
        if errors.Is(err, ErrRefreshTokenReused) {
            // ★ 重放攻击检测: 立即踢出该用户所有设备
            //    session 为 nil，通过索引反查 userID（修复原版 nil panic）
            if uid, lookupErr := m.sessions.lookupRefreshTokenUserID(ctx, jti); lookupErr == nil && uid != "" {
                go m.blacklist.BlacklistAllUserTokens(context.Background(), uid)
            }
            return nil, apperrors.NewDefault(apperrors.ErrTokenInvalid)
        }
        return nil, apperrors.NewDefault(apperrors.ErrTokenExpired)
    }

    // ★ 查 DB 取最新 role（修复原版刷新后 role 丢失的 bug）
    role := ""
    if m.roleFetcher != nil {
        if r, err := m.roleFetcher.GetRole(ctx, session.UserID); err == nil {
            role = r
        }
    }

    // 签发新 Token 对（旋转）
    return m.IssueTokens(ctx, session.UserID, role, session.DeviceID)
}

// ValidateAccess 验证 Access Token（含黑名单检查）
func (m *TokenManager) ValidateAccess(ctx context.Context, tokenStr string) (*Claims, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
        // ★ 强制 HMAC（防 alg=none 攻击）
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return m.secret, nil
    })
    if err != nil {
        // 区分过期 vs 无效
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, apperrors.NewDefault(apperrors.ErrTokenExpired)
        }
        return nil, apperrors.NewDefault(apperrors.ErrTokenInvalid)
    }
    if !token.Valid {
        return nil, apperrors.NewDefault(apperrors.ErrTokenInvalid)
    }

    // 黑名单检查
    blacklisted, err := m.blacklist.IsBlacklisted(ctx, claims.ID)
    if err != nil {
        // Redis 不可用 → 放行（可用性优先，可根据安全要求改为拒绝）
        return claims, nil
    }
    if blacklisted {
        return nil, apperrors.NewDefault(apperrors.ErrTokenInvalid)
    }

    return claims, nil
}

// Logout 登出 — Access Token jti 入黑名单
func (m *TokenManager) Logout(ctx context.Context, claims *Claims) error {
    if err := m.blacklist.Add(ctx, claims.ID, claims.ExpiresAt.Time); err != nil {
        return err
    }
    return m.sessions.removeAccessJTI(ctx, claims.Subject, claims.DeviceID, claims.ID)
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
    // 格式: "jti.random_hex"
    for i := len(raw) - 1; i >= 0; i-- {
        if raw[i] == '.' {
            return raw[:i], raw[i+1:], nil
        }
    }
    return "", "", errors.New("invalid refresh token format")
}
```

### 7.5 RBAC 权限矩阵 (`pkg/auth/rbac.go`)

```go
package auth

import (
    "github.com/gin-gonic/gin"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
    "github.com/your-org/grand-canal-guardian/pkg/response"
)

// Permission 权限位（位掩码）
type Permission uint64

const (
    PermReadPost Permission = 1 << iota
    PermCreatePost
    PermDeletePost
    PermComment
    PermLike
    PermReadMap
    PermPlayQuiz
    PermViewLeaderboard
    PermUploadGarbage
    PermViewMonitorData
    PermReviewPost
    PermManageUser
    PermSystemConfig
)

// RolePermissions 角色 → 权限映射
var RolePermissions = map[string]Permission{
    "user": PermReadPost | PermCreatePost | PermComment | PermLike |
        PermReadMap | PermPlayQuiz | PermViewLeaderboard,
    "monitor": PermReadPost | PermCreatePost | PermComment | PermLike |
        PermReadMap | PermPlayQuiz | PermViewLeaderboard |
        PermUploadGarbage | PermViewMonitorData,
    "admin": ^Permission(0), // 全部权限
}

// HasPermission 检查角色是否拥有指定权限
func HasPermission(role string, perm Permission) bool {
    perms, ok := RolePermissions[role]
    return ok && perms&perm != 0
}

// RequirePermission 权限校验中间件工厂
func RequirePermission(perm Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !HasPermission(GetRole(c), perm) {
            response.AbortWithError(c, errors.NewDefault(errors.ErrForbidden))
            return
        }
        c.Next()
    }
}

// RequireRole 角色校验中间件工厂（兼容旧风格）
func RequireRole(roles ...string) gin.HandlerFunc {
    set := make(map[string]bool, len(roles))
    for _, r := range roles {
        set[r] = true
    }
    return func(c *gin.Context) {
        if !set[GetRole(c)] {
            response.AbortWithError(c, errors.NewDefault(errors.ErrForbidden))
            return
        }
        c.Next()
    }
}
```

### 7.6 Auth 中间件 (`pkg/auth/middleware.go`)

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
    ClaimsKey = "auth_claims"
    UserIDKey = "auth_user_id"
    RoleKey   = "auth_role"
)

type AuthMiddleware struct {
    tm           *TokenManager
    skipPaths    map[string]bool
    skipPrefixes []string
}

func NewAuthMiddleware(tm *TokenManager, skipPaths, skipPrefixes []string) *AuthMiddleware {
    sp := make(map[string]bool, len(skipPaths))
    for _, p := range skipPaths {
        sp[p] = true
    }
    return &AuthMiddleware{tm: tm, skipPaths: sp, skipPrefixes: skipPrefixes}
}

// Handler 返回 Gin 中间件
func (a *AuthMiddleware) Handler() gin.HandlerFunc {
    return func(c *gin.Context) {
        if a.shouldSkip(c.Request.URL.Path) {
            c.Next()
            return
        }

        tokenStr := extractToken(c)
        if tokenStr == "" {
            response.AbortWithError(c, errors.NewDefault(errors.ErrTokenInvalid))
            return
        }

        claims, err := a.tm.ValidateAccess(c.Request.Context(), tokenStr)
        if err != nil {
            response.AbortWithError(c, toAppError(err))
            return
        }

        c.Set(ClaimsKey, claims)
        c.Set(UserIDKey, claims.Subject) // ★ 用 Subject
        c.Set(RoleKey, claims.Role)
        c.Next()
    }
}

func (a *AuthMiddleware) shouldSkip(path string) bool {
    if a.skipPaths[path] {
        return true
    }
    for _, p := range a.skipPrefixes {
        if strings.HasPrefix(path, p) {
            return true
        }
    }
    return false
}

func extractToken(c *gin.Context) string {
    h := c.GetHeader("Authorization")
    if h == "" || !strings.HasPrefix(h, "Bearer ") {
        return ""
    }
    return strings.TrimPrefix(h, "Bearer ")
}

func toAppError(err error) *errors.AppError {
    if appErr, ok := err.(*errors.AppError); ok {
        return appErr
    }
    return errors.NewDefault(errors.ErrTokenExpired)
}

// ---- Handler 辅助函数（全局唯一，pkg/middleware 中不再重复定义）----

func GetUserID(c *gin.Context) string {
    if v, ok := c.Get(UserIDKey); ok {
        if s, ok := v.(string); ok {
            return s
        }
    }
    return ""
}

func GetRole(c *gin.Context) string {
    if v, ok := c.Get(RoleKey); ok {
        if s, ok := v.(string); ok {
            return s
        }
    }
    return ""
}

func GetClaims(c *gin.Context) *Claims {
    if v, ok := c.Get(ClaimsKey); ok {
        if claims, ok := v.(*Claims); ok {
            return claims
        }
    }
    return nil
}
```

---

## 8. pkg/ratelimit

> 令牌桶限流（生产级）：Redis Lua 原子操作 + 三级限流（IP / 用户 / 接口）。
>
> **注意**: 原 doc 04 中 `pkg/middleware/rate_limit.go`（INCR 固定窗口）已废弃，统一使用本包。

### 8.1 Lua 令牌桶脚本 (`pkg/ratelimit/token_bucket.lua`)

```lua
-- KEYS[1]: bucket key
-- ARGV[1]: capacity     (桶容量)
-- ARGV[2]: rate         (令牌填充速率, token/s)
-- ARGV[3]: requested    (本次消耗令牌数)
-- ARGV[4]: now_ms       (当前时间毫秒)
-- 返回: {allowed(0/1), remaining, reset_at_ms, retry_after_ms}

local key        = KEYS[1]
local capacity   = tonumber(ARGV[1])
local rate       = tonumber(ARGV[2])
local requested  = tonumber(ARGV[3])
local now_ms     = tonumber(ARGV[4])

local bucket = redis.call('HMGET', key, 'tokens', 'last_fill_ms')
local tokens       = tonumber(bucket[1])
local last_fill_ms = tonumber(bucket[2])

if tokens == nil then
    tokens       = capacity
    last_fill_ms = now_ms
end

local elapsed_ms = now_ms - last_fill_ms
if elapsed_ms < 0 then elapsed_ms = 0 end
local new_tokens = math.min(capacity, tokens + (elapsed_ms / 1000.0) * rate)

local allowed = 0
if new_tokens >= requested then
    new_tokens = new_tokens - requested
    allowed = 1
end

local fill_time_ms = 0
if new_tokens < capacity then
    fill_time_ms = math.ceil((capacity - new_tokens) / rate * 1000)
end
local reset_at_ms = now_ms + fill_time_ms

local retry_after_ms = 0
if allowed == 0 then
    retry_after_ms = math.ceil((requested - new_tokens) / rate * 1000)
end

redis.call('HMSET', key, 'tokens', new_tokens, 'last_fill_ms', now_ms)
redis.call('PEXPIRE', key, math.max(60000, fill_time_ms + 10000))

return {allowed, math.floor(new_tokens), reset_at_ms, retry_after_ms}
```

### 8.2 限流器 (`pkg/ratelimit/limiter.go`)

```go
package ratelimit

import (
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
    Name      string  // 标识 (如 "ip", "user", "llm_chat")
    Capacity  int     // 桶容量（最大突发）
    Rate      float64 // 填充速率（令牌/秒）
}

// Limiter Redis 令牌桶限流器
type Limiter struct {
    rdb    *redis.Client
    script *redis.Script
}

func NewLimiter(rdb *redis.Client) *Limiter {
    return &Limiter{rdb: rdb, script: redis.NewScript(tokenBucketScript)}
}

// Result 限流结果
type Result struct {
    Allowed      bool
    Remaining    int
    ResetAt      time.Time
    RetryAfterMs int
}

// Allow 检查是否允许请求。Redis 不可用时 fail-open（可用性优先）。
func (l *Limiter) Allow(ctx context.Context, key string, cfg BucketConfig, cost int) (*Result, error) {
    nowMs := time.Now().UnixMilli()

    result, err := l.script.Run(ctx, l.rdb, []string{key},
        cfg.Capacity, cfg.Rate, cost, nowMs,
    ).Result()
    if err != nil {
        // fail-open
        return &Result{Allowed: true, Remaining: cfg.Capacity - cost}, nil
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

> **注**: 上面的 `Allow` 签名含 `context.Context`，import 需补 `"context"`。完整 import 块见附录。

### 8.3 三级限流配置 (`pkg/ratelimit/config.go`)

```go
package ratelimit

// TierConfig 三级限流配置
type TierConfig struct {
    IP   BucketConfig
    User BucketConfig
    API  map[string]BucketConfig // path → config
}

// DefaultTierConfig 默认三级限流。
// 阈值与架构文档 01 §7.1 对齐: 单 IP 100 req/s、LLM 50 req/s。
// （原 doc 05 的 20/s、2/s 为误值，已修正）
func DefaultTierConfig() TierConfig {
    return TierConfig{
        IP:   {Name: "ip", Capacity: 200, Rate: 100.0},   // 100/s 持续，200 突发
        User: {Name: "user", Capacity: 100, Rate: 50.0},
        API: map[string]BucketConfig{
            "/api/v1/llm/chat":         {Name: "llm_chat", Capacity: 100, Rate: 50.0},
            "/api/v1/vision/classify":   {Name: "vision", Capacity: 40, Rate: 10.0},
            "/api/v1/quiz/submit":       {Name: "quiz", Capacity: 60, Rate: 20.0},
            "/api/v1/auth/login":        {Name: "login", Capacity: 10, Rate: 2.0},  // 防爆破
            "/api/v1/upload/image":      {Name: "upload", Capacity: 40, Rate: 10.0},
        },
    }
}
```

### 8.4 三级限流中间件 (`pkg/ratelimit/middleware.go`)

```go
package ratelimit

import (
    "fmt"

    "github.com/gin-gonic/gin"
    pkgauth "github.com/your-org/grand-canal-guardian/pkg/auth"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
    "github.com/your-org/grand-canal-guardian/pkg/response"
)

// Middleware 三级令牌桶限流中间件
func Middleware(limiter *Limiter, cfg TierConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()

        // 1. IP 限流
        r, _ := limiter.Allow(ctx, fmt.Sprintf("ratelimit:ip:%s", c.ClientIP()), cfg.IP, 1)
        if !r.Allowed {
            setRateLimitHeaders(c, r)
            response.AbortWithError(c, errors.NewDefault(errors.ErrTooManyRequest))
            return
        }

        // 2. 用户限流（如已登录）
        if userID := pkgauth.GetUserID(c); userID != "" {
            r, _ := limiter.Allow(ctx, fmt.Sprintf("ratelimit:user:%s", userID), cfg.User, 1)
            if !r.Allowed {
                setRateLimitHeaders(c, r)
                response.AbortWithError(c, errors.NewDefault(errors.ErrTooManyRequest))
                return
            }
        }

        // 3. 接口级限流
        if apiCfg, ok := cfg.API[c.FullPath()]; ok {
            r, _ := limiter.Allow(ctx, fmt.Sprintf("ratelimit:api:%s", c.FullPath()), apiCfg, 1)
            if !r.Allowed {
                setRateLimitHeaders(c, r)
                response.AbortWithError(c, errors.NewDefault(errors.ErrTooManyRequest))
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

## 9. pkg/database

> 连接池 + 读写分离 + Prometheus 指标 + 慢查询告警。
>
> **修复的关键 bug**:
> - Prometheus 累计值（WaitCount/WaitDuration）改用 Gauge，不再用 Counter.Add（否则每 tick 累加全量，rate 全错）
> - GORM Callback 注册改用正确名称（`gorm:create` 等），并在 NewDB 中实际调用

### 9.1 连接池 (`pkg/database/pool.go`)

```go
package database

import (
    "context"
    "fmt"
    "log"
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
    // Gauge: 瞬时值（连接池状态每次采样覆盖）
    dbConnOpen = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "db_connections_open",
        Help: "当前打开的数据库连接数",
    }, []string{"db", "mode"})

    dbConnIdle = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "db_connections_idle",
        Help: "当前空闲连接数",
    }, []string{"db", "mode"})

    // ★ Gauge（非 Counter）: WaitCount/WaitDuration 是进程累计值，
    //   用 Gauge 直接采样，rate 时用 deriv()/delta() 计算增量
    dbConnWaitCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "db_connections_wait_total",
        Help: "累计等待连接次数（Gauge，采样累计值；用 deriv 计算速率）",
    }, []string{"db", "mode"})

    dbConnWaitDuration = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "db_connections_wait_seconds_total",
        Help: "累计等待连接总秒数（Gauge）",
    }, []string{"db", "mode"})

    // Histogram: 查询耗时分布
    dbQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "db_query_duration_seconds",
        Help:    "查询耗时",
        Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
    }, []string{"db", "operation"})

    // Counter: 慢查询次数
    dbSlowQueryCount = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "db_slow_query_total",
        Help: "慢查询计数",
    }, []string{"db"})
)

// PoolConfig 连接池配置
type PoolConfig struct {
    Name string

    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration

    SlowThreshold time.Duration // 慢查询阈值

    WriteDSN string
    ReadDSNs []string // 从库列表
}

func DefaultPoolConfig(name string) PoolConfig {
    return PoolConfig{
        Name: name,
        MaxOpenConns: 25, MaxIdleConns: 10,
        ConnMaxLifetime: time.Hour, ConnMaxIdleTime: 10 * time.Minute,
        SlowThreshold: 200 * time.Millisecond,
    }
}

// NewDB 创建数据库连接（主库 + 读写分离 + 监控）
func NewDB(cfg PoolConfig) (*gorm.DB, error) {
    gormCfg := &gorm.Config{
        Logger:                                   newGormLogger(cfg.Name, cfg.SlowThreshold),
        NowFunc:                                  func() time.Time { return time.Now().UTC() },
        DisableForeignKeyConstraintWhenMigrating: true,
    }

    db, err := gorm.Open(postgres.New(postgres.Config{
        DSN:                  cfg.WriteDSN,
        PreferSimpleProtocol: true,
    }), gormCfg)
    if err != nil {
        return nil, fmt.Errorf("连接主库失败: %w", err)
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("获取 *sql.DB 失败: %w", err)
    }
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

    // 读写分离（仅设从库；写走主库默认连接）
    if len(cfg.ReadDSNs) > 0 {
        replicas := make([]gorm.Dialector, len(cfg.ReadDSNs))
        for i, dsn := range cfg.ReadDSNs {
            replicas[i] = postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true})
        }
        if err := db.Use(dbresolver.Register(dbresolver.Config{
            Replicas: replicas,
            Policy:   &dbresolver.RandomPolicy{},
        }).SetConnMaxIdleTime(cfg.ConnMaxIdleTime).
            SetConnMaxLifetime(cfg.ConnMaxLifetime).
            SetMaxIdleConns(cfg.MaxIdleConns).
            SetMaxOpenConns(cfg.MaxOpenConns)); err != nil {
            return nil, fmt.Errorf("配置读写分离失败: %w", err)
        }
    }

    // ★ 注册查询耗时监控（修复原版 registerCallbacks 未被调用的问题）
    registerCallbacks(db, cfg.Name)

    // 启动连接池监控
    go monitorPool(cfg.Name, sqlDB)

    return db, nil
}

// monitorPool 定期采样连接池指标
func monitorPool(name string, db *sql.DB) {
    ticker := time.NewTicker(15 * time.Second)
    defer ticker.Stop()
    for range ticker.C {
        stats := db.Stats()
        dbConnOpen.WithLabelValues(name, "write").Set(float64(stats.OpenConnections))
        dbConnIdle.WithLabelValues(name, "write").Set(float64(stats.Idle))
        // ★ Gauge 采样累计值（不再 Add，否则每 tick 重复累加）
        dbConnWaitCount.WithLabelValues(name, "write").Set(float64(stats.WaitCount))
        dbConnWaitDuration.WithLabelValues(name, "write").Set(stats.WaitDuration.Seconds())

        maxOpen := stats.MaxOpenConnections
        if maxOpen > 0 && stats.OpenConnections > int(float64(maxOpen)*0.8) {
            log.Printf("[WARN] %s 连接池使用率 > 80%%: %d/%d, 等待: %d",
                name, stats.OpenConnections, maxOpen, stats.WaitCount)
        }
    }
}
```

> **注**: `monitorPool` 用到 `*sql.DB`，import 需补 `"database/sql"`。

### 9.2 GORM Callbacks (`pkg/database/callbacks.go`)

```go
package database

import (
    "fmt"
    "time"

    "gorm.io/gorm"
)

// registerCallbacks 注册所有 CRUD 操作的耗时监控
// ★ 修复原版: 用正确的 GORM callback 名称（gorm:create 等，非 "gorm:after_*" 通配）
func registerCallbacks(db *gorm.DB, dbName string) {
    registerDurationCallback(db, dbName, "create")
    registerDurationCallback(db, dbName, "query")
    registerDurationCallback(db, dbName, "update")
    registerDurationCallback(db, dbName, "delete")
    registerDurationCallback(db, dbName, "row")
    registerDurationCallback(db, dbName, "raw")
}

func registerDurationCallback(db *gorm.DB, dbName, op string) {
    cb := db.Callback()
    name := fmt.Sprintf("metrics:%s:%s", dbName, op)

    var after func() *gorm.Callback
    switch op {
    case "create":
        after = cb.Create().After("gorm:create")
    case "query":
        after = cb.Query().After("gorm:query")
    case "update":
        after = cb.Update().After("gorm:update")
    case "delete":
        after = cb.Delete().After("gorm:delete")
    case "row":
        after = cb.Row().After("gorm:row")
    case "raw":
        after = cb.Raw().After("gorm:raw")
    default:
        return
    }
    _ = after.Register(name, func(d *gorm.DB) {
        duration := time.Since(d.Statement.StartTime).Seconds()
        dbQueryDuration.WithLabelValues(dbName, op).Observe(duration)
    })
}
```

### 9.3 GORM Logger (`pkg/database/logger.go`)

```go
package database

import (
    "context"
    "log"
    "time"

    "gorm.io/gorm/logger"
)

type gormLogger struct {
    dbName        string
    slowThreshold time.Duration
    level         logger.LogLevel
}

func newGormLogger(dbName string, slowThreshold time.Duration) logger.Interface {
    return &gormLogger{
        dbName: dbName, slowThreshold: slowThreshold, level: logger.Warn,
    }
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
    cp := *l
    cp.level = level
    return &cp
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{})    {}
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
    log.Printf("[GORM WARN][%s] %s %v", l.dbName, msg, data)
}
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
    log.Printf("[GORM ERROR][%s] %s %v", l.dbName, msg, data)
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
    elapsed := time.Since(begin)
    sql, rows := fc()

    if elapsed > l.slowThreshold {
        dbSlowQueryCount.WithLabelValues(l.dbName).Inc() // ★ 用 dbName，非 "unknown"
        log.Printf("[SLOW QUERY][%s] %v | rows:%d | %s", l.dbName, elapsed, rows, sql)
    }
    if err != nil {
        log.Printf("[DB ERROR][%s] %v | %s | %v", l.dbName, elapsed, sql, err)
    }
}
```

### 9.4 健康检查 (`pkg/database/health.go`)

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

## 10. pkg/app

> 应用启动器 + 优雅关闭 + 健康检查。
>
> **修复的关键 bug**:
> - 健康检查路由在 Auth 之前注册 + 自动加入白名单（避免 K8s 探针被 401）
> - 优雅关闭移除多余的 `<-ctx.Done()`（Shutdown 已阻塞到完成）
> - 不再自带 logger 工厂，委托 `pkg/log`

### 10.1 引擎工厂 (`pkg/app/app.go`)

```go
package app

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    pkgauth "github.com/your-org/grand-canal-guardian/pkg/auth"
    "github.com/your-org/grand-canal-guardian/pkg/log"
    "github.com/your-org/grand-canal-guardian/pkg/middleware"
    pkgratelimit "github.com/your-org/grand-canal-guardian/pkg/ratelimit"
)

// App 应用实例
type App struct {
    Engine *gin.Engine
    Server *http.Server
    Logger *zap.Logger
    Config Config
}

// Config 应用配置
type Config struct {
    Name string
    Host string
    Port int
    Mode string // debug | release | test

    // 中间件依赖
    Logger    *log.Config       // 日志配置（nil 用默认）
    Auth      *pkgauth.AuthMiddleware
    RateLimit gin.HandlerFunc
    CORS      middleware.CORSConfig
    Trace     middleware.TraceConfig
}

// New 创建应用。中间件依赖全部通过依赖注入传入（显式、可测）。
func New(cfg Config) (*App, error) {
    gin.SetMode(cfg.Mode)

    // 创建 Logger（委托 pkg/log，不再自带实现）
    logCfg := log.DefaultConfig(cfg.Name)
    if cfg.Logger != nil {
        logCfg = *cfg.Logger
    }
    logger, err := log.New(logCfg)
    if err != nil {
        return nil, fmt.Errorf("创建 logger 失败: %w", err)
    }

    engine := gin.New()

    // ★ 健康检查路由先注册（在 Apply 之前），不经过任何中间件
    //    这样 K8s 探针无需 Token 也能访问
    engine.GET("/health", HealthHandler)
    engine.GET("/ready", ReadyHandler)

    // 注册中间件链
    middleware.Apply(engine, middleware.Dependencies{
        Logger:    logger,
        Auth:      cfg.Auth,
        RateLimit: cfg.RateLimit,
        CORS:      cfg.CORS,
        Trace:     cfg.Trace,
    })

    return &App{Engine: engine, Logger: logger, Config: cfg}, nil
}

// MountRoutes 挂载业务路由
func (a *App) MountRoutes(register func(r *gin.RouterGroup)) {
    v1 := a.Engine.Group("/api/v1")
    register(v1)
}
```

### 10.2 优雅关闭 (`pkg/app/shutdown.go`)

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

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    select {
    case err := <-errCh:
        return fmt.Errorf("服务启动失败: %w", err)
    case sig := <-quit:
        a.Logger.Info("收到关闭信号", zap.String("signal", sig.String()))
    }

    // 优雅关闭: Shutdown 会阻塞到所有在飞请求处理完或 ctx 超时
    a.Logger.Info("正在优雅关闭...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := a.Server.Shutdown(ctx); err != nil {
        a.Logger.Error("强制关闭", zap.Error(err))
        return err
    }
    // ★ 不再 <-ctx.Done()（原版多余且会多等 30s）
    a.Logger.Info("服务已关闭")
    return nil
}
```

### 10.3 健康检查 (`pkg/app/health.go`)

```go
package app

import (
    "github.com/gin-gonic/gin"
)

// 全局就绪检查器注册表
var readinessChecks []func() error

// RegisterReadinessCheck 注册就绪检查（各服务 init 时调用）
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
            c.JSON(503, gin.H{"status": "not_ready", "reason": err.Error()})
            return
        }
    }
    c.JSON(200, gin.H{"status": "ready"})
}
```

---

## 11. pkg/transaction

> 事务装饰器。
>
> **★ 重大修复**: 原版 `TxFunc` 签名只收 `ctx`，闭包内用的是 `s.db`（非事务连接），两条 Create 根本不在同一事务里。本版改为把 `*gorm.DB`（事务句柄）显式传给闭包。

(`pkg/transaction/tx.go`)

```go
package transaction

import (
    "context"

    "gorm.io/gorm"

    "github.com/your-org/grand-canal-guardian/pkg/errors"
)

// TxFunc 事务函数签名 — 接收事务句柄 tx，所有 DB 操作必须用 tx 而非原 *gorm.DB
type TxFunc func(ctx context.Context, tx *gorm.DB) error

// WithTx 在事务中执行业务逻辑（自动 commit/rollback）。
// 用法:
//
//	err := transaction.WithTx(ctx, s.db, func(ctx context.Context, tx *gorm.DB) error {
//	    if e := tx.WithContext(ctx).Create(post).Error; e != nil {
//	        return errors.WrapDefault(errors.ErrDatabaseError, e)
//	    }
//	    if e := tx.WithContext(ctx).Create(review).Error; e != nil {
//	        return errors.WrapDefault(errors.ErrDatabaseError, e)
//	    }
//	    return nil
//	})
func WithTx(ctx context.Context, db *gorm.DB, fn TxFunc) error {
    tx := db.WithContext(ctx).Begin()
    if tx.Error != nil {
        return errors.WrapDefault(errors.ErrDatabaseError, tx.Error)
    }

    committed := false
    defer func() {
        if !committed {
            // panic 或 return error 时回滚
            _ = tx.Rollback().Error
        }
    }()

    if err := fn(ctx, tx); err != nil {
        return err // defer 会 rollback
    }

    if err := tx.Commit().Error; err != nil {
        return errors.WrapDefault(errors.ErrDatabaseError, err)
    }
    committed = true
    return nil
}

// TxResultFunc 带返回值的事务函数
type TxResultFunc[T any] func(ctx context.Context, tx *gorm.DB) (T, error)

func WithTxResult[T any](ctx context.Context, db *gorm.DB, fn TxResultFunc[T]) (T, error) {
    var zero T
    tx := db.WithContext(ctx).Begin()
    if tx.Error != nil {
        return zero, errors.WrapDefault(errors.ErrDatabaseError, tx.Error)
    }

    committed := false
    defer func() {
        if !committed {
            _ = tx.Rollback().Error
        }
    }()

    result, err := fn(ctx, tx)
    if err != nil {
        return zero, err
    }

    if err := tx.Commit().Error; err != nil {
        return zero, errors.WrapDefault(errors.ErrDatabaseError, err)
    }
    committed = true
    return result, nil
}
```

**Service 层正确用法**:

```go
func (s *ContentService) CreatePostWithReview(ctx context.Context, req *CreatePostReq) (*Post, error) {
    return transaction.WithTxResult(ctx, s.db, func(ctx context.Context, tx *gorm.DB) (*Post, error) {
        post := &Post{AuthorID: req.AuthorID, Title: req.Title, Content: req.Content, Status: "pending"}
        if err := tx.WithContext(ctx).Create(post).Error; err != nil { // ★ 用 tx
            return nil, errors.WrapDefault(errors.ErrDatabaseError, err)
        }
        review := &Review{AuditID: post.ID, Status: "pending"}
        if err := tx.WithContext(ctx).Create(review).Error; err != nil { // ★ 用 tx
            return nil, errors.WrapDefault(errors.ErrDatabaseError, err)
        }
        return post, nil
    })
}
```

---

## 12. pkg/grpc

> 跨服务错误传播：AppError ↔ gRPC Status。

### 12.1 错误转换 (`pkg/grpc/error_convert.go`)

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
        return status.Error(httpToGRPCCode(appErr.Code.HTTPStatus()), appErr.Message)
    }
    return status.Errorf(codes.Internal, "内部错误")
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
    return errors.New(gRPCCodeToErrorCode(st.Code()), st.Message())
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
    case httpStatus == 409:
        return codes.AlreadyExists
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
    case codes.AlreadyExists:
        return errors.ErrConflict
    case codes.ResourceExhausted:
        return errors.ErrTooManyRequest
    case codes.Unavailable:
        return errors.ErrLLMUnavailable
    default:
        return errors.ErrInternal
    }
}
```

### 12.2 拦截器 (`pkg/grpc/interceptors.go`)

```go
package grpc

import (
    "context"
    "time"

    "go.uber.org/zap"
    "google.golang.org/grpc"
)

// UnaryServerInterceptor 服务端: 日志 + 错误转换
func UnaryServerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        resp, err := handler(ctx, req)
        logger.Info("gRPC 调用完成",
            zap.String("method", info.FullMethod),
            zap.Duration("latency", time.Since(start)),
            zap.Error(err),
        )
        if err != nil {
            return resp, ToGRPCStatus(err)
        }
        return resp, nil
    }
}

// UnaryClientInterceptor 客户端: gRPC Status → AppError
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        err := invoker(ctx, method, req, reply, cc, opts...)
        if err != nil {
            return FromGRPCStatus(err)
        }
        return nil
    }
}
```

---

## 13. swaggo

> OpenAPI 自动生成。

### 13.1 注解示例

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
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterRequest  true  "注册信息"
// @Success      201   {object}  response.R
// @Failure      409   {object}  response.R
// @Router       /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) { /* ... */ }

// GetProfile godoc
// @Summary      获取当前用户信息
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.R
// @Router       /users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) { /* ... */ }
```

### 13.2 Makefile

```makefile
SWAGGER_SERVICES = user-service content-service map-service quiz-service

swagger-%:
	cd services/$* && swag init -g cmd/main.go -o docs --parseDependency --parseInternal

swagger-all:
	@for svc in $(SWAGGER_SERVICES); do \
		echo ">>> 生成 $$svc swagger..."; \
		(cd services/$$svc && swag init -g cmd/main.go -o docs --parseDependency --parseInternal); \
	done
```

---

## 14. 统一中间件链注册

> 所有服务通过 `middleware.Apply` 注册中间件，顺序固定，依赖显式注入。
> 本节是 `pkg/app.New` 内部调用的展开说明。

**顺序**: `Recovery → RequestID → Logger → CORS → Trace → Auth → RateLimit`

```
请求进入
   │
   ▼
1. Recovery        (pkg/log)       — 捕获所有 panic，HTTP 响应不泄露
   │
   ▼
2. RequestID       (pkg/middleware) — 注入/透传 X-Request-ID
   │
   ▼
3. Logger          (pkg/log)        — 结构化请求日志（需要 request_id）
   │
   ▼
4. CORS            (pkg/middleware) — 跨域
   │
   ▼
5. Trace           (pkg/middleware) — OpenTelemetry 链路追踪
   │
   ▼
6. Auth            (pkg/auth)       — JWT 鉴权（含黑名单 + RBAC；白名单跳过）
   │
   ▼
7. RateLimit       (pkg/ratelimit)  — 三级令牌桶限流
   │
   ▼
Handler (业务逻辑)
```

**关键点**:
- `/health`、`/ready` 在 `middleware.Apply` 之前注册，不经过任何中间件
- `Recovery` 最内层（第一个注册），确保后续所有中间件和 handler 的 panic 都被捕获
- `Auth`、`RateLimit` 最外层（最后注册），尽早拦截无效请求

---

## 15. 完整服务示例 — user-service

### 15.1 main.go

```go
// services/user-service/cmd/main.go
package main

import (
    "log"
    "os"
    "time"

    "github.com/gin-gonic/gin"

    pkgauth "github.com/your-org/grand-canal-guardian/pkg/auth"
    "github.com/your-org/grand-canal-guardian/pkg/app"
    pkgerrors "github.com/your-org/grand-canal-guardian/pkg/errors"
    "github.com/your-org/grand-canal-guardian/pkg/middleware"
    pkgratelimit "github.com/your-org/grand-canal-guardian/pkg/ratelimit"
    "github.com/your-org/grand-canal-guardian/pkg/validator"
    "github.com/your-org/grand-canal-guardian/services/user-service/internal/config"
    "github.com/your-org/grand-canal-guardian/services/user-service/internal/handler"
    "github.com/your-org/grand-canal-guardian/services/user-service/internal/wire"
)

func main() {
    cfg := config.Load()

    // 1. 设置错误脱敏模式
    if cfg.GinMode == "release" {
        pkgerrors.SetMode("release")
    } else {
        pkgerrors.SetMode("debug")
    }

    // 2. 注册自定义校验器
    validator.RegisterCustomValidators()

    // 3. 依赖注入（DB / Redis / TokenManager / Service）
    deps, err := wire.InitializeApp(cfg)
    if err != nil {
        log.Fatalf("初始化依赖失败: %v", err)
    }

    // 4. 构造 Auth 中间件（白名单: 公开路由）
    authMW := pkgauth.NewAuthMiddleware(
        deps.TokenManager,
        []string{"/api/v1/auth/register", "/api/v1/auth/login", "/api/v1/auth/refresh"},
        []string{"/api/v1/public/"}, // 公开前缀
    )

    // 5. 构造 RateLimit 中间件
    rateLimitMW := pkgratelimit.Middleware(deps.Limiter, pkgratelimit.DefaultTierConfig())

    // 6. 创建应用
    application, err := app.New(app.Config{
        Name:      "user-service",
        Host:      "0.0.0.0",
        Port:      cfg.Port,
        Mode:      cfg.GinMode,
        Auth:      authMW,
        RateLimit: rateLimitMW,
        CORS:      middleware.DefaultCORSConfig(),
        Trace:     middleware.TraceConfig{Enabled: true, ServiceName: "user-service"},
    })
    if err != nil {
        log.Fatalf("创建应用失败: %v", err)
    }

    // 7. 注册就绪检查
    app.RegisterReadinessCheck(deps.DB.Ping)
    app.RegisterReadinessCheck(deps.Redis.Ping)

    // 8. 挂载路由
    application.MountRoutes(func(v1 *gin.RouterGroup) {
        h := handler.NewUserHandler(deps.UserService)
        h.RegisterRoutes(v1)
    })

    // 9. 启动 → 等待信号 → 优雅关闭
    if err := application.Run(); err != nil {
        log.Fatalf("服务异常退出: %v", err)
    }
}

func ginMode() string {
    if m := os.Getenv("GIN_MODE"); m != "" {
        return m
    }
    return "debug"
}

// RoleFetcher 实现，供 TokenManager 在 refresh 时查 DB
type roleFetcher struct{ repo interface{ GetRole(string) (string, error) } }

// 用法（在 wire 中）:
//   tm := auth.NewTokenManager(cfg.JWTSecret, rdb,
//       15*time.Minute, 7*24*time.Hour, &roleFetcher{repo})
```

### 15.2 Handler 层

```go
// services/user-service/internal/handler/user_handler.go
package handler

import (
    "github.com/gin-gonic/gin"

    "github.com/your-org/grand-canal-guardian/pkg/errors"
    "github.com/your-org/grand-canal-guardian/pkg/response"
    pkgauth "github.com/your-org/grand-canal-guardian/pkg/auth"
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
        auth.POST("/logout", h.Logout) // 需要 Auth 中间件（白名单不含 logout）
    }

    users := r.Group("/users")
    users.Use(pkgauth.RequirePermission(0)) // 占位: 实际用 RequireRole("user","monitor","admin")
    {
        users.GET("/me", h.GetProfile)
        users.PUT("/me", h.UpdateProfile)
        users.GET("/:id", h.GetUserByID)
    }

    admin := r.Group("/admin/users")
    admin.Use(pkgauth.RequireRole("admin"))
    {
        admin.GET("", h.ListUsers)
        admin.PUT("/:id/role", h.ChangeRole)
        admin.POST("/:id/ban", h.BanUser)
    }
}

func (h *UserHandler) Register(c *gin.Context) {
    var req model.RegisterRequest
    if err := validator.BindAndValidate(c, &req); err != nil {
        response.Error(c, err)
        return
    }
    user, tokens, err := h.svc.Register(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err)
        return
    }
    response.Created(c, gin.H{
        "user":          user,
        "access_token":  tokens.AccessToken,
        "refresh_token": tokens.RefreshToken,
        "token_type":    tokens.TokenType,
        "expires_in":    tokens.ExpiresIn,
    })
}

func (h *UserHandler) GetProfile(c *gin.Context) {
    profile, err := h.svc.GetProfile(c.Request.Context(), pkgauth.GetUserID(c))
    if err != nil {
        response.Error(c, err)
        return
    }
    response.OK(c, profile)
}

// ListUsers 管理员: 用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
    page := validator.QueryInt(c, "page", 1)
    pageSize := validator.QueryInt(c, "page_size", 20)
    role := c.Query("role")

    users, total, err := h.svc.ListUsers(c.Request.Context(), page, pageSize, role)
    if err != nil {
        response.Error(c, err)
        return
    }
    response.OKList(c, users, page, pageSize, total)
}
```

### 15.3 Service 层

```go
// services/user-service/internal/service/user_service.go
package service

import (
    "context"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"

    pkgauth "github.com/your-org/grand-canal-guardian/pkg/auth"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
    "github.com/your-org/grand-canal-guardian/services/user-service/internal/model"
    "github.com/your-org/grand-canal-guardian/services/user-service/internal/repository"
)

type UserService struct {
    repo  *repository.UserRepo
    tm    *pkgauth.TokenManager
}

func NewUserService(repo *repository.UserRepo, tm *pkgauth.TokenManager) *UserService {
    return &UserService{repo: repo, tm: tm}
}

func (s *UserService) Register(ctx context.Context, req *model.RegisterRequest) (*model.UserProfile, *pkgauth.TokenPair, error) {
    exists, err := s.repo.ExistsByUsername(ctx, req.Username)
    if err != nil {
        return nil, nil, errors.WrapDefault(errors.ErrDatabaseError, err)
    }
    if exists {
        return nil, nil, errors.NewDefault(errors.ErrUsernameExists)
    }

    hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
    if err != nil {
        return nil, nil, errors.Internal(err)
    }

    user := &model.User{
        Username: req.Username,
        Password: string(hashed),
        Email:    req.Email,
        Nickname: req.Nickname,
        Role:     "user", // ★ 固定 user，不接受客户端传入的 role（防越权）
    }
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, nil, errors.WrapDefault(errors.ErrDatabaseError, err)
    }

    tokens, err := s.tm.IssueTokens(ctx, user.ID, user.Role, "default")
    if err != nil {
        return nil, nil, err
    }
    return user.ToProfile(), tokens, nil
}

func (s *UserService) Login(ctx context.Context, req *model.LoginRequest) (*model.UserProfile, *pkgauth.TokenPair, error) {
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

    tokens, err := s.tm.IssueTokens(ctx, user.ID, user.Role, "default")
    if err != nil {
        return nil, nil, err
    }
    return user.ToProfile(), tokens, nil
}

func (s *UserService) Logout(ctx context.Context, claims *pkgauth.Claims) error {
    return s.tm.Logout(ctx, claims)
}
```

> **说明**: Token 签发已收敛到 `pkg/auth.TokenManager`，原 doc 04 中 `generateTokens` 自行用 `jwt.NewWithClaims` 的版本已移除（避免 Claims 字段冲突）。

---

## 16. 附录：go.mod 依赖清单

```go
// pkg/go.mod
module github.com/your-org/grand-canal-guardian/pkg

go 1.22

require (
    github.com/gin-contrib/cors                    v1.7.2
    github.com/gin-gonic/gin                       v1.10.0
    github.com/go-playground/locales               v0.14.1
    github.com/go-playground/universal-translator  v0.18.1
    github.com/go-playground/validator/v10         v10.22.0
    github.com/golang-jwt/jwt/v5                   v5.2.1
    github.com/google/uuid                         v1.6.0
    github.com/prometheus/client_golang            v1.20.0
    github.com/redis/go-redis/v9                   v9.5.4
    go.opentelemetry.io/otel                       v1.28.0
    go.opentelemetry.io/otel/trace                 v1.28.0
    go.uber.org/zap                                v1.27.0
    golang.org/x/crypto                            v0.25.0
    google.golang.org/grpc                         v1.65.0
    gopkg.in/natefinch/lumberjack.v2               v2.2.1
    gorm.io/driver/postgres                        v1.5.9
    gorm.io/gorm                                   v1.25.11
    gorm.io/plugin/dbresolver                      v1.5.3
)
```

### 完整 import 参考（补全各文件缺失的 import）

```go
// pkg/ratelimit/limiter.go
import (
    "context"
    _ "embed"
    "fmt"
    "strconv"
    "time"
    "github.com/redis/go-redis/v9"
)

// pkg/database/pool.go
import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/plugin/dbresolver"
)

// pkg/log/sanitize.go
import (
    "fmt"
    "regexp"
    "github.com/your-org/grand-canal-guardian/pkg/errors"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)
```

---

## 变更摘要

本次合并相对原三份草稿的关键变更:

| 变更 | 原文档 | 本文档 |
|------|--------|--------|
| Auth 中间件 | doc 04 `pkg/middleware/auth.go` + doc 05 `pkg/auth/middleware.go` 两套 | 仅 `pkg/auth/middleware.go`，删除 doc 04 版本 |
| RateLimit | doc 04 `pkg/middleware/rate_limit.go`（固定窗口）+ doc 05 Lua 令牌桶 | 仅 `pkg/ratelimit/`，删除 doc 04 版本 |
| Logger 工厂 | doc 04 `pkg/app/logger.go` + doc 05 `pkg/log/logger.go` | 仅 `pkg/log/`，`pkg/app` 委托调用 |
| Recovery | doc 04 + doc 05 `GinRecovery` + doc 06 patch 三份 | 合并进 `pkg/log/middleware.go`（脱敏版） |
| 错误脱敏 | doc 06 为补丁 | 内置进 `pkg/errors/app_error.go` + `sanitize.go` |
| Response.Error | doc 04 单模式 + doc 06 双模式 | 内置双模式 |
| JWT Claims | 含 `UserID json:"sub"` 与 `RegisteredClaims.Subject` 冲突 | 删除 `UserID`，统一用 `Subject` |
| Refresh 重放检测 | 访问 nil session panic | 通过 jti 索引反查 userID |
| Refresh 后 role | 丢失（session 未存 role） | 查 DB 取最新 role |
| 事务装饰器 | 闭包用 `s.db`，事务形同虚设 | 闭包接收 `tx *gorm.DB` |
| 限流阈值 | 100/s vs 20/s 矛盾 | 统一 100/s（对齐架构文档） |
| Prometheus 累计值 | Counter.Add 全量 → rate 错 | Gauge 采样累计值 |
| GORM Callback | 注册名错误且未被调用 | 正确名称 + NewDB 中调用 |
| 健康检查 | 在 Auth 之后注册 → 探针 401 | 在 Auth 之前注册 |
| 优雅关闭 | 多余 `<-ctx.Done()` 多等 30s | 删除 |
| `queryInt` | 永远返回默认值 | 真正解析 |
| validator | `zh := zh.New()` 变量遮蔽 | 改名 `zhLocale` |
| `go.work` + `replace` | 冗余并存 | 仅 `go.work` |

---

> **下一步建议**: 原 `05-production-grade-infra.md` 与 `06-error-desensitization-patch.md` 可归档至 `docs/archive/` 或删除。`03-developer-guide.md` 第 3 节（Go 后端开发）应引用本文件，删除其中重复的工程化代码片段。
