# 大运河生态与文化保护平台 — 架构设计文档

> **版本**: v2.0 | **日期**: 2026-06-20 | **作者**: Architecture Team
>
> **项目代号**: Grand Canal Guardian (GCG)
> **架构模式**: 模块化单体 (Modular Monolith)

---

## 目录

1. [系统概述](#1-系统概述)
2. [模块化单体架构](#2-模块化单体架构)
3. [技术选型](#3-技术选型)
4. [模块详细设计](#4-模块详细设计)
5. [数据层架构](#5-数据层架构)
6. [缓存策略](#6-缓存策略)
7. [高并发设计](#7-高并发设计)
8. [安全架构](#8-安全架构)
9. [部署架构](#9-部署架构)
10. [监控与可观测性](#10-监控与可观测性)

---

## 1. 系统概述

### 1.1 项目定位

面向公众、生态监测员、管理员的大运河生态与文化保护综合平台，融合 GIS 时空地图、AI 科普互动、社区分享、AI 桌宠、趣味问答等功能。

### 1.2 为什么是单体而非微服务

| 维度 | 本项目实际情况 | 结论 |
|------|---------------|------|
| 团队规模 | 小团队 / 个人开发 | 微服务运维成本远超收益 |
| 业务耦合度 | 用户-帖子-问答-地图高度关联 | 拆服务 = 分布式事务噩梦 |
| 流量规模 | 峰值 QPS 预估 < 1000 | 单体 + 水平扩展完全够用 |
| 部署复杂度 | 需要快速迭代上线 | 单体一个二进制搞定 |
| 未来演进 | 真有瓶颈时再拆，不晚 | YAGNI (You Ain't Gonna Need It) |

**一句话**：大运河平台的功能量级和数据规模，用微服务是杀鸡用牛刀。模块化单体（Modular Monolith）——代码按模块组织、部署为单一进程——是最理性的选择。

### 1.3 核心功能矩阵

| 模块 | 用户角色 | 描述 |
|------|---------|------|
| 用户认证 | 普通用户 / 监测员 / 管理员 | 三角色 RBAC，JWT 鉴权 |
| 时空地图 | 全部 | 大运河历史变迁 GIS 交互地图 |
| 科普故事会 | 全部 | LLM 驱动强交互叙事体验 |
| 分享广场 | 普通用户 / 监测员 | 图文帖子、评论、点赞 |
| 运河小精灵 | 普通用户 | 页面桌宠 + LLM 对话 |
| 趣味问答 | 普通用户 | 积分排名、段位系统 |
| 垃圾识别 | 监测员 | YOLO 视觉分类 + 上报 |
| 管理后台 | 管理员 | 用户管理、内容审核、数据看板 |

### 1.4 非功能需求

- **并发**: 峰值 QPS 1000+
- **可用性**: 99.5% SLA（单体多实例水平扩展）
- **响应时间**: P99 < 200ms（读），P99 < 500ms（写）
- **数据安全**: HTTPS + AES-256 敏感字段加密 + RBAC
- **扩展性**: 无状态，水平扩展多实例 + Nginx 负载均衡

---

## 2. 模块化单体架构

### 2.1 整体拓扑

```
                         ┌──────────────────────────────┐
                         │        Nginx / Caddy          │
                         │  (HTTPS + 静态文件 + 负载均衡)   │
                         └──────────────┬───────────────┘
                                        │
                    ┌───────────────────▼────────────────────┐
                    │          Go 单体应用 (Gin)               │
                    │  ┌───────────────────────────────────┐  │
                    │  │          Middleware 层             │  │
                    │  │  JWT/RBAC | CORS | RateLimit      │  │
                    │  │  Recovery | RequestID | Logger    │  │
                    │  ├───────────────────────────────────┤  │
                    │  │          Handler 层                │  │
                    │  │  /api/v1/auth/*    → authHandler  │  │
                    │  │  /api/v1/users/*   → userHandler  │  │
                    │  │  /api/v1/posts/*   → postHandler  │  │
                    │  │  /api/v1/quiz/*    → quizHandler  │  │
                    │  │  /api/v1/map/*     → mapHandler   │  │
                    │  │  /api/v1/llm/*     → llmHandler   │  │
                    │  │  /api/v1/admin/*   → adminHandler │  │
                    │  │  /ws               → wsHandler    │  │
                    │  ├───────────────────────────────────┤  │
                    │  │          Service 层 (业务逻辑)      │  │
                    │  │  UserSvc | PostSvc | QuizSvc     │  │
                    │  │  MapSvc  | LLMSvc  | VisionSvc   │  │
                    │  ├───────────────────────────────────┤  │
                    │  │          Repository 层 (GORM)     │  │
                    │  │  UserRepo | PostRepo | QuizRepo  │  │
                    │  │  MapRepo  | AdminRepo            │  │
                    │  ├───────────────────────────────────┤  │
                    │  │          Async Worker 池          │  │
                    │  │  goroutine pool + channel         │  │
                    │  │  审核任务 | 图片处理 | 积分结算     │  │
                    │  └───────────────────────────────────┘  │
                    └──┬──────────────┬──────────────┬───────┘
                       │              │              │
         ┌─────────────▼──┐  ┌───────▼───┐  ┌──────▼──────┐
         │  PostgreSQL 16  │  │  Redis 7  │  │  MinIO      │
         │  + PostGIS      │  │  (缓存/   │  │  (图片/     │
         │  (主数据库)      │  │   会话)   │  │   文件)     │
         └─────────────────┘  └───────────┘  └─────────────┘
```

### 2.2 模块间通信

| 场景 | 方式 | 说明 |
|------|------|------|
| 模块间调用 | 直接 Go 函数调用 | 同进程，零开销 |
| 异步任务 | goroutine pool + channel | 发帖审核、图片处理、积分结算 |
| 实时推送 | WebSocket (同进程) | 桌宠对话 SSE 流、问答排名推送 |
| 外部 AI 调用 | HTTP/2 + JSON | 通义千问 API / DeepSeek API |

### 2.3 模块依赖方向

```
                    ┌──────────┐
                    │  Handler  │  ← HTTP 入口，调用 Service
                    └────┬─────┘
                         │
                    ┌────▼─────┐
                    │  Service │  ← 业务逻辑，调用 Repository + 外部 API
                    └────┬─────┘
                         │
                    ┌────▼──────┐
                    │ Repository│  ← 数据访问，GORM 封装
                    └────┬──────┘
                         │
                    ┌────▼──────┐
                    │   Model   │  ← 数据结构定义
                    └───────────┘

依赖规则：上层依赖下层，下层不依赖上层。同层模块通过接口解耦。
```

---

## 3. 技术选型

### 3.1 后端

| 组件 | 选型 | 理由 |
|------|------|------|
| Web 框架 | Go + Gin | 高性能 HTTP，中间件生态丰富 |
| 数据库 ORM | GORM | Go 最成熟 ORM |
| 缓存 | go-redis | Redis 客户端 |
| 配置管理 | Viper | YAML/ENV 多源配置 |
| 日志 | Zap | 高性能结构化日志 |
| 参数校验 | go-playground/validator | 结构体 tag 校验 |
| 迁移 | golang-migrate | SQL 版本迁移 |
| 任务队列 | 自建 goroutine pool | 无需外部 MQ，降低运维复杂度 |

### 3.2 前端

| 组件 | 选型 |
|------|------|
| 框架 | Vue 3 + TypeScript |
| 地图 | Leaflet (CDN 加载，OSM 瓦片) |
| 状态管理 | Pinia |
| UI 库 | Element Plus |
| 桌宠 | 自研 Canvas 渲染 |
| 构建 | Vite |

### 3.3 数据层

| 组件 | 选型 | 用途 |
|------|------|------|
| 主数据库 | PostgreSQL 16 + PostGIS | 所有业务数据（用户、帖子、积分、GIS） |
| 缓存 | Redis 7 | 会话、热点数据、排行榜、限流计数 |
| 对象存储 | MinIO (S3 兼容) | 图片/文件 |
| 全文搜索 | PostgreSQL tsvector | 帖子搜索（小规模场景无需 ES） |
| 搜索引擎 | Elasticsearch (可选) | 帖子量 > 10w 时启用 |

### 3.4 AI 集成

| 能力 | 方案 | 说明 |
|------|------|------|
| LLM 对话/科普 | 通义千问 API / DeepSeek API | HTTP 调用，Service 层封装 |
| 垃圾分类识别 | YOLOv8 (ultralytics) | Python 脚本/Golang CGO 调用，或外部 API |
| 内容审核 | LLM API + 敏感词库 | 异步 goroutine 处理 |

### 3.5 基础设施

| 组件 | 选型 |
|------|------|
| 容器化 | Docker Compose (dev) / 单 Dockerfile (prod) |
| CI/CD | GitHub Actions |
| 监控 | Prometheus + Grafana |
| 日志 | Zap → stdout → Docker logs / Loki |

---

## 4. 模块详细设计

### 4.1 用户模块 (User)

- **职责**: 注册、登录、RBAC 权限、个人资料
- **路由前缀**: `/api/v1/auth`, `/api/v1/users`

```
POST   /api/v1/auth/register        # 注册
POST   /api/v1/auth/login           # 登录 → JWT
POST   /api/v1/auth/refresh         # 刷新 Token
POST   /api/v1/auth/logout          # 登出 → Token 入黑名单
GET    /api/v1/users/me             # 当前用户信息
PUT    /api/v1/users/me             # 更新资料
GET    /api/v1/admin/users          # 管理员: 用户列表
PUT    /api/v1/admin/users/:id/role # 管理员: 改角色
```

**JWT Payload**:
```json
{
  "sub": "user_uuid",
  "role": "user|monitor|admin",
  "exp": 1718515200,
  "iat": 1718428800
}
```

### 4.2 内容模块 (Content/Posts)

- **职责**: 分享广场帖子、评论、点赞
- **路由前缀**: `/api/v1/posts`

```
GET    /api/v1/posts                 # 帖子列表（分页+搜索）
GET    /api/v1/posts/:id            # 帖子详情
POST   /api/v1/posts                # 发帖
PUT    /api/v1/posts/:id            # 编辑
DELETE /api/v1/posts/:id            # 删除
POST   /api/v1/posts/:id/like      # 点赞/取消
POST   /api/v1/posts/:id/comments  # 评论
GET    /api/v1/posts/:id/comments  # 评论列表
```

**审核流程**: 发帖 → 同步写 DB (status=pending) → 提交 goroutine pool → LLM 审核 + 敏感词过滤 → 更新状态 (approved/rejected)

### 4.3 地图模块 (Map)

- **职责**: 大运河时空 GIS 数据、POI、历史图层
- **路由前缀**: `/api/v1/map`

```
GET    /api/v1/map/layers                   # 可用图层列表
GET    /api/v1/map/pois?lat=&lng=&radius=   # 周边 POI
GET    /api/v1/map/timeline?year=           # 历年运河数据
GET    /api/v1/map/routes/:id               # 运河路线详情
```

**图层设计**（前端 Leaflet + GeoJSON）:
- 隋唐运河 (581-907)
- 元明清运河 (1271-1911)
- 现代南水北调东线
- 生态监测站点
- 文化遗产 POI

### 4.4 问答模块 (Quiz)

- **职责**: 趣味问答、积分、排行榜、段位
- **路由前缀**: `/api/v1/quiz`

```
GET    /api/v1/quiz/questions         # 获取题目（分难度）
POST   /api/v1/quiz/submit           # 提交答案
GET    /api/v1/quiz/leaderboard      # 排行榜（日/周/总）
GET    /api/v1/quiz/users/:id/stats  # 个人统计
```

**积分设计**:
```
积分 = 正确数 × 难度系数 × 连续正确加成
段位: 青铜(0) → 白银(500) → 黄金(1500) → 钻石(5000) → 运河守护者(15000)
```

**防作弊**: 题目从 Redis 预载池随机抽取，答案 HMAC 签名防篡改

### 4.5 AI 模块 (LLM)

- **职责**: 运河小精灵对话、科普故事生成、内容审核
- **路由前缀**: `/api/v1/llm`
- **实现**: 封装 LLM API 调用（通义千问 / DeepSeek 可配置切换）

```
POST   /api/v1/llm/chat              # 桌宠对话 (SSE 流式)
POST   /api/v1/llm/story/generate    # 生成科普故事
GET    /api/v1/llm/health            # 健康检查
```

**桌宠 System Prompt 示例**:
```
你是"运河小精灵"，一只生活在京杭大运河畔的千年水精灵。
你活泼可爱，对大运河的历史、生态、文化了如指掌。
你说话带点古风，喜欢用成语和诗句，但也能和小朋友亲切交流。
你的使命是引导人们了解和保护大运河。
```

**SSE 流式响应** (桌宠对话):
```
event: token
data: {"token": "***", "index": 0}

event: token
data: {"token": "***", "index": 1}

event: done
data: {"total_tokens": 45}
```

### 4.6 视觉模块 (Vision)

- **职责**: 垃圾分类识别、图片审核
- **路由前缀**: `/api/v1/vision`
- **实现**: 调用 ultralytics Python 子进程 / ONNX Runtime / 或外部视觉 API

```
POST   /api/v1/vision/classify       # 垃圾图片分类
GET    /api/v1/vision/health         # 健康检查
```

**分类类别** (8 类):
```
可回收物, 有害垃圾, 厨余垃圾, 其他垃圾,
塑料袋, 玻璃瓶, 废电池, 废弃药品
```

### 4.7 异步任务模块 (Async Worker)

- **职责**: 处理非实时任务，避免阻塞 HTTP 响应
- **实现**: 自建 goroutine pool（buffer channel 实现），无需 RabbitMQ

```
任务类型:
├── content_review     # 发帖审核 (LLM 审核 + 敏感词)
├── image_process      # 图片缩略图/压缩
├── score_settle       # 积分结算/周榜计算
└── notification       # WebSocket 推送通知
```

**Worker 池配置**:
- Worker 数量: 10 (可通过配置调整)
- 任务队列容量: 1000
- 优雅关闭: 等待队列任务完成 (最长 30s)

---

## 5. 数据层架构

### 5.1 单数据库设计

所有模块共用同一个 PostgreSQL 数据库，通过 schema 或表前缀逻辑隔离。

```
PostgreSQL 16 (grand_canal_db)
├── public schema (默认)
│   ├── users              # 用户表
│   ├── user_sessions      # 会话表
│   ├── posts              # 帖子表
│   ├── comments           # 评论表
│   ├── likes              # 点赞表
│   ├── questions          # 题目表
│   ├── quiz_records       # 答题记录
│   ├── quiz_leaderboard   # 排行榜
│   ├── map_layers         # 地图图层元数据
│   ├── map_pois           # POI 数据
│   ├── map_routes         # 运河路线
│   ├── notifications      # 通知
│   └── audit_logs         # 审计日志
├── postgis extension      # 空间数据支持
└── pg_trgm extension      # 模糊搜索支持
```

### 5.2 为什么不需要分库

| 微服务分库理由 | 本项目实际情况 |
|---------------|---------------|
| 独立扩缩容 | 单库完全够用，PostgreSQL 能支撑百万级数据 |
| 技术栈异构 | 全部用 PostgreSQL，无异构需求 |
| 故障隔离 | 单体应用已共享进程，分库无意义 |
| 团队独立 | 小团队，无组织隔离需求 |

### 5.3 索引设计

| 表 | 索引 | 类型 |
|----|------|------|
| users | (username) UNIQUE, (email) UNIQUE, (role) | B-tree |
| posts | (user_id, created_at), (status, created_at) | B-tree |
| posts | (title, content) | GIN (tsvector) |
| comments | (post_id, created_at) | B-tree |
| quiz_records | (user_id, created_at) | B-tree |
| map_pois | (location) | GiST (PostGIS) |

### 5.4 连接池配置

```yaml
database:
  max_open_conns: 25       # 最大连接数
  max_idle_conns: 10       # 最大空闲连接
  conn_max_lifetime: 5m    # 连接最大生命周期
  conn_max_idle_time: 1m   # 空闲连接超时
```

---

## 6. 缓存策略

### 6.1 两级缓存架构

```
  Request
     │
     ▼
┌─────────────┐  命中   ┌─────────────┐
│ L1: 本地内存  │◄───────│  Response   │
│ (go-cache)   │        └─────────────┘
│ TTL: 10s    │
└──────┬──────┘
       │ 未命中
       ▼
┌─────────────┐  命中   ┌─────────────┐
│ L2: Redis    │◄───────│  Response   │
│ TTL: 5min   │        └─────────────┘
└──────┬──────┘
       │ 未命中
       ▼
┌─────────────┐
│ PostgreSQL  │ ──────► 回写 L2 + L1
└─────────────┘
```

### 6.2 Redis 数据结构

| Key Pattern | 类型 | TTL | 说明 |
|-------------|------|-----|------|
| `user:{id}:profile` | Hash | 1h | 用户信息缓存 |
| `post:{id}:detail` | String(JSON) | 30min | 帖子详情 |
| `leaderboard:daily` | Sorted Set | 24h | 日排行 |
| `leaderboard:weekly` | Sorted Set | 7d | 周排行 |
| `leaderboard:total` | Sorted Set | ∞ | 总排行 |
| `quiz:pool:{difficulty}` | List | ∞ | 题池预载 |
| `rate_limit:{ip}:{endpoint}` | String | 1min | 限流计数 |
| `session:{token_id}` | String | Token TTL | JWT 黑名单 |
| `lock:{resource}` | String (NX) | 30s | 分布式锁 |

### 6.3 缓存更新策略

| 场景 | 策略 | 说明 |
|------|------|------|
| 用户资料 | Write-Through | 写 DB 同时更新 Redis |
| 帖子详情 | Cache-Aside | 读时缓存，写时失效 |
| 排行榜 | 定时刷新 | 每 5min 从 DB 计算 Top100 |
| 限流 | Redis INCR + EXPIRE | 滑动窗口 |

---

## 7. 高并发设计

### 7.1 限流策略

```
Token Bucket (令牌桶)
├── 全局: 1000 req/s
├── 单 IP:  100 req/s
├── LLM 接口:  50 req/s (成本控制)
└── 存储: Redis rate_limit:{key}
```

### 7.2 异步化

| 操作 | 模式 |
|------|------|
| 发帖审核 | 同步写 DB → goroutine pool 异步审核 |
| 积分结算 | 同步写 Redis → goroutine 异步落 DB |
| 图片处理 | 上传 MinIO → goroutine 异步缩略图 |
| 通知推送 | goroutine 异步 WebSocket 推送 |

### 7.3 降级策略

```
LLM API 不可用:
  ├── 桌宠 → 预设对话回复（降级文案库）
  ├── 科普生成 → 返回静态故事模板
  └── 内容审核 → 降级为本地敏感词过滤

Vision 不可用:
  └── 垃圾分类 → 提示用户手动选择分类
```

---

## 8. 安全架构

### 8.1 认证流程

```
Client                         Gin 应用内
  │                            │
  │  POST /auth/login          │
  │  {username, password}      │
  │───────────────────────────►│ 验证密码 (bcrypt)
  │                            │ 生成 Access Token (15min)
  │                            │ 生成 Refresh Token (7d)
  │  {access, refresh}         │
  │◄───────────────────────────│
  │                            │
  │  (后续请求)                 │
  │  Authorization: Bearer xxx │
  │───────────────────────────►│ JWT 中间件验证
  │                            │ RBAC 中间件检查角色
```

### 8.2 RBAC 权限矩阵

| 接口 | 普通用户 | 监测员 | 管理员 |
|------|:---:|:---:|:---:|
| 注册/登录 | ✓ | ✓ | ✓ |
| 查看帖子/评论 | ✓ | ✓ | ✓ |
| 发帖/评论 | ✓ | ✓ | ✓ |
| 时空地图 | ✓ | ✓ | ✓ |
| 科普故事 | ✓ | ✓ | ✓ |
| 桌宠对话 | ✓ | ✓ | ✓ |
| 趣味问答 | ✓ | ✓ | ✓ |
| 查看排行榜 | ✓ | ✓ | ✓ |
| 上传垃圾图片 | ✗ | ✓ | ✓ |
| 查看监测数据 | ✗ | ✓ | ✓ |
| 审核帖子 | ✗ | ✗ | ✓ |
| 管理用户 | ✗ | ✗ | ✓ |

### 8.3 其他安全措施

- **HTTPS**: TLS 1.3, HSTS
- **密码**: bcrypt (cost=12)
- **SQL 注入**: GORM 参数化查询
- **XSS**: CSP Header + 前端输出转义
- **CSRF**: SameSite Cookie + CSRF Token
- **文件上传**: 类型白名单 + 大小限制

---

## 9. 部署架构

### 9.1 Docker Compose (开发环境)

```yaml
# 极简：4 个容器
services:
  postgres:    # PostgreSQL 16 + PostGIS (5432)
  redis:       # Redis 7 (6379)
  minio:       # MinIO (9000/9001)
  app:         # Go 单体 + Vue 静态文件 (8080)
```

### 9.2 生产部署

```
                    ┌──────────────────┐
                    │   Nginx / Caddy  │
                    │   (HTTPS + LB)   │
                    └────────┬─────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
         ┌────▼────┐   ┌────▼────┐   ┌────▼────┐
         │  App-1  │   │  App-2  │   │  App-3  │
         │  :8080  │   │  :8080  │   │  :8080  │
         └────┬────┘   └────┬────┘   └────┬────┘
              │              │              │
              └──────────────┼──────────────┘
                             │
                    ┌────────▼────────┐
                    │  PostgreSQL     │
                    │  + Redis + MinIO│
                    └─────────────────┘

水平扩展：Nginx 反向代理 → 多实例 Go 单体（无状态，共享 PostgreSQL + Redis + MinIO）
```

### 9.3 CI/CD 流水线

```
Git Push (main)
    │
    ▼
GitHub Actions
    ├── Lint (golangci-lint)
    ├── Test (go test)
    ├── Build (go build → 单二进制)
    ├── Docker Build & Push
    └── Deploy (rsync / docker pull + restart)
```

---

## 10. 监控与可观测性

### 10.1 监控栈

| 层 | 工具 | 指标 |
|----|------|------|
| 基础设施 | Prometheus + Node Exporter | CPU/Mem/Disk/Net |
| 应用 | Prometheus + gin-prometheus | QPS/Latency/Error Rate |
| 业务埋点 | Prometheus Counter/Gauge | 注册数/发帖数/答题数 |
| 日志 | Zap → stdout | Docker logs / Loki |

### 10.2 告警规则

| 告警 | 条件 | 级别 |
|------|------|------|
| 应用宕机 | up == 0 | Critical |
| P99 > 1s | histogram_quantile(0.99) > 1 | Warning |
| 错误率 > 5% | rate(errors[5m]) > 0.05 | Critical |
| DB 连接池满 | pool_wait > 3s | Warning |
| Redis 内存 > 80% | used_memory > 80% | Warning |

---

## 附录 A: 项目结构（单体）

```
grand-canal-guardian/
├── cmd/
│   └── server/
│       └── main.go              # ★ 唯一入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置加载 (Viper)
│   ├── middleware/
│   │   ├── auth.go              # JWT 鉴权中间件
│   │   ├── rbac.go              # RBAC 权限中间件
│   │   ├── cors.go              # CORS
│   │   ├── ratelimit.go         # 令牌桶限流
│   │   ├── recovery.go          # Panic 恢复
│   │   └── request_id.go        # 请求 ID 追踪
│   ├── handler/
│   │   ├── auth_handler.go      # 认证相关 handler
│   │   ├── user_handler.go      # 用户 handler
│   │   ├── post_handler.go      # 帖子 handler
│   │   ├── quiz_handler.go      # 问答 handler
│   │   ├── map_handler.go       # 地图 handler
│   │   ├── llm_handler.go       # AI 对话 handler
│   │   ├── admin_handler.go     # 管理后台 handler
│   │   └── ws_handler.go        # WebSocket handler
│   ├── service/
│   │   ├── user_service.go      # 用户业务逻辑
│   │   ├── post_service.go      # 帖子业务逻辑
│   │   ├── quiz_service.go      # 问答业务逻辑
│   │   ├── map_service.go       # 地图业务逻辑
│   │   ├── llm_service.go       # LLM 调用封装
│   │   └── vision_service.go    # 视觉识别封装
│   ├── repository/
│   │   ├── user_repo.go         # 用户数据访问
│   │   ├── post_repo.go         # 帖子数据访问
│   │   ├── quiz_repo.go         # 问答数据访问
│   │   └── map_repo.go          # 地图数据访问
│   ├── model/
│   │   ├── user.go              # 用户模型
│   │   ├── post.go              # 帖子模型
│   │   ├── question.go          # 题目模型
│   │   └── map.go               # 地图模型
│   ├── worker/
│   │   ├── pool.go              # Goroutine 池
│   │   ├── review_task.go       # 审核任务
│   │   └── image_task.go        # 图片处理任务
│   └── router/
│       └── router.go            # 路由注册
├── pkg/
│   ├── apperror/                # 统一错误码
│   ├── response/                # 统一响应格式
│   ├── log/                     # Zap 日志封装
│   └── validator/               # 参数校验
├── web/                         # Vue 3 前端
├── migrations/                  # SQL 迁移文件
├── deploy/
│   ├── Dockerfile
│   └── docker-compose.yml
├── docs/                        # 文档
├── go.mod                       # ★ 唯一 go.mod
├── go.sum
├── Makefile
└── README.md
```

## 附录 B: 数据流示例 — 用户发帖 (单体版)

```
1. 用户提交帖子 → Gin Handler (JWT 中间件验证身份)
2. Handler → PostService.Create():
   a. 参数校验
   b. PostRepo.Insert() → PostgreSQL (status=pending)
   c. 失效 Redis 缓存: DEL posts:list:*
   d. Worker.Submit(reviewTask) → goroutine pool 异步审核
3. 审核 Worker:
   a. LLMService.Moderate() → 调用 LLM API 内容审核
   b. VisionService.Classify() → 图片审核 (如有)
   c. PostRepo.UpdateStatus() → 更新帖子状态
   d. WebSocket 推送审核结果给用户
4. 全部在同一进程内完成，零网络开销
```

---

> **下一份文档**: `02-openapi-spec.yaml` — 完整 OpenAPI 3.0 规范
