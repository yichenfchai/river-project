package errors

import (
	"fmt"
	"regexp"
	"strings"
)

var Mode = "release"

func SetMode(mode string) { Mode = mode }

type SafeDetail struct {
	Type    string `json:"type,omitempty"`
	Summary string `json:"summary,omitempty"`
}

type AppError struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Detail  string      `json:"-"`
	Safe    *SafeDetail `json:"-"`
	Cause   error       `json:"-"`
}

func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Cause }

func New(code ErrorCode, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func Newf(code ErrorCode, format string, args ...any) *AppError {
	return &AppError{Code: code, Message: fmt.Sprintf(format, args...)}
}

func NewDefault(code ErrorCode) *AppError {
	return New(code, code.GetMessage())
}

func Wrap(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Detail:  sanitizeDetail(code, cause),
		Safe:    classifyError(code, cause),
	}
}

func WrapDefault(code ErrorCode, cause error) *AppError {
	return Wrap(code, code.GetMessage(), cause)
}

func BadRequest(msg string) *AppError  { return New(ErrBadRequest, msg) }
func NotFound(resource string) *AppError { return Newf(ErrNotFound, "%s 不存在", resource) }
func Unauthorized(msg string) *AppError  { return New(ErrUnauthorized, msg) }
func Forbidden(msg string) *AppError     { return New(ErrForbidden, msg) }
func Conflict(msg string) *AppError      { return New(ErrConflict, msg) }

func Internal(cause error) *AppError {
	return &AppError{
		Code:    ErrInternal,
		Message: ErrInternal.GetMessage(),
		Cause:   cause,
		Detail:  "internal_error",
		Safe:    &SafeDetail{Type: "internal", Summary: "internal_error"},
	}
}

var leakPatterns = []struct {
	Pattern *regexp.Regexp
	Replace string
}{
	{regexp.MustCompile(`"[a-z_]+_[a-z_]+_key"`), `"***_key"`},
	{regexp.MustCompile(`relation "([a-z_]+)"`), `relation "***"`},
	{regexp.MustCompile(`column "([a-z_]+)"`), `column "***"`},
	{regexp.MustCompile(`Key \(([^)]+)\)=\([^)]*\)`), `Key ($1)=(***)`},
	{regexp.MustCompile(`Duplicate entry '([^']+)' for key`), `Duplicate entry '***' for key`},
	{regexp.MustCompile(`host=[^\s]+`), `host=***`},
	{regexp.MustCompile(`dbname=[^\s]+`), `dbname=***`},
	{regexp.MustCompile(`user=[^\s]+`), `user=***`},
	{regexp.MustCompile(`password=[^\s]+`), `password=***`},
	{regexp.MustCompile(`index out of range \[\d+:\d+\]`), `index out of range [N:N]`},
	{regexp.MustCompile(`invalid memory address 0x[0-9a-f]+`), `invalid memory address ***`},
	{regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`), `<uuid>`},
}

func sanitizeDetail(code ErrorCode, cause error) string {
	if cause == nil {
		return code.GetMessage()
	}
	if code == ErrInternal || code == ErrDatabaseError {
		return code.GetMessage()
	}
	raw := cause.Error()
	for _, rule := range leakPatterns {
		raw = rule.Pattern.ReplaceAllString(raw, rule.Replace)
	}
	if len(raw) > 200 {
		raw = raw[:200] + "..."
	}
	return raw
}

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
