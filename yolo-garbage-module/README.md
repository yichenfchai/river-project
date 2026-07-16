# YOLO 垃圾分类识别模块

大运河生态与文化保护平台 (Grand Canal Guardian) 的独立垃圾分类识别服务。

基于 **YOLOv8n** 微调，13 类细分类映射到中国国标 4 大类。与 Go 主项目解耦，可独立部署运行。

## 环境

| 组件 | 版本/型号 |
|------|----------|
| Python | 3.10+ |
| PyTorch | 2.6+ CUDA 12.4 |
| GPU | NVIDIA RTX 4060 8GB |
| Conda 环境 | `yolo-gpu`（`D:\anaconda3\envs\yolo-gpu`） |
| 推理设备 | CUDA (`DEVICE='cuda'`) |

## 快速启动

```bash
# 激活 conda 环境（首次使用前创建，见下方"环境配置"）
conda activate yolo-gpu

# 安装依赖
pip install -r requirements.txt

# 启动服务（GPU 推理）
python app.py
```

服务端口：

| 地址 | 用途 |
|------|------|
| `http://localhost:8081` | Web UI（上传图片测试） |
| `http://localhost:8081/docs` | Swagger API 文档 |
| `http://localhost:8081/api/v1/vision/health` | 健康检查 |

### 环境配置（首次）

```bash
# 创建 conda 环境
conda create -n yolo-gpu python=3.10 -y
conda activate yolo-gpu

# PyTorch CUDA 版（conda 镜像可能缺 nvidia 包，用 pip 走官方 CDN）
pip install torch torchvision --index-url https://download.pytorch.org/whl/cu124

# 项目依赖
pip install -r requirements.txt
```

## 模型

| 项目 | 说明 |
|------|------|
| 架构 | YOLOv8n（nano，轻量版） |
| 细分类 | 13 类（TACO 子类别） |
| 大类映射 | 可回收物 / 有害垃圾 / 厨余垃圾 / 其他垃圾 |
| 模型文件 | `models/garbage-yolov8n-13cls.pt` |

### 13 类 → 4 大类映射

```
可回收物: 可回收塑料, 可回收玻璃, 可回收金属, 可回收纸类
有害垃圾: 电池, 气雾罐, 铝箔药板
厨余垃圾: 厨余垃圾
其他垃圾: 一次性塑料品, 塑料包装膜袋, 泡沫发泡类, 复合杂项材料, 卫生日用品
```

## 训练

```bash
conda activate yolo-gpu

# 训练（必须在独立进程中运行，Jupyter/VSCode 交互窗口会崩）
python train.py

# 或后台运行（Windows）
start /min python train.py
```

训练参数（`train.py`）：

| 参数 | 值 | 说明 |
|------|-----|------|
| `data` | `TACO/dataset-13cls.yaml` | 数据集配置 |
| `epochs` | `100` | 训练轮数 |
| `device` | `0` | GPU 设备编号 |
| `workers` | `0` | Windows 下必须设为 0，否则 DataLoader 报错 |

输出目录: `runs/detect/train/weights/best.pt`

### 添加数据集

1. 下载 TACO: `git clone https://github.com/pedropro/TACO.git`
2. 运行 `convert_to_yolo.py` 生成 YOLO 格式标注
3. 编写 `dataset.yaml` 指定 train/val 路径和类别列表
4. **务必确保 `.gitignore` 已排除 `images/`、`labels/`、`*.pt` 等大文件后再 `git add`**

## API

### POST /api/v1/vision/classify

垃圾分类识别。

```
Content-Type: multipart/form-data
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `image` | File | 是 | jpg/png/webp, <=10MB |

成功响应 (200):

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
  "processing_time_ms": 45,
  "advice": "塑料瓶属于可回收物，请清洗后投入蓝色可回收物垃圾桶。"
}
```

### GET /api/v1/vision/health

```json
{
  "status": "healthy",
  "model_loaded": true,
  "gpu_available": true,
  "device": "cuda"
}
```

## 与主项目集成

模块独立运行，Go 主项目通过 Nginx 反向代理接入：

```nginx
location /api/v1/vision/ {
    proxy_pass http://localhost:8081/api/v1/vision/;
    client_max_body_size 10M;
}
```

Vue 前端无需修改。

## 项目结构

```
yolo-garbage-module/
├── app.py              # FastAPI 入口
├── model.py            # YOLOv8 封装 + 13类→4类映射 + 投放建议
├── train.py            # 训练脚本
├── requirements.txt    # Python 依赖
├── .gitignore          # 排除数据集图片/标注/模型权重
├── static/
│   └── index.html      # Web UI
├── models/
│   └── garbage-yolov8n-13cls.pt   # 训练好的权重
├── TACO/               # 数据集（git 忽略，本地存放）
│   ├── dataset-13cls.yaml
│   ├── convert_to_yolo.py
│   ├── images/
│   └── labels/
└── runs/               # 训练输出（git 忽略）
    └── detect/train/weights/best.pt
```

## Git 注意事项

- `*.pt`、`*.pth` 模型权重文件已在 `.gitignore` 中排除
- `TACO/images/`、`TACO/labels/` 数据集目录已排除
- **提交前务必 `git status` 确认没有误加二进制大文件**
- 如果误提交了数据集，用 `git filter-branch` 清理历史（本次事件已修复）