package errors

// DefaultMessages 错误码 → 默认中文消息
var DefaultMessages = map[ErrorCode]string{
	// 通用
	ErrInternal:       "服务器内部错误",
	ErrBadRequest:     "请求参数错误",
	ErrNotFound:       "资源不存在",
	ErrConflict:       "资源冲突，请稍后重试",
	ErrTooManyRequest: "请求过于频繁，请稍后再试",

	// 认证
	ErrUnauthorized:  "请先登录",
	ErrTokenExpired:  "登录已过期，请重新登录",
	ErrTokenInvalid:  "Token 无效",
	ErrForbidden:     "无权执行此操作",
	ErrPasswordWrong: "用户名或密码错误",
	ErrUserBanned:    "账号已被封禁",

	// 用户
	ErrUserNotFound:   "用户不存在",
	ErrUsernameExists: "用户名已被注册",
	ErrEmailExists:    "邮箱已被注册",
	ErrInvalidRole:    "无效角色",

	// 内容
	ErrPostNotFound:     "帖子不存在",
	ErrPostNotOwner:     "只能操作自己的帖子",
	ErrCommentNotFound:  "评论不存在",
	ErrContentSensitive: "内容包含违规信息，请修改后重试",

	// 问答
	ErrQuestionNotFound: "题目不存在",
	ErrDuplicateAnswer:  "不能重复提交",
	ErrSessionExpired:   "答题会话已过期",

	// LLM
	ErrLLMTimeout:     "AI 响应超时，请稍后重试",
	ErrLLMRateLimit:   "AI 服务繁忙，请稍后再试",
	ErrLLMUnavailable: "AI 服务暂时不可用",

	// 外部依赖
	ErrDatabaseError: "数据库错误",
	ErrRedisError:    "缓存服务异常",
}

// GetMessage 获取错误码对应的默认消息
func (c ErrorCode) GetMessage() string {
	if msg, ok := DefaultMessages[c]; ok {
		return msg
	}
	return "未知错误"
}
