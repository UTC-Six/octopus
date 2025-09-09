.PHONY: build example init-config clean test deps start-nacos stop-nacos restart-nacos status-nacos sync-config watch-config

# 构建项目
build:
	go mod tidy
	go build -o bin/example ./cmd/example
	go build -o bin/simple-example ./cmd/simple-example
	go build -o bin/file-example ./cmd/file-example
	go build -o bin/demo ./cmd/demo
	go build -o bin/init-config ./cmd/init-config
	go build -o bin/sync-config ./cmd/sync-config
	go build -o bin/watch-config ./cmd/watch-config

# 启动 Nacos
start-nacos:
	@echo "Starting Nacos..."
	docker-compose up -d nacos
	@echo "Nacos container started. Please wait for it to fully initialize..."
	@echo "You can check the status with: make status-nacos"
	@echo "Or view logs with: docker-compose logs -f nacos"
	@echo "Console will be available at: http://localhost:8848/nacos (nacos/nacos)"

# 停止 Nacos
stop-nacos:
	@echo "Stopping Nacos..."
	docker-compose down

# 重启 Nacos
restart-nacos: stop-nacos start-nacos

# 查看 Nacos 状态
status-nacos:
	@echo "Nacos status:"
	docker-compose ps nacos
	@echo ""
	@echo "Health check:"
	@curl -f http://localhost:8848/nacos/v1/console/health/readiness 2>/dev/null && echo "✅ Nacos is healthy" || echo "❌ Nacos is not responding"

# 同步本地配置到 Nacos
sync-config: build
	@echo "Syncing local config to Nacos..."
	./bin/sync-config

# 初始化配置（推送到 Nacos）
init-config: build
	./bin/init-config

# 运行示例
example: build
	./bin/example

# 运行简单示例（不依赖 Nacos）
simple-example: build
	./bin/simple-example

# 运行文件配置示例（支持热更新）
file-example: build
	./bin/file-example

# 运行演示程序（推荐）
demo: build
	./bin/demo

# 监控配置变化
watch-config: build
	./bin/watch-config

# 完整启动流程（启动 Nacos + 同步配置 + 运行示例）
start: start-nacos sync-config
	@echo "Setup complete! You can now:"
	@echo "1. View Nacos console: http://localhost:8848/nacos (nacos/nacos)"
	@echo "2. Run example: make example"
	@echo "3. Sync config changes: make sync-config"

# 清理构建文件
clean:
	rm -rf bin/

# 清理所有（包括 Docker 容器和卷）
clean-all: clean
	docker-compose down -v
	docker system prune -f

# 运行测试
test:
	go test ./...

# 安装依赖
deps:
	go mod download