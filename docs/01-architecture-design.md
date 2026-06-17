# 大运河生态与文化保护平台 — 架构设计文档

> **实现阶段**: Phase 1 (核心服务) 已完成 | Phase 2 (Map/LLM/Vision) 规划中

> **版本**: v1.0 | **日期**: 2026-06-16 | **作者**: Architecture Team
>
> **项目代号**: Grand Canal Guardian (GCG)

---

## 目录

1. [系统概述](#1-系统概述)
2. [微服务架构拓扑](#2-微服务架构拓扑)
3. [技术选型](#3-技术选型)
4. [服务详细设计](#4-服务详细设计)
5. [数据层架构](#5-数据层架构)
6. [缓存策略（Redis 多级缓存）](#6-缓存策略)
7. [高并发设计](#7-高并发设计)
8. [安全架构](#8-安全架构)
9. [部署架构](#9-部署架构)
10. [监控与可观测性](#10-监控与可观测性)

---

## 1. 系统概述

### 1.1 项目定位

面向公众、生态监测员、管理员的大运河生态与文化保护综合平台，融合 GIS 时空地图、AI 科普互动、社区分享、AI 桌宠、趣味问答等功能。

### 1.2 核心功能矩阵

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

### 1.3 非功能需求

- **并发**: 峰值 QPS 5000+
- **可用性**: 99.9% SLA
- **响应时间**: P99 < 200ms（读），P99 < 500ms（写）
- **数据安全**: HTTPS + AES-256 敏感字段加密 + RBAC
- **扩展性**: 无状态服务水平扩展

---

## 2. 微服务架构拓扑

```
                            ┌─────────────────────────────────────┐
                            │           Nginx / Kong              │
                            │     (Reverse Proxy + Rate Limit)    │
                            └──────┬──────────────────┬───────────┘
                                   │                  │
                    ┌──────────────▼──────┐   ┌───────▼────────────┐
                    │   API Gateway      │   │   WebSocket GW    │
                    │   (Go - Gin)       │   │   (Go - Gorilla)  │
                    │   鉴权/路由/限流     │   │   实时推送         │
                    └──┬───┬───┬───┬───┬─┘   └───────────────────┘
                       │   │   │   │   │
        ┌──────────────┼───┼───┼───┼───┼──────────────┐
        │              │   │   │   │   │              │
        ▼              ▼   ▼   ▼   ▼   ▼              ▼
┌───────────┐ ┌───────────┐ ┌───────────┐ ┌────────────────┐
│  User     │ │  Content  │ │   Map     │ │   Quiz/Game    │
│  Service  │ │  Service  │ │  Service  │ │   Service      │
│  (Go)     │ │  (Go)     │ │  (Go)     │ │   (Go)         │
│  认证/RBAC│ │ 帖子/评论  │ │ GIS/时空   │ │  问答/积分/排名 │
└─────┬─────┘ └─────┬─────┘ └─────┬─────┘ └──────┬─────────┘
      │             │             │              │
      └─────────────┼─────────────┼──────────────┘
                    │             │
         ┌──────────▼──┐   ┌─────▼────────┐
         │  RabbitMQ   │   │  gRPC Bus    │
         │  (异步消息)  │   │  (同步调用)   │
         └──────┬──────┘   └──────┬───────┘
                │                 │
    ┌───────────▼─────┐  ┌───────▼──────────────┐
    │  LLM Service    │  │  Vision Service      │
    │  (Python/FastAPI)│  │  (Python/FastAPI)    │
    │  通义千问/DeepSeek│  │  YOLOv8 + OpenCV     │
    │  桌宠对话/科普生成│  │  垃圾分类识别         │
    └─────────────────┘  └──────────────────────┘
```

### 2.1 Phase 1 vs Phase 2

| Phase | 服务 | 状态 |
|-------|------|------|
| Phase 1 | main-service (Go 单体) | ✅ 已实现 — 合并 Auth/Users/Posts/Quiz/Admin |
| Phase 2 | Vision Service (Python) | 🚧 规划中 — YOLO 垃圾分类 |
| Phase 2 | LLM Service (Python) | 🚧 规划中 — AI 对话/科普 |
| Phase 2 | Map Service (Go) | 🚧 规划中 — GIS 时空地图 |

### 2.2 服务间通信

| 场景 | 方式 | 说明 |
|------|------|------|
| 同步调用 | gRPC (protobuf) | 低延迟 RPC，Go 服务间调用 |
| Go → Python | HTTP/2 + JSON | LLM/Vision 走 REST 接口 |
| 异步事件 | RabbitMQ | 帖子审核、积分变更通知、监测上报 |
| 实时推送 | WebSocket | 桌宠对话流、问答实时排名 |
| 服务发现 | Consul | 健康检查 + 动态路由 |

---

## 3. 技术选型

### 3.1 后端

| 组件 | 选型 | 理由 |
|------|------|------|
| API 网关 | Go + Gin | 高性能 HTTP 框架，中间件生态丰富 |
| 业务服务 | Go 1.22+ | 高并发（goroutine），编译型部署简单 |
| AI 服务 | Python 3.11 + FastAPI | ML 生态，async 支持好 |
| RPC 框架 | gRPC + protobuf | 强类型契约，跨语言支持 |
| ORM | GORM (Go) / SQLAlchemy (Py) | 成熟稳定 |
| 消息队列 | RabbitMQ | Phase 2 启用 (异步审核/通知) — Phase 1 同步处理 |

### 3.2 前端

| 组件 | 选型 |
|------|------|
| 框架 | Vue 3 + TypeScript |
| 地图 | MapLibre GL JS (OSM 瓦片) |
| 状态管理 | Pinia |
| UI 库 | Element Plus |
| 桌宠 | 自研 Canvas/WebGL 渲染 |
| 构建 | Vite |

### 3.3 数据层

| 组件 | 选型 | 用途 |
|------|------|------|
| 主数据库 | PostgreSQL 16 (主) | 业务数据（用户、帖子、积分） |
| 从数据库 | PostgreSQL 16 (只读副本) | 读流量卸载 |
| 空间数据 | PostGIS 扩展 | 大运河 GIS 数据 |
| 缓存 | Redis 7 (Cluster) | 多级缓存 |
| 搜索引擎 | Elasticsearch | 帖子全文搜索 (Phase 2 — Phase 1 使用 PostgreSQL LIKE) |
| 对象存储 | MinIO (S3 兼容) | 图片/文件 |
| 时序数据 | InfluxDB (可选) | 生态监测数据 |

### 3.4 AI 模型

| 模型 | 框架 | 部署 |
|------|------|------|
| LLM (对话/科普) | 通义千问-Plus / DeepSeek-V3 | API 调用 |
| YOLOv8 (垃圾分类) | ultralytics + ONNX Runtime | 自部署 GPU (T4/A10) |
| Embedding | text2vec-large-chinese | 自部署 |

### 3.5 基础设施

| 组件 | 选型 |
|------|------|
| 容器编排 | Docker Compose | Phase 1 仅 Docker Compose (K8s → Phase 2) |
| CI/CD | GitHub Actions |
| 监控 | Prometheus + Grafana |
| 日志 | Loki + Promtail |
| 链路追踪 | Jaeger (OpenTelemetry) |

---

## 4. 服务详细设计

### 4.1 User Service（用户服务）

- **职责**: 注册、登录、RBAC 权限、个人资料
- **端口**: 8001 (HTTP) / 9001 (gRPC)
- **数据库**: PostgreSQL `user_db`

```
POST   /api/v1/auth/register        # 注册
POST   /api/v1/auth/login           # 登录 → JWT
POST   /api/v1/auth/refresh         # 刷新 Token
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

### 4.2 Content Service（内容服务）

- **职责**: 分享广场帖子、评论、点赞、收藏
- **端口**: 8002 (HTTP) / 9002 (gRPC)
- **数据库**: PostgreSQL `content_db`
- **搜索**: Elasticsearch

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

**审核流程**: 发帖 → RabbitMQ `content.review` → 敏感词过滤 + LLM 审核 → 通过/驳回

### 4.3 Map Service（地图服务）

- **职责**: 大运河时空 GIS 数据、POI、历史图层
- **端口**: 8003 (HTTP) / 9003 (gRPC)
- **数据库**: PostgreSQL + PostGIS `map_db`
- **瓦片服务**: Martin (Rust) 动态矢量瓦片

```
GET    /api/v1/map/tiles/{z}/{x}/{y}.pbf    # 矢量瓦片
GET    /api/v1/map/layers                   # 可用图层列表
GET    /api/v1/map/pois?lat=&lng=&radius=   # 周边 POI
GET    /api/v1/map/timeline?year=           # 历年运河数据
GET    /api/v1/map/routes/:id               # 运河路线详情
```

**图层设计**:
- 隋唐运河 (581-907)
- 元明清运河 (1271-1911)
- 现代南水北调东线
- 生态监测站点
- 文化遗产 POI

### 4.4 Quiz/Game Service（问答服务）

- **职责**: 趣味问答、积分、排行榜、段位
- **端口**: 8004 (HTTP) / 9004 (gRPC)
- **数据库**: PostgreSQL `quiz_db`

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

### 4.5 LLM Service（大模型服务 - Python）

- **职责**: 运河小精灵对话、科普故事生成、内容审核
- **端口**: 8101
- **框架**: FastAPI + asyncio
- **LLM**: 通义千问 / DeepSeek（可配置切换）

```
POST   /api/v1/llm/chat              # 桌宠对话 (SSE 流式)
POST   /api/v1/llm/story/generate    # 生成科普故事
POST   /api/v1/llm/moderate          # 内容审核
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
data: {"token": "你", "index": 0}

event: token
data: {"token": "好", "index": 1}

event: done
data: {"total_tokens": 45}
```

### 4.6 Vision Service（视觉服务 - Python）

- **职责**: YOLO 垃圾分类识别、图片审核
- **端口**: 8102
- **框架**: FastAPI + ultralytics + ONNX Runtime
- **GPU**: NVIDIA T4 / A10 (CUDA 12)

```
POST   /api/v1/vision/classify       # 垃圾图片分类
POST   /api/v1/vision/detect         # 通用目标检测
GET    /api/v1/vision/health         # 健康检查
```

**分类类别** (8 类):
```
可回收物, 有害垃圾, 厨余垃圾, 其他垃圾,
塑料袋, 玻璃瓶, 废电池, 废弃药品
```

**响应格式**:
```json
{
  "image_id": "uuid",
  "detections": [
    {
      "class": "塑料瓶",
      "category": "可回收物",
      "confidence": 0.96,
      "bbox": [x, y, w, h]
    }
  ],
  "processing_time_ms": 120
}
```

### 4.7 Notification Service（通知服务）

- **职责**: 实时消息推送、系统通知
- **端口**: 8005 (HTTP) / WebSocket 8005
- **消息通道**: RabbitMQ `notification.*`

---

## 5. 数据层架构

### 5.1 数据库主从架构

```
          ┌─────────────────┐
          │  HAProxy/PgBouncer│  ← 连接池 + 读写分离
          └────┬────────┬────┘
               │        │
       ┌───────▼──┐ ┌──▼──────────┐
       │ Master   │ │ Slave-1     │ (同步复制)
       │ (Read/   │─┤ (Read Only) │
       │  Write)  │ │             │
       └──────────┘ └─────────────┘
                         │
                    ┌────▼──────────┐
                    │ Slave-2       │ (异步复制，备份)
                    │ (Read Only)   │
                    └───────────────┘
```

- **主库**: 写操作，强一致性
- **从库**: 读操作（各服务配置读写分离 DSN）
- **连接池**: PgBouncer，每个服务最大 25 连接
- **复制延迟监控**: < 100ms (P99)

### 5.2 数据分区策略

| 表 | 分区键 | 策略 |
|----|--------|------|
| posts | created_at | 按月 RANGE 分区 |
| quiz_records | created_at | 按月 RANGE 分区 |
| notifications | user_id | 按 HASH 分区 (16 片) |

---

## 6. 缓存策略（Redis 多级缓存）

### 6.1 三级缓存架构

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
│ L3: PostgreSQL│ ──────► 回写 L2 + L1
└─────────────┘
```

### 6.2 Redis 数据结构设计

| Key Pattern | 类型 | TTL | 说明 |
|-------------|------|-----|------|
| `user:{id}:profile` | Hash | 1h | 用户信息 |
| `post:{id}:detail` | String(JSON) | 30min | 帖子详情 |
| `leaderboard:daily` | Sorted Set | 24h | 日排行 |
| `leaderboard:weekly` | Sorted Set | 7d | 周排行 |
| `leaderboard:total` | Sorted Set | ∞ | 总排行 |
| `quiz:pool:{difficulty}` | List | ∞ | 题池 |
| `rate_limit:{ip}:{endpoint}` | String | 1min | 限流计数 |
| `session:{token}` | String | Token TTL | JWT 黑名单 |
| `lock:{resource}` | String (NX) | 30s | 分布式锁 |
| `map:tile:{z}:{x}:{y}` | Binary | 1h | 瓦片缓存 |

### 6.3 缓存更新策略

| 场景 | 策略 |
|------|------|
| 用户资料 | Write-Through（写 DB 同时更新 Redis） |
| 帖子详情 | Cache-Aside（读时缓存，写时失效） |
| 排行榜 | 定时刷新（每 5min 从 DB 计算 Top100） |
| 限流 | Redis INCR + EXPIRE（滑动窗口） |

### 6.4 缓存击穿防护

- **热点 Key 永不过期**: 逻辑过期 + 异步刷新
- **布隆过滤器**: 防缓存穿透（空值查询）
- **互斥锁**: `SETNX lock:rebuild:{key}` 防缓存雪崩（单线程重建）

---

## 7. 高并发设计

### 7.1 限流策略

```
┌──────────────────────────────────────────┐
│              Token Bucket                 │
│                                            │
│  API Gateway 层:                            │
│  - 全局: 5000 req/s (令牌桶)                │
│  - 单 IP:  100 req/s                       │
│  - LLM 接口: 50 req/s (成本控制)            │
│                                            │
│  存储: Redis `rate_limit:{key}`             │
└──────────────────────────────────────────┘
```

### 7.2 异步化

| 操作 | 模式 |
|------|------|
| 发帖审核 | 同步写 DB → 异步送审 RabbitMQ |
| 积分变更 | 同步写 Redis → 异步落 DB (MQ) |
| 监测数据上报 | 批量写入 (每 100 条/30s) |
| 图片处理 | 上传 → MinIO → MQ → 缩略图生成 |

### 7.3 连接池配置

| 组件 | 配置 |
|------|------|
| PgBouncer | pool_size=25, max_client_conn=500 |
| Redis Pool | MinIdle=10, MaxIdle=50, MaxActive=200 |
| HTTP Client | MaxIdleConns=100, Timeout=30s |

### 7.4 降级与熔断

```
LLM Service 不可用:
  ├── 桌宠 → 预设对话回复（降级文案库）
  ├── 科普生成 → 返回静态故事模板
  └── 内容审核 → 降级为关键词过滤

Vision Service 不可用:
  └── 垃圾分类 → 提示用户手动选择分类
```

熔断器配置 (go-sircuitbreaker):
- 错误率阈值: 50%
- 半开状态探测间隔: 30s
- 冷却时间: 60s

---

## 8. 安全架构

### 8.1 认证流程

```
Client                    Gateway                   User Service
  │                          │                          │
  │  POST /auth/login        │                          │
  │  {username, password}    │                          │
  │─────────────────────────►│─────────────────────────►│
  │                          │                          │ 验证密码 (bcrypt)
  │                          │                          │ 生成 Access Token (15min)
  │                          │                          │ 生成 Refresh Token (7d)
  │                          │  {access, refresh}       │
  │◄─────────────────────────│◄─────────────────────────│
  │                          │                          │
  │  (后续请求)               │                          │
  │  Authorization: Bearer   │                          │
  │─────────────────────────►│ 验证 JWT                  │
  │                          │ 检查角色权限               │
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
| 系统配置 | ✗ | ✗ | ✓ |

### 8.3 其他安全措施

- **HTTPS**: TLS 1.3, HSTS
- **密码**: bcrypt (cost=12)
- **SQL 注入**: GORM 参数化查询
- **XSS**: CSP Header + 前端输出转义
- **CSRF**: SameSite Cookie + CSRF Token
- **文件上传**: 类型白名单 + 病毒扫描 (ClamAV)

---

## 9. 部署架构

### 9.1 Docker Compose (开发环境)

```
docker-compose.yml
├── postgres-master    (5432)
├── postgres-slave     (5433)
├── redis              (6379)
├── rabbitmq           (5672/15672)
├── minio              (9000/9001)
├── elasticsearch      (9200)
├── consul             (8500)
├── api-gateway        (8080)
├── user-service       (8001)
├── content-service    (8002)
├── map-service        (8003)
├── quiz-service       (8004)
├── notification-svc   (8005)
├── llm-service        (8101)
├── vision-service     (8102)
└── frontend           (3000)
```

### 9.2 K8s 生产部署

```
┌─────────────────────────────────────────────┐
│                  Ingress                     │
│            (Nginx + cert-manager)            │
└──────────────────┬──────────────────────────┘
                   │
     ┌─────────────┼─────────────┐
     │             │             │
┌────▼────┐  ┌────▼────┐  ┌────▼────┐
│ Frontend│  │ API GW  │  │  WS GW  │
│  (3x)   │  │  (2x)   │  │  (2x)   │
└─────────┘  └────┬────┘  └─────────┘
                  │
    ┌─────────────┼─────────────┐
    │             │             │
┌───▼───┐ ┌──────▼───┐ ┌──────▼───┐
│User   │ │Content   │ │  Map     │ ...
│Svc 2x │ │Svc   2x  │ │  Svc 2x  │
└───────┘ └──────────┘ └──────────┘
```

### 9.3 CI/CD 流水线

```
Git Push (main)
    │
    ▼
GitHub Actions
    ├── Lint (golangci-lint, ruff)
    ├── Test (go test, pytest)
    ├── Build (Docker images)
    ├── Push to Registry (Harbor)
    └── Deploy
        ├── Dev  (自动)
        ├── Staging (审批)
        └── Prod (审批 + 金丝雀发布)
```

---

## 10. 监控与可观测性

### 10.1 监控栈

| 层 | 工具 | 指标 |
|----|------|------|
| 基础设施 | Prometheus + Node Exporter | CPU/Mem/Disk/Net |
| 应用 | Prometheus + 自定义 Metrics | QPS/Latency/Error Rate |
| 业务 | Prometheus + 埋点 | 注册数/发帖数/答题数 |
| 日志 | Loki + Promtail | 集中日志 + 全文搜索 |
| 链路 | Jaeger + OpenTelemetry | 请求全链路追踪 |

### 10.2 告警规则

| 告警 | 条件 | 级别 |
|------|------|------|
| 服务宕机 | up == 0 | Critical |
| P99 > 1s | histogram_quantile(0.99) > 1 | Warning |
| 错误率 > 5% | rate(errors[5m]) > 0.05 | Critical |
| DB 连接池满 | pool_wait > 3s | Warning |
| Redis 内存 > 80% | used_memory > 80% | Warning |

---

## 附录 A: 项目结构

```
grand-canal-guardian/
├── pkg/                      # ★ 公共库 — Go 工程化基础设施
│   ├── errors/               #   统一错误码 + AppError
│   ├── response/             #   统一 {code, message, data} 响应
│   ├── auth/                 #   ★ JWT 鉴权 (双Token+黑名单+Refresh旋转+RBAC位掩码)
│   ├── ratelimit/            #   ★ 令牌桶限流 (Redis Lua原子操作+三级IP/User/API)
│   ├── database/             #   ★ 连接池管理 (Prometheus指标+慢查询+GORM Callback)
│   ├── log/                  #   ★ 日志系统 (Zap+轮转+脱敏+链路关联)
│   ├── middleware/            #   通用中间件 (RequestID/CORS/Trace)
│   ├── validator/            #   go-playground + 中文翻译
│   ├── app/                  #   Gin 启动器 + 优雅关闭 + 健康检查
│   ├── transaction/          #   WithTx 事务装饰器
│   └── grpc/                 #   AppError ↔ gRPC Status 转换
├── services/
│   ├── api-gateway/          # Go - Gin
│   ├── user-service/         # Go (复用 pkg/)
│   ├── content-service/      # Go (复用 pkg/)
│   ├── map-service/          # Go (复用 pkg/)
│   ├── quiz-service/         # Go (复用 pkg/)
│   ├── notification-service/ # Go (复用 pkg/)
│   ├── llm-service/          # Python - FastAPI
│   └── vision-service/       # Python - FastAPI
├── web/                      # Vue 3 + TypeScript
├── proto/                    # protobuf 定义
├── deploy/
│   ├── docker-compose/
│   └── k8s/
├── docs/
│   ├── 01-architecture-design.md
│   ├── 02-openapi-spec.yaml
│   ├── 03-developer-guide.md
│   ├── 04-go-engineering-standards.md
│   └── 05-production-grade-infra.md
├── go.work                   # Go workspace (多模块)
└── scripts/
    ├── init-db.sql
    └── seed-data.sh
```

## 附录 B: 数据流示例 — 用户发帖

```
1. 用户提交帖子 → API Gateway (JWT 验证)
2. Gateway → Content Service (gRPC CreatePost)
3. Content Service:
   a. 写入 PostgreSQL 主库
   b. 失效 Redis 缓存: DEL posts:list:*
   c. 发送 RabbitMQ: content.review
4. 审核 Consumer:
   a. LLM Service 文本审核
   b. Vision Service 图片审核 (如有)
   c. 更新帖子状态 → RabbitMQ: notification.post_approved
5. Notification Service:
   a. WebSocket 推送审核结果
```

---

> **下一份文档**: `02-openapi-spec.yaml` — 完整 OpenAPI 3.0 规范
