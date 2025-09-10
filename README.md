# Octopus - 大模型路由系统

一个基于 Go 的大模型路由系统，支持从 Nacos 配置中心动态加载模型配置，实现优先级和权重-based 的模型选择。

## 功能特性

- 🚀 **动态配置管理**: 从 Nacos 配置中心获取模型配置
- ⚡ **热更新**: 支持配置实时更新，无需重启服务
- 🎯 **智能路由**: 基于优先级和权重的模型选择算法
- 🔧 **灵活配置**: 支持模型启用/禁用、优先级调整、权重分配
- 📊 **状态监控**: 实时显示模型状态和配置信息

## 架构设计

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Nacos 配置中心  │────│   配置中心接口     │────│   模型路由器      │
│  (远程服务器)    │    │  ConfigCenter    │    │  ModelRouter    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                       │
                                                       ▼
                                               ┌─────────────────┐
                                               │   聊天服务       │
                                               │  ChatService    │
                                               └─────────────────┘
```

## 快速开始

### 1. 环境要求

- Go 1.21+
- 访问远程 Nacos 服务器 (172.31.0.16:8848)

### 2. 安装依赖

```bash
make deps
```

### 3. 构建项目

```bash
make build
```

### 4. 运行示例

```bash
make example
```

## 配置说明

### Nacos 配置

系统连接到远程 Nacos 服务器，配置如下：

- **服务器地址**: 172.31.0.16:8848
- **GRPC 端口**: 9848
- **命名空间**: bfdb58fb-11fd-4e95-a78c-3432cdcfa104
- **配置组**: DEFAULT_GROUP
- **配置ID**: chat-models

### 配置文件

项目使用 `config.yaml` 文件管理 Nacos 连接配置：

```yaml
nacos:
  server_configs:
    - ip_addr: "172.31.0.16"
      port: 8848
      grpc_port: 9848
  client_config:
    namespace_id: "bfdb58fb-11fd-4e95-a78c-3432cdcfa104"
    timeout_ms: 5000
    not_load_cache_at_start: true
    log_dir: "/tmp/nacos/log"
    cache_dir: "/tmp/nacos/cache"
    log_level: "debug"
  namespace_id: "bfdb58fb-11fd-4e95-a78c-3432cdcfa104"
  group: "DEFAULT_GROUP"
  data_id: "chat-models"
```

### 模型配置格式

在 Nacos 中创建配置，DataId: `chat-models`, Group: `DEFAULT_GROUP`，内容格式如下：

```json
[
  {
    "name": "gpt-3.5-turbo",
    "url": "https://api.openai.com/v1/chat/completions",
    "app_key": "sk-your-api-key",
    "priority": 1,
    "weight": 80,
    "enabled": true
  },
  {
    "name": "gpt-4",
    "url": "https://api.openai.com/v1/chat/completions", 
    "app_key": "sk-your-api-key",
    "priority": 1,
    "weight": 20,
    "enabled": true
  },
  {
    "name": "claude-3",
    "url": "https://api.anthropic.com/v1/messages",
    "app_key": "sk-ant-your-api-key",
    "priority": 2,
    "weight": 100,
    "enabled": true
  }
]
```

### 配置字段说明

- `name`: 模型名称，用于标识和选择模型
- `url`: 模型的 API 端点
- `app_key`: API 密钥
- `priority`: 优先级，数值越小优先级越高
- `weight`: 权重，用于同优先级模型的选择
- `enabled`: 是否启用该模型

## 使用示例

### 基本使用

```go
package main

import (
    "context"
    "github.com/UTC-Six/octopus/internal/config"
    "github.com/UTC-Six/octopus/internal/router"
    "github.com/UTC-Six/octopus/internal/service"
    "github.com/UTC-Six/octopus/internal/types"
)

func main() {
    // 创建配置中心
    configCenter, err := config.NewNacosConfigCenter()
    if err != nil {
        log.Fatal(err)
    }

    // 创建模型路由器
    modelRouter := router.NewDefaultModelRouter(configCenter)
    
    // 启动路由器
    if err := modelRouter.Start(); err != nil {
        log.Fatal(err)
    }
    defer modelRouter.Stop()

    // 创建聊天服务
    chatService := service.NewChatService(modelRouter)

    // 自动选择模型
    messages := []types.Message{
        {Role: "user", Content: "Hello!"},
    }
    
    response, err := chatService.Chat(context.Background(), "", messages)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Model: %s\n", response.Model)
    fmt.Printf("Response: %s\n", response.Content)
}
```

### 指定模型

```go
// 使用指定模型
response, err := chatService.Chat(context.Background(), "gpt-4", messages)
```

### 获取可用模型

```go
models := chatService.GetAvailableModels()
for _, model := range models {
    fmt.Printf("Model: %s, Priority: %d, Weight: %d, Available: %v\n",
        model.GetName(), model.GetPriority(), model.GetWeight(), model.IsAvailable())
}
```

## 模型选择算法

系统使用以下算法选择模型：

1. **优先级筛选**: 选择优先级最高的模型
2. **权重选择**: 在相同优先级的模型中，按权重随机选择
3. **可用性检查**: 只选择启用的模型

## 常用命令

```bash
# 构建项目
make build

# 运行示例
make example

# 清理构建文件
make clean

# 运行测试
make test

# 安装依赖
make deps
```

## 项目结构

```
octopus/
├── cmd/
│   └── example/           # 示例程序
├── internal/
│   ├── config/           # 配置中心实现
│   │   └── nacos.go      # Nacos 配置中心
│   ├── models/           # 模型实现
│   │   └── default_chat_model.go
│   ├── router/           # 模型路由器
│   │   └── model_router.go
│   ├── service/          # 服务层
│   │   └── chat_service.go
│   └── types/            # 类型定义
│       └── types.go
├── config.yaml           # 配置文件
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 开发说明

### 添加新的配置中心

实现 `types.ConfigCenter` 接口：

```go
type ConfigCenter interface {
    GetModelConfigs() ([]ModelConfig, error)
    WatchConfigChanges(callback func([]ModelConfig)) error
    Close() error
}
```

### 添加新的模型实现

实现 `types.ChatModel` 接口：

```go
type ChatModel interface {
    GetName() string
    GetPriority() int
    GetWeight() int
    IsAvailable() bool
    Chat(ctx context.Context, messages []Message) (*ChatResponse, error)
}
```

## 技术栈

- **Go**: 1.21+
- **Nacos SDK**: v2.3.3
- **YAML**: gopkg.in/yaml.v3

## 许可证

MIT License