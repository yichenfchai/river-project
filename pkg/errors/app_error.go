package errors

import "fmt"

// AppError 统一应用错误
type AppError struct {
	Code    ErrorCode   `json:"code"`              // 错误码
	Message string      `json:"message"`           // 用户可读消息
	Detail  string      `json:"detail,omitempty"`  // 调试详情 (仅 debug 模式序列化)
	Safe    *SafeDetail `json:"safe,omitempty"`    // 生产环境安全摘要
	Cause   error       `json:"-"`                 // 原始错误 (永不序列化)
}

// SafeDetail 生产环境安全分类
type SafeDetail struct {
	Type    string `json:"type,omitempty"`    // 错误类型 (如 "database", "validation")
	Summary string `json:"summary,omitempty"` // 安全摘要
}

// Mode 环境模式
var Mode = "release"

// SetMode 由启动时调用
func SetMode(mode string) { Mode = mode }

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
	return &AppError{
		Code:    code,
		Message: code.GetMessage(),
		Cause:   cause,
		Detail:  sanitizeDetail(code, cause),
		Safe:    classifyError(code, cause),
	}
}

func Wrapf(code ErrorCode, cause error, format string, args ...any) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Cause:   cause,
		Detail:  sanitizeDetail(code, cause),
		Safe:    classifyError(code, cause),
	}
}

// ---- 快捷构造 (使用默认中文消息) ----

func BadRequest(msg string) *AppError  { return New(ErrBadRequest, msg) }
func NotFound(r string) *AppError      { return Newf(ErrNotFound, "%s 不存在", r) }
func Unauthorized(msg string) *AppError { return New(ErrUnauthorized, msg) }
func Forbidden(msg string) *AppError   { return New(ErrForbidden, msg) }
func Conflict(msg string) *AppError    { return New(ErrConflict, msg) }
func TooManyRequest() *AppError        { return NewDefault(ErrTooManyRequest) }
func Internal(cause error) *AppError {
	return &AppError{
		Code:    ErrInternal,
		Message: ErrInternal.GetMessage(),
		Cause:   cause,
		Detail:  "internal_error",
		Safe:    &SafeDetail{Type: "internal", Summary: "internal_server_error"},
	}
}

func IsAppErr(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	var appErr *AppError
	// Try direct type assertion first, then unwrap
	if e, ok := err.(*AppError); ok {
		return e, true
	}
	// Try unwrapping to find AppError in chain
	for {
		if e, ok := err.(*AppError); ok {
			appErr = e
			break
		}
		err = fmt.Errorf("unwrap: %w", err)
		if err == nil {
			break
		}
	}
	return appErr, appErr != nil
}
