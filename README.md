# 大运河生态与文化保护平台 (Grand Canal Guardian)

京杭大运河生态与文化保护综合平台 — **模块化单体架构**。

## 架构

```
┌──────────────────────────────────────────┐
│         Go 单体应用 (Gin)                 │
│  /api/v1/auth/*    认证 (注册/登录/刷新)   │
│  /api/v1/users/*   用户 (资料 CRUD)       │
│  /api/v1/posts/*   帖子 (CRUD + 点赞+评论) │
│  /api/v1/quiz/*    问答 (出题/答题/排行)    │
│  /api/v1/map/*     地图 (GIS/时空)        │
│  /api/v1/llm/*     AI 对话 (SSE 流式)     │
│  /api/v1/admin/*   管理后台              │
│  /ws                WebSocket            │
├──────────────────────────────────────────┤
│  PostgreSQL 16 + PostGIS  |  Redis 7  |  MinIO  │
└──────────────────────────────────────────┘
```

> 无 gRPC、无 RabbitMQ、无 Consul。模块间直接函数调用，异步用 goroutine 池。

## 快速启动

```bash
# 1. 环境变量
cp .env.example .env
# 编辑 .env 填入 JWT_SECRET 和 LLM API Key

# 2. 基础设施
docker compose -f deploy/docker-compose/docker-compose.yml up -d
# PostgreSQL:5432, Redis:6379, MinIO:9000

# 3. 启动后端
go run ./cmd/server         # 或: make run

# 4. 前端 (新终端)
cd web && npm install && npm run dev

# 5. 验证
curl http://localhost:8080/health
# → {"status": "ok"}
```

访问:
- 前端: http://localhost:5173
- API: http://localhost:8080
- MinIO Console: http://localhost:9001

## 项目结构

```
├── cmd/server/main.go       # 唯一入口
├── internal/
│   ├── handler/             # HTTP 处理器
│   ├── service/             # 业务逻辑
│   ├── repository/          # 数据访问 (GORM)
│   ├── model/               # 数据模型
│   ├── middleware/          # 鉴权/限流/恢复
│   ├── config/              # 配置加载
│   └── worker/              # Goroutine 异步任务池
├── pkg/                     # 公共基础设施库
│   ├── apperror/            # 统一错误码
│   ├── response/            # 统一响应格式
│   ├── auth/                # JWT 双Token + 黑名单 + RBAC
│   ├── log/                 # Zap 日志
│   └── validator/           # 参数校验 + 中文翻译
├── web/                     # Vue 3 + TypeScript 前端
├── migrations/              # SQL 迁移
├── deploy/
│   ├── Dockerfile           # 多阶段构建
│   └── docker-compose/
│       └── docker-compose.yml
├── docs/                    # 架构/API/开发文档
└── scripts/init-db.sql      # 数据库初始化
```

## 生产构建

```bash
cd web && npm run build           # 前端构建 → web/dist/
CGO_ENABLED=0 go build -o gcg ./cmd/server  # 后端编译为单二进制
# 部署: ./gcg (内嵌前端静态文件, 单文件部署)
```

## 文档

- [架构设计](docs/01-architecture-design.md)
- [OpenAPI 规范](docs/02-openapi-spec.yaml)
- [开发者指南](docs/03-developer-guide.md)
- [Go 工程化规范](docs/04-go-engineering-standards.md)
