package repository

import (
	"context"

	"gorm.io/gorm"

	"novel2comic/backend/internal/models"
)

// GenerationRepository 生成记录数据访问接口
type GenerationRepository interface {
	Create(ctx context.Context, record *models.GenerationRecord) error
	Update(ctx context.Context, record *models.GenerationRecord) error
	FindByID(ctx context.Context, id uint64) (*models.GenerationRecord, error)
	FindByUserID(ctx context.Context, userID uint64, page, pageSize int) ([]models.GenerationRecord, int64, error)
	Delete(ctx context.Context, id uint64) error
	CountPendingByUserID(ctx context.Context, userID uint64) (int64, error)
}

type generationRepo struct {
	db *gorm.DB
}

// NewGenerationRepository 创建生成记录 Repository
func NewGenerationRepository(db *gorm.DB) GenerationRepository {
	return &generationRepo{db: db}
}

func (r *generationRepo) Create(ctx context.Context, record *models.GenerationRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *generationRepo) Update(ctx context.Context, record *models.GenerationRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *generationRepo) FindByID(ctx context.Context, id uint64) (*models.GenerationRecord, error) {
	var record models.GenerationRecord
	err := r.db.WithContext(ctx).Preload("Style").First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *generationRepo) FindByUserID(ctx context.Context, userID uint64, page, pageSize int) ([]models.GenerationRecord, int64, error) {
	var records []models.GenerationRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&models.GenerationRecord{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("Style").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&records).Error

	return records, total, err
}

func (r *generationRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.GenerationRecord{}, id).Error
}

func (r *generationRepo) CountPendingByUserID(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.GenerationRecord{}).
		Where("user_id = ? AND status IN ?", userID, []string{models.StatusPending, models.StatusProcessing}).
		Count(&count).Error
	return count, err
}
