package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"novel2comic/backend/internal/config"
	"novel2comic/backend/internal/models"
	"novel2comic/backend/internal/repository"
)

// GenerationService 图片生成服务
type GenerationService struct {
	genRepo  repository.GenerationRepository
	styleRepo repository.StyleRepository
	sdClient *sdClient
	cfg      *config.Config
	sem      chan struct{} // 并发控制信号量
}

// NewGenerationService 创建生成 Service
func NewGenerationService(
	genRepo repository.GenerationRepository,
	styleRepo repository.StyleRepository,
	sdClient *sdClient,
	cfg *config.Config,
) *GenerationService {
	return &GenerationService{
		genRepo:   genRepo,
		styleRepo: styleRepo,
		sdClient:  sdClient,
		cfg:       cfg,
		sem:       make(chan struct{}, cfg.SD.MaxConcurrent),
	}
}

// GetActiveStyles 获取所有活跃的风格
func (s *GenerationService) GetActiveStyles(ctx context.Context) ([]models.ImageStyle, error) {
	return s.styleRepo.FindActive(ctx)
}

// GenerateSimple 普通模式生成：使用预设风格
func (s *GenerationService) GenerateSimple(ctx context.Context, userID uint64, req *models.SimpleGenerateRequest) (*models.GenerateResponse, error) {
	// 获取风格
	style, err := s.styleRepo.FindByID(ctx, req.StyleID)
	if err != nil {
		return nil, fmt.Errorf("风格不存在: %w", err)
	}
	if !style.IsActive {
		return nil, fmt.Errorf("风格已禁用")
	}

	// 从风格中获取参数
	prompt := req.Prompt
	negativePrompt := ""
	width := 512
	height := 512
	cfgScale := 7.0
	steps := 20
	sampler := "Euler a"

	if style.PresetParams != nil {
		if v, ok := style.PresetParams["negative_prompt"].(string); ok {
			negativePrompt = v
		}
	}

	// 创建数据库记录
	record := &models.GenerationRecord{
		UserID:         userID,
		Mode:           models.ModeSimple,
		Prompt:         prompt,
		NegativePrompt: negativePrompt,
		StyleID:        &req.StyleID,
		Width:          width,
		Height:         height,
		CFGScale:       cfgScale,
		Steps:          steps,
		Sampler:        sampler,
		Seed:           -1,
		Status:         models.StatusPending,
	}

	if err := s.genRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("创建生成记录失败: %w", err)
	}

	// 异步生成
	go s.processGeneration(record.ID)

	return &models.GenerateResponse{
		RecordID: record.ID,
		Status:   models.StatusPending,
	}, nil
}

// GeneratePro 专业模式生成：自定义参数
func (s *GenerationService) GeneratePro(ctx context.Context, userID uint64, req *models.ProGenerateRequest) (*models.GenerateResponse, error) {
	// 填充默认值
	if req.Width == 0 {
		req.Width = 512
	}
	if req.Height == 0 {
		req.Height = 512
	}
	if req.CFGScale == 0 {
		req.CFGScale = 7.0
	}
	if req.Steps == 0 {
		req.Steps = 20
	}
	if req.Sampler == "" {
		req.Sampler = "Euler a"
	}
	if req.Seed == 0 {
		req.Seed = -1
	}

	// 创建数据库记录
	record := &models.GenerationRecord{
		UserID:         userID,
		Mode:           models.ModePro,
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Width:          req.Width,
		Height:         req.Height,
		CFGScale:       req.CFGScale,
		Steps:          req.Steps,
		Sampler:        req.Sampler,
		Seed:           req.Seed,
		Status:         models.StatusPending,
	}

	if err := s.genRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("创建生成记录失败: %w", err)
	}

	// 异步生成
	go s.processGeneration(record.ID)

	return &models.GenerateResponse{
		RecordID: record.ID,
		Status:   models.StatusPending,
	}, nil
}

// GetStatus 查询生成状态
func (s *GenerationService) GetStatus(ctx context.Context, id uint64) (*models.GenerationRecord, error) {
	return s.genRepo.FindByID(ctx, id)
}

// processGeneration 后台生成流程
func (s *GenerationService) processGeneration(recordID uint64) {
	// 并发控制
	s.sem <- struct{}{}
	defer func() { <-s.sem }()

	ctx := context.Background()

	// 更新状态为 processing
	record, err := s.genRepo.FindByID(ctx, recordID)
	if err != nil {
		return
	}
	record.Status = models.StatusProcessing
	s.genRepo.Update(ctx, record)

	startTime := time.Now()

	// 调用 SD 生成
	result, err := s.sdClient.Generate(ctx, &GenerateParams{
		Prompt:         record.Prompt,
		NegativePrompt: record.NegativePrompt,
		Width:          record.Width,
		Height:         record.Height,
		Steps:          record.Steps,
		CFGScale:       record.CFGScale,
		Sampler:        record.Sampler,
		Seed:           record.Seed,
	})

	duration := time.Since(startTime).Milliseconds()
	record.DurationMs = duration

	if err != nil {
		record.Status = models.StatusFailed
		record.ErrorMessage = err.Error()
		s.genRepo.Update(ctx, record)
		return
	}

	// 保存图片到本地
	imageURL, err := saveGeneratedImage(s.cfg.Upload.Dir, record.UserID, recordID, result.ImageData)
	if err != nil {
		record.Status = models.StatusFailed
		record.ErrorMessage = fmt.Sprintf("保存图片失败: %v", err)
		s.genRepo.Update(ctx, record)
		return
	}

	// 更新记录为完成
	record.Status = models.StatusCompleted
	record.ImageURL = imageURL
	record.Seed = result.Seed
	s.genRepo.Update(ctx, record)
}

// saveGeneratedImage 保存生成的图片到本地
func saveGeneratedImage(uploadDir string, userID, recordID uint64, imageData []byte) (string, error) {
	// 确保目录存在
	dir := filepath.Join(uploadDir, fmt.Sprintf("%d", userID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	// 生成文件名
	filename := fmt.Sprintf("%d_%s.png", time.Now().UnixMilli(), uuid.New().String()[:8])
	filePath := filepath.Join(dir, filename)

	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		return "", err
	}

	// 返回相对路径
	return filepath.Join(fmt.Sprintf("%d", userID), filename), nil
}
