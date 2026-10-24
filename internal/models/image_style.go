package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONMap 自定义 JSON 类型，用于存储 SD 预设参数
type JSONMap map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("类型断言为 []byte 失败")
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// ImageStyle 图片风格预设模型
type ImageStyle struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"name"`
	DisplayName  string    `gorm:"type:varchar(128);not null" json:"display_name"`
	Description  string    `gorm:"type:text" json:"description"`
	PresetParams JSONMap   `gorm:"type:json" json:"preset_params"`
	SortOrder    int       `gorm:"default:0" json:"sort_order"`
	PreviewURL   string    `gorm:"type:varchar(512)" json:"preview_url"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 自定义表名
func (ImageStyle) TableName() string {
	return "image_styles"
}
