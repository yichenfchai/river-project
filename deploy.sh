#!/bin/bash
# deploy.sh — 一键部署到 你的服务器IP
# 用法: bash deploy.sh [--frontend] [--backend] [--all]

set -e
SERVER="ubuntu@你的服务器IP"
REMOTE_DIR="/opt/gcg"

FRONTEND=false
BACKEND=false

# 默认全部部署
if [ $# -eq 0 ]; then
    FRONTEND=true
    BACKEND=true
fi

for arg in "$@"; do
    case $arg in
        --frontend) FRONTEND=true ;;
        --backend)  BACKEND=true ;;
        --all)      FRONTEND=true; BACKEND=true ;;
    esac
done

# ── 前端 ──
if $FRONTEND; then
    echo "=== [1/3] 构建前端 ==="
    cd web && npm run build && cd ..
    echo "=== [2/3] 上传前端 ==="
    ssh "$SERVER" "sudo systemctl stop gcg 2>/dev/null; true"
    scp -r web/dist "$SERVER:$REMOTE_DIR/web/dist"
    echo "前端已上传"
fi

# ── 后端 ──
if $BACKEND; then
    echo "=== [1/3] 交叉编译 Go ==="
    export CGO_ENABLED=0 GOOS=linux GOARCH=amd64
    go build -ldflags="-s -w" -o gcg ./cmd/server
    echo "=== [2/3] 上传后端 ==="
    ssh "$SERVER" "sudo systemctl stop gcg 2>/dev/null; true"
    scp gcg "$SERVER:$REMOTE_DIR/gcg"
    echo "后端已上传"
fi

# ── 重启服务 ──
echo "=== [3/3] 重启服务 ==="
ssh "$SERVER" "chmod +x $REMOTE_DIR/gcg && sudo systemctl start gcg && sleep 1 && sudo systemctl status gcg --no-pager | head -5"
echo ""
echo "部署完成 ✓"
echo "验证: curl http://你的服务器IP:8081/health"
