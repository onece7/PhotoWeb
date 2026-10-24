package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"novel2comic/backend/internal/config"
	"novel2comic/backend/internal/models"
	"novel2comic/backend/internal/repository"
	jwtpkg "novel2comic/backend/pkg/jwt"
	redisCli "novel2comic/backend/pkg/redis"
)

// AuthService 认证业务逻辑
type AuthService struct {
	userRepo repository.UserRepository
	redis    *redisCli.Client
	jwt      *jwtpkg.Manager
}

// NewAuthService 创建认证 Service
func NewAuthService(userRepo repository.UserRepository, redis *redisCli.Client, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		redis:    redis,
		jwt:      jwtpkg.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL),
	}
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.LoginResponse, error) {
	// 检查用户名是否已存在
	if _, err := s.userRepo.FindByUsername(ctx, req.Username); err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查手机号是否已存在
	if _, err := s.userRepo.FindByPhone(ctx, req.Phone); err == nil {
		return nil, errors.New("手机号已注册")
	}

	// bcrypt 哈希密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	user := &models.User{
		Username:     req.Username,
		Phone:        req.Phone,
		PasswordHash: string(passwordHash),
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 生成 Token
	return s.generateTokenPair(ctx, user.ID, user.Username)
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// 查找用户
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 比对密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	return s.generateTokenPair(ctx, user.ID, user.Username)
}

// RefreshToken 刷新 Token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
	// 计算 refresh token hash
	hash := hashToken(refreshToken)
	key := fmt.Sprintf("refresh:%s", hash)

	// 从 Redis 获取 user_id
	userIDStr, err := s.redis.Get(ctx, key)
	if err != nil {
		return nil, errors.New("无效的 Refresh Token")
	}

	// 删除旧的 Refresh Token (轮换)
	s.redis.Del(ctx, key)

	// 查找用户
	var userID uint64
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		return nil, errors.New("无效的 Refresh Token")
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	return s.generateTokenPair(ctx, user.ID, user.Username)
}

// Logout 登出
func (s *AuthService) Logout(ctx context.Context, accessToken string, refreshToken string) error {
	// 将 Access Token 加入黑名单
	claims, err := s.jwt.ParseAccessToken(accessToken)
	if err == nil && claims != nil {
		// 计算剩余有效期
		remainingTTL := time.Until(claims.ExpiresAt.Time)
		if remainingTTL > 0 {
			blackKey := fmt.Sprintf("blacklist:%s", accessToken)
			s.redis.Set(ctx, blackKey, "1", remainingTTL)
		}
	}

	// 删除 Refresh Token
	if refreshToken != "" {
		hash := hashToken(refreshToken)
		key := fmt.Sprintf("refresh:%s", hash)
		s.redis.Del(ctx, key)
	}

	return nil
}

// IsTokenBlacklisted 检查 Token 是否在黑名单中
func (s *AuthService) IsTokenBlacklisted(ctx context.Context, token string) bool {
	key := fmt.Sprintf("blacklist:%s", token)
	exists, _ := s.redis.Exists(ctx, key)
	return exists > 0
}

// generateTokenPair 生成 Token 对
func (s *AuthService) generateTokenPair(ctx context.Context, userID uint64, username string) (*models.LoginResponse, error) {
	// 生成 Access Token
	accessToken, expiresIn, err := s.jwt.GenerateAccessToken(userID, username)
	if err != nil {
		return nil, fmt.Errorf("生成 Access Token 失败: %w", err)
	}

	// 生成 Refresh Token
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshHash := hashToken(refreshToken)
	key := fmt.Sprintf("refresh:%s", refreshHash)

	if err := s.redis.Set(ctx, key, fmt.Sprintf("%d", userID), s.jwt.RefreshTokenTTL()); err != nil {
		return nil, fmt.Errorf("存储 Refresh Token 失败: %w", err)
	}

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// generateRefreshToken 生成随机 Refresh Token
func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成随机令牌失败: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// hashToken 对 Token 做 SHA256 哈希
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
