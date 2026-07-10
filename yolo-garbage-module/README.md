# YOLO 垃圾分类识别模块

大运河生态与文化保护平台 (Grand Canal Guardian) 的独立垃圾分类识别服务。

基于 **YOLOv8** 实现，与 Go 主项目完全分离，可单独运行。

## 快速开始

### 环境要求

- Python 3.10+
- pip

### 安装与启动

```bash
# 1. 进入目录
cd yolo-garbage-module

# 2. 安装依赖
pip install -r requirements.txt

# 3. 启动服务
python app.py
```

服务启动后：
- **Web UI**: http://localhost:8081
- **API 文档**: http://localhost:8081/docs
- **健康检查**: http://localhost:8081/api/v1/vision/health

### 训练自定义模型

1. 下载 TACO 数据集: `git clone git@github.com:pedropro/TACO.git`
2. 生成 YOLO 格式标注: 运行 `convert_taco_to_yolo.py`
3. 训练: `yolo train model=yolov8n.pt data=TACO/dataset.yaml epochs=100`
4. 训练完成后将 `runs/detect/train/weights/best.pt` 复制到 `models/garbage-yolov8n.pt`

## API

### 垃圾分类识别

```http
POST /api/v1/vision/classify
Content-Type: multipart/form-data
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| image | File | 是 | 图片文件 (jpg/png/webp, ≤10MB) |

**成功响应 (200)：**

```json
{
  "image_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "detections": [
    {
      "class_name": "塑料瓶",
      "category": "可回收物",
      "confidence": 0.96,
      "bbox": { "x": 120, "y": 80, "w": 210, "h": 350 }
    }
  ],
  "processing_time_ms": 120,
  "advice": "塑料瓶属于可回收物，请清洗后投入蓝色可回收物垃圾桶。"
}
```

### 健康检查

```http
GET /api/v1/vision/health
```

**响应：**

```json
{
  "status": "healthy",
  "model_loaded": true,
  "gpu_available": false,
  "using_fallback": true
}
```

## 垃圾分类体系

基于 TACO (Trash Annotations in Context) 数据集，60 个细分类，映射到中国垃圾分类 4 大类：

| 大类 | 示例类别 | 垃圾桶颜色 |
|------|----------|-----------|
| 可回收物 | 塑料瓶、玻璃瓶、易拉罐、纸箱、报纸 | 🔵 蓝色 |
| 有害垃圾 | 电池、气雾罐、铝塑药板 | 🔴 红色 |
| 厨余垃圾 | 食物残渣 | 🟢 绿色 |
| 其他垃圾 | 烟头、塑料袋/膜、泡沫、一次性餐具 | ⚫ 灰色 |

> 完整 60 类见 model.py CLASS_NAMES

## 技术方案

- **模型**: YOLOv8n + TACO 60 类数据集微调训练
- **框架**: FastAPI + uvicorn
- **前端**: 纯 HTML/CSS/JS，零构建步骤
- **端口**: 8081（避免与 Go 主服务 8080 冲突）

## 与主项目集成

模块独立运行。需要与 Go 主项目联动时，在 Nginx 中添加反向代理：

```nginx
location /api/v1/vision/ {
    proxy_pass http://localhost:8081/api/v1/vision/;
    client_max_body_size 10M;
}
```

Vue 前端无需任何修改。

## 项目结构

```
yolo-garbage-module/
├── app.py              # FastAPI 入口
├── model.py            # YOLOv8 模型封装 + 分类映射
├── requirements.txt    # Python 依赖
├── static/
│   └── index.html      # Web UI
├── models/             # 自定义模型存放目录
└── README.md
```
