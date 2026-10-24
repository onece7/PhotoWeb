package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 错误码定义
const (
	CodeSuccess       = 0
	CodeBadRequest    = 40000
	CodeUnauthorized  = 40100
	CodeTokenExpired  = 40101
	CodeForbidden     = 40300
	CodeNotFound      = 40400
	CodeConflict      = 40900
	CodeGenFailed     = 42000
	CodeGenTimeout    = 42001
	CodeGenBusy       = 42002
	CodeRateLimited   = 42900
	CodeHistoryNotFound = 44000
	CodeInternal      = 50000
)

// 错误码对应的默认消息
var messages = map[int]string{
	CodeSuccess:       "success",
	CodeBadRequest:    "请求参数错误",
	CodeUnauthorized:  "未认证或认证已过期",
	CodeTokenExpired:  "Token 已过期",
	CodeForbidden:     "无权限访问",
	CodeNotFound:      "资源不存在",
	CodeConflict:      "资源冲突",
	CodeGenFailed:     "图片生成失败",
	CodeGenTimeout:    "图片生成超时",
	CodeGenBusy:       "生成队列繁忙，请稍后重试",
	CodeRateLimited:   "请求过于频繁，请稍后重试",
	CodeHistoryNotFound: "记录不存在",
	CodeInternal:      "服务器内部错误",
}

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Init 初始化 (可扩展)
func Init() {}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: messages[CodeSuccess],
		Data:    data,
	})
}

// SuccessWithMessage 成功响应 (自定义消息)
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Created 201 创建成功
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    CodeSuccess,
		Message: messages[CodeSuccess],
		Data:    data,
	})
}

// Accepted 202 已接收 (异步处理)
func Accepted(c *gin.Context, data interface{}) {
	c.JSON(http.StatusAccepted, Response{
		Code:    CodeSuccess,
		Message: messages[CodeSuccess],
		Data:    data,
	})
}

// Error 通用错误响应
func Error(c *gin.Context, httpStatus int, code int, message string) {
	if message == "" {
		message = messages[code]
	}
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest 400 参数错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, CodeBadRequest, message)
}

// Unauthorized 401 未认证
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden 403 无权限
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, CodeForbidden, message)
}

// TooManyRequests 429 请求过于频繁
func TooManyRequests(c *gin.Context, message string) {
	Error(c, http.StatusTooManyRequests, CodeRateLimited, message)
}

// NotFound 404 资源不存在
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, CodeNotFound, message)
}

// Conflict 409 资源冲突
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, CodeConflict, message)
}

// InternalError 500 服务器内部错误
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, CodeInternal, message)
}
