package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"novel2comic/backend/internal/config"
	"novel2comic/backend/internal/models"
	"novel2comic/backend/internal/repository"
	"novel2comic/backend/internal/router"
	"novel2comic/backend/internal/service"
	"novel2comic/backend/pkg/redis"
	"novel2comic/backend/pkg/response"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(
		&models.User{},
		&models.ImageStyle{},
		&models.GenerationRecord{},
	); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 种子数据：首次启动时插入预设风格
	seedStyles(db)

	// 初始化 Redis
	rdb, err := redis.NewClient(cfg)
	if err != nil {
		log.Fatalf("连接 Redis 失败: %v", err)
	}

	// 初始化统一响应
	response.Init()

	// 依赖注入 - Repository 层
	userRepo := repository.NewUserRepository(db)
	styleRepo := repository.NewStyleRepository(db)
	generationRepo := repository.NewGenerationRepository(db)

	// 依赖注入 - Service 层
	authService := service.NewAuthService(userRepo, rdb, cfg)
	sdClient := service.NewSDClient(cfg)
	generationService := service.NewGenerationService(generationRepo, styleRepo, sdClient, cfg)
	historyService := service.NewHistoryService(generationRepo)

	// 初始化路由
	r := router.Setup(cfg, authService, generationService, historyService)

	// 启动服务器
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// 优雅关闭
	go func() {
		fmt.Printf("🚀 服务启动: http://localhost:%d\n", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\n🛑 正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务关闭失败: %v", err)
	}
	fmt.Println("✅ 服务已安全关闭")
}

// seedStyles 首次启动时插入预设风格种子数据
func seedStyles(db *gorm.DB) {
	var count int64
	if err := db.Model(&models.ImageStyle{}).Count(&count).Error; err != nil {
		log.Printf("检查风格数据失败: %v", err)
		return
	}
	if count > 0 {
		return
	}

	styles := []models.ImageStyle{
		{
			Name: "anime", DisplayName: "动漫风格", Description: "日系动漫风格，色彩鲜艳，线条清晰",
			PresetParams: models.JSONMap{"negative_prompt": "realistic, photo, 3d render", "cfg_scale": 7.0, "steps": float64(20), "sampler": "Euler a"},
			SortOrder: 1, IsActive: true,
		},
		{
			Name: "realistic", DisplayName: "写实风格", Description: "照片级写实风格，细节丰富",
			PresetParams: models.JSONMap{"negative_prompt": "cartoon, anime, illustration", "cfg_scale": 7.5, "steps": float64(30), "sampler": "DPM++ 2M Karras"},
			SortOrder: 2, IsActive: true,
		},
		{
			Name: "watercolor", DisplayName: "水彩风格", Description: "柔和的水彩画风格",
			PresetParams: models.JSONMap{"negative_prompt": "photo, realistic, 3d", "cfg_scale": 6.5, "steps": float64(25), "sampler": "Euler a"},
			SortOrder: 3, IsActive: true,
		},
		{
			Name: "oil_painting", DisplayName: "油画风格", Description: "古典油画风格，笔触明显",
			PresetParams: models.JSONMap{"negative_prompt": "photo, digital art, anime", "cfg_scale": 7.0, "steps": float64(30), "sampler": "DPM++ 2M Karras"},
			SortOrder: 4, IsActive: true,
		},
		{
			Name: "pixel_art", DisplayName: "像素风格", Description: "复古像素艺术风格",
			PresetParams: models.JSONMap{"negative_prompt": "realistic, smooth, high resolution", "cfg_scale": 7.0, "steps": float64(20), "sampler": "Euler a"},
			SortOrder: 5, IsActive: true,
		},
		{
			Name: "sketch", DisplayName: "素描风格", Description: "黑白素描/线稿风格",
			PresetParams: models.JSONMap{"negative_prompt": "color, painting, realistic", "cfg_scale": 6.0, "steps": float64(20), "sampler": "Euler a"},
			SortOrder: 6, IsActive: true,
		},
	}

	if err := db.Create(&styles).Error; err != nil {
		log.Printf("插入种子风格数据失败: %v", err)
		return
	}
	log.Println("✅ 已插入 6 条预设风格种子数据")
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Info
	if cfg.Server.Mode == "release" {
		logLevel = logger.Warn
	}

	db, err := gorm.Open(mysql.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	return db, nil
}
