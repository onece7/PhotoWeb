package repository

import (
	"context"

	"gorm.io/gorm"

	"photo-web/internal/models"
)

// StyleRepository 风格数据访问接口
type StyleRepository interface {
	FindAll(ctx context.Context) ([]models.ImageStyle, error)
	FindActive(ctx context.Context) ([]models.ImageStyle, error)
	FindByID(ctx context.Context, id uint64) (*models.ImageStyle, error)
	FindByName(ctx context.Context, name string) (*models.ImageStyle, error)
}

type styleRepo struct {
	db *gorm.DB
}

// NewStyleRepository 创建风格 Repository
func NewStyleRepository(db *gorm.DB) StyleRepository {
	return &styleRepo{db: db}
}

func (r *styleRepo) FindAll(ctx context.Context) ([]models.ImageStyle, error) {
	styles, err := gorm.G[models.ImageStyle](r.db).Order("sort_order ASC").Find(ctx)
	return styles, err
}

func (r *styleRepo) FindActive(ctx context.Context) ([]models.ImageStyle, error) {
	styles, err := gorm.G[models.ImageStyle](r.db).Where("is_active = ?", true).Order("sort_order ASC").Find(ctx)
	return styles, err
}

func (r *styleRepo) FindByID(ctx context.Context, id uint64) (*models.ImageStyle, error) {
	style, err := gorm.G[models.ImageStyle](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		return nil, err
	}
	return &style, nil
}

func (r *styleRepo) FindByName(ctx context.Context, name string) (*models.ImageStyle, error) {
	style, err := gorm.G[models.ImageStyle](r.db).Where("name = ?", name).First(ctx)
	if err != nil {
		return nil, err
	}
	return &style, nil
}
