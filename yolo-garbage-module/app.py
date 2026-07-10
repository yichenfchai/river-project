"""
大运河生态保护平台 — YOLO 垃圾分类识别服务

独立可运行的 FastAPI 应用，提供：
- POST /api/v1/vision/classify  图片上传 + 垃圾分类
- GET  /api/v1/vision/health    服务健康检查
- GET  /                        独立 Web UI

启动方式:
    pip install -r requirements.txt
    python app.py
    浏览器打开 http://localhost:8081
"""

import logging
from pathlib import Path

from fastapi import FastAPI, File, UploadFile, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse, HTMLResponse

from model import detector

# ============================================================
# 日志
# ============================================================
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)
logger = logging.getLogger("garbage-vision")

# ============================================================
# FastAPI 应用
# ============================================================
app = FastAPI(
    title="Grand Canal Guardian — Garbage Classification API",
    version="1.0.0",
    description=(
        "基于 YOLOv8 + TACO 数据集的垃圾分类识别服务。\n\n"
        "**60 个细分类**: TACO (Trash Annotations in Context) 标准类别\n\n"
        "**4 个大类**: 可回收物、有害垃圾、厨余垃圾、其他垃圾"
    ),
    docs_url="/docs",
    redoc_url="/redoc",
)

# CORS — 开发阶段全开
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# ============================================================
# 上传校验
# ============================================================
ALLOWED_CONTENT_TYPES: set[str] = {
    "image/jpeg",
    "image/png",
    "image/webp",
}
MAX_FILE_SIZE: int = 10 * 1024 * 1024  # 10 MB

# ============================================================
# API 端点（必须在静态文件挂载之前注册）
# ============================================================

@app.post(
    "/api/v1/vision/classify",
    summary="垃圾分类识别",
    description="上传垃圾图片，返回 YOLOv8 检测结果（类别、置信度、检测框、分类建议）。",
    tags=["Vision"],
)
async def classify_image(image: UploadFile = File(..., description="图片文件 (jpg/png/webp, ≤10MB)")):
    # ---- 校验 Content-Type ----
    if image.content_type is None or image.content_type not in ALLOWED_CONTENT_TYPES:
        raise HTTPException(
            status_code=400,
            detail=f"不支持的图片格式: {image.content_type}。请上传 JPG、PNG 或 WebP 格式。",
        )

    # ---- 读取图片字节 ----
    try:
        contents = await image.read()
    except Exception:
        raise HTTPException(status_code=400, detail="读取图片失败，请重新上传。")

    # ---- 校验文件大小 ----
    if len(contents) > MAX_FILE_SIZE:
        raise HTTPException(
            status_code=413,
            detail=f"图片过大 ({len(contents) / 1024 / 1024:.1f}MB)，请上传小于 10MB 的图片。",
        )

    if len(contents) == 0:
        raise HTTPException(status_code=400, detail="图片为空，请选择有效的图片文件。")

    # ---- 执行检测 ----
    logger.info(
        "收到识别请求: filename=%s size=%dKB content_type=%s",
        image.filename,
        len(contents) // 1024,
        image.content_type,
    )

    try:
        result = detector.detect(contents)
    except Exception as exc:
        logger.exception("识别失败")
        raise HTTPException(
            status_code=500,
            detail=f"AI 识别服务异常: {str(exc)}。请稍后重试。",
        )

    detection_count = len(result["detections"])
    logger.info(
        "识别完成: image_id=%s detections=%d time=%dms",
        result["image_id"],
        detection_count,
        result["processing_time_ms"],
    )

    return JSONResponse(content=result)


@app.get(
    "/api/v1/vision/health",
    summary="健康检查",
    description="检查垃圾分类识别服务是否正常运行。",
    tags=["Vision"],
)
async def health_check():
    return JSONResponse(content=detector.health())


# ============================================================
# 静态页面（独立 Web UI）
# ============================================================
STATIC_DIR = Path(__file__).parent / "static"
INDEX_PATH = STATIC_DIR / "index.html"


@app.get("/", response_class=HTMLResponse, tags=["UI"])
async def serve_ui():
    """返回垃圾分类识别的独立 Web UI。"""
    if INDEX_PATH.exists():
        return HTMLResponse(content=INDEX_PATH.read_text(encoding="utf-8"))
    return HTMLResponse(content="<h1>UI not found</h1>", status_code=404)


# ============================================================
# 直接运行
# ============================================================
if __name__ == "__main__":
    import uvicorn

    logger.info("启动垃圾分类识别服务...")
    logger.info("API 文档: http://localhost:8081/docs")
    logger.info("Web   UI: http://localhost:8081")

    uvicorn.run(
        "app:app",
        host="0.0.0.0",
        port=8081,
        reload=True,
        log_level="info",
    )
