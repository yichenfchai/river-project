"""
YOLOv8 垃圾分类检测模型封装。

负责模型加载、推理、结果解析与分类建议生成。
模型：
- TACO 60 类自定义训练模型，直接映射到中国垃圾分类 4 大类
- 训练命令: yolo train model=yolov8n.pt data=TACO/dataset.yaml epochs=100
"""

import time
import uuid
from io import BytesIO
from pathlib import Path
from typing import Optional

import numpy as np
from PIL import Image
from ultralytics import YOLO

# ============================================================
# 分类定义（TACO 60 类 → 4 个中国垃圾分类大类）
# ============================================================

# TACO 60 类垃圾分类标签 (按 TACO category_id 顺序)
CLASS_NAMES = [
    "Aluminium foil",           # 0
    "Battery",                  # 1
    "Aluminium blister pack",   # 2
    "Carded blister pack",      # 3
    "Other plastic bottle",     # 4
    "Clear plastic bottle",     # 5
    "Glass bottle",             # 6
    "Plastic bottle cap",       # 7
    "Metal bottle cap",         # 8
    "Broken glass",             # 9
    "Food Can",                 # 10
    "Aerosol",                  # 11
    "Drink can",                # 12
    "Toilet tube",              # 13
    "Other carton",             # 14
    "Egg carton",               # 15
    "Drink carton",             # 16
    "Corrugated carton",        # 17
    "Meal carton",              # 18
    "Pizza box",                # 19
    "Paper cup",                # 20
    "Disposable plastic cup",   # 21
    "Foam cup",                 # 22
    "Glass cup",                # 23
    "Other plastic cup",        # 24
    "Food waste",               # 25
    "Glass jar",                # 26
    "Plastic lid",              # 27
    "Metal lid",                # 28
    "Other plastic",            # 29
    "Magazine paper",           # 30
    "Tissues",                  # 31
    "Wrapping paper",           # 32
    "Normal paper",             # 33
    "Paper bag",                # 34
    "Plastified paper bag",     # 35
    "Plastic film",             # 36
    "Six pack rings",           # 37
    "Garbage bag",              # 38
    "Other plastic wrapper",    # 39
    "Single-use carrier bag",   # 40
    "Polypropylene bag",        # 41
    "Crisp packet",             # 42
    "Spread tub",               # 43
    "Tupperware",               # 44
    "Disposable food container",# 45
    "Foam food container",      # 46
    "Other plastic container",  # 47
    "Plastic gloves",           # 48
    "Plastic utensils",         # 49
    "Pop tab",                  # 50
    "Rope & strings",           # 51
    "Scrap metal",              # 52
    "Shoe",                     # 53
    "Squeezable tube",          # 54
    "Plastic straw",            # 55
    "Paper straw",              # 56
    "Styrofoam piece",          # 57
    "Unlabeled litter",         # 58
    "Cigarette",                # 59
]

# TACO 类别 → 中国垃圾分类 4 大类
CLASS_TO_CATEGORY: dict[str, str] = {
    # === 可回收物 (Blue bin) ===
    "Clear plastic bottle":     "可回收物",
    "Other plastic bottle":     "可回收物",
    "Glass bottle":             "可回收物",
    "Glass cup":                "可回收物",
    "Glass jar":                "可回收物",
    "Drink can":                "可回收物",
    "Food Can":                 "可回收物",
    "Metal bottle cap":         "可回收物",
    "Metal lid":                "可回收物",
    "Pop tab":                  "可回收物",
    "Scrap metal":              "可回收物",
    "Aluminium foil":           "可回收物",
    "Normal paper":             "可回收物",
    "Magazine paper":           "可回收物",
    "Wrapping paper":           "可回收物",
    "Paper bag":                "可回收物",
    "Corrugated carton":        "可回收物",
    "Other carton":             "可回收物",
    "Drink carton":             "可回收物",
    "Meal carton":              "可回收物",
    "Pizza box":                "可回收物",
    "Egg carton":               "可回收物",
    "Tupperware":               "可回收物",
    "Other plastic container":  "可回收物",

    # === 有害垃圾 (Red bin) ===
    "Battery":                  "有害垃圾",
    "Aerosol":                  "有害垃圾",
    "Aluminium blister pack":   "有害垃圾",

    # === 厨余垃圾 (Green bin) ===
    "Food waste":               "厨余垃圾",

    # === 其他垃圾 (Grey bin) ===
    "Cigarette":                "其他垃圾",
    "Styrofoam piece":          "其他垃圾",
    "Plastic film":             "其他垃圾",
    "Plastic bottle cap":       "其他垃圾",
    "Plastic lid":              "其他垃圾",
    "Broken glass":             "其他垃圾",
    "Carded blister pack":      "其他垃圾",
    "Toilet tube":              "其他垃圾",
    "Tissues":                  "其他垃圾",
    "Six pack rings":           "其他垃圾",
    "Garbage bag":              "其他垃圾",
    "Other plastic wrapper":    "其他垃圾",
    "Single-use carrier bag":   "其他垃圾",
    "Polypropylene bag":        "其他垃圾",
    "Crisp packet":             "其他垃圾",
    "Spread tub":               "其他垃圾",
    "Disposable food container": "其他垃圾",
    "Foam food container":      "其他垃圾",
    "Disposable plastic cup":   "其他垃圾",
    "Foam cup":                 "其他垃圾",
    "Other plastic cup":        "其他垃圾",
    "Paper cup":                "其他垃圾",
    "Plastic straw":            "其他垃圾",
    "Paper straw":              "其他垃圾",
    "Plastic gloves":           "其他垃圾",
    "Plastic utensils":         "其他垃圾",
    "Rope & strings":           "其他垃圾",
    "Shoe":                     "其他垃圾",
    "Squeezable tube":          "其他垃圾",
    "Unlabeled litter":         "其他垃圾",
    "Other plastic":            "其他垃圾",
    "Plastified paper bag":     "其他垃圾",
}

# 垃圾分类投放建议（按 4 大类统一，不再逐类编写）
CLASS_ADVICE: dict[str, str] = {}

def _get_advice(category: str) -> str:
    """根据大类返回投放建议。"""
    advices = {
        "可回收物": (
            "可回收物是指适宜回收和资源化利用的废弃物。"
            "请清洗干净、沥干水分后投入蓝色可回收物垃圾桶。"
            "回收利用可有效减少资源浪费和碳排放。"
        ),
        "有害垃圾": (
            "有害垃圾是指对人体健康或自然环境造成直接或潜在危害的废弃物。"
            "请投入红色有害垃圾桶，切勿混入其他垃圾。"
            "一粒纽扣电池可污染 60 万升水，请务必正确投放。"
        ),
        "厨余垃圾": (
            "厨余垃圾是指家庭产生的有机易腐垃圾。"
            "请沥干水分后投入绿色厨余垃圾桶。"
            "厨余垃圾经生物处理后可变废为宝，制成有机肥料。"
        ),
        "其他垃圾": (
            "其他垃圾是指除可回收物、有害垃圾、厨余垃圾之外的废弃物。"
            "请投入灰色其他垃圾桶。"
            "减少使用一次性物品，从源头减少垃圾产生。"
        ),
    }
    return advices.get(category, "请根据当地垃圾分类标准进行投放。")

# ============================================================
# （使用 TACO 自定义模型，无需 COCO 映射）
MODEL_DIR = Path(__file__).parent / "models"
MODEL_PATH = MODEL_DIR / "garbage-yolov8n.pt"  # TACO 60 类自定义训练模型
USE_PRETRAINED_FALLBACK = True


class GarbageDetector:
    """YOLO 垃圾分类检测器，惰性加载模型。"""

    def __init__(self) -> None:
        self._model: Optional[YOLO] = None
        self._loaded: bool = False
        self._using_fallback: bool = False

    @property
    def model(self) -> YOLO:
        """获取模型实例（首次访问时自动加载）。"""
        if self._model is None:
            self._load_model()
        return self._model  # type: ignore[return-value]

    def _load_model(self) -> None:
        """加载 YOLO 模型。"""
        MODEL_DIR.mkdir(parents=True, exist_ok=True)

        if MODEL_PATH.exists():
            self._model = YOLO(str(MODEL_PATH))
            self._using_fallback = False  # TACO 自定义模型，直接用 CLASS_NAMES
        else:
            raise FileNotFoundError(
                f"模型文件未找到: {MODEL_PATH}\n"
                "请将训练好的 YOLOv8 模型放置到该路径，"
                "或在 model.py 中设置 USE_PRETRAINED_FALLBACK = True"
            )
        self._loaded = True

    @property
    def is_loaded(self) -> bool:
        return self._loaded and self._model is not None

    @property
    def using_fallback(self) -> bool:
        return self._using_fallback

    def detect(self, image_bytes: bytes) -> dict:
        """
        对原始图片字节执行垃圾检测推理。

        Args:
            image_bytes: 图片文件的原始字节（jpg/png/webp）

        Returns:
            dict 符合 OpenAPI VisionClassifyResponse 规范:
            {
                "image_id": "uuid",
                "detections": [{class_name, category, confidence, bbox: {x,y,w,h}}],
                "processing_time_ms": 120,
                "advice": "分类建议文案"
            }
        """
        image_id = str(uuid.uuid4())
        start = time.perf_counter()

        # 将 bytes 转为 PIL Image，再转为 numpy array（ultralytics 接受的格式）
        pil_image = Image.open(BytesIO(image_bytes))
        pil_image = pil_image.convert("RGB")  # 确保是 RGB 三通道
        image_array = np.array(pil_image)

        results = self.model(image_array, verbose=False)
        elapsed_ms = int((time.perf_counter() - start) * 1000)

        detections: list[dict] = []
        detected_classes: set[str] = set()

        result = results[0]
        boxes = result.boxes

        if boxes is not None and len(boxes) > 0:
            for i in range(len(boxes)):
                cls_id = int(boxes.cls[i].item())
                conf = float(boxes.conf[i].item())
                x1, y1, x2, y2 = boxes.xyxy[i].tolist()

                # 类别映射
                if self._using_fallback:
                    class_name = COCO_GARBAGE_MAP.get(cls_id)
                    if class_name is None:
                        continue  # 跳过非垃圾 COCO 类别
                else:
                    class_name = (
                        CLASS_NAMES[cls_id]
                        if cls_id < len(CLASS_NAMES)
                        else "其他垃圾"
                    )

                category = CLASS_TO_CATEGORY.get(class_name, "其他垃圾")
                detected_classes.add(class_name)

                detections.append({
                    "class_name": class_name,
                    "category": category,
                    "confidence": round(conf, 4),
                    "bbox": {
                        "x": max(0, int(x1)),
                        "y": max(0, int(y1)),
                        "w": max(0, int(x2 - x1)),
                        "h": max(0, int(y2 - y1)),
                    },
                })

        # 按置信度降序排列
        detections.sort(key=lambda d: d["confidence"], reverse=True)

        # 生成建议文案
        advice = self._build_advice(detections)

        return {
            "image_id": image_id,
            "detections": detections,
            "processing_time_ms": elapsed_ms,
            "advice": advice,
        }

    def _build_advice(self, detections: list[dict]) -> str:
        """根据检测结果生成分类投放建议。"""
        if not detections:
            return (
                "未检测到明确垃圾类别。请确保图片清晰、光线充足，"
                "或切换到手动分类模式。"
            )

        # 主建议：置信度最高的检测项
        primary = detections[0]
        primary_category = primary.get("category", "其他垃圾")
        advice = _get_advice(primary_category)

        # 多类别检测时追加提示
        unique_classes = {d["class_name"] for d in detections}
        unique_categories = {d["category"] for d in detections}

        if len(unique_classes) > 1:
            advice += (
                f" 本次共检测到 {len(unique_classes)} 种物品，"
                f"涉及 {len(unique_categories)} 个分类，请按类别分别投放。"
            )

        return advice

    def health(self) -> dict:
        """健康检查，返回服务状态。"""
        gpu_available = self._check_gpu()
        return {
            "status": "healthy",
            "model_loaded": self.is_loaded,
            "gpu_available": gpu_available,
            "using_fallback": self._using_fallback,
        }

    @staticmethod
    def _check_gpu() -> bool:
        """检测是否有可用的 CUDA GPU。"""
        try:
            import torch
            return torch.cuda.is_available()
        except ImportError:
            return False


# ============================================================
# 模块级单例（FastAPI app 导入时创建，首次推理时加载模型）
# ============================================================
detector = GarbageDetector()
