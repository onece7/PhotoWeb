-- novel2comic 数据库初始化脚本
-- 使用方法: mysql -u root -p < backend/migrations/001_init.sql

CREATE DATABASE IF NOT EXISTS novel2comic
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE novel2comic;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username    VARCHAR(64)   NOT NULL,
    phone       VARCHAR(20)   NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_username (username),
    UNIQUE KEY uk_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 图片风格表 (普通模式预设)
CREATE TABLE IF NOT EXISTS image_styles (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name          VARCHAR(64)   NOT NULL,
    display_name  VARCHAR(128)  NOT NULL,
    description   TEXT,
    preset_params JSON,
    sort_order    INT           NOT NULL DEFAULT 0,
    preview_url   VARCHAR(512),
    is_active     TINYINT(1)    NOT NULL DEFAULT 1,
    created_at    TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 生成记录表
CREATE TABLE IF NOT EXISTS generation_records (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL,
    mode            ENUM('simple','pro') NOT NULL,
    prompt          TEXT           NOT NULL,
    negative_prompt TEXT,
    style_id        BIGINT UNSIGNED DEFAULT NULL,
    width           INT            NOT NULL DEFAULT 512,
    height          INT            NOT NULL DEFAULT 512,
    cfg_scale       DECIMAL(4,2)   NOT NULL DEFAULT 7.00,
    steps           INT            NOT NULL DEFAULT 20,
    sampler         VARCHAR(64)    NOT NULL DEFAULT 'Euler a',
    seed            BIGINT         NOT NULL DEFAULT -1,
    image_url       VARCHAR(512),
    thumbnail_url   VARCHAR(512),
    status          ENUM('pending','processing','completed','failed') NOT NULL DEFAULT 'pending',
    error_message   TEXT,
    duration_ms     BIGINT         NOT NULL DEFAULT 0,
    created_at      TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    CONSTRAINT fk_record_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_record_style FOREIGN KEY (style_id) REFERENCES image_styles(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入预设风格数据
INSERT INTO image_styles (name, display_name, description, preset_params, sort_order) VALUES
('anime',       '动漫风格', '日系动漫风格，色彩鲜艳，线条清晰',
  '{"negative_prompt": "realistic, photo, 3d render", "cfg_scale": 7, "steps": 20, "sampler": "Euler a"}', 1),
('realistic',   '写实风格', '照片级写实风格，细节丰富',
  '{"negative_prompt": "cartoon, anime, illustration", "cfg_scale": 7.5, "steps": 30, "sampler": "DPM++ 2M Karras"}', 2),
('watercolor',  '水彩风格', '柔和的水彩画风格',
  '{"negative_prompt": "photo, realistic, 3d", "cfg_scale": 6.5, "steps": 25, "sampler": "Euler a"}', 3),
('oil_painting','油画风格', '古典油画风格，笔触明显',
  '{"negative_prompt": "photo, digital art, anime", "cfg_scale": 7, "steps": 30, "sampler": "DPM++ 2M Karras"}', 4),
('pixel_art',   '像素风格', '复古像素艺术风格',
  '{"negative_prompt": "realistic, smooth, high resolution", "cfg_scale": 7, "steps": 20, "sampler": "Euler a"}', 5),
('sketch',      '素描风格', '黑白素描/线稿风格',
  '{"negative_prompt": "color, painting, realistic", "cfg_scale": 6, "steps": 20, "sampler": "Euler a"}', 6);
