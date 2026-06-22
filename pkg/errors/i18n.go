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
	ErrInvalidRole:    "无效的用户角色",

	ErrPostNotFound:     "帖子不存在",
	ErrPostNotOwner:     "只能操作自己的帖子",
	ErrCommentNotFound:  "评论不存在",
	ErrContentSensitive: "内容包含违规信息，请修改后重试",

	ErrPOINotFound:   "POI 不存在",
	ErrRouteNotFound: "路线不存在",

	ErrQuestionNotFound: "题目不存在",
	ErrDuplicateAnswer:  "请勿重复提交",

	ErrLLMTimeout:     "AI 响应超时，请稍后重试",
	ErrLLMRateLimit:   "AI 服务繁忙，请稍后再试",
	ErrLLMUnavailable: "AI 服务暂时不可用",

	ErrImageTooLarge:    "图片文件过大，请压缩后重试",
	ErrImageFormat:      "不支持的图片格式",
	ErrDetectionFailed:  "识别失败，请重新拍摄",

	ErrItemNotFound:       "商品不存在",
	ErrInsufficientPoints: "积分不足",
	ErrOutOfStock:         "商品已售罄",

	ErrDatabaseError: "数据库错误",
	ErrRedisError:    "缓存服务错误",
	ErrMinIOError:    "文件存储错误",
}

func (c ErrorCode) GetMessage() string {
	if msg, ok := DefaultMessages[c]; ok {
		return msg
	}
	return "未知错误"
}
