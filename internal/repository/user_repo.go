package repository

import (
	"context"

	"gorm.io/gorm"

	"novel2comic/backend/internal/models"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uint64) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByPhone(ctx context.Context, phone string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository 创建用户 Repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) FindByID(ctx context.Context, id uint64) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}
