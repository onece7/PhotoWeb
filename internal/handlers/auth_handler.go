package handlers

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"novel2comic/backend/internal/models"
	"novel2comic/backend/internal/service"
	"novel2comic/backend/pkg/response"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证 Handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register 用户注册
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入有效的用户名、手机号和密码")
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}

	response.Created(c, resp)
}

// Login 用户登录
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入用户名和密码")
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// RefreshToken 刷新 Token
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供 Refresh Token")
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// Logout 登出
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从 Authorization Header 获取 Access Token
	accessToken := extractBearerToken(c)
	if accessToken == "" {
		response.BadRequest(c, "未提供 Token")
		return
	}

	var refreshToken string
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err == nil {
		refreshToken = req.RefreshToken
	}

	if err := h.authService.Logout(c.Request.Context(), accessToken, refreshToken); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// extractBearerToken 从 Authorization Header 提取 Bearer Token
func extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

// GetUserID 从 context 中获取当前用户 ID
func GetUserID(c *gin.Context) (uint64, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, errors.New("未找到用户信息")
	}
	id, ok := userID.(uint64)
	if !ok {
		return 0, errors.New("用户信息类型错误")
	}
	return id, nil
}
