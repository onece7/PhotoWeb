package models

// --- 认证相关 DTO ---

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Phone    string `json:"phone" binding:"required,len=11"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录/Tokens 响应
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // 秒
}

// RefreshRequest 刷新 Token 请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// --- 生成相关 DTO ---

// SimpleGenerateRequest 普通模式生成请求
type SimpleGenerateRequest struct {
	Prompt  string `json:"prompt" binding:"required,min=1,max=2000"`
	StyleID uint64 `json:"style_id" binding:"required"`
}

// ProGenerateRequest 专业模式生成请求
type ProGenerateRequest struct {
	Prompt         string  `json:"prompt" binding:"required,min=1,max=2000"`
	NegativePrompt string  `json:"negative_prompt" binding:"max=2000"`
	Width          int     `json:"width" binding:"min=64,max=2048"`
	Height         int     `json:"height" binding:"min=64,max=2048"`
	CFGScale       float64 `json:"cfg_scale" binding:"min=1,max=30"`
	Steps          int     `json:"steps" binding:"min=1,max=150"`
	Sampler        string  `json:"sampler"`
	Seed           int64   `json:"seed"`
}

// GenerateResponse 生成请求响应 (202)
type GenerateResponse struct {
	RecordID uint64 `json:"record_id"`
	Status   string `json:"status"`
}

// GenerationStatusResponse 生成状态响应
type GenerationStatusResponse struct {
	RecordID       uint64  `json:"record_id"`
	Status         string  `json:"status"`
	ImageURL       string  `json:"image_url,omitempty"`
	ThumbnailURL   string  `json:"thumbnail_url,omitempty"`
	Seed           int64   `json:"seed,omitempty"`
	DurationMs     int64   `json:"duration_ms,omitempty"`
	ErrorMessage   string  `json:"error_message,omitempty"`
	Prompt         string  `json:"prompt,omitempty"`
	NegativePrompt string  `json:"negative_prompt,omitempty"`
	Width          int     `json:"width,omitempty"`
	Height         int     `json:"height,omitempty"`
	CFGScale       float64 `json:"cfg_scale,omitempty"`
	Steps          int     `json:"steps,omitempty"`
	Sampler        string  `json:"sampler,omitempty"`
}

// --- 历史相关 DTO ---

// HistoryListResponse 历史列表响应
type HistoryListResponse struct {
	Records    []GenerationRecord `json:"records"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// PaginationQuery 分页请求参数
type PaginationQuery struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}
