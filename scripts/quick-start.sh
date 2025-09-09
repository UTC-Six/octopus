#!/bin/bash

# 快速启动脚本

set -e

echo "🚀 Octopus 大模型路由服务快速启动"
echo "=================================="

# 检查 Docker 是否运行
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# 检查 Docker Compose 是否可用
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed. Please install it first."
    exit 1
fi

echo "✅ Docker environment check passed"

# 启动 Nacos
echo "📦 Starting Nacos..."
make start-nacos

# 等待一下确保 Nacos 完全启动
sleep 5

# 同步配置
echo "📋 Syncing configuration..."
make sync-config

# 显示状态
echo ""
echo "🎉 Setup complete!"
echo ""
echo "📊 Available commands:"
echo "  make status-nacos    - Check Nacos status"
echo "  make sync-config     - Sync local config to Nacos"
echo "  make watch-config    - Watch config changes"
echo "  make example         - Run example program"
echo "  make stop-nacos      - Stop Nacos"
echo ""
echo "🌐 Nacos Console: http://localhost:8848/nacos"
echo "   Username: nacos"
echo "   Password: nacos"
echo ""
echo "💡 You can now:"
echo "   1. Run 'make example' to test the service"
echo "   2. Run 'make watch-config' to monitor config changes"
echo "   3. Edit configs/chat-models.json and run 'make sync-config' to update"
echo "   4. Use Nacos console to modify configs directly"
