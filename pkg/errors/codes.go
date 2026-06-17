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
	ErrUnauthorized  ErrorCode = 10101 // 未登录
	ErrTokenExpired  ErrorCode = 10102 // Token 已过期
	ErrTokenInvalid  ErrorCode = 10103 // Token 无效
	ErrForbidden     ErrorCode = 10104 // 无权限
	ErrPasswordWrong ErrorCode = 10105 // 密码错误
	ErrUserBanned    ErrorCode = 10106 // 账号已封禁

	// ---- 用户 (02) ----
	ErrUserNotFound   ErrorCode = 10201 // 用户不存在
	ErrUsernameExists ErrorCode = 10202 // 用户名已存在
	ErrEmailExists    ErrorCode = 10203 // 邮箱已注册
	ErrInvalidRole    ErrorCode = 10204 // 无效角色

	// ---- 内容 (03) ----
	ErrPostNotFound     ErrorCode = 10301 // 帖子不存在
	ErrPostNotOwner     ErrorCode = 10302 // 非帖子作者
	ErrCommentNotFound  ErrorCode = 10303 // 评论不存在
	ErrContentSensitive ErrorCode = 10304 // 内容违规

	// ---- 问答 (05) ----
	ErrQuestionNotFound ErrorCode = 10501 // 题目不存在
	ErrDuplicateAnswer  ErrorCode = 10502 // 重复提交
	ErrSessionExpired   ErrorCode = 10503 // 答题会话过期

	// ---- LLM (06) ----
	ErrLLMTimeout     ErrorCode = 10601 // LLM 超时
	ErrLLMRateLimit   ErrorCode = 10602 // LLM 限流
	ErrLLMUnavailable ErrorCode = 10603 // LLM 服务不可用

	// ---- 外部依赖 (08) ----
	ErrDatabaseError ErrorCode = 10801 // 数据库错误
	ErrRedisError    ErrorCode = 10802 // Redis 错误
)

// HTTPStatus 返回对应的 HTTP 状态码
func (c ErrorCode) HTTPStatus() int {
	switch {
	case c == ErrUnauthorized || c == ErrTokenExpired || c == ErrTokenInvalid:
		return 401
	case c == ErrForbidden || c == ErrPostNotOwner:
		return 403
	case c == ErrNotFound || c == ErrUserNotFound ||
		c == ErrPostNotFound || c == ErrCommentNotFound ||
		c == ErrQuestionNotFound:
		return 404
	case c == ErrConflict || c == ErrUsernameExists || c == ErrEmailExists ||
		c == ErrDuplicateAnswer:
		return 409
	case c == ErrTooManyRequest:
		return 429
	case c == ErrLLMTimeout || c == ErrLLMUnavailable || c == ErrLLMRateLimit:
		return 503
	case c == ErrBadRequest:
		return 400
	default:
		return 500
	}
}
