package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/UTC-Six/octopus/internal/types"
)

// FileConfigCenter 基于文件的配置中心
type FileConfigCenter struct {
	configFile string
	mu         sync.RWMutex
	callbacks  []func([]types.ModelConfig)
	stopCh     chan struct{}
}

// NewFileConfigCenter 创建基于文件的配置中心
func NewFileConfigCenter(configFile string) *FileConfigCenter {
	return &FileConfigCenter{
		configFile: configFile,
		callbacks:  make([]func([]types.ModelConfig), 0),
		stopCh:     make(chan struct{}),
	}
}

// GetModelConfigs 获取模型配置
func (f *FileConfigCenter) GetModelConfigs() ([]types.ModelConfig, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// 检查文件是否存在
	if _, err := os.Stat(f.configFile); os.IsNotExist(err) {
		// 如果文件不存在，创建默认配置
		return f.createDefaultConfig()
	}

	// 读取配置文件
	data, err := ioutil.ReadFile(f.configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var configs []types.ModelConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return configs, nil
}

// WatchConfigChanges 监听配置变化
func (f *FileConfigCenter) WatchConfigChanges(callback func([]types.ModelConfig)) error {
	f.mu.Lock()
	f.callbacks = append(f.callbacks, callback)
	f.mu.Unlock()

	// 启动文件监听
	go f.watchFile()

	return nil
}

// watchFile 监听文件变化
func (f *FileConfigCenter) watchFile() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-f.stopCh:
			return
		case <-ticker.C:
			// 检查文件是否存在
			if _, err := os.Stat(f.configFile); err != nil {
				continue
			}

			// 重新加载配置
			configs, err := f.GetModelConfigs()
			if err != nil {
				continue
			}

			// 通知所有回调
			f.mu.RLock()
			for _, cb := range f.callbacks {
				go cb(configs)
			}
			f.mu.RUnlock()
		}
	}
}

// Close 关闭配置中心
func (f *FileConfigCenter) Close() error {
	close(f.stopCh)
	return nil
}

// PublishConfig 发布配置
func (f *FileConfigCenter) PublishConfig(configs []types.ModelConfig) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 确保目录存在
	dir := filepath.Dir(f.configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := ioutil.WriteFile(f.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// createDefaultConfig 创建默认配置
func (f *FileConfigCenter) createDefaultConfig() ([]types.ModelConfig, error) {
	defaultConfigs := []types.ModelConfig{
		{
			Name:     "gpt-3.5-turbo",
			URL:      "https://api.openai.com",
			AppKey:   "sk-mock-key-1",
			Priority: 1,
			Weight:   80,
			Enabled:  true,
		},
		{
			Name:     "gpt-4",
			URL:      "https://api.openai.com",
			AppKey:   "sk-mock-key-2",
			Priority: 1,
			Weight:   20,
			Enabled:  true,
		},
		{
			Name:     "claude-3",
			URL:      "https://api.anthropic.com",
			AppKey:   "sk-mock-key-3",
			Priority: 2,
			Weight:   100,
			Enabled:  true,
		},
	}

	// 创建默认配置文件
	if err := f.PublishConfig(defaultConfigs); err != nil {
		return nil, err
	}

	return defaultConfigs, nil
}
