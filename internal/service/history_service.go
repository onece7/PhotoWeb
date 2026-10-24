package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"novel2comic/backend/internal/models"
	"novel2comic/backend/internal/repository"
)

// HistoryService 历史记录服务
type HistoryService struct {
	genRepo repository.GenerationRepository
}

// NewHistoryService 创建历史 Service
func NewHistoryService(genRepo repository.GenerationRepository) *HistoryService {
	return &HistoryService{genRepo: genRepo}
}

// GetHistory 获取用户历史列表 (分页)
func (s *HistoryService) GetHistory(ctx context.Context, userID uint64, page, pageSize int) (*models.HistoryListResponse, error) {
	records, total, err := s.genRepo.FindByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询历史记录失败: %w", err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &models.HistoryListResponse{
		Records:    records,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetHistoryDetail 获取历史详情
func (s *HistoryService) GetHistoryDetail(ctx context.Context, id, userID uint64) (*models.GenerationRecord, error) {
	record, err := s.genRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("记录不存在: %w", err)
	}

	// 验证归属
	if record.UserID != userID {
		return nil, fmt.Errorf("无权访问此记录")
	}

	return record, nil
}

// DeleteHistory 删除历史记录及图片文件
func (s *HistoryService) DeleteHistory(ctx context.Context, id, userID uint64, uploadDir string) error {
	// 先查记录确认归属
	record, err := s.genRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("记录不存在: %w", err)
	}
	if record.UserID != userID {
		return fmt.Errorf("无权删除此记录")
	}

	// 删除数据库记录
	if err := s.genRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除记录失败: %w", err)
	}

	// 删除图片文件
	if record.ImageURL != "" {
		imagePath := filepath.Join(uploadDir, record.ImageURL)
		os.Remove(imagePath) // 忽略错误，文件可能不存在
	}

	return nil
}
