package models

import "time"

// 生成模式
const (
	ModeSimple = "simple"
	ModePro    = "pro"
)

// 生成状态
const (
	StatusPending   = "pending"
	StatusProcessing = "processing"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

// GenerationRecord 生成记录模型
type GenerationRecord struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint64    `gorm:"index;not null" json:"user_id"`
	Mode           string    `gorm:"type:enum('simple','pro');not null" json:"mode"`
	Prompt         string    `gorm:"type:text;not null" json:"prompt"`
	NegativePrompt string    `gorm:"type:text" json:"negative_prompt"`
	StyleID        *uint64   `gorm:"default:null" json:"style_id"`
	Width          int       `gorm:"default:512" json:"width"`
	Height         int       `gorm:"default:512" json:"height"`
	CFGScale       float64   `gorm:"default:7.0" json:"cfg_scale"`
	Steps          int       `gorm:"default:20" json:"steps"`
	Sampler        string    `gorm:"type:varchar(64);default:'Euler a'" json:"sampler"`
	Seed           int64     `gorm:"default:-1" json:"seed"`
	ImageURL       string    `gorm:"type:varchar(512)" json:"image_url"`
	ThumbnailURL   string    `gorm:"type:varchar(512)" json:"thumbnail_url"`
	Status         string    `gorm:"type:enum('pending','processing','completed','failed');default:'pending';index" json:"status"`
	ErrorMessage   string    `gorm:"type:text" json:"error_message"`
	DurationMs     int64     `gorm:"default:0" json:"duration_ms"`
	CreatedAt      time.Time `gorm:"autoCreateTime;index" json:"created_at"`

	// 关联
	User  User       `gorm:"foreignKey:UserID" json:"-"`
	Style *ImageStyle `gorm:"foreignKey:StyleID" json:"style,omitempty"`
}

// TableName 自定义表名
func (GenerationRecord) TableName() string {
	return "generation_records"
}
