# 错误链脱敏补丁

> ⚠️ **本文档已废弃（v1.1, 2026-06-18）**
>
> 本补丁的内容已内置进 [`04-go-engineering-standards.md`](./04-go-engineering-standards.md) 的：
> - `pkg/errors/app_error.go`（`Detail`/`Safe`/`Mode` 双模式）
> - `pkg/errors/sanitize.go`（脱敏引擎）
> - `pkg/response/response.go`（双模式 `Error`）
> - `pkg/log/middleware.go`（脱敏版 `Recovery`）
>
> 同时修复了原补丁中 `regexp.MustCompile` 参数错误（漏括号导致无法编译）。
>
> **请以 `04-go-engineering-standards.md` 为准。**
>
> ---
>
> 原始记录（供参考）:
> **适用范围**: `pkg/errors/` + `pkg/response/` + `pkg/middleware/recovery.go`
>
> 当前问题: `WrapDefault` 无条件将原始 error 字符串塞入 `Detail`，日志和异常监控直接暴露数据库结构。

---

## 修复 1: AppError — Detail 分级 (`pkg/errors/app_error.go` 补丁)

```go
// ★ 新增: 环境模式控制
var Mode = "release" // debug | release (由 main.go 设置)

// SetMode 由启动时调用
func SetMode(mode string) {
    Mode = mode
}

// SafeDetail 生产环境安全摘要 (替换原始 error 文本)
type SafeDetail struct {
    Type    string `json:"type,omitempty"`    // 错误类型 (如 "database", "validation")
    Summary string `json:"summary,omitempty"` // 安全摘要
}

// AppError (补丁: 新增 Safe + 生产环境 Detail 清理)
type AppError struct {
    Code    ErrorCode   `json:"code"`
    Message string      `json:"message"`
    Detail  string      `json:"detail,omitempty"` // ★ 仅 debug 模式序列化
    Safe    *SafeDetail `json:"safe,omitempty"`   // ★ 新增: 生产安全摘要
    Cause   error       `json:"-"`                // 永远不序列化
}

// ★ 修复: WrapDefault — Detail 不再裸存原始 error
func WrapDefault(code ErrorCode, cause error) *AppError {
    return &AppError{
        Code:    code,
        Message: code.GetMessage(),
        Cause:   cause,
        Detail:  sanitizeDetail(code, cause), // ★ 脱敏后存储
        Safe:    classifyError(code, cause),  // ★ 安全分类
    }
}

// ★ 修复: Wrap — 同上
func Wrap(code ErrorCode, message string, cause error) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
        Cause:   cause,
        Detail:  sanitizeDetail(code, cause),
        Safe:    classifyError(code, cause),
    }
}

// ★ 新增: Internal — 不传原始 error 文本
func Internal(cause error) *AppError {
    return &AppError{
        Code:    ErrInternal,
        Message: ErrInternal.GetMessage(),
        Cause:   cause,
        Detail:  "internal_error",  // ★ 固定值，不含原始错误
        Safe:    &SafeDetail{Type: "internal"},
    }
}
```

## 修复 2: 脱敏引擎 (`pkg/errors/sanitize.go` — 新文件)

```go
package errors

import (
    "regexp"
    "strings"
)

// 泄漏模式 → 替换
var leakPatterns = []struct {
    Pattern *regexp.Regexp
    Replace string
}{
    // PostgreSQL 约束名
    {regexp.MustCompile(`"([a-z_]+)_([a-z_]+)_key"`), `"***_key"`},
    // PostgreSQL 表名+列名
    {regexp.MustCompile(`relation "([a-z_]+)"`), `relation "***"`},
    {regexp.MustCompile(`column "([a-z_]+)"`), `column "***"`},
    // 重复键值
    {regexp.MustCompile(`Key \(([^)]+)\)=\((.*?)\)`), `Key ($1)=(***)`},
    // MySQL 表名
    {regexp.MustCompile(`Duplicate entry '([^']+)' for key`), `Duplicate entry '***' for key`},
    // 连接信息 (host/port/dbname)
    {regexp.MustCompile(`(host=[^\s]+)`, `host=***`},
    {regexp.MustCompile(`(dbname=[^\s]+)`, `dbname=***`},
    // Panic 中的具体值 (保留类型，隐去值)
    {regexp.MustCompile(`index out of range \[\d+\]`), `index out of range [N]`},
    {regexp.MustCompile(`invalid memory address (0x[0-9a-f]+)`), `invalid memory address ***`},
    // UUID/数字型 ID
    {regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`), `<uuid>`},
}

// sanitizeDetail 脱敏错误详情
func sanitizeDetail(code ErrorCode, cause error) string {
    if cause == nil {
        return code.GetMessage()
    }

    raw := cause.Error()

    // 内部错误: 不返回任何原始消息
    if code == ErrInternal || code == ErrDatabaseError {
        return code.GetMessage()
    }

    // 其他错误: 执行脱敏后返回
    for _, rule := range leakPatterns {
        raw = rule.Pattern.ReplaceAllString(raw, rule.Replace)
    }

    // 截断长度 (防超长注入)
    if len(raw) > 200 {
        raw = raw[:200] + "..."
    }

    return raw
}

// classifyError 生产环境安全分类 (用于监控聚合，不含具体数据)
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

## 修复 3: response.Error — 生产/调试双模式 (`pkg/response/response.go` 补丁)

```go
// ★ 改造: Error 支持 Detail/Safe 双模式
func Error(c *gin.Context, err *errors.AppError) {
    status := err.Code.HTTPStatus()

    body := R{
        Code:    int(err.Code),
        Message: err.Message,
    }

    // debug 模式: 返回脱敏后的 Detail
    if errors.Mode == "debug" && err.Detail != "" {
        body.Data = map[string]interface{}{
            "detail": err.Detail,
        }
    }

    // 生产模式: 返回 Safe (用于前端错误分类，不泄露任何原始信息)
    if errors.Mode == "release" && err.Safe != nil {
        // Safe 通过扩展字段传递 (OpenAPI 中定义为 optional)
        body.Data = map[string]interface{}{
            "error_type": err.Safe.Type,
        }
    }

    c.JSON(status, body)
    c.Abort()
}
```

## 修复 4: Recovery — 不向外泄露 (`pkg/middleware/recovery.go` 补丁)

```go
// ★ 修复: Recovery — panic 仅记日志，HTTP 响应不含任何原始信息
func Recovery(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                stack := string(debug.Stack())

                // ★ 日志记录完整信息 (服务器内部，受访问控制保护)
                logger.Error("panic recovered",
                    zap.String("request_id", GetRequestID(c)),
                    zap.String("method", c.Request.Method),
                    zap.String("path", c.Request.URL.Path),
                    zap.Any("panic", r),
                    zap.String("stack", stack),
                )

                // ★ HTTP 响应: 只返回通用内部错误 (不含任何原始信息)
                response.Error(c, &errors.AppError{
                    Code:    errors.ErrInternal,
                    Message: "服务器内部错误",
                    // ★ Detail 和 Cause 均为空 — 不泄露
                    Safe: &errors.SafeDetail{
                        Type:    "panic",
                        Summary: "internal_server_error",
                    },
                })
                c.Abort()
            }
        }()
        c.Next()
    }
}
```

## 修复 5: 日志脱敏中间件 (`pkg/log/sanitize.go` — 追加)

```go
// ★ 追加: 脱敏所有日志中的 AppError
// 在 Logger 中间件中使用

// SanitizeError 脱敏 error 字段 (用于 zap.Error)
func SanitizeError(err error) zap.Field {
    if appErr, ok := err.(*errors.AppError); ok {
        // 日志中只输出 Code + Message，不含 Detail/Cause
        return zap.Object("error", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
            enc.AddInt("code", int(appErr.Code))
            enc.AddString("message", appErr.Message)
            if appErr.Safe != nil {
                enc.AddString("error_type", appErr.Safe.Type)
            }
            return nil
        }))
    }
    // 非 AppError → 只输出类型，不输出消息 (可能是 driver 原始错误)
    return zap.String("error_type", fmt.Sprintf("%T", err))
}
```

## 修复 6: 异常监控 (Sentry) 集成保护

```go
// pkg/errors/sentry.go
// 如果要接入 Sentry/异常监控，必须在发送前脱敏

func ToSentryBreadcrumb(err *AppError) *sentry.Breadcrumb {
    return &sentry.Breadcrumb{
        Category: "app_error",
        Data: map[string]interface{}{
            "code":       int(err.Code),
            "message":    err.Message,
            "error_type": safeType(err),
            // ★ 不含 Detail, Cause, Stack
        },
    }
}

func safeType(err *AppError) string {
    if err.Safe != nil {
        return err.Safe.Type
    }
    return "unknown"
}
```

---

## 脱敏前后对比

```
场景: 用户注册重复用户名 (admin 已存在)

┌─ 脱敏前 ─────────────────────────────────────────────┐
│ 日志: [10801] 数据库错误: ERROR: duplicate key value │
│       violates unique constraint "users_username_key"│
│       (SQLSTATE 23505)                                │
│                                                       │
│ Sentry: 同上完整错误                                   │
│ HTTP:   {"code":10801, "message":"数据库错误"}        │
│         (Detail 不在 R 中，但日志/Sentry 已泄漏)       │
└───────────────────────────────────────────────────────┘

┌─ 脱敏后 ─────────────────────────────────────────────┐
│ 日志: {"code":10801, "message":"数据库错误",          │
│        "error_type":"unique_violation"}               │
│                                                       │
│ Sentry: {code:10801, error_type:"unique_violation"}   │
│ HTTP:   {"code":10801, "message":"数据库错误",        │
│          "data":{"error_type":"unique_violation"}}    │
│                                                       │
│ ★ 前端可根据 error_type 展示友好提示:                  │
│   "该用户名已被注册，请更换"                           │
└───────────────────────────────────────────────────────┘
```

---

## 主函数初始化

```go
// cmd/main.go
func main() {
    // ★ 第一行: 设置错误脱敏模式
    ginMode := os.Getenv("GIN_MODE")
    if ginMode == "release" {
        errors.SetMode("release")
    } else {
        errors.SetMode("debug")
    }

    // ... 其余初始化
}
```
