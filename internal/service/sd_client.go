package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"novel2comic/backend/internal/config"
)

// SDProvider SD 生图接口抽象
type SDProvider interface {
	Generate(ctx context.Context, params *GenerateParams) (*GenerateResult, error)
}

// GenerateParams 生成参数
type GenerateParams struct {
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	Steps          int     `json:"steps"`
	CFGScale       float64 `json:"cfg_scale"`
	Sampler        string  `json:"sampler_name"`
	Seed           int64   `json:"seed"`
}

// GenerateResult 生成结果
type GenerateResult struct {
	ImageData []byte `json:"image_data"` // 解码后的图片数据
	Seed      int64  `json:"seed"`
}

type mockProvider struct {
	dir   string
	delay time.Duration
}

// sdClient SD 客户端实现
type sdClient struct {
	cfg      *config.Config
	provider SDProvider
}

// NewSDClient 创建 SD 客户端 (根据配置选择实现)
func NewSDClient(cfg *config.Config) *sdClient {
	c := &sdClient{cfg: cfg}

	switch strings.ToLower(strings.TrimSpace(cfg.SD.Provider)) {
	case "mock":
		c.provider = &mockProvider{
			dir:   cfg.SD.MockImageDir,
			delay: time.Duration(cfg.SD.MockDelayMs) * time.Millisecond,
		}
	case "a1111":
		c.provider = &a1111Provider{baseURL: cfg.SD.BaseURL}
	case "replicate":
		c.provider = &replicateProvider{apiKey: cfg.SD.APIKey}
	case "":
		if cfg.SD.APIKey != "" {
			c.provider = &replicateProvider{apiKey: cfg.SD.APIKey}
		} else if cfg.SD.BaseURL != "" {
			c.provider = &a1111Provider{baseURL: cfg.SD.BaseURL}
		}
	}

	return c
}

// Generate 调用 SD 生成图片
func (c *sdClient) Generate(ctx context.Context, params *GenerateParams) (*GenerateResult, error) {
	if c.provider == nil {
		return nil, fmt.Errorf("SD 后端未配置: 请设置 sd.base_url (A1111) 或 sd.api_key (Replicate)")
	}
	return c.provider.Generate(ctx, params)
}

// IsConfigured 检查 SD 是否已配置
func (c *sdClient) IsConfigured() bool {
	return c.provider != nil
}

// ============ mock 实现 ============

func (mp *mockProvider) Generate(ctx context.Context, params *GenerateParams) (*GenerateResult, error) {
	select {
	case <-time.After(mp.delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	files, err := imageFiles(mp.dir)
	if err != nil {
		return nil, fmt.Errorf("读取演示图片目录失败: %w", err)
	}
	if len(files) == 0 {
		return nil, errors.New("演示图片目录为空")
	}

	index := stableIndex(params, len(files))
	data, err := os.ReadFile(files[index])
	if err != nil {
		return nil, err
	}

	return &GenerateResult{ImageData: data, Seed: int64(index)}, nil

}

func imageFiles(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	pathLists := make([]string, 0, len(files))

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(f.Name())) {
		case ".jpg", ".jpeg", ".png", ".webp":
			pathLists = append(pathLists, filepath.Join(dir, f.Name()))
		default:
			continue
		}
	}
	sort.Strings(pathLists)
	return pathLists, nil
}

func stableIndex(params *GenerateParams, size int) int {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%d|%d",
		params.Prompt,
		params.NegativePrompt,
		params.Width,
		params.Height,
	)

	sum := h.Sum(nil)
	value := binary.BigEndian.Uint64(sum[:8])
	return int(value % uint64(size))
}

// ============ A1111 实现 ============

type a1111Provider struct {
	baseURL string
}

// a1111Request A1111 txt2img 请求结构
type a1111Request struct {
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	Steps          int     `json:"steps"`
	CFGScale       float64 `json:"cfg_scale"`
	SamplerName    string  `json:"sampler_name"`
	Seed           int64   `json:"seed"`
}

// a1111Response A1111 txt2img 响应结构
type a1111Response struct {
	Images     []string               `json:"images"` // base64 编码的图片
	Parameters map[string]interface{} `json:"parameters"`
	Info       string                 `json:"info"`
}

func (p *a1111Provider) Generate(ctx context.Context, params *GenerateParams) (*GenerateResult, error) {
	reqBody := a1111Request{
		Prompt:         params.Prompt,
		NegativePrompt: params.NegativePrompt,
		Width:          params.Width,
		Height:         params.Height,
		Steps:          params.Steps,
		CFGScale:       params.CFGScale,
		SamplerName:    params.Sampler,
		Seed:           params.Seed,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/sdapi/v1/txt2img", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 SD API 失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SD API 返回错误 (%d): %s", resp.StatusCode, string(respBytes))
	}

	var a1111Resp a1111Response
	if err := json.Unmarshal(respBytes, &a1111Resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(a1111Resp.Images) == 0 {
		return nil, fmt.Errorf("SD 未返回图片")
	}

	// 解码 base64 图片
	imageData, err := base64.StdEncoding.DecodeString(a1111Resp.Images[0])
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %w", err)
	}

	// 从 info 中提取 seed
	seed := params.Seed
	if a1111Resp.Info != "" {
		var info map[string]interface{}
		if json.Unmarshal([]byte(a1111Resp.Info), &info) == nil {
			if s, ok := info["seed"].(float64); ok {
				seed = int64(s)
			}
		}
	}

	return &GenerateResult{
		ImageData: imageData,
		Seed:      seed,
	}, nil
}

// ============ Replicate 实现 ============

type replicateProvider struct {
	apiKey string
}

func (p *replicateProvider) Generate(ctx context.Context, params *GenerateParams) (*GenerateResult, error) {
	// TODO: 实现 Replicate API 调用
	// POST https://api.replicate.com/v1/predictions
	// 轮询直到完成，下载输出图片
	return nil, fmt.Errorf("Replicate 后端暂未实现")
}
