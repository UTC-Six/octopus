package types

import (
	"context"
	"time"
)

// ChatModel 大模型接口
type ChatModel interface {
	GetName() string
	GetURL() string
	GetAppKey() string
	GetPriority() int
	GetWeight() int
	IsAvailable() bool
	Chat(ctx context.Context, messages []Message) (*ChatResponse, error)
}

// Message 聊天消息
type Message struct {
	Role    string `json:"role"`    // user, assistant, system
	Content string `json:"content"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Content   string    `json:"content"`
	Model     string    `json:"model"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// ModelConfig 大模型配置
type ModelConfig struct {
	Name     string `yaml:"name" json:"name"`         // 模型名称
	URL      string `yaml:"url" json:"url"`           // 模型URL
	AppKey   string `yaml:"app_key" json:"app_key"`   // 应用密钥
	Priority int    `yaml:"priority" json:"priority"` // 优先级（越小越高）
	Weight   int    `yaml:"weight" json:"weight"`     // 权重
	Enabled  bool   `yaml:"enabled" json:"enabled"`   // 是否启用
}

// ConfigCenter 配置中心接口
type ConfigCenter interface {
	GetModelConfigs() ([]ModelConfig, error)
	WatchConfigChanges(callback func([]ModelConfig)) error
	Close() error
}

// ModelRouter 大模型路由接口
type ModelRouter interface {
	GetModel(modelName string) (ChatModel, error)
	GetAvailableModel() (ChatModel, error)
	GetAllModels() []ChatModel
	UpdateConfigs(configs []ModelConfig) error
	Start() error
	Stop() error
}