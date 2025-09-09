#!/bin/bash

# 等待 Nacos 启动的脚本

set -e

MAX_ATTEMPTS=60
ATTEMPT=0
SLEEP_INTERVAL=3

echo "⏳ Waiting for Nacos to be ready..."

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
    # 首先检查端口是否开放
    if nc -z localhost 8848 2>/dev/null; then
        # 然后检查基本的 HTTP 响应
        if curl -s http://localhost:8848/nacos/ 2>/dev/null | grep -q "nacos"; then
            echo "✅ Nacos is ready!"
            exit 0
        fi
    fi
    
    ATTEMPT=$((ATTEMPT + 1))
    echo "   Attempt $ATTEMPT/$MAX_ATTEMPTS - Nacos not ready yet, waiting ${SLEEP_INTERVAL}s..."
    sleep $SLEEP_INTERVAL
done

echo "❌ Nacos failed to start within $((MAX_ATTEMPTS * SLEEP_INTERVAL)) seconds"
echo "   Please check the logs with: docker-compose logs nacos"
echo "   Or check if the container is running with: docker-compose ps"
exit 1
