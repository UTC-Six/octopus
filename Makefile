.PHONY: build example init-config clean test deps

# 构建项目
build:
	go mod tidy
	go build -o bin/example ./cmd/example
	go build -o bin/init-config ./cmd/init-config

# 初始化配置（推送到 Nacos）
init-config: build
	./bin/init-config

# 运行示例
example: build
	./bin/example

# 清理构建文件
clean:
	rm -rf bin/

# 运行测试
test:
	go test ./...

# 安装依赖
deps:
	go mod download