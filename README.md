# 大运河生态与文化保护平台 (Grand Canal Guardian)

微服务架构的京杭大运河生态与文化保护综合平台。

## 快速启动

```bash
# 1. 复制环境变量
cp .env.example .env

# 2. 启动基础设施
docker compose -f deploy/docker-compose/docker-compose.yml up -d postgres redis rabbitmq

# 3. 初始化数据库
docker compose -f deploy/docker-compose/docker-compose.yml exec postgres psql -U canal -d grand_canal -f /docker-entrypoint-initdb.d/init.sql

# 4. 启动所有服务 (需要 Go 1.22+)
make run-all

# 5. 验证
curl http://localhost:8080/health
```

## 项目结构

```
├── pkg/           # 公共库 (errors, response, auth, middleware, app...)
├── services/      # 微服务 (user, content, quiz, api-gateway)
├── proto/         # Protobuf 定义
├── deploy/        # Docker Compose + K8s 部署
└── docs/          # 架构文档
```

## 文档

- [架构设计](docs/01-architecture-design.md)
- [OpenAPI 规范](docs/02-openapi-spec.yaml)
- [开发者指南](docs/03-developer-guide.md)
- [Go 工程化规范](docs/04-go-engineering-standards.md)
- [生产级基础设施](docs/05-production-grade-infra.md)
