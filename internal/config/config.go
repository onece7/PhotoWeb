package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	SD       SDConfig       `mapstructure:"sd"`
	Upload   UploadConfig   `mapstructure:"upload"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug | release | test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 秒
}

// DSN 返回 MySQL 连接字符串
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBName)
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Addr 返回 Redis 地址
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret           string `mapstructure:"secret"`
	AccessTokenTTL   int    `mapstructure:"access_token_ttl"`   // 分钟
	RefreshTokenTTL  int    `mapstructure:"refresh_token_ttl"`  // 小时
}

// SDConfig Stable Diffusion 配置
type SDConfig struct {
	BaseURL        string `mapstructure:"base_url"`        // A1111 地址
	APIKey         string `mapstructure:"api_key"`          // Replicate API Key
	MaxConcurrent  int    `mapstructure:"max_concurrent"`   // 最大并发数
	RequestTimeout int    `mapstructure:"request_timeout"`  // 请求超时(秒)
}

// UploadConfig 上传配置
type UploadConfig struct {
	Dir      string `mapstructure:"dir"`       // 图片存储目录
	MaxSize  int64  `mapstructure:"max_size"`  // 最大文件大小(MB)
}

// Load 加载配置: 先读 config.yaml，环境变量可覆盖
func Load() (*Config, error) {
	v := viper.New()

	// 配置文件路径
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./backend")

	// 环境变量覆盖
	v.AutomaticEnv()
	v.SetEnvPrefix("N2C")

	// 设置默认值
	setDefaults(v)

	// 读取配置文件 (可选，不存在时使用默认值+环境变量)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
		// 配置文件不存在时继续
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")

	v.SetDefault("database.host", "127.0.0.1")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.user", "root")
	v.SetDefault("database.password", "")
	v.SetDefault("database.dbname", "novel2comic")
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.conn_max_lifetime", 3600)

	v.SetDefault("redis.host", "127.0.0.1")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.access_token_ttl", 15)
	v.SetDefault("jwt.refresh_token_ttl", 168) // 7 天

	v.SetDefault("sd.max_concurrent", 2)
	v.SetDefault("sd.request_timeout", 300)

	v.SetDefault("upload.dir", "uploads/generated")
	v.SetDefault("upload.max_size", 10)
}
