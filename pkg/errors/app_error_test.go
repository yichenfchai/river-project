package errors

import (
	"errors"
	"testing"
)

// ─── 基础构造函数 ───

func TestNew(t *testing.T) {
	e := New(ErrBadRequest, "参数错误")
	if e.Code != ErrBadRequest {
		t.Errorf("Code = %d, want %d", e.Code, ErrBadRequest)
	}
	if e.Message != "参数错误" {
		t.Errorf("Message = %q, want %q", e.Message, "参数错误")
	}
	if e.Cause != nil {
		t.Error("Cause should be nil")
	}
}

func TestNewf(t *testing.T) {
	e := Newf(ErrNotFound, "%s 不存在", "用户")
	if e.Message != "用户 不存在" {
		t.Errorf("Message = %q, want %q", e.Message, "用户 不存在")
	}
}

func TestNewDefault(t *testing.T) {
	e := NewDefault(ErrUnauthorized)
	if e.Message != ErrUnauthorized.GetMessage() {
		t.Errorf("Message = %q, want %q", e.Message, ErrUnauthorized.GetMessage())
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("原始错误")
	e := Wrap(ErrDatabaseError, "数据库查询失败", cause)
	
	if e.Message != "数据库查询失败" {
		t.Errorf("Message = %q", e.Message)
	}
	if e.Cause != cause {
		t.Error("Cause should be the original error")
	}
	if e.Safe == nil {
		t.Error("Safe should be set for database errors")
	}
}

func TestWrapDefault(t *testing.T) {
	cause := errors.New("sql: connection refused")
	e := WrapDefault(ErrDatabaseError, cause)
	
	if e.Message != ErrDatabaseError.GetMessage() {
		t.Errorf("Message = %q, want default message", e.Message)
	}
}

// ─── 便捷函数 ───

func TestBadRequest(t *testing.T) {
	e := BadRequest("字段缺失")
	if e.Code != ErrBadRequest {
		t.Errorf("Code = %d", e.Code)
	}
}

func TestNotFound(t *testing.T) {
	e := NotFound("商品")
	if e.Code != ErrNotFound {
		t.Errorf("Code = %d", e.Code)
	}
	if e.Message != "商品 不存在" {
		t.Errorf("Message = %q", e.Message)
	}
}

func TestUnauthorized(t *testing.T) {
	e := Unauthorized("密码错误")
	if e.Code != ErrUnauthorized {
		t.Errorf("Code = %d", e.Code)
	}
}

func TestForbidden(t *testing.T) {
	e := Forbidden("权限不足")
	if e.Code != ErrForbidden {
		t.Errorf("Code = %d", e.Code)
	}
}

func TestConflict(t *testing.T) {
	e := Conflict("用户名已存在")
	if e.Code != ErrConflict {
		t.Errorf("Code = %d", e.Code)
	}
}

func TestInternal(t *testing.T) {
	cause := errors.New("nil pointer dereference")
	e := Internal(cause)
	
	if e.Code != ErrInternal {
		t.Errorf("Code = %d, want %d", e.Code, ErrInternal)
	}
	if e.Safe == nil {
		t.Error("Internal error should have Safe detail")
	}
	if e.Safe.Type != "internal" {
		t.Errorf("Safe.Type = %q, want %q", e.Safe.Type, "internal")
	}
}

// ─── Error() 格式化 ───

func TestAppError_Error_WithDetail(t *testing.T) {
	e := &AppError{
		Code:    ErrUserNotFound,
		Message: "用户不存在",
		Detail:  "user_id=123",
	}
	expected := "[10201] 用户不存在: user_id=123"
	if e.Error() != expected {
		t.Errorf("Error() = %q, want %q", e.Error(), expected)
	}
}

func TestAppError_Error_NoDetail(t *testing.T) {
	e := NewDefault(ErrInternal)
	if e.Error() != "[10001] 服务器内部错误" {
		t.Errorf("Error() = %q", e.Error())
	}
}

// ─── Unwrap ───

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("底层错误")
	e := Wrap(ErrDatabaseError, "封装", cause)
	
	if !errors.Is(e, cause) {
		t.Error("errors.Is should find the wrapped cause")
	}
	
	var appErr *AppError
	if !errors.As(e, &appErr) {
		t.Error("errors.As should find AppError")
	}
}

// ─── 错误码 ───

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		code    ErrorCode
		message string
	}{
		{ErrInternal, "服务器内部错误"},
		{ErrBadRequest, "请求参数错误"},
		{ErrNotFound, "资源不存在"},
		{ErrUnauthorized, "请先登录"},
		{ErrForbidden, "无权执行此操作"},
		{ErrTokenExpired, "登录已过期，请重新登录"},
		{ErrUserNotFound, "用户不存在"},
		{ErrUsernameExists, "用户名已被注册"},
		{ErrPasswordWrong, "用户名或密码错误"},
		{ErrTooManyRequest, "请求过于频繁，请稍后再试"},
	}

	for _, tt := range tests {
		msg := tt.code.GetMessage()
		if msg != tt.message {
			t.Errorf("Code %d: GetMessage() = %q, want %q", tt.code, msg, tt.message)
		}
	}
}

// ─── 模式切换 ───

func TestSetMode(t *testing.T) {
	old := Mode
	defer func() { Mode = old }()

	SetMode("debug")
	if Mode != "debug" {
		t.Errorf("Mode = %q, want %q", Mode, "debug")
	}

	SetMode("release")
	if Mode != "release" {
		t.Errorf("Mode = %q, want %q", Mode, "release")
	}
}

// ─── 泄漏模式脱敏 ───

func TestSanitizeDetail_DuplicateKey(t *testing.T) {
	cause := errors.New(`ERROR: duplicate key value violates unique constraint "users_username_key" (SQLSTATE 23505)`)
	e := Wrap(ErrConflict, "用户名已存在", cause)
	
	if e.Detail == "" {
		t.Error("Detail should not be empty")
	}
	// Should NOT leak column name
	if contains(e.Detail, "users_username") {
		t.Errorf("Detail leaked constraint name: %q", e.Detail)
	}
}

func TestSanitizeDetail_ConnectionRefused(t *testing.T) {
	cause := errors.New("dial tcp: connection refused host=db.example.com user=admin password=secret")
	e := Wrap(ErrDatabaseError, "DB 连接失败", cause)
	
	if contains(e.Detail, "db.example.com") {
		t.Errorf("Detail leaked host: %q", e.Detail)
	}
	if contains(e.Detail, "admin") {
		t.Errorf("Detail leaked user: %q", e.Detail)
	}
}

// ─── classifyError ───

func TestClassifyError_UniqueViolation(t *testing.T) {
	cause := errors.New("duplicate key value violates unique constraint")
	sd := classifyError(ErrConflict, cause)
	
	if sd == nil {
		t.Fatal("SafeDetail should not be nil")
	}
	if sd.Type != "unique_violation" {
		t.Errorf("Type = %q, want unique_violation", sd.Type)
	}
}

func TestClassifyError_ForeignKey(t *testing.T) {
	cause := errors.New("violates foreign key constraint")
	sd := classifyError(ErrBadRequest, cause)
	
	if sd == nil {
		t.Fatal("SafeDetail should not be nil")
	}
	if sd.Type != "foreign_key_violation" {
		t.Errorf("Type = %q, want foreign_key_violation", sd.Type)
	}
}

func TestClassifyError_ConnectionRefused(t *testing.T) {
	cause := errors.New("connection refused")
	sd := classifyError(ErrDatabaseError, cause)
	
	if sd == nil {
		t.Fatal("SafeDetail should not be nil")
	}
	if sd.Type != "connection_error" {
		t.Errorf("Type = %q, want connection_error", sd.Type)
	}
}

func TestClassifyError_Deadlock(t *testing.T) {
	cause := errors.New("deadlock detected")
	sd := classifyError(ErrDatabaseError, cause)
	
	if sd == nil {
		t.Fatal("SafeDetail should not be nil")
	}
	if sd.Type != "deadlock" {
		t.Errorf("Type = %q, want deadlock", sd.Type)
	}
}

func TestClassifyError_NoRows(t *testing.T) {
	cause := errors.New("sql: no rows in result set")
	sd := classifyError(ErrNotFound, cause)
	
	if sd == nil {
		t.Fatal("SafeDetail should not be nil")
	}
	if sd.Type != "not_found" {
		t.Errorf("Type = %q, want not_found", sd.Type)
	}
}

func TestClassifyError_NilCause(t *testing.T) {
	sd := classifyError(ErrBadRequest, nil)
	if sd != nil {
		t.Error("SafeDetail should be nil for nil cause")
	}
}

// ─── helper ───

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
