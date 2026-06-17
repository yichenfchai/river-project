# 大运河生态与文化保护平台 (Grand Canal Guardian)

京杭大运河生态与文化保护综合平台。

## 架构 (单体)

```
main-service/          ← Go 单体 (Gin + GORM + Redis)
├── /api/v1/auth/*    认证 (注册/登录/刷新/登出)
├── /api/v1/users/*   用户 (资料CRUD)
├── /api/v1/posts/*   帖子 (CRUD + 点赞 + 评论)
├── /api/v1/quiz/*    问答 (出题/答题/排行榜)
├── /api/v1/leaderboard 排行榜
├── /api/v1/admin/*   管理后台
│
pkg/                   ← 公共库 (errors/response/auth/middleware/app)
│
Vision Service (Phase 2)  ← Python/FastAPI (YOLO 识别)
LLM Service   (Phase 2)  ← Python/FastAPI (AI 对话)
```

## 快速启动

```bash
# 1. 复制环境变量
cp .env.example .env
# 编辑 .env 填入 JWT_SECRET

# 2. 启动基础设施 + 主服务
make run

# 3. 初始化数据库
make db-init

# 4. 测试
curl http://localhost:8080/health
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test1234","email":"test@test.com"}'
```

## 项目结构

```
├── pkg/                     公共库
├── services/main-service/   单体主服务
│   ├── cmd/main.go          统一入口 + 路由注册
│   └── internal/            handler/service/repository/model
├── proto/                   Protobuf 定义
├── deploy/docker-compose/   Postgres + Redis
├── scripts/init-db.sql      DDL + 种子数据
└── docs/                    架构文档
```

├── web/                     Vue 3 前端
│   ├── src/views/          页面 (Home/Posts/Quiz/Profile/Admin)
│   ├── src/api/            API 客户端 (auth/posts/quiz/admin)
│   ├── src/stores/         Pinia 状态 (auth)
│   └── src/router/         路由 + 守卫

## 开发

```bash
# 后端
cp .env.example .env
make run

# 前端
cd web
npm install
npm run dev       # :3000, 自动代理 API 到 :8080
```

## 生产构建

```bash
cd web && npm run build    # 输出到 web/dist/
cd .. && make run          # Go 直接服务前端静态文件
```

## 文档

- [架构设计](docs/01-architecture-design.md)
- [OpenAPI 规范](docs/02-openapi-spec.yaml)
- [开发者指南](docs/03-developer-guide.md)
- [Go 工程化规范](docs/04-go-engineering-standards.md)
