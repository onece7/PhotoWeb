package jwt

import (
	"errors"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// Claims JWT 声明
type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwtv5.RegisteredClaims
}

// Manager JWT 管理器
type Manager struct {
	secret        []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// NewManager 创建 JWT 管理器
func NewManager(secret string, accessTTLMinutes, refreshTTLHours int) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  time.Duration(accessTTLMinutes) * time.Minute,
		refreshTTL: time.Duration(refreshTTLHours) * time.Hour,
	}
}

// GenerateAccessToken 生成访问令牌
func (m *Manager) GenerateAccessToken(userID uint64, username string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(m.accessTTL)
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwtv5.RegisteredClaims{
			IssuedAt:  jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(m.secret)
	if err != nil {
		return "", 0, err
	}

	return signedToken, int64(m.accessTTL.Seconds()), nil
}

// ParseAccessToken 解析访问令牌
func (m *Manager) ParseAccessToken(tokenString string) (*Claims, error) {
	token, err := jwtv5.ParseWithClaims(tokenString, &Claims{}, func(token *jwtv5.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名算法")
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("无效的 Token")
	}

	return claims, nil
}

// AccessTokenTTL 返回 Access Token 有效期
func (m *Manager) AccessTokenTTL() time.Duration {
	return m.accessTTL
}

// RefreshTokenTTL 返回 Refresh Token 有效期
func (m *Manager) RefreshTokenTTL() time.Duration {
	return m.refreshTTL
}
