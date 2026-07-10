"""
TACO COCO JSON → YOLO 格式转换 + 图片下载 + 数据集划分

用法:
    python convert_to_yolo.py

输出结构:
    TACO/
    ├── images/train/    # 80% 训练图片
    ├── images/val/      # 20% 验证图片
    ├── labels/train/    # YOLO .txt 标注
    ├── labels/val/
    └── dataset.yaml     # YOLO 训练配置
"""

import json
import os
import sys
import random
import shutil
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor, as_completed

import requests
from PIL import Image
from io import BytesIO

# ============================================================
# 配置
# ============================================================
BASE_DIR = Path(__file__).parent
ANNOTATIONS_PATH = BASE_DIR / "data" / "annotations.json"
IMAGES_DIR = BASE_DIR / "images"
LABELS_DIR = BASE_DIR / "labels"
TRAIN_RATIO = 0.8
RANDOM_SEED = 42
MAX_WORKERS = 8  # 并行下载线程数

random.seed(RANDOM_SEED)

# ============================================================
# 1. 加载标注
# ============================================================
print("[1/5] 加载标注文件...")
with open(ANNOTATIONS_PATH, "r", encoding="utf-8") as f:
    coco = json.load(f)

images_info = {img["id"]: img for img in coco["images"]}
categories = {cat["id"]: cat for cat in coco["categories"]}

# 按 image_id 分组标注
anns_by_image: dict[int, list[dict]] = {}
for ann in coco["annotations"]:
    img_id = ann["image_id"]
    anns_by_image.setdefault(img_id, []).append(ann)

print(f"  图片: {len(images_info)}  标注: {len(coco['annotations'])}  类别: {len(categories)}")

# ============================================================
# 2. 下载图片
# ============================================================
print("\n[2/5] 下载图片...")

IMAGES_DIR.mkdir(parents=True, exist_ok=True)

def download_image(img: dict) -> tuple[int, bool]:
    """下载单张图片，返回 (image_id, success)"""
    img_id = img["id"]
    file_name = img["file_name"]
    save_path = IMAGES_DIR / file_name
    save_path.parent.mkdir(parents=True, exist_ok=True)

    if save_path.exists():
        return img_id, True  # 已存在

    for url_key in ["flickr_url", "flickr_640_url"]:
        url = img.get(url_key)
        if not url:
            continue
        try:
            resp = requests.get(url, timeout=30)
            if resp.status_code == 200:
                img_obj = Image.open(BytesIO(resp.content))
                img_obj.save(save_path)
                return img_id, True
        except Exception:
            continue

    return img_id, False

success_count = 0
fail_count = 0
image_list = list(images_info.values())

with ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
    futures = {executor.submit(download_image, img): img for img in image_list}
    for i, future in enumerate(as_completed(futures)):
        img_id, ok = future.result()
        if ok:
            success_count += 1
        else:
            fail_count += 1
        if (i + 1) % 100 == 0 or (i + 1) == len(image_list):
            print(f"  [{i+1}/{len(image_list)}] 成功:{success_count} 失败:{fail_count}")

print(f"  下载完成: {success_count}/{len(image_list)} 张成功, {fail_count} 张失败")

# ============================================================
# 3. 转换标注 → YOLO 格式
# ============================================================
print("\n[3/5] 转换标注为 YOLO 格式...")

LABELS_DIR.mkdir(parents=True, exist_ok=True)

converted = 0
skipped_no_img = 0
skipped_no_ann = 0

for img_id, img_info in images_info.items():
    file_name = img_info["file_name"]
    img_path = IMAGES_DIR / file_name

    # 检查图片是否下载成功
    if not img_path.exists():
        skipped_no_img += 1
        continue

    # 获取该图片的标注
    anns = anns_by_image.get(img_id, [])
    if not anns:
        skipped_no_ann += 1
        continue

    img_w = img_info["width"]
    img_h = img_info["height"]

    # 生成 YOLO 标注文件 (.txt)
    label_name = Path(file_name).stem + ".txt"
    label_path = LABELS_DIR / label_name
    label_path.parent.mkdir(parents=True, exist_ok=True)

    lines = []
    for ann in anns:
        cat_id = ann["category_id"]
        bbox = ann["bbox"]  # COCO: [x, y, w, h]

        x, y, w, h = bbox

        # YOLO: x_center y_center width height (归一化)
        x_center = (x + w / 2.0) / img_w
        y_center = (y + h / 2.0) / img_h
        norm_w = w / img_w
        norm_h = h / img_h

        # 边界裁剪
        x_center = max(0.0, min(1.0, x_center))
        y_center = max(0.0, min(1.0, y_center))
        norm_w = max(0.0, min(1.0, norm_w))
        norm_h = max(0.0, min(1.0, norm_h))

        # TACO category_id 就是 YOLO class_id（0-59）
        lines.append(f"{cat_id} {x_center:.6f} {y_center:.6f} {norm_w:.6f} {norm_h:.6f}")

    with open(label_path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines))

    converted += 1

print(f"  转换完成: {converted} 张 (跳过了 {skipped_no_img} 张无图片, {skipped_no_ann} 张无标注)")

# ============================================================
# 4. 划分 train/val
# ============================================================
print("\n[4/5] 划分训练集/验证集...")

# 收集所有已转换的图片
valid_items = []
for img_id, img_info in images_info.items():
    file_name = img_info["file_name"]
    img_path = IMAGES_DIR / file_name
    label_name = Path(file_name).stem + ".txt"
    label_path = LABELS_DIR / label_name

    if img_path.exists() and label_path.exists():
        valid_items.append((file_name, label_name))

random.shuffle(valid_items)
split_idx = int(len(valid_items) * TRAIN_RATIO)
train_items = valid_items[:split_idx]
val_items = valid_items[split_idx:]

print(f"  训练集: {len(train_items)}  验证集: {len(val_items)}")

# 创建目录结构
for split in ["train", "val"]:
    (BASE_DIR / "images" / split).mkdir(parents=True, exist_ok=True)
    (BASE_DIR / "labels" / split).mkdir(parents=True, exist_ok=True)

# 移动文件
for items, split in [(train_items, "train"), (val_items, "val")]:
    for file_name, label_name in items:
        # 图片
        src_img = IMAGES_DIR / file_name
        dst_img = BASE_DIR / "images" / split / file_name
        if not dst_img.parent.exists():
            dst_img.parent.mkdir(parents=True, exist_ok=True)
        shutil.move(str(src_img), str(dst_img))

        # 标注
        src_lbl = LABELS_DIR / label_name
        dst_lbl = BASE_DIR / "labels" / split / label_name
        if not dst_lbl.parent.exists():
            dst_lbl.parent.mkdir(parents=True, exist_ok=True)
        shutil.move(str(src_lbl), str(dst_lbl))

# 清理空目录
for d in [IMAGES_DIR / "batch_1", IMAGES_DIR / "batch_2", IMAGES_DIR / "batch_3",
          IMAGES_DIR / "batch_4", IMAGES_DIR / "batch_5", IMAGES_DIR / "batch_6",
          IMAGES_DIR / "batch_7", IMAGES_DIR / "batch_8", IMAGES_DIR / "batch_9",
          IMAGES_DIR / "batch_10", IMAGES_DIR / "batch_11", IMAGES_DIR / "batch_12",
          IMAGES_DIR / "batch_13", IMAGES_DIR / "batch_14", IMAGES_DIR / "batch_15",
          LABELS_DIR]:
    try:
        if d.exists() and not any(d.iterdir()):
            d.rmdir()
    except OSError:
        pass

print("  数据划分完成")

# ============================================================
# 5. 生成 dataset.yaml
# ============================================================
print("\n[5/5] 生成 dataset.yaml...")

# 类别名（TACO 标准名称）
class_names = [categories[cid]["name"] for cid in sorted(categories.keys())]

yaml_content = f"""# TACO 垃圾分类数据集 - YOLOv8 训练配置
# 自动生成 by convert_to_yolo.py

path: {BASE_DIR.as_posix()}
train: images/train
val: images/val

nc: {len(class_names)}
names:
"""
for i, name in enumerate(class_names):
    yaml_content += f"  {i}: {name}\n"

yaml_path = BASE_DIR / "dataset.yaml"
with open(yaml_path, "w", encoding="utf-8") as f:
    f.write(yaml_content)

print(f"  dataset.yaml 已生成: {yaml_path}")
print(f"\n{'='*50}")
print(f"转换完成！")
print(f"  图片: {len(train_items)} train + {len(val_items)} val")
print(f"  类别: {len(class_names)}")
print(f"\n训练命令:")
print(f"  yolo train model=yolov8n.pt data={yaml_path} epochs=100 imgsz=640")
print(f"{'='*50}")
