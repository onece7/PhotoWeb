package handlers

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"photo-web/pkg/response"
)

// ServeGeneratedFile 提供生成的图片文件
// GET /api/v1/files/*filepath
func ServeGeneratedFile(uploadDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestedPath := strings.TrimPrefix(c.Param("filepath"), string(os.PathSeparator))
		cleanPath := filepath.Clean(requestedPath)

		// 安全检查：防止路径遍历
		if cleanPath == "." ||
			cleanPath == ".." ||
			strings.HasPrefix(cleanPath, ".."+string(os.PathSeparator)) {
			response.BadRequest(c, "无效的文件路径")
			return
		}

		resolvedPath := filepath.Join(uploadDir, cleanPath)
		c.File(resolvedPath)
	}
}
