package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/UTC-Six/octopus/internal/config"
	"github.com/UTC-Six/octopus/internal/types"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
)

func main() {
	// 从环境变量或默认值获取配置
	nacosServer := getEnv("NACOS_SERVER", "127.0.0.1:8848")
	namespace := getEnv("NACOS_NAMESPACE", "public")
	group := getEnv("NACOS_GROUP", "DEFAULT_GROUP")
	dataId := getEnv("NACOS_DATA_ID", "chat-models")
	configFile := getEnv("CONFIG_FILE", "configs/chat-models.json")

	// 解析 Nacos 服务器地址
	serverIP, serverPort := parseServerAddress(nacosServer)

	// 创建 Nacos 配置中心
	nacosConfig := config.NacosConfig{
		ServerConfigs: []constant.ServerConfig{
			{
				IpAddr: serverIP,
				Port:   uint64(serverPort),
			},
		},
		ClientConfig: constant.ClientConfig{
			NamespaceId: namespace,
			TimeoutMs:   5000,
		},
		NamespaceId: namespace,
		Group:       group,
		DataId:      dataId,
	}

	configCenter, err := config.NewNacosConfigCenter(nacosConfig)
	if err != nil {
		log.Fatalf("Failed to create config center: %v", err)
	}
	defer configCenter.Close()

	// 读取本地配置文件
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file %s: %v", configFile, err)
	}

	// 解析配置
	var modelConfigs []types.ModelConfig
	if err := json.Unmarshal(configData, &modelConfigs); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	// 验证配置
	if err := validateConfigs(modelConfigs); err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}

	// 发布配置到 Nacos
	if err := configCenter.PublishConfig(modelConfigs); err != nil {
		log.Fatalf("Failed to publish config: %v", err)
	}

	// 打印结果
	fmt.Printf("✅ Successfully synced %d models to Nacos\n", len(modelConfigs))
	fmt.Printf("📋 Config details:\n")
	fmt.Printf("   Server: %s\n", nacosServer)
	fmt.Printf("   Namespace: %s\n", namespace)
	fmt.Printf("   Group: %s\n", group)
	fmt.Printf("   DataId: %s\n", dataId)
	fmt.Printf("   Source: %s\n", configFile)

	fmt.Printf("\n📊 Model summary:\n")
	for _, config := range modelConfigs {
		status := "❌ Disabled"
		if config.Enabled {
			status = "✅ Enabled"
		}
		fmt.Printf("   %s %s (priority: %d, weight: %d)\n",
			status, config.Name, config.Priority, config.Weight)
	}

	fmt.Printf("\n🌐 Nacos Console: http://%s/nacos (nacos/nacos)\n", nacosServer)
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseServerAddress 解析服务器地址
func parseServerAddress(server string) (string, int) {
	// 简单的解析，假设格式为 "ip:port"
	for i, c := range server {
		if c == ':' {
			ip := server[:i]
			port := server[i+1:]
			return ip, parseInt(port)
		}
	}
	// 如果没有端口，使用默认端口
	return server, 8848
}

// parseInt 简单的字符串转整数
func parseInt(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}

// validateConfigs 验证配置
func validateConfigs(configs []types.ModelConfig) error {
	if len(configs) == 0 {
		return fmt.Errorf("no models configured")
	}

	names := make(map[string]bool)
	for _, config := range configs {
		if config.Name == "" {
			return fmt.Errorf("model name cannot be empty")
		}
		if names[config.Name] {
			return fmt.Errorf("duplicate model name: %s", config.Name)
		}
		names[config.Name] = true

		if config.URL == "" {
			return fmt.Errorf("model %s URL cannot be empty", config.Name)
		}
		if config.AppKey == "" {
			return fmt.Errorf("model %s AppKey cannot be empty", config.Name)
		}
		if config.Priority < 0 {
			return fmt.Errorf("model %s priority cannot be negative", config.Name)
		}
		if config.Weight < 0 {
			return fmt.Errorf("model %s weight cannot be negative", config.Name)
		}
	}

	return nil
}
