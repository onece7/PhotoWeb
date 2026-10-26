# novel2comic - 后端服务

文本转图片 Web 应用的后端 API 服务。基于 Go + Gin + GORM，提供用户认证、风格选择、Stable Diffusion 生图、历史记录与图片下载能力。

## 技术栈

- Go 1.21+
- Gin（HTTP 框架）
- GORM + MySQL 8.0（数据持久化）
- Redis 7（Token 缓存 / 并发控制）
- Stable Diffusion（生图，支持 A1111 / Replicate）

## 功能

- 用户注册 / 登录（bcrypt + JWT 双 Token）
- 普通模式：文字 + 风格一键出图
- 专业模式：自定义全部 SD 参数
- 生成历史：分页查询、详情、删除
- 图片下载与静态文件访问

## 目录结构

```
backend/
├── cmd/server/           # 服务入口
├── internal/
│   ├── config/           # 配置加载
│   ├── handlers/         # HTTP 处理器
│   ├── middleware/       # 认证、CORS、日志、限流
│   ├── models/           # 数据模型
│   ├── repository/       # 数据访问层
│   ├── router/           # 路由注册
│   └── service/          # 业务逻辑层
├── migrations/           # 数据库初始化 SQL
├── pkg/                  # JWT、Redis、统一响应等公共包
├── uploads/generated/    # 生成的图片存储目录
├── config.yaml.example   # YAML 配置示例
├── .env.example          # 环境变量示例
└── Dockerfile
```

## 快速开始

### 1. 环境要求

- Go 1.21+
- MySQL 8.0
- Redis 7

### 2. 准备配置

```bash
# 方式一：环境变量
cp .yaml.example .env

# 方式二：YAML 配置
cp config.yaml.example config.yaml
```

两种方式都可以，环境变量会覆盖 YAML 中的同名配置。

### 3. 启动服务

```bash
go mod tidy
go run ./cmd/server/main.go
```

服务默认监听 `http://localhost:8080`，健康检查地址为 `GET /api/v1/health`。

### 4. Docker 运行

```bash
docker build -t novel2comic-backend .
docker run --rm -p 8080:8080 \
  -e N2C_DATABASE_HOST=host.docker.internal \
  -e N2C_REDIS_HOST=host.docker.internal \
  -e N2C_JWT_SECRET=your-secret \
  novel2comic-backend
```

## API 概览

所有接口统一响应格式：`{"code": 0, "message": "success", "data": {...}}`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/health | 健康检查 |
| POST | /api/v1/auth/register | 注册 |
| POST | /api/v1/auth/login | 登录 |
| POST | /api/v1/auth/refresh | 刷新 Token |
| POST | /api/v1/auth/logout | 登出 |
| GET | /api/v1/generate/styles | 风格列表 |
| POST | /api/v1/generate/simple | 普通生成 |
| POST | /api/v1/generate/pro | 专业生成 |
| GET | /api/v1/generate/status/:id | 查询生成状态 |
| GET | /api/v1/history | 历史列表 |
| GET | /api/v1/history/:id | 历史详情 |
| DELETE | /api/v1/history/:id | 删除记录 |
| GET | /api/v1/files/*filepath | 访问生成的图片 |

## 配置项

完整配置项见 [config.yaml.example](./config.yaml.example) 与 [.env.example](.yaml.example)。生产环境请务必修改 `N2C_JWT_SECRET` 与数据库、Redis 密码。
