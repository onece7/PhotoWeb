package repository

import (
	"context"

	"gorm.io/gorm"

	"photo-web/internal/models"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uint64) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByPhone(ctx context.Context, phone string) (*models.User, error)
	Update(ctx context.Context, user models.User) error
}

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository 创建用户 Repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	err := gorm.G[models.User](r.db).Create(ctx, user)
	return err
}

func (r *userRepo) FindByID(ctx context.Context, id uint64) (*models.User, error) {
	user, err := gorm.G[models.User](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := gorm.G[models.User](r.db).Where("username = ?", username).First(ctx)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	user, err := gorm.G[models.User](r.db).Where("phone = ?", phone).First(ctx)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) Update(ctx context.Context, user models.User) error {
	_, err := gorm.G[models.User](r.db).Where("id = ?", user.ID).Select("*").Updates(ctx, user)
	return err
}
