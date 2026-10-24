package models

import "time"

// User 用户模型
type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Phone        string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 自定义表名
func (User) TableName() string {
	return "users"
}
