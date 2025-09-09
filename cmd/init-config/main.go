package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/UTC-Six/octopus/internal/config"
	"github.com/UTC-Six/octopus/internal/types"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
)

func main() {
	// 创建 Nacos 配置中心
	nacosConfig := config.NacosConfig{
		ServerConfigs: []constant.ServerConfig{
			{
				IpAddr: "127.0.0.1",
				Port:   8848,
			},
		},
		ClientConfig: constant.ClientConfig{
			NamespaceId: "public",
			TimeoutMs:   5000,
		},
		NamespaceId: "public",
		Group:       "DEFAULT_GROUP",
		DataId:      "chat-models",
	}

	configCenter, err := config.NewNacosConfigCenter(nacosConfig)
	if err != nil {
		log.Fatalf("Failed to create config center: %v", err)
	}
	defer configCenter.Close()

	// 示例配置
	modelConfigs := []types.ModelConfig{
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
		{
			Name:     "gemini-pro",
			URL:      "https://generativelanguage.googleapis.com",
			AppKey:   "sk-mock-key-4",
			Priority: 3,
			Weight:   50,
			Enabled:  false, // 禁用的模型
		},
	}

	// 发布配置到 Nacos
	if err := configCenter.PublishConfig(modelConfigs); err != nil {
		log.Fatalf("Failed to publish config: %v", err)
	}

	// 打印配置内容
	configJSON, _ := json.MarshalIndent(modelConfigs, "", "  ")
	fmt.Println("Configuration published to Nacos:")
	fmt.Println(string(configJSON))

	fmt.Println("\nConfiguration details:")
	for _, config := range modelConfigs {
		status := "❌ Disabled"
		if config.Enabled {
			status = "✅ Enabled"
		}
		fmt.Printf("%s %s (priority: %d, weight: %d)\n",
			status, config.Name, config.Priority, config.Weight)
	}

	fmt.Println("\nYou can now run the example with: make example")
}
