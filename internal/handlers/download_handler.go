package handlers

import (
	"path/filepath"

	"github.com/gin-gonic/gin"

	"novel2comic/backend/pkg/response"
)

// ServeGeneratedFile 提供生成的图片文件
// GET /api/v1/files/:filename
func ServeGeneratedFile(uploadDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")

		// 安全检查：防止路径遍历
		cleanPath := filepath.Clean(filename)
		if cleanPath != filename {
			response.BadRequest(c, "无效的文件路径")
			return
		}

		filePath := filepath.Join(uploadDir, filename)
		c.File(filePath)
	}
}
