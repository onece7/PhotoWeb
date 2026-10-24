package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 请求日志中间件
// 记录: 方法、路径、状态码、耗时、客户端 IP
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 记录日志
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		if query != "" {
			path = path + "?" + query
		}

		log.Printf("[%s] %s %s %d %v | IP: %s",
			time.Now().Format("2006-01-02 15:04:05"),
			method,
			path,
			statusCode,
			latency,
			clientIP,
		)
	}
}
