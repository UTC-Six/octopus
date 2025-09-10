package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/UTC-Six/octopus/internal/types"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/zeromicro/go-zero/core/logx"
	"gopkg.in/yaml.v3"
)

// NacosConfigCenter Nacos配置中心实现
type NacosConfigCenter struct {
	client      config_client.IConfigClient
	namespaceId string
	group       string
	dataId      string
	mu          sync.RWMutex
	callbacks   []func([]types.ModelConfig)
}

// NacosConfig Nacos配置
type NacosConfig struct {
	ServerConfigs []ServerConfig `yaml:"server_configs"`
	ClientConfig  ClientConfig   `yaml:"client_config"`
	NamespaceId   string         `yaml:"namespace_id"`
	Group         string         `yaml:"group"`
	DataId        string         `yaml:"data_id"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	IpAddr   string `yaml:"ip_addr"`
	Port     uint64 `yaml:"port"`
	GrpcPort uint64 `yaml:"grpc_port"`
}

// ClientConfig 客户端配置
type ClientConfig struct {
	NamespaceId         string `yaml:"namespace_id"`
	TimeoutMs           uint64 `yaml:"timeout_ms"`
	NotLoadCacheAtStart bool   `yaml:"not_load_cache_at_start"`
	LogDir              string `yaml:"log_dir"`
	CacheDir            string `yaml:"cache_dir"`
	LogLevel            string `yaml:"log_level"`
}

// Config 应用配置
type Config struct {
	Nacos NacosConfig `yaml:"nacos"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// NewNacosConfigCenter 创建Nacos配置中心
func NewNacosConfigCenter() (*NacosConfigCenter, error) {
	// 加载配置文件
	config, err := LoadConfig("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 转换服务器配置
	serverConfigs := make([]constant.ServerConfig, len(config.Nacos.ServerConfigs))
	for i, s := range config.Nacos.ServerConfigs {
		serverConfigs[i] = constant.ServerConfig{
			IpAddr:   s.IpAddr,
			Port:     s.Port,
			GrpcPort: s.GrpcPort,
		}
	}

	// 转换客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         config.Nacos.ClientConfig.NamespaceId,
		TimeoutMs:           config.Nacos.ClientConfig.TimeoutMs,
		NotLoadCacheAtStart: config.Nacos.ClientConfig.NotLoadCacheAtStart,
		LogDir:              config.Nacos.ClientConfig.LogDir,
		CacheDir:            config.Nacos.ClientConfig.CacheDir,
		LogLevel:            config.Nacos.ClientConfig.LogLevel,
	}

	// 创建配置客户端
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create nacos client: %w", err)
	}

	center := &NacosConfigCenter{
		client:      client,
		namespaceId: config.Nacos.NamespaceId,
		group:       config.Nacos.Group,
		dataId:      config.Nacos.DataId,
		callbacks:   make([]func([]types.ModelConfig), 0),
	}

	return center, nil
}

// GetModelConfigs 获取模型配置
func (n *NacosConfigCenter) GetModelConfigs() ([]types.ModelConfig, error) {
	content, err := n.client.GetConfig(vo.ConfigParam{
		DataId: n.dataId,
		Group:  n.group,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get config from nacos: %w", err)
	}

	var configs []types.ModelConfig
	if err := json.Unmarshal([]byte(content), &configs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return configs, nil
}

// WatchConfigChanges 监听配置变化
func (n *NacosConfigCenter) WatchConfigChanges(callback func([]types.ModelConfig)) error {
	n.mu.Lock()
	n.callbacks = append(n.callbacks, callback)
	n.mu.Unlock()

	err := n.client.ListenConfig(vo.ConfigParam{
		DataId: n.dataId,
		Group:  n.group,
		OnChange: func(namespace, group, dataId, data string) {
			logx.Infof("Config changed: namespace=%s, group=%s, dataId=%s", namespace, group, dataId)

			var configs []types.ModelConfig
			if err := json.Unmarshal([]byte(data), &configs); err != nil {
				logx.Errorf("Failed to unmarshal changed config: %v", err)
				return
			}

			n.mu.RLock()
			for _, cb := range n.callbacks {
				go cb(configs)
			}
			n.mu.RUnlock()
		},
	})

	if err != nil {
		return fmt.Errorf("failed to listen config changes: %w", err)
	}

	return nil
}

// Close 关闭配置中心
func (n *NacosConfigCenter) Close() error {
	// Nacos客户端没有显式的Close方法，这里可以做一些清理工作
	return nil
}

// PublishConfig 发布配置（用于测试）
func (n *NacosConfigCenter) PublishConfig(configs []types.ModelConfig) error {
	data, err := json.Marshal(configs)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	success, err := n.client.PublishConfig(vo.ConfigParam{
		DataId:  n.dataId,
		Group:   n.group,
		Content: string(data),
	})
	if err != nil {
		return fmt.Errorf("failed to publish config: %w", err)
	}

	if !success {
		return fmt.Errorf("failed to publish config to nacos")
	}

	return nil
}
