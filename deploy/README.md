# GCG 部署文件清单

## 文件结构
```
deploy/
  .env          → 服务器端 /opt/gcg/.env
  gcg.service   → 服务器端 /etc/systemd/system/gcg.service
  gcg.conf      → 服务器端 /opt/1panel/apps/openresty/openresty/conf/vhost/gcg.conf
```

## 一、本地构建 & 上传（Windows）

```powershell
cd D:\zhuomian\r-dvp

# 1. 构建前端
cd web && npm run build && cd ..

# 2. 交叉编译 Go 二进制
$env:CGO_ENABLED="0"
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -ldflags="-s -w" -o gcg ./cmd/server

# 3. 上传到服务器（替换 IP）
scp gcg ubuntu@你的服务器IP:/opt/gcg/gcg
scp -r web/dist ubuntu@你的服务器IP:/opt/gcg/web/dist
scp deploy/.env ubuntu@你的服务器IP:/opt/gcg/.env
scp deploy/gcg.service ubuntu@你的服务器IP:/tmp/gcg.service
scp deploy/gcg.conf ubuntu@你的服务器IP:/tmp/gcg.conf
```

## 二、服务器端部署

```bash
# 创建目录 & 权限
sudo mkdir -p /opt/gcg/web
sudo chown -R ubuntu:ubuntu /opt/gcg

# 安装 systemd 服务
sudo cp /tmp/gcg.service /etc/systemd/system/gcg.service
sudo systemctl daemon-reload
sudo systemctl enable --now gcg
sudo systemctl status gcg

# Nginx 配置（手动方式）
sudo cp /tmp/gcg.conf /opt/1panel/apps/openresty/openresty/conf/vhost/gcg.conf
# 然后在 1Panel 中重载 OpenResty
```

## 三、验证

```bash
# 健康检查
curl http://127.0.0.1:8080/health

# 数据库迁移检查
psql -U canal -d grand_canal -h 127.0.0.1 -c "\dt"
```
