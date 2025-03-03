package utils

// ErrorCode 错误码定义
type ErrorCode struct {
	Code    int    `json:"code"`    // 错误码
	Message string `json:"message"` // 错误信息
	Type    string `json:"type"`    // 错误类型
}

// 错误类型定义
const (
	ErrorTypeSystem   = "system"   // 系统错误
	ErrorTypeBusiness = "business" // 业务错误
	ErrorTypeValidate = "validate" // 验证错误
)

// 系统错误码 (1000-1999)
var (
	ErrInternalServer = &ErrorCode{Code: 1000, Message: "内部服务器错误", Type: ErrorTypeSystem}
	ErrServiceUnavailable = &ErrorCode{Code: 1001, Message: "服务不可用", Type: ErrorTypeSystem}
	ErrTimeout = &ErrorCode{Code: 1002, Message: "请求超时", Type: ErrorTypeSystem}
	ErrTooManyRequests = &ErrorCode{Code: 1003, Message: "请求过于频繁", Type: ErrorTypeSystem}
)

// 认证和授权错误码 (2000-2999)
var (
	ErrUnauthorized = &ErrorCode{Code: 2000, Message: "未授权访问", Type: ErrorTypeBusiness}
	ErrForbidden = &ErrorCode{Code: 2001, Message: "禁止访问", Type: ErrorTypeBusiness}
	ErrTokenExpired = &ErrorCode{Code: 2002, Message: "令牌已过期", Type: ErrorTypeBusiness}
	ErrTokenInvalid = &ErrorCode{Code: 2003, Message: "令牌无效", Type: ErrorTypeBusiness}
)

// 请求参数错误码 (3000-3999)
var (
	ErrBadRequest = &ErrorCode{Code: 3000, Message: "请求参数错误", Type: ErrorTypeValidate}
	ErrValidation = &ErrorCode{Code: 3001, Message: "数据验证失败", Type: ErrorTypeValidate}
	ErrMissingParam = &ErrorCode{Code: 3002, Message: "缺少必要参数", Type: ErrorTypeValidate}
	ErrInvalidParam = &ErrorCode{Code: 3003, Message: "参数值无效", Type: ErrorTypeValidate}
)

// 资源错误码 (4000-4999)
var (
	ErrNotFound = &ErrorCode{Code: 4000, Message: "资源不存在", Type: ErrorTypeBusiness}
	ErrAlreadyExists = &ErrorCode{Code: 4001, Message: "资源已存在", Type: ErrorTypeBusiness}
	ErrConflict = &ErrorCode{Code: 4002, Message: "资源冲突", Type: ErrorTypeBusiness}
)

// 数据库错误码 (5000-5999)
var (
	ErrDatabase = &ErrorCode{Code: 5000, Message: "数据库操作失败", Type: ErrorTypeSystem}
	ErrDatabaseConnection = &ErrorCode{Code: 5001, Message: "数据库连接失败", Type: ErrorTypeSystem}
	ErrDatabaseQuery = &ErrorCode{Code: 5002, Message: "数据库查询失败", Type: ErrorTypeSystem}
)

// 第三方服务错误码 (6000-6999)
var (
	ErrThirdPartyService = &ErrorCode{Code: 6000, Message: "第三方服务调用失败", Type: ErrorTypeSystem}
	ErrThirdPartyTimeout = &ErrorCode{Code: 6001, Message: "第三方服务超时", Type: ErrorTypeSystem}
)