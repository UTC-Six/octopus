# Octopus - 大模型路由服务

Octopus 是一个大模型路由服务，支持多个大模型的统一管理和智能路由。通过 Nacos 配置中心实现配置的动态更新，支持基于优先级和权重的智能模型选择。

## 功能特性

- **多模型支持**: 支持多个大模型的同时管理
- **智能路由**: 基于优先级和权重的智能模型选择
- **配置热更新**: 通过 Nacos 配置中心实现配置的动态更新
- **健康检查**: 自动检测模型可用性
- **内部调用**: 提供简单的内部调用接口

## 项目结构

```
octopus/
├── internal/              # 内部实现
│   ├── config/           # 配置中心
│   │   └── nacos.go
│   ├── models/           # 大模型实现
│   │   └── default_chat_model.go
│   ├── router/           # 路由逻辑
│   │   └── model_router.go
│   ├── service/          # 服务层
│   │   └── chat_service.go
│   └── types/            # 类型定义
│       └── types.go
├── cmd/                  # 应用程序入口
│   └── example/          # 示例程序
│       └── main.go
├── etc/                  # 配置文件
│   └── config.yaml
└── Makefile             # 构建配置
```

## 核心接口

### ChatModel 接口
```go
type ChatModel interface {
    GetName() string
    GetURL() string
    GetAppKey() string
    GetPriority() int
    GetWeight() int
    IsAvailable() bool
    Chat(ctx context.Context, messages []Message) (*ChatResponse, error)
}
```

### ModelRouter 接口
```go
type ModelRouter interface {
    GetModel(modelName string) (ChatModel, error)
    GetAvailableModel() (ChatModel, error)
    GetAllModels() []ChatModel
    UpdateConfigs(configs []ModelConfig) error
    Start() error
    Stop() error
}
```

## 配置结构

### ModelConfig 模型配置
```go
type ModelConfig struct {
    Name     string `yaml:"name" json:"name"`         // 模型名称
    URL      string `yaml:"url" json:"url"`           // 模型URL
    AppKey   string `yaml:"app_key" json:"app_key"`   // 应用密钥
    Priority int    `yaml:"priority" json:"priority"` // 优先级（越小越高）
    Weight   int    `yaml:"weight" json:"weight"`     // 权重
    Enabled  bool   `yaml:"enabled" json:"enabled"`   // 是否启用
}
```

## 路由机制

### 优先级选择
1. **优先级排序**: 按优先级从小到大排序（数值越小优先级越高）
2. **同优先级权重选择**: 在相同优先级内按权重随机选择
3. **健康检查**: 只选择可用的模型

### 选择算法
```go
// 1. 按优先级分组
priorityGroups := make(map[int][]ChatModel)
for _, model := range availableModels {
    priority := model.GetPriority()
    priorityGroups[priority] = append(priorityGroups[priority], model)
}

// 2. 选择最高优先级组
minPriority := getMinPriority(priorityGroups)

// 3. 在最高优先级组中按权重选择
models := priorityGroups[minPriority]
selectedModel := selectByWeight(models)
```

## 快速开始

### 方式一：一键启动（推荐）

```bash
# 快速启动（自动启动 Nacos + 同步配置）
./scripts/quick-start.sh
```

### 方式二：分步启动

#### 1. 安装依赖

```bash
make deps
```

#### 2. 启动 Nacos

```bash
# 启动 Nacos 容器
make start-nacos

# 检查状态
make status-nacos
```

#### 3. 同步配置

```bash
# 同步本地配置到 Nacos
make sync-config
```

#### 4. 运行示例

```bash
# 运行示例程序
make example

# 监控配置变化
make watch-config
```

### 常用命令

```bash
# 启动 Nacos
make start-nacos

# 停止 Nacos
make stop-nacos

# 重启 Nacos
make restart-nacos

# 查看状态
make status-nacos

# 同步配置
make sync-config

# 监控配置变化
make watch-config

# 运行示例
make example

# 清理所有
make clean-all
```

## 使用示例

### 基本使用

```go
// 创建配置中心
configCenter := createConfigCenter()

// 创建模型路由器
modelRouter := router.NewDefaultModelRouter(configCenter)

// 启动路由器
modelRouter.Start()
defer modelRouter.Stop()

// 创建聊天服务
chatService := service.NewChatService(modelRouter)

// 自动选择模型
response, err := chatService.Chat(ctx, "", messages)

// 指定模型
response, err := chatService.Chat(ctx, "gpt-3.5-turbo", messages)
```

### 配置示例

```yaml
# Nacos 配置
nacos:
  server_configs:
    - ip_addr: "127.0.0.1"
      port: 8848
  client_config:
    namespace_id: "public"
    timeout_ms: 5000
  namespace_id: "public"
  group: "DEFAULT_GROUP"
  data_id: "chat-models"
```

### 模型配置示例

配置文件位置：`configs/chat-models.json`

```json
[
  {
    "name": "gpt-3.5-turbo",
    "url": "https://api.openai.com",
    "app_key": "sk-xxx",
    "priority": 1,
    "weight": 80,
    "enabled": true
  },
  {
    "name": "gpt-4",
    "url": "https://api.openai.com",
    "app_key": "sk-xxx",
    "priority": 1,
    "weight": 20,
    "enabled": true
  },
  {
    "name": "claude-3",
    "url": "https://api.anthropic.com",
    "app_key": "sk-xxx",
    "priority": 2,
    "weight": 100,
    "enabled": true
  }
]
```

### 配置管理

#### 方式一：通过 Nacos 控制台管理

1. 访问 Nacos 控制台：http://localhost:8848/nacos
2. 用户名/密码：nacos/nacos
3. 进入"配置管理" -> "配置列表"
4. 找到 `chat-models` 配置项进行编辑
5. 修改后点击"发布"即可生效

#### 方式二：通过本地文件同步

1. 编辑 `configs/chat-models.json` 文件
2. 运行同步命令：
   ```bash
   make sync-config
   ```

#### 方式三：通过代码更新配置

```go
// 更新配置
newConfigs := []types.ModelConfig{
    // ... 新的配置
}
configCenter.PublishConfig(newConfigs)
```

#### 配置监控

实时监控配置变化：

```bash
# 启动配置监控
make watch-config
```

#### 配置热更新

- 服务会自动监听 Nacos 配置变化
- 配置更新后无需重启服务即可生效
- 支持通过控制台或代码两种方式更新

#### 环境变量配置

可以通过环境变量自定义 Nacos 连接参数：

```bash
export NACOS_SERVER=127.0.0.1:8848
export NACOS_NAMESPACE=public
export NACOS_GROUP=DEFAULT_GROUP
export NACOS_DATA_ID=chat-models
export CONFIG_FILE=configs/chat-models.json
```

## 技术栈

- **Go 1.24.6**: 编程语言
- **Nacos**: 配置中心
- **HTTP Client**: 大模型 API 调用

## 依赖管理

项目使用 Go modules 进行依赖管理，主要依赖包括：

- `github.com/nacos-group/nacos-sdk-go`: Nacos 配置中心客户端

## 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 待办事项

- [ ] 添加更多大模型支持
- [ ] 实现模型调用统计和监控
- [ ] 添加模型调用重试机制
- [ ] 支持流式响应
- [ ] 添加模型调用限流功能
