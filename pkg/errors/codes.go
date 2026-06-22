package errors

type ErrorCode int

const (
	ErrInternal       ErrorCode = 10001
	ErrBadRequest     ErrorCode = 10002
	ErrNotFound       ErrorCode = 10003
	ErrConflict       ErrorCode = 10004
	ErrTooManyRequest ErrorCode = 10005

	ErrUnauthorized  ErrorCode = 10101
	ErrTokenExpired  ErrorCode = 10102
	ErrTokenInvalid  ErrorCode = 10103
	ErrForbidden     ErrorCode = 10104
	ErrPasswordWrong ErrorCode = 10105
	ErrUserBanned    ErrorCode = 10106

	ErrUserNotFound   ErrorCode = 10201
	ErrUsernameExists ErrorCode = 10202
	ErrEmailExists    ErrorCode = 10203
	ErrInvalidRole    ErrorCode = 10204

	ErrPostNotFound     ErrorCode = 10301
	ErrPostNotOwner     ErrorCode = 10302
	ErrCommentNotFound  ErrorCode = 10303
	ErrContentSensitive ErrorCode = 10304

	ErrPOINotFound   ErrorCode = 10401
	ErrRouteNotFound ErrorCode = 10402

	ErrQuestionNotFound ErrorCode = 10501
	ErrDuplicateAnswer  ErrorCode = 10502

	ErrLLMTimeout     ErrorCode = 10601
	ErrLLMRateLimit   ErrorCode = 10602
	ErrLLMUnavailable ErrorCode = 10603

	ErrImageTooLarge    ErrorCode = 10701
	ErrImageFormat      ErrorCode = 10702
	ErrDetectionFailed  ErrorCode = 10703

	ErrItemNotFound        ErrorCode = 10901
	ErrInsufficientPoints  ErrorCode = 10902
	ErrOutOfStock          ErrorCode = 10903

	ErrDatabaseError ErrorCode = 10801
	ErrRedisError    ErrorCode = 10802
	ErrMinIOError    ErrorCode = 10804
)

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
		c == ErrQuestionNotFound || c == ErrItemNotFound:
		return 404
	case c == ErrTooManyRequest || c == ErrDuplicateAnswer:
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
