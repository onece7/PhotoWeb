package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"novel2comic/backend/pkg/response"
)

// RateLimiter 简单的令牌桶限流器
// 生产环境建议使用 Redis 实现分布式限流
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // 每秒允许请求数
	burst    int           // 突发允许数
	cleanup  time.Duration // 清理间隔
}

type visitor struct {
	tokens    float64
	lastCheck time.Time
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rate, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    burst,
		cleanup:  5 * time.Minute,
	}
	go rl.cleanupRoutine()
	return rl
}

// Handler 返回 Gin 中间件
func (rl *RateLimiter) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()

		rl.mu.Lock()
		v, exists := rl.visitors[key]
		if !exists {
			v = &visitor{
				tokens:    float64(rl.burst),
				lastCheck: time.Now(),
			}
			rl.visitors[key] = v
		}

		// 计算令牌补充
		now := time.Now()
		elapsed := now.Sub(v.lastCheck).Seconds()
		v.tokens += elapsed * float64(rl.rate)
		if v.tokens > float64(rl.burst) {
			v.tokens = float64(rl.burst)
		}
		v.lastCheck = now

		// 消费令牌
		if v.tokens < 1 {
			rl.mu.Unlock()
			response.TooManyRequests(c, "")
			c.Abort()
			return
		}
		v.tokens--
		rl.mu.Unlock()

		c.Next()
	}
}

// cleanupRoutine 定期清理过期条目
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanup)
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastCheck) > rl.cleanup {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}
