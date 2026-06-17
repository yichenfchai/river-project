# 大运河生态与文化保护平台 — 开发者文档

> **架构**: Go 单体 (services/main-service) + Python 旁路 (Vision/LLM Phase 2)
> **Phase 1**: pkg/ 公共库 + 主服务 (Auth/Users/Posts/Quiz/Admin)
> **Phase 2**: LLM/Vision Python 服务 + Elasticsearch + K8s + RabbitMQ

> **版本**: v1.0 | **日期**: 2026-06-16
>
> 面向后端、前端、AI 开发者的完整开发指南。

---

## 目录

1. [快速开始](#1-快速开始)
2. [项目结构](#2-项目结构)
3. [Go 后端开发](#3-go-后端开发)
4. [Python AI 服务开发](#4-python-ai-服务开发)
5. [前端开发](#5-前端开发)
6. [数据库](#6-数据库)
7. [Redis 缓存开发](#7-redis-缓存开发)
8. [gRPC 与 Protobuf](#8-grpc-与-protobuf)
9. [消息队列 (RabbitMQ)](#9-消息队列-rabbitmq)
10. [测试规范](#10-测试规范)
11. [CI/CD 流程](#11-cicd-流程)
12. [部署指南](#12-部署指南)
13. [故障排查](#13-故障排查)

---

## 1. 快速开始

### 1.1 环境要求

| 工具 | 版本 | 说明 |
|------|------|------|
| Docker | 24.x+ | 本地开发环境 |
| Docker Compose | v2.x+ | 编排 |
| Go | 1.22+ | 后端服务 |
| Python | 3.11+ | AI 服务 (conda 管理) |
| Node.js | 20 LTS | 前端 |
| protoc | 25.x+ | protobuf 编译 |
| golangci-lint | 1.55+ | Go 代码检查 |
| ruff | 0.3+ | Python 代码检查 |

### 1.2 一键启动（3 分钟）

```bash
# 1. 克隆仓库
git clone git@github.com:your-org/grand-canal-guardian.git
cd grand-canal-guardian

# 2. 复制环境变量
cp .env.example .env
# 编辑 .env 填入 LLM API Key (QWEN_API_KEY / DEEPSEEK_API_KEY)

# 3. 启动所有基础设施 + 服务
docker compose up -d

# 4. 初始化数据库
docker compose exec postgres-master bash -c \
  "psql -U canal -d grand_canal < /docker-entrypoint-initdb.d/init.sql"

# 5. 加载种子数据（运河 POI、问答题目）
go run scripts/seed/main.go

# 6. 验证
curl http://localhost:8080/api/v1/map/layers  # 应返回图层列表
curl http://localhost:8080/health             # 应返回 ok
```

**访问地址**:
- 前端: http://localhost:3000
- API Gateway: http://localhost:8080
- RabbitMQ 管理: http://localhost:15672 (guest/guest)
- MinIO Console: http://localhost:9001 (minioadmin/minioadmin)

### 1.3 单独启动某个服务（调试用）

```bash
# Go 服务示例
cd services/main-service
cp .env.example .env
go run cmd/main.go

# Python 服务示例
cd services/llm-service
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8101
```

---

## 2. 项目结构

> **工程化规范详见**: `04-go-engineering-standards.md` — 包含 pkg/ 公共库全部代码设计

```
grand-canal-guardian/
│
├── pkg/                           # ★ 公共库 — 所有 Go 服务共享 (零重复)
│   ├── errors/                    #   统一错误封装
│   │   ├── codes.go               #     错误码枚举 (5位: 模块+分类+序号)
│   │   ├── app_error.go           #     AppError 结构体 (Code + Message + Cause)
│   │   └── i18n.go                #     中文错误消息映射
│   ├── response/                  #   统一响应格式
│   │   └── response.go            #     {code, message, data} + 分页
│   ├── auth/                      # ★ 鉴权 (生产级) → 详见 05-production-grade-infra.md
│   │   ├── middleware.go           #   Auth 中间件
│   │   ├── token.go               #   TokenManager (双 Token + 黑名单)
│   │   ├── blacklist.go           #   Redis Token 黑名单
│   │   ├── session.go             #   多设备 Session 管理
│   │   └── rbac.go                #   RBAC 位掩码权限矩阵
│   ├── ratelimit/                 # ★ 令牌桶限流 (生产级)
│   │   ├── limiter.go             #   Redis Lua 令牌桶
│   │   ├── token_bucket.lua       #   Lua 脚本 (原子操作)
│   │   └── middleware.go           #   三级限流中间件 (IP/User/API)
│   ├── database/                  # ★ 连接池 (生产级)
│   │   ├── pool.go                #   连接池管理 + Prometheus 指标
│   │   └── health.go              #   健康检查注册
│   ├── log/                       # ★ 日志系统 (生产级)
│   │   ├── logger.go              #   Zap 工厂 + 轮转 + 脱敏
│   │   └── middleware.go           #   Gin 日志中间件
│   ├── middleware/                 # 通用中间件 (RequestID/CORS/Trace)
│   ├── validator/                 #   参数校验
│   │   └── validator.go           #     go-playground + 中文翻译
│   ├── app/                       #   启动器
│   │   ├── app.go                 #     Gin 引擎工厂 + 中间件注册
│   │   ├── shutdown.go            #     优雅关闭 (SIGINT/SIGTERM)
│   │   ├── health.go              #     /health + /ready (K8s 探针)
│   │   └── logger.go              #     Zap 工厂
│   ├── transaction/               #   事务封装
│   │   └── tx.go                  #     WithTx / WithTxResult 装饰器
│   ├── grpc/                      #   gRPC 工具
│   │   ├── error_convert.go       #     AppError ↔ gRPC Status
│   │   └── interceptors.go        #     Unary 拦截器
│   ├── log/                       #   日志
│   │   └── logger.go
│   └── go.mod
│
├── services/                     # 微服务
│   ├── api-gateway/              # 统一网关 (Go + Gin)
│   │   ├── cmd/main.go           # 入口 (用 pkg/app 启动)
│   │   ├── internal/
│   │   │   ├── proxy/            # 反向代理路由
│   │   │   └── config/           # 配置加载
│   │   ├── docs/                 # swaggo 自动生成
│   │   ├── Dockerfile
│   │   └── go.mod
│   │
│   ├── user-service/             # 用户 & 认证 (Go)
│   │   ├── cmd/main.go           # app.New() + wire.InitializeApp() + app.Run()
│   │   ├── internal/
│   │   │   ├── handler/          # HTTP handler — 只调 response.OK/Error
│   │   │   ├── service/          # 业务逻辑 — 只返回 *errors.AppError
│   │   │   ├── repository/       # 数据访问 — 返回原始 error
│   │   │   ├── model/            # 数据模型 + Request DTO
│   │   │   ├── config/           # Viper 配置
│   │   │   └── wire/             # ★ 依赖注入 (wire.go + wire_gen.go)
│   │   ├── migrations/           # 数据库迁移
│   │   ├── docs/                 # swaggo 自动生成
│   │   ├── Dockerfile
│   │   └── go.mod
│   │
│   ├── content-service/          # 内容 & 帖子 (Go)
│   │   └── ...                   # (同 user-service 结构，复用 pkg/)
│   │
│   ├── map-service/              # GIS 地图 (Go)
│   │   └── ...
│   │
│   ├── quiz-service/             # 问答 & 积分 (Go)
│   │   └── ...
│   │
│   ├── notification-service/     # 实时通知 (Go + WebSocket)
│   │   └── ...
│   │
│   ├── llm-service/              # LLM 对话 & 科普 (Python + FastAPI)
│   │   ├── app/
│   │   │   ├── main.py           # FastAPI 入口
│   │   │   ├── routers/          # 路由
│   │   │   │   ├── chat.py       # 桌宠对话 (SSE)
│   │   │   │   └── story.py      # 科普故事
│   │   │   ├── services/         # 业务逻辑
│   │   │   │   ├── llm_client.py # LLM SDK 封装
│   │   │   │   └── prompt.py     # Prompt 模板
│   │   │   ├── models/           # Pydantic 模型
│   │   │   └── exceptions.py     # Python 侧统一异常 (对齐 Go 错误码)
│   │   ├── requirements.txt
│   │   └── Dockerfile
│   │
│   └── vision-service/           # YOLO 视觉 (Python + FastAPI)
│       ├── app/
│       │   ├── main.py
│       │   ├── routers/
│       │   │   └── classify.py   # 垃圾分类
│       │   ├── services/
│       │   │   ├── yolo_detector.py  # YOLO 推理
│       │   │   └── image_utils.py    # 图片预处理
│       │   └── models/
│       │       └── garbage_yolov8n.onnx
│       ├── requirements.txt
│       └── Dockerfile
│
├── web/                          # 前端 (Vue 3 + TypeScript)
│   ├── src/
│   │   ├── pages/                # 页面 (.vue)
│   │   │   ├── Home.vue
│   │   │   ├── MapView.vue       # 时空地图
│   │   │   ├── StoryPage.vue     # 科普故事会
│   │   │   ├── PlazaPage.vue     # 分享广场
│   │   │   ├── QuizPage.vue      # 趣味问答
│   │   │   ├── LeaderboardPage.vue # 排行榜
│   │   │   └── AdminPage.vue     # 管理后台
│   │   ├── components/           # 通用组件
│   │   │   ├── CanalSprite/      # 运河小精灵桌宠
│   │   │   │   ├── SpriteCanvas.vue
│   │   │   │   └── ChatBubble.vue
│   │   │   ├── MapViewer/        # 地图组件
│   │   │   └── Layout/
│   │   ├── composables/          # 组合式函数 (代替 React hooks)
│   │   ├── stores/               # Pinia 状态管理
│   │   ├── api/                  # API 调用封装
│   │   └── utils/
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
│
├── proto/                        # Protobuf 定义
│   ├── user.proto
│   ├── content.proto
│   ├── map.proto
│   ├── quiz.proto
│   └── gen/                      # 生成的代码 (git-ignored)
│       ├── go/
│       └── python/
│
├── deploy/
│   ├── docker-compose/
│   │   ├── docker-compose.yml
│   │   ├── docker-compose.dev.yml
│   │   └── .env.example
│   └── k8s/
│       ├── base/
│       └── overlays/
│           ├── dev/
│           ├── staging/
│           └── prod/
│
├── docs/
│   ├── 01-architecture-design.md
│   ├── 02-openapi-spec.yaml
│   └── 03-developer-guide.md     # (本文档)
│
├── scripts/
│   ├── init-db.sql               # 初始化 DDL
│   ├── seed/                     # 种子数据
│   ├── proto-gen.sh              # protobuf 代码生成
│   └── deploy.sh                 # 部署脚本
│
├── Makefile                      # 统一命令入口
├── .env.example
├── .golangci.yml
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── deploy.yml
└── README.md
```

---

## 3. Go 后端开发

> **鉴权架构**: Auth 中间件直接在进程内校验 JWT，通过 `c.Set("user_id", ...)` 注入 gin.Context。所有 handler 通过 `c.GetString("user_id")` 获取。无需独立 Gateway 进程。

> **★ 工程化规范**: 所有 Go 服务必须复用 `pkg/` 公共库。
> - 基础框架: [`04-go-engineering-standards.md`](./04-go-engineering-standards.md) — 错误/响应/校验/启动器/事务/gRPC
> - 生产强化: [`05-production-grade-infra.md`](./05-production-grade-infra.md) — 鉴权(黑名单+Refresh旋转) / 令牌桶(Lua) / 连接池(Prometheus) / 日志(Zap+轮转+脱敏)
>
> **关键架构原则**:
> - Handler 层: 只调 `response.OK()` / `response.Error()`，不感知 HTTP 状态码
> - Service 层: 只返回 `*errors.AppError`，不引入 `gin.Context`
> - Repository 层: 返回 Go 原始 `error`，由 Service 用 `errors.WrapDefault()` 包装
> - 所有 panic 由 `middleware.Recovery` 统一捕获并转换为 AppError

### 3.1 项目初始化（新建服务）

```bash
# 以 user-service 为例
mkdir -p services/main-service
cd services/main-service
go mod init github.com/your-org/grand-canal-guardian/services/main-service

# 安装核心依赖
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/redis/go-redis/v9
go get github.com/golang-jwt/jwt/v5
go get google.golang.org/grpc
go get github.com/streadway/amqp      # RabbitMQ
go get go.uber.org/zap                # 日志
go get github.com/prometheus/client_golang  # 指标
```

### 3.2 项目分层规范

```
handler/    ← HTTP 层：参数校验、序列化、响应状态码
   ↓ 调用
service/    ← 业务逻辑层：事务编排、权限校验、缓存策略
   ↓ 调用
repository/ ← 数据访问层：SQL、Redis 操作、ES 查询
   ↓ 使用
model/      ← 数据模型定义 (GORM struct)
```

**规则**:
- handler 不直接调用 repository
- service 不 import `gin.Context`
- repository 不包含业务逻辑

### 3.3 配置管理

使用 Viper + 环境变量：

```go
// internal/config/config.go
package config

import "github.com/spf13/viper"

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    JWT      JWTConfig
}

type ServerConfig struct {
    Port int    `mapstructure:"SERVER_PORT"`
    Mode string `mapstructure:"GIN_MODE"` // debug | release
}

type DatabaseConfig struct {
    Host     string `mapstructure:"DB_HOST"`
    Port     int    `mapstructure:"DB_PORT"`
    User     string `mapstructure:"DB_USER"`
    Password string `mapstructure:"DB_PASSWORD"`
    Name     string `mapstructure:"DB_NAME"`
    // 读写分离 DSN
    ReadDSN  string // 从库
    WriteDSN string // 主库
}

type JWTConfig struct {
    Secret          string `mapstructure:"JWT_SECRET"`
    AccessTokenTTL  int    `mapstructure:"JWT_ACCESS_TTL"`  // 秒，默认 900
    RefreshTokenTTL int    `mapstructure:"JWT_REFRESH_TTL"` // 秒，默认 604800
}

func Load() *Config {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()
    viper.ReadInConfig()

    cfg := &Config{}
    viper.Unmarshal(cfg)

    // 构造读写分离 DSN
    cfg.Database.WriteDSN = fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        cfg.Database.Host, cfg.Database.Port,
        cfg.Database.User, cfg.Database.Password, cfg.Database.Name,
    )
    cfg.Database.ReadDSN = fmt.Sprintf(
        "host=%s-slave port=%d user=%s password=%s dbname=%s sslmode=disable",
        cfg.Database.Host, cfg.Database.Port,
        cfg.Database.User, cfg.Database.Password, cfg.Database.Name,
    )
    return cfg
}
```

### 3.4 读写分离 GORM 配置

```go
// internal/repository/db.go
package repository

import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/plugin/dbresolver"
)

func NewDB(cfg DatabaseConfig) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(cfg.WriteDSN), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, err
    }

    // 配置读写分离
    db.Use(dbresolver.Register(dbresolver.Config{
        Sources:  []gorm.Dialector{postgres.Open(cfg.WriteDSN)},
        Replicas: []gorm.Dialector{postgres.Open(cfg.ReadDSN)},
        Policy:   dbresolver.RandomPolicy{}, // 从库随机选择
    }))

    // 连接池配置
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(25)
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetConnMaxLifetime(time.Hour)

    return db, nil
}
```

**使用方式**:
```go
// 写操作自动走主库
db.Create(&post)

// 读操作自动走从库
db.Where("id = ?", id).First(&post)

// 强制走主库（事务/读自己刚写的数据）
db.Clauses(dbresolver.Write).Where("id = ?", id).First(&post)
```

### 3.5 多级缓存实现（Go）

```go
// internal/service/cache_service.go
package service

import (
    "context"
    "encoding/json"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/patrickmn/go-cache" // L1 本地缓存
)

type CacheService struct {
    local *cache.Cache         // L1: 本地内存 (TTL 10s)
    redis *redis.Client        // L2: Redis (TTL 5min)
    db    *gorm.DB             // L3: PostgreSQL
}

func NewCacheService(rdb *redis.Client, db *gorm.DB) *CacheService {
    return &CacheService{
        local: cache.New(10*time.Second, 30*time.Second),
        redis: rdb,
        db:    db,
    }
}

// GetPost 三级缓存读取
func (c *CacheService) GetPost(ctx context.Context, postID string) (*Post, error) {
    cacheKey := fmt.Sprintf("post:%s:detail", postID)

    // L1: 本地缓存
    if val, found := c.local.Get(cacheKey); found {
        return val.(*Post), nil
    }

    // L2: Redis
    data, err := c.redis.Get(ctx, cacheKey).Bytes()
    if err == nil {
        var post Post
        json.Unmarshal(data, &post)
        c.local.Set(cacheKey, &post, cache.DefaultExpiration) // 回写 L1
        return &post, nil
    }

    // L3: 数据库（带防击穿互斥锁）
    lockKey := "lock:rebuild:" + cacheKey
    if !c.redis.SetNX(ctx, lockKey, 1, 30*time.Second).Val() {
        time.Sleep(50 * time.Millisecond) // 等待重建完成
        return c.GetPost(ctx, postID)     // 重试
    }
    defer c.redis.Del(ctx, lockKey)

    var post Post
    if err := c.db.First(&post, "id = ?", postID).Error; err != nil {
        return nil, err
    }

    // 回写缓存
    go c.writeCache(context.Background(), cacheKey, &post)

    return &post, nil
}

func (c *CacheService) writeCache(ctx context.Context, key string, post *Post) {
    data, _ := json.Marshal(post)
    c.redis.Set(ctx, key, data, 5*time.Minute)
    c.local.Set(key, post, cache.DefaultExpiration)
}

// InvalidatePost 更新时失效缓存
func (c *CacheService) InvalidatePost(ctx context.Context, postID string) {
    key := "post:" + postID + ":detail"
    c.redis.Del(ctx, key)
    c.local.Delete(key)
}
```

### 3.6 限流中间件

```go
// internal/middleware/rate_limit.go
func RateLimitMiddleware(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := fmt.Sprintf("rate_limit:%s:%s", c.ClientIP(), c.FullPath())
        
        count, err := rdb.Incr(c.Request.Context(), key).Result()
        if err != nil {
            c.Next()
            return
        }
        
        if count == 1 {
            rdb.Expire(c.Request.Context(), key, window)
        }
        
        if count > int64(limit) {
            c.AbortWithStatusJSON(429, gin.H{
                "code": 42900, "message": "请求过于频繁，请稍后再试",
            })
            return
        }
        
        c.Next()
    }
}

// 使用: router.Use(RateLimitMiddleware(rdb, 100, time.Minute))
```

### 3.7 gRPC 服务端/客户端示例

```go
// 服务端
lis, _ := net.Listen("tcp", ":9001")
s := grpc.NewServer(
    grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
)
pb.RegisterUserServiceServer(s, &UserServer{})
s.Serve(lis)

// 客户端（带连接池）
conn, _ := grpc.Dial("content-service:9002",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
)
client := pb.NewContentServiceClient(conn)
```

---

## 4. Python AI 服务开发

### 4.1 LLM Service (llm-service)

**依赖** (`requirements.txt`):
```
fastapi==0.110.0
uvicorn[standard]==0.29.0
httpx==0.27.0           # 异步 HTTP 客户端
openai==1.33.0          # 通义千问 / DeepSeek 兼容接口
sse-starlette==2.0.0    # SSE 流式响应
pydantic==2.7.0
python-dotenv==1.0.1
redis[hiredis]==5.0.0
tenacity==8.3.0         # 重试
```

**LLM Client 封装**:
```python
# app/services/llm_client.py
import os
import httpx
from tenacity import retry, stop_after_attempt, wait_exponential

class LLMClient:
    """统一的 LLM 客户端，支持通义千问 / DeepSeek 切换"""

    # 运河小精灵 System Prompt
    SPRITE_SYSTEM_PROMPT = """你是"运河小精灵"，一只生活在京杭大运河畔的千年水精灵。
你活泼可爱，对大运河的历史、生态、文化了如指掌。
你说话带点古风，喜欢用成语和诗句，但也能和小朋友亲切交流。
你的使命是引导人们了解和保护大运河，让这条千年水道永葆生机。

知识范围：
- 大运河 2500 年历史（从邗沟到南水北调）
- 运河流域生态系统（鱼类 100+、鸟类 200+、湿地植物 300+）
- 非物质文化遗产（船工号子、年画、美食）
- 水利工程知识（船闸、水库、水柜）
- 环保知识（垃圾分类、水质保护）
"""

    STORY_SYSTEM_PROMPT = """你是一位大运河科普作家，擅长用生动的故事传递知识。
请根据用户指定的话题和年龄层，创作一篇引人入胜的科普故事。
故事需要：真实历史/科学知识为基础 + 引人入胜的叙事 + 互动性强的结尾提问。"""

    def __init__(self):
        self.provider = os.getenv("LLM_PROVIDER", "qwen")  # qwen | deepseek
        self.api_key = os.getenv("QWEN_API_KEY") if self.provider == "qwen" else os.getenv("DEEPSEEK_API_KEY")
        self.base_url = os.getenv("LLM_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1")
        self.model = os.getenv("LLM_MODEL", "qwen-plus")

    @retry(stop=stop_after_attempt(3), wait=wait_exponential(multiplier=1, min=2, max=10))
    async def chat_stream(self, messages: list, **kwargs):
        """流式对话（返回 async generator）"""
        async with httpx.AsyncClient(timeout=60.0) as client:
            async with client.stream(
                "POST",
                f"{self.base_url}/chat/completions",
                headers={
                    "Authorization": f"Bearer {self.api_key}",
                    "Content-Type": "application/json",
                },
                json={
                    "model": self.model,
                    "messages": messages,
                    "stream": True,
                    "temperature": 0.8,
                    "max_tokens": 1024,
                    **kwargs,
                },
            ) as response:
                response.raise_for_status()
                async for line in response.aiter_lines():
                    if line.startswith("data: "):
                        data = line[6:]
                        if data == "[DONE]":
                            break
                        yield json.loads(data)

    @retry(stop=stop_after_attempt(2), wait=wait_exponential(multiplier=1, min=2, max=10))
    async def chat(self, messages: list, **kwargs) -> str:
        """非流式对话"""
        async with httpx.AsyncClient(timeout=60.0) as client:
            response = await client.post(
                f"{self.base_url}/chat/completions",
                headers={
                    "Authorization": f"Bearer {self.api_key}",
                    "Content-Type": "application/json",
                },
                json={
                    "model": self.model,
                    "messages": messages,
                    "temperature": 0.7,
                    "max_tokens": 2048,
                    **kwargs,
                },
            )
            response.raise_for_status()
            data = response.json()
            return data["choices"][0]["message"]["content"]
```

**SSE 路由**:
```python
# app/routers/chat.py
from fastapi import APIRouter, Request
from sse_starlette.sse import EventSourceResponse
import json

router = APIRouter(prefix="/api/v1/llm", tags=["LLM"])

@router.post("/chat")
async def chat(request: Request, body: ChatRequest):
    """桌宠流式对话 (SSE)"""
    llm = request.app.state.llm_client

    messages = [
        {"role": "system", "content": LLMClient.SPRITE_SYSTEM_PROMPT},
        *get_history(body.session_id),  # 从 Redis 获取历史
        {"role": "user", "content": body.message},
    ]

    async def event_generator():
        full_response = ""
        try:
            async for chunk in llm.chat_stream(messages):
                if "choices" in chunk and chunk["choices"]:
                    delta = chunk["choices"][0].get("delta", {})
                    token = delta.get("content", "")
                    if token:
                        full_response += token
                        yield {
                            "event": "token",
                            "data": json.dumps({"token": token}, ensure_ascii=False),
                        }
            # 保存到 Redis
            save_history(body.session_id, body.message, full_response)
            yield {
                "event": "done",
                "data": json.dumps({"total_tokens": len(full_response)}),
            }
        except Exception as e:
            yield {
                "event": "error",
                "data": json.dumps({"message": str(e)}),
            }

    return EventSourceResponse(event_generator())
```

### 4.2 Vision Service (vision-service)

**依赖** (`requirements.txt`):
```
fastapi==0.110.0
uvicorn[standard]==0.29.0
python-multipart==0.0.9      # 文件上传
ultralytics==8.2.0            # YOLOv8
onnxruntime-gpu==1.18.0       # ONNX GPU 推理 (或用 onnxruntime CPU 版)
opencv-python-headless==4.9.0
Pillow==10.3.0
numpy==1.26.4
```

**YOLO 检测器**:
```python
# app/services/yolo_detector.py
from ultralytics import YOLO
import numpy as np
from PIL import Image
import io

class GarbageDetector:
    """YOLO 垃圾分类检测器"""

    # 类别映射（根据训练模型调整）
    CATEGORY_MAP = {
        0: ("塑料瓶", "可回收物"),
        1: ("玻璃瓶", "可回收物"),
        2: ("废纸", "可回收物"),
        3: ("废电池", "有害垃圾"),
        4: ("废弃药品", "有害垃圾"),
        5: ("厨余垃圾", "厨余垃圾"),
        6: ("塑料袋", "其他垃圾"),
        7: ("烟蒂", "其他垃圾"),
    }

    ADVICE_MAP = {
        "可回收物": "请投入蓝色垃圾桶，回收前请清洗干净",
        "有害垃圾": "请投入红色垃圾桶，注意避免破损",
        "厨余垃圾": "请投入绿色垃圾桶，沥干水分",
        "其他垃圾": "请投入灰色垃圾桶",
    }

    def __init__(self, model_path: str = "models/garbage_yolov8n.onnx"):
        self.model = YOLO(model_path)

    def detect(self, image_bytes: bytes, conf_threshold: float = 0.5) -> dict:
        """检测图片中的垃圾"""
        image = Image.open(io.BytesIO(image_bytes)).convert("RGB")
        results = self.model(image, conf=conf_threshold)

        detections = []
        if results and results[0].boxes:
            for box in results[0].boxes:
                cls_id = int(box.cls[0])
                conf = float(box.conf[0])
                xywh = box.xywh[0].tolist()
                name, category = self.CATEGORY_MAP.get(cls_id, ("未知", "其他垃圾"))

                detections.append({
                    "class_name": name,
                    "category": category,
                    "confidence": round(conf, 4),
                    "bbox": {
                        "x": int(xywh[0]),
                        "y": int(xywh[1]),
                        "w": int(xywh[2]),
                        "h": int(xywh[3]),
                    },
                })

        # 合并建议
        categories = set(d["category"] for d in detections)
        advice = "；".join(self.ADVICE_MAP[c] for c in categories)

        return {
            "image_id": str(uuid.uuid4()),
            "detections": detections,
            "processing_time_ms": round(results[0].speed["inference"], 1) if results else 0,
            "advice": advice or "未检测到垃圾，请重新拍摄",
        }
```

**路由**:
```python
# app/routers/classify.py
from fastapi import APIRouter, UploadFile, File, Form
import time

router = APIRouter(prefix="/api/v1/vision", tags=["Vision"])

@router.post("/classify")
async def classify_garbage(
    image: UploadFile = File(...),
    lat: float = Form(None),
    lng: float = Form(None),
):
    """垃圾分类识别"""
    contents = await image.read()

    start = time.time()
    result = app.state.detector.detect(contents)
    result["processing_time_ms"] = round((time.time() - start) * 1000)

    # 保存上报记录
    if lat and lng:
        await save_report(result["image_id"], lat, lng, result["detections"])

    return result
```

### 4.3 Poetry 依赖管理（可选）

```bash
cd services/llm-service
poetry init
poetry add fastapi uvicorn httpx openai sse-starlette redis tenacity

cd services/vision-service
poetry init
poetry add fastapi uvicorn ultralytics onnxruntime-gpu opencv-python-headless pillow
```

---

## 5. 前端开发

### 5.1 安装与启动

```bash
cd web
npm install

# 开发模式（带 HMR）
npm run dev
# → http://localhost:3000

# 生产构建
npm run build
npm run preview
```

### 5.2 API 调用封装

```typescript
// src/api/client.ts
const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080/api/v1';

class ApiClient {
  private getToken(): string | null {
    return localStorage.getItem('access_token');
  }

  async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const token = this.getToken();
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    };

    const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

    if (res.status === 401) {
      // Token 过期，尝试刷新
      const refreshed = await this.refreshToken();
      if (refreshed) return this.request(path, options);
      // 刷新失败，跳转登录
      window.location.href = '/login';
      throw new Error('请重新登录');
    }

    if (!res.ok) {
      const error = await res.json().catch(() => ({}));
      throw new ApiError(res.status, error.message || '请求失败');
    }

    return res.json();
  }

  async uploadFile(path: string, formData: FormData): Promise<any> {
    const token = this.getToken();
    const res = await fetch(`${API_BASE}${path}`, {
      method: 'POST',
      headers: token ? { Authorization: `Bearer ${token}` } : {},
      body: formData,
    });
    return res.json();
  }
}

export const api = new ApiClient();
```

### 5.3 桌宠组件

```vue
<!-- src/components/CanalSprite/SpriteCanvas.vue -->
<script setup lang="ts">
import { ref } from 'vue'
import ChatBubble from './ChatBubble.vue'

const chatOpen = ref(false)
const messages = ref<ChatMessage[]>([])
const streaming = ref(false)

const sendMessage = async (text: string) => {
  messages.value.push({ role: 'user', content: text })
  streaming.value = true

  const response = await fetch(`${API_BASE}/llm/chat`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ message: text, session_id: sessionId }),
  })

  const reader = response.body!.getReader()
  const decoder = new TextDecoder()
  let assistantMsg = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    const lines = decoder.decode(value).split('\n')
    for (const line of lines) {
      if (line.startsWith('data: ')) {
        const data = JSON.parse(line.slice(6))
        if (data.token) {
          assistantMsg += data.token
          const last = messages.value[messages.value.length - 1]
          if (last?.role === 'assistant') {
            last.content = assistantMsg
          } else {
            messages.value.push({ role: 'assistant', content: assistantMsg })
          }
        }
      }
    }
  }
  streaming.value = false
}
</script>

<template>
  <div class="canal-sprite">
    <canvas ref="canvasRef" @click="chatOpen = !chatOpen" />
    <ChatBubble v-if="chatOpen" :messages="messages" :streaming="streaming" @send="sendMessage" />
  </div>
</template>
```

### 5.4 状态管理 (Pinia)

```typescript
// src/stores/auth.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('access_token'))

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isMonitor = computed(() => user.value?.role === 'monitor')

  async function login(username: string, password: string) {
    const res = await api.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
    localStorage.setItem('access_token', res.access_token)
    localStorage.setItem('refresh_token', res.refresh_token)
    user.value = res.user
    token.value = res.access_token
  }

  function logout() {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    user.value = null
    token.value = null
  }

  return { user, token, isLoggedIn, isAdmin, isMonitor, login, logout }
})
```

### 5.5 原生能力 (定位 / 摄像头 / 文件上传)

纯 Web API，Android + Windows 通用，无需任何第三方库。

#### 定位 — `navigator.geolocation`

```typescript
// src/composables/useLocation.ts
export function useLocation() {
  async function getCurrentPosition(options?: {
    enableHighAccuracy?: boolean
    timeout?: number
  }): Promise<{ lat: number; lng: number; accuracy: number }> {
    return new Promise((resolve, reject) => {
      if (!navigator.geolocation) {
        reject(new Error('浏览器不支持定位'))
        return
      }
      navigator.geolocation.getCurrentPosition(
        (pos) => resolve({
          lat: pos.coords.latitude,
          lng: pos.coords.longitude,
          accuracy: pos.coords.accuracy,
        }),
        (err) => {
          const messages: Record<number, string> = {
            1: '定位权限被拒绝，请在设置中开启',
            2: '无法获取位置信息',
            3: '定位超时，请重试',
          }
          reject(new Error(messages[err.code] || '定位失败'))
        },
        {
          enableHighAccuracy: options?.enableHighAccuracy ?? false,
          timeout: options?.timeout ?? 10000,
        }
      )
    })
  }

  function watchPosition(
    callback: (pos: { lat: number; lng: number }) => void
  ): () => void {
    const id = navigator.geolocation.watchPosition(
      (pos) => callback({ lat: pos.coords.latitude, lng: pos.coords.longitude }),
      () => {},
      { enableHighAccuracy: true }
    )
    return () => navigator.geolocation.clearWatch(id)
  }

  return { getCurrentPosition, watchPosition }
}
```

#### 摄像头 — 移动端 `<input capture>` / 桌面端 `MediaDevices`

```vue
<!-- src/components/CameraCapture.vue -->
<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const emit = defineEmits<{ capture: [blob: Blob] }>()
const isMobile = /Android|iPhone/i.test(navigator.userAgent)
const videoRef = ref<HTMLVideoElement>()
const canvasRef = ref<HTMLCanvasElement>()
const stream = ref<MediaStream | null>(null)

onMounted(async () => {
  if (isMobile) return
  stream.value = await navigator.mediaDevices.getUserMedia({
    video: { facingMode: 'environment', width: 1920, height: 1080 },
    audio: false,
  })
  videoRef.value!.srcObject = stream.value
})

onUnmounted(() => stream.value?.getTracks().forEach(t => t.stop()))

function capture() {
  const video = videoRef.value!
  const canvas = canvasRef.value!
  canvas.width = video.videoWidth
  canvas.height = video.videoHeight
  canvas.getContext('2d')!.drawImage(video, 0, 0)
  canvas.toBlob((blob) => { if (blob) emit('capture', blob) }, 'image/jpeg', 0.85)
}

function onMobileCapture(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (file) emit('capture', file)
}
</script>

<template>
  <input v-if="isMobile" type="file" accept="image/*" capture="environment"
    @change="onMobileCapture" />
  <template v-else>
    <video ref="videoRef" autoplay playsinline />
    <button @click="capture">拍照</button>
    <canvas ref="canvasRef" style="display:none" />
  </template>
</template>
```

#### 文件上传 — `<input type="file">`

```vue
<input type="file" accept="image/*,.pdf" multiple @change="onPick" />
```

#### 监测员上报完整流程

```vue
<script setup lang="ts">
import { useLocation } from '@/composables/useLocation'
import { api } from '@/api/client'

const { getCurrentPosition } = useLocation()

async function submitGarbageReport(photoBlob: Blob) {
  const [pos] = await Promise.allSettled([
    getCurrentPosition({ enableHighAccuracy: true, timeout: 8000 }),
  ])

  const formData = new FormData()
  formData.append('image', photoBlob, 'garbage.jpg')
  if (pos.status === 'fulfilled') {
    formData.append('lat', String(pos.value.lat))
    formData.append('lng', String(pos.value.lng))
  }

  return api.uploadFile('/vision/classify', formData)
}
</script>
```

> **Android 注意**: `getUserMedia` 需要 HTTPS（`localhost` 除外）。生产部署务必配 SSL。

---

## 6. 数据库

### 6.1 DDL 核心表

```sql
-- init.sql (简化版，完整版见 scripts/init-db.sql)

-- 用户表
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username    VARCHAR(32) NOT NULL UNIQUE,
    password    VARCHAR(255) NOT NULL,      -- bcrypt hash
    email       VARCHAR(128) NOT NULL UNIQUE,
    nickname    VARCHAR(64) DEFAULT '',
    avatar_url  TEXT DEFAULT '',
    bio         VARCHAR(200) DEFAULT '',
    role        VARCHAR(16) NOT NULL DEFAULT 'user',  -- user | monitor | admin
    banned      BOOLEAN DEFAULT FALSE,
    banned_reason TEXT DEFAULT '',
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_created ON users(created_at);

-- 帖子表（按月分区）
CREATE TABLE posts (
    id          UUID DEFAULT gen_random_uuid(),
    author_id   UUID NOT NULL REFERENCES users(id),
    title       VARCHAR(128) NOT NULL,
    content     TEXT NOT NULL,
    images      TEXT[] DEFAULT '{}',
    tags        VARCHAR(32)[] DEFAULT '{}',
    topic       VARCHAR(16) NOT NULL DEFAULT 'share',
    status      VARCHAR(16) NOT NULL DEFAULT 'pending',  -- pending | approved | rejected
    like_count  INT DEFAULT 0,
    comment_count INT DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- 创建分区（每月一个）
CREATE TABLE posts_2026_06 PARTITION OF posts
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE posts_2026_07 PARTITION OF posts
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
-- ... 执行 scripts/create-partitions.sql 自动生成未来 12 个月分区

CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_topic ON posts(topic);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_created ON posts(created_at DESC);
CREATE INDEX idx_posts_tags ON posts USING GIN(tags);

-- 评论表
CREATE TABLE comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id     UUID NOT NULL,
    author_id   UUID NOT NULL REFERENCES users(id),
    content     VARCHAR(2000) NOT NULL,
    parent_id   UUID REFERENCES comments(id),
    reply_to_id UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_comments_post ON comments(post_id, created_at);

-- 点赞表
CREATE TABLE likes (
    user_id UUID NOT NULL REFERENCES users(id),
    post_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, post_id)
);

-- 问答题目表
CREATE TABLE questions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question    TEXT NOT NULL,
    options     TEXT[] NOT NULL,
    answer      VARCHAR(8) NOT NULL,
    explanation TEXT DEFAULT '',
    difficulty  VARCHAR(8) NOT NULL DEFAULT 'easy',  -- easy | medium | hard
    category    VARCHAR(32) NOT NULL,
    created_by  UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_questions_diff_cat ON questions(difficulty, category);

-- 答题记录表（按月分区）
CREATE TABLE quiz_records (
    id          UUID DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    question_id UUID NOT NULL,
    user_answer VARCHAR(8) NOT NULL,
    is_correct  BOOLEAN NOT NULL,
    points_earned INT DEFAULT 0,
    answer_time_ms INT DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- 积分表
CREATE TABLE user_points (
    user_id         UUID PRIMARY KEY REFERENCES users(id),
    total_points    INT DEFAULT 0,
    total_answers   INT DEFAULT 0,
    correct_answers INT DEFAULT 0,
    current_streak  INT DEFAULT 0,
    max_streak      INT DEFAULT 0,
    rank_title      VARCHAR(32) DEFAULT '青铜守护者',
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 大运河 POI 表
CREATE TABLE canal_pois (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(128) NOT NULL,
    type        VARCHAR(32) NOT NULL,
    lat         DOUBLE PRECISION NOT NULL,
    lng         DOUBLE PRECISION NOT NULL,
    era         VARCHAR(32),
    description TEXT,
    image_url   TEXT,
    route_id    VARCHAR(32),
    geom        GEOMETRY(Point, 4326),
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_pois_geom ON canal_pois USING GIST(geom);
CREATE INDEX idx_pois_type ON canal_pois(type);
CREATE INDEX idx_pois_route ON canal_pois(route_id);

-- 运河路线表
CREATE TABLE canal_routes (
    id          VARCHAR(32) PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    era         VARCHAR(32),
    length_km   DOUBLE PRECISION,
    geojson     JSONB NOT NULL,
    description TEXT
);

-- 科普故事表
CREATE TABLE stories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(256) NOT NULL,
    content     TEXT NOT NULL,
    topic       VARCHAR(32) NOT NULL,
    target_age  VARCHAR(16) NOT NULL,
    style       VARCHAR(32) DEFAULT 'storytelling',
    poi_related UUID[] DEFAULT '{}',
    tags        VARCHAR(32)[] DEFAULT '{}',
    cover_prompt TEXT,
    like_count  INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- 垃圾分类上报记录
CREATE TABLE garbage_reports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id UUID NOT NULL REFERENCES users(id),
    image_url   TEXT NOT NULL,
    detections  JSONB DEFAULT '[]',
    lat         DOUBLE PRECISION,
    lng         DOUBLE PRECISION,
    status      VARCHAR(16) DEFAULT 'pending',  -- pending | verified | dismissed
    reported_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 6.2 从库复制配置

```sql
-- Master 配置 (postgresql.conf)
wal_level = replica
max_wal_senders = 5
wal_keep_size = 1024   -- 1GB WAL 保留

-- Slave 配置
-- 1. 基础备份
pg_basebackup -h master -D /var/lib/postgresql/data -U replicator -P -R

-- 2. 启动后自动从 WAL 恢复
```

### 6.3 查询示例

```sql
-- 周边 POI 查询（5000m 半径）
SELECT id, name, type, ST_Distance(geom, ST_SetSRID(ST_MakePoint(119.42, 32.39), 4326)) AS dist
FROM canal_pois
WHERE ST_DWithin(geom, ST_SetSRID(ST_MakePoint(119.42, 32.39), 4326), 0.045)  -- ≈ 5000m
ORDER BY dist;

-- 热门帖子（走从库）
SELECT p.*, u.nickname, u.avatar_url
FROM posts p JOIN users u ON p.author_id = u.id
WHERE p.status = 'approved'
ORDER BY p.like_count DESC
LIMIT 20;
```

---

## 7. Redis 缓存开发

### 7.1 Redis Cluster 配置

```yaml
# docker-compose.yml 片段
redis-node-0:
  image: redis:7-alpine
  command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf
  ports: ["6379:6379"]

redis-node-1:
  image: redis:7-alpine
  command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf

# 初始化集群
docker exec redis-node-0 redis-cli --cluster create \
  node0:6379 node1:6379 node2:6379 --cluster-replicas 0
```

### 7.2 缓存 Key 设计规范

```
# 用户
user:{id}:profile         → Hash    → 用户信息
user:{id}:sessions        → Set     → 在线 session
user:{id}:liked_posts     → Set     → 点赞过的帖子

# 帖子
post:{id}:detail          → String  → 帖子详情 (JSON)
posts:list:{topic}:{page}  → String  → 列表缓存
posts:hot                  → ZSet    → 热门帖子 (score=like_count)

# 排行榜
leaderboard:daily         → ZSet    → member=user_id, score=points
leaderboard:weekly        → ZSet
leaderboard:total         → ZSet

# 问答
quiz:pool:easy             → List    → 简单题 ID 池
quiz:pool:medium           → List
quiz:pool:hard             → List
quiz:session:{id}         → Hash    → 答题会话

# 限流
rate_limit:{ip}:{endpoint} → String  → 计数

# Token 黑名单
blacklist:{jti}            → String  → 1, TTL=Token 剩余有效期

# 分布式锁
lock:{resource}            → String  → NX, TTL=30s

# 地图瓦片
map:tile:{layer}:{z}:{x}:{y} → Binary → MVT/PBF
```

---

## 8. gRPC 与 Protobuf

### 8.1 proto 定义示例

```protobuf
// proto/user.proto
syntax = "proto3";
option go_package = "github.com/your-org/grand-canal-guardian/proto/gen/go/userpb";

service UserService {
  rpc GetUser(GetUserRequest) returns (UserResponse);
  rpc VerifyToken(VerifyTokenRequest) returns (VerifyTokenResponse);
}

message GetUserRequest {
  string user_id = 1;
}

message UserResponse {
  string id = 1;
  string username = 2;
  string nickname = 3;
  string avatar_url = 4;
  string role = 5;
}

message VerifyTokenRequest {
  string token = 1;
}

message VerifyTokenResponse {
  bool valid = 1;
  string user_id = 2;
  string role = 3;
}
```

### 8.2 代码生成

```bash
# scripts/proto-gen.sh
#!/bin/bash
PROTO_DIR=proto
OUT_GO=${PROTO_DIR}/gen/go
OUT_PY=${PROTO_DIR}/gen/python

# Go
protoc --proto_path=${PROTO_DIR} \
  --go_out=${OUT_GO} --go_opt=paths=source_relative \
  --go-grpc_out=${OUT_GO} --go-grpc_opt=paths=source_relative \
  ${PROTO_DIR}/*.proto

# Python
protoc --proto_path=${PROTO_DIR} \
  --python_out=${OUT_PY} \
  --grpc_python_out=${OUT_PY} \
  ${PROTO_DIR}/*.proto
```

---

## 9. 消息队列 (RabbitMQ)

### 9.1 交换机与队列设计

```
Exchange: content.events (topic)
  ├── Queue: content.review      ← 帖子审核 (routing: review.#)
  ├── Queue: content.notify      ← 帖子通知 (routing: notify.#)
  └── Queue: content.es_index    ← ES 同步 (routing: index.#)

Exchange: quiz.events (topic)
  ├── Queue: quiz.score_update   ← 积分更新 (routing: score.#)
  └── Queue: quiz.rank_refresh   ← 排行榜刷新 (routing: rank.#)

Exchange: vision.events (topic)
  ├── Queue: vision.classify_result ← 分类结果 (routing: result.#)
  └── Queue: vision.report_store   ← 上报存储 (routing: report.#)
```

### 9.2 发布/消费示例 (Go)

```go
// 发布
ch.Publish(
    "content.events",       // exchange
    "review.post.created",  // routing key
    false, false,
    amqp.Publishing{
        ContentType: "application/json",
        Body:        jsonBody,
    },
)

// 消费（带重试 + 死信）
msgs, _ := ch.Consume(
    "content.review", // queue
    "", false,        // auto-ack = false (手动确认)
    false, false, false, nil,
)
for msg := range msgs {
    if err := processReview(msg.Body); err != nil {
        msg.Nack(false, true) // requeue
    } else {
        msg.Ack(false)
    }
}
```

---

## 10. 测试规范

### 10.1 Go 测试

```bash
# 单元测试
go test ./... -v -cover

# 带竞态检测
go test ./... -race

# 查看覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

```go
// internal/service/post_service_test.go
func TestCreatePost(t *testing.T) {
    // Setup
    db := setupTestDB(t)       // 测试数据库
    redis := setupTestRedis(t) // 测试 Redis
    svc := NewPostService(db, redis)

    // Test
    post, err := svc.CreatePost(context.Background(), &CreatePostReq{
        AuthorID: "test-user-id",
        Title:    "测试帖子",
        Content:  "这是一篇测试内容",
        Topic:    "share",
    })

    assert.NoError(t, err)
    assert.Equal(t, "测试帖子", post.Title)
    assert.Equal(t, "pending", post.Status)

    // 验证缓存
    cached, err := svc.GetPost(context.Background(), post.ID)
    assert.NoError(t, err)
    assert.Equal(t, post.Title, cached.Title)
}
```

### 10.2 Python 测试

```bash
# 运行测试
cd services/llm-service
pytest -v --cov=app --cov-report=term-missing

cd services/vision-service
pytest -v --cov=app
```

```python
# tests/test_garbage_detector.py
import pytest
from app.services.yolo_detector import GarbageDetector

@pytest.fixture
def detector():
    return GarbageDetector(model_path="models/garbage_yolov8n.onnx")

def test_detect_plastic_bottle(detector):
    with open("tests/fixtures/plastic_bottle.jpg", "rb") as f:
        result = detector.detect(f.read())

    assert len(result["detections"]) > 0
    assert result["detections"][0]["category"] == "可回收物"
    assert result["detections"][0]["confidence"] > 0.5

def test_empty_image(detector):
    with open("tests/fixtures/clean_street.jpg", "rb") as f:
        result = detector.detect(f.read())

    assert len(result["detections"]) == 0
```

### 10.3 前端测试

```bash
cd web
npm test            # vitest
npm run test:e2e    # Playwright (可选)
```

### 10.4 集成测试

```bash
# 启动完整环境后运行
docker compose -f docker-compose.test.yml up -d
go test ./... -tags=integration
```

---

## 11. CI/CD 流程

### 11.1 GitHub Actions 配置

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  lint-go:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: golangci-lint run ./services/...

  lint-py:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with: { python-version: '3.11' }
      - run: pip install ruff
      - run: ruff check services/llm-service services/vision-service

  test-go:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgis/postgis:16-3.4
        env: { POSTGRES_USER: canal, POSTGRES_PASSWORD: test, POSTGRES_DB: testdb }
      redis:
        image: redis:7-alpine
    steps:
      - uses: actions/checkout@v4
      - run: go test ./... -race -cover

  test-py:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: |
          pip install -r services/llm-service/requirements.txt
          cd services/llm-service && pytest

  build:
    needs: [lint-go, lint-py, test-go, test-py]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: |
          docker build -t gcg-api-gateway services/main-service (原 Gateway 逻辑已合并)
          docker build -t gcg-user-service services/main-service
          # ... 构建所有服务
```

---

## 12. 部署指南

### 12.1 开发环境 (Docker Compose)

```bash
# 完整启动
docker compose -f deploy/docker-compose/docker-compose.yml \
               -f deploy/docker-compose/docker-compose.dev.yml \
               up -d --build

# 查看日志
docker compose logs -f api-gateway

# 重启单个服务
docker compose restart user-service
```

### 12.2 生产环境 (K8s)

```bash
# 1. 构建并推送镜像
docker build -t harbor.example.com/gcg/api-gateway:v1.0.0 services/main-service (原 Gateway 逻辑已合并)
docker push harbor.example.com/gcg/api-gateway:v1.0.0

# 2. 部署到 K8s
kubectl apply -k deploy/k8s/overlays/prod

# 3. 金丝雀发布
kubectl set image deployment/api-gateway api-gateway=harbor.example.com/gcg/api-gateway:v1.1.0-canary
# 验证新版本健康
kubectl rollout status deployment/api-gateway
# 全量发布
kubectl set image deployment/api-gateway api-gateway=harbor.example.com/gcg/api-gateway:v1.1.0

# 4. 回滚
kubectl rollout undo deployment/api-gateway
```

### 12.3 环境变量清单

```bash
# .env.example
# ---- LLM ----
LLM_PROVIDER=qwen                # qwen | deepseek
LLM_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
LLM_MODEL=qwen-plus              # qwen-plus | qwen-max | deepseek-chat
QWEN_API_KEY=sk-xxxxxxxx
DEEPSEEK_API_KEY=sk-xxxxxxxx     # 备选

# ---- Database ----
DB_HOST=postgres-master
DB_PORT=5432
DB_USER=canal
DB_PASSWORD=your_secure_password
DB_NAME=grand_canal

# ---- Redis ----
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0

# ---- JWT ----
JWT_SECRET=your-256-bit-secret-key-change-in-production
JWT_ACCESS_TTL=900               # 15 分钟
JWT_REFRESH_TTL=604800           # 7 天

# ---- RabbitMQ ----
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/

# ---- MinIO (S3) ----
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=gcg-uploads

# ---- Elasticsearch ----
ES_ADDR=http://elasticsearch:9200

# ---- Consul ----
CONSUL_ADDR=consul:8500

# ---- Monitoring ----
PROMETHEUS_PORT=9090
OTEL_EXPORTER_JAEGER_ENDPOINT=http://jaeger:14268/api/traces
```

---

## 13. 故障排查

### 13.1 常见问题

| 问题 | 可能原因 | 解决 |
|------|----------|------|
| `connect ECONNREFUSED 127.0.0.1:8080` | API Gateway 未启动 | `docker compose ps` 检查状态 |
| `401 Unauthorized` | Token 过期 | 检查 JWT_ACCESS_TTL，刷新 Token |
| `429 Too Many Requests` | 限流触发 | 检查 Redis 限流计数，调整阈值 |
| `connection pool timeout` | DB 连接池满 | 增加 `max_open_conns`，检查慢查询 |
| `Redis OOM` | 缓存未设 TTL | 检查所有 Key 是否有 EXPIRE |
| LLM 返回乱码 | Prompt 语言问题 | 确认 System Prompt 为中文 |
| YOLO 检测慢 (>500ms) | CPU 推理 | 切换到 ONNX GPU 推理 |
| 主从复制延迟大 | 大事务阻塞 | 拆分大事务，增加从库资源 |
| 帖子搜索不准确 | ES 索引未更新 | 检查 ES 同步消费者状态 |

### 13.2 调试命令

```bash
# 查看服务日志
docker compose logs -f --tail=100 user-service

# 进入容器调试
docker compose exec user-service sh

# 检查数据库连接
docker compose exec postgres-master psql -U canal -d grand_canal -c "SELECT count(*) FROM users;"

# 检查 Redis 缓存
docker compose exec redis redis-cli
> KEYS post:*
> GET post:xxx:detail

# 检查 RabbitMQ 队列
curl -u guest:guest http://localhost:15672/api/queues | jq '.[] | {name, messages, consumers}'

# 性能测试（wrk）
wrk -t4 -c100 -d30s --latency http://localhost:8080/api/v1/leaderboard

# 链路追踪
# 在 Jaeger UI (http://localhost:16686) 查看请求链路
```

### 13.3 性能调优检查清单

- [ ] PostgreSQL: `EXPLAIN ANALYZE` 确认索引命中
- [ ] Redis: `INFO stats` 查看命中率（应 > 90%）
- [ ] gRPC: 使用连接复用（非每次新建）
- [ ] Go: `pprof` 分析 CPU/内存热点
- [ ] Python: `async` 非阻塞 IO，避免同步调用
- [ ] 前端: Code splitting + 图片懒加载

---

## 附录

### A. Makefile

```makefile
.PHONY: up down build test lint seed

up:
	docker compose -f deploy/docker-compose/docker-compose.yml up -d

down:
	docker compose -f deploy/docker-compose/docker-compose.yml down

build:
	docker compose -f deploy/docker-compose/docker-compose.yml build

test:
	go test ./... -race -count=1
	cd services/llm-service && pytest
	cd services/vision-service && pytest
	cd web && npm test

lint:
	golangci-lint run ./services/...
	ruff check services/llm-service services/vision-service
	cd web && npm run lint

proto:
	bash scripts/proto-gen.sh

seed:
	go run scripts/seed/main.go

db-reset:
	docker compose exec postgres-master psql -U canal -d grand_canal -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	docker compose exec postgres-master psql -U canal -d grand_canal < scripts/init-db.sql
	make seed
```

### B. Git 提交规范

```
feat: 新功能
fix: Bug 修复
docs: 文档
style: 代码格式（不影响功能）
refactor: 重构
perf: 性能优化
test: 测试
chore: 构建/工具

示例:
feat(user-service): add RBAC middleware
fix(quiz): prevent duplicate answer submission
docs: add API reference for vision service
```

### C. 代码审查检查点

- [ ] 是否遵循分层架构（handler → service → repository）
- [ ] 是否有适当的错误处理和日志
- [ ] 数据库查询是否使用索引
- [ ] 缓存是否设置了合理的 TTL
- [ ] 是否有 SQL 注入风险（使用参数化查询）
- [ ] 敏感信息是否硬编码（密码、Key）
- [ ] 是否有适当的限流和权限检查
- [ ] 是否添加了单元测试

---

> **文档维护**: 本文档随项目迭代持续更新。任何架构/流程变更必须同步更新本文档。
