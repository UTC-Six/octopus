package config

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/UTC-Six/octopus/internal/types"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
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
	ServerConfigs []constant.ServerConfig `yaml:"server_configs"`
	ClientConfig  constant.ClientConfig   `yaml:"client_config"`
	NamespaceId   string                  `yaml:"namespace_id"`
	Group         string                  `yaml:"group"`
	DataId        string                  `yaml:"data_id"`
}

// NewNacosConfigCenter 创建Nacos配置中心
func NewNacosConfigCenter(config NacosConfig) (*NacosConfigCenter, error) {
	// 创建配置客户端
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &config.ClientConfig,
			ServerConfigs: config.ServerConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create nacos client: %w", err)
	}

	center := &NacosConfigCenter{
		client:      client,
		namespaceId: config.NamespaceId,
		group:       config.Group,
		dataId:      config.DataId,
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
			log.Printf("Config changed: namespace=%s, group=%s, dataId=%s", namespace, group, dataId)

			var configs []types.ModelConfig
			if err := json.Unmarshal([]byte(data), &configs); err != nil {
				log.Printf("Failed to unmarshal changed config: %v", err)
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
