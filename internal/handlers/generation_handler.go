package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"novel2comic/backend/internal/models"
	"novel2comic/backend/internal/service"
	"novel2comic/backend/pkg/response"
)

// GenerationHandler 图片生成处理器
type GenerationHandler struct {
	genService *service.GenerationService
}

// NewGenerationHandler 创建生成 Handler
func NewGenerationHandler(genService *service.GenerationService) *GenerationHandler {
	return &GenerationHandler{genService: genService}
}

// GetStyles 获取可用风格列表
// GET /api/v1/generate/styles
func (h *GenerationHandler) GetStyles(c *gin.Context) {
	styles, err := h.genService.GetActiveStyles(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "获取风格列表失败")
		return
	}
	response.Success(c, styles)
}

// SimpleGenerate 普通模式生成
// POST /api/v1/generate/simple
func (h *GenerationHandler) SimpleGenerate(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "")
		return
	}

	var req models.SimpleGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入提示词并选择风格")
		return
	}

	resp, err := h.genService.GenerateSimple(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, 400, response.CodeGenFailed, err.Error())
		return
	}

	response.Accepted(c, resp)
}

// ProGenerate 专业模式生成
// POST /api/v1/generate/pro
func (h *GenerationHandler) ProGenerate(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "")
		return
	}

	var req models.ProGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写有效的生成参数")
		return
	}

	resp, err := h.genService.GeneratePro(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, 400, response.CodeGenFailed, err.Error())
		return
	}

	response.Accepted(c, resp)
}

// GetGenerationStatus 查询生成状态
// GET /api/v1/generate/status/:id
func (h *GenerationHandler) GetGenerationStatus(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}

	record, err := h.genService.GetStatus(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "记录不存在")
		return
	}

	// 验证归属
	if record.UserID != userID {
		response.Forbidden(c, "无权查看此记录")
		return
	}

	statusResp := models.GenerationStatusResponse{
		RecordID:       record.ID,
		Status:         record.Status,
		Seed:           record.Seed,
		DurationMs:     record.DurationMs,
		ErrorMessage:   record.ErrorMessage,
		Prompt:         record.Prompt,
		NegativePrompt: record.NegativePrompt,
		Width:          record.Width,
		Height:         record.Height,
		CFGScale:       record.CFGScale,
		Steps:          record.Steps,
		Sampler:        record.Sampler,
	}

	if record.Status == models.StatusCompleted && record.ImageURL != "" {
		statusResp.ImageURL = "/api/v1/files/" + record.ImageURL
	}

	response.Success(c, statusResp)
}
