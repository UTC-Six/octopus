.PHONY: build example clean test deps

# 构建项目
build:
	go mod tidy
	go build -o bin/example ./cmd/example

# 运行示例
example: build
	./bin/example

# 清理构建文件
clean:
	rm -rf bin/

# 清理所有生成的文件
clean-all: clean
	go clean -cache

# 运行测试
test:
	go test ./...

# 安装依赖
deps:
	go mod download
	go mod tidy