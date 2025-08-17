package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`     // 业务码
	Message string      `json:"msg"`      // 提示信息
	Data    interface{} `json:"data"`     // 数据
	TraceID string      `json:"trace_id"` // 追踪ID
}

// ResponseOption 响应选项
type ResponseOption func(*Response)

// WithTraceID 设置追踪ID
func WithTraceID(traceID string) ResponseOption {
	return func(r *Response) {
		r.TraceID = traceID
	}
}

// WithMessage 设置自定义错误信息
func WithMessage(message string) ResponseOption {
	return func(r *Response) {
		r.Message = message
	}
}

// Success 成功响应
func Success(c *gin.Context, data interface{}, opts ...ResponseOption) {
	resp := &Response{
		Code:    1,
		Message: "success",
		Data:    data,
	}

	// 获取并设置 trace_id
	if traceID, exists := c.Get("trace_id"); exists {
		resp.TraceID = traceID.(string)
	}

	// 应用选项
	for _, opt := range opts {
		opt(resp)
	}

	c.JSON(http.StatusOK, resp)
}

// Error 错误响应
func Error(c *gin.Context, errCode *ErrorCode, opts ...ResponseOption) {
	resp := &Response{
		Code:    errCode.Code,
		Message: errCode.Message,
	}

	// 获取并设置 trace_id
	if traceID, exists := c.Get("trace_id"); exists {
		resp.TraceID = traceID.(string)
	}

	// 应用选项
	for _, opt := range opts {
		opt(resp)
	}

	// 根据错误类型设置HTTP状态码
	httpStatus := http.StatusInternalServerError
	if errCode.Type == ErrorTypeBusiness || errCode.Type == ErrorTypeValidate {
		httpStatus = http.StatusOK
	}
	switch errCode.Type {
	case ErrorTypeValidate:
		httpStatus = http.StatusBadRequest
	case ErrorTypeBusiness:
		if errCode.Code >= 2000 && errCode.Code < 3000 {
			switch errCode.Code {
			case 2000: // ErrUnauthorized
				httpStatus = http.StatusUnauthorized
			case 2001: // ErrForbidden
				httpStatus = http.StatusForbidden
			default:
				httpStatus = http.StatusBadRequest
			}
		} else if errCode.Code >= 4000 && errCode.Code < 5000 {
			if errCode.Code == 4000 { // ErrNotFound
				httpStatus = http.StatusNotFound
			} else {
				httpStatus = http.StatusBadRequest
			}
		} else if errCode.Code >= 10000003 && errCode.Code <= 10000007 {
			// 认证相关错误码处理
			// 根据错误码定义：
			// 10000003 - Unauthorized (鉴权失败)
			// 10000004 - UnauthorizedAuthNotExist (鉴权失败, Token不存在)
			// 10000005 - UnauthorizedTokenError (鉴权失败, Token错误)
			// 10000006 - UnauthorizedTokenTimeout (鉴权失败, Token超时)
			// 10000007 - UnauthorizedTokenGenerate (鉴权失败, Token生成失败)
			// 这些都应该返回 401 Unauthorized 状态码。
			httpStatus = http.StatusUnauthorized
		}
	case ErrorTypeSystem:
		httpStatus = http.StatusInternalServerError
	}

	c.JSON(httpStatus, resp)
}

// BadRequest 请求参数错误响应
func BadRequest(c *gin.Context, opts ...ResponseOption) {
	Error(c, ErrBadRequest, opts...)
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, opts ...ResponseOption) {
	c.Header("WWW-Authenticate", "Bearer")
	Error(c, ErrUnauthorized, opts...)
}

// Forbidden 禁止访问响应
func Forbidden(c *gin.Context, opts ...ResponseOption) {
	Error(c, ErrForbidden, opts...)
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context, opts ...ResponseOption) {
	Error(c, ErrNotFound, opts...)
}

// ServerError 服务器错误响应
func ServerError(c *gin.Context, opts ...ResponseOption) {
	Error(c, ErrInternalServer, opts...)
}
