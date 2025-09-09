package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// 获取初始配置
	initialConfigs, err := configCenter.GetModelConfigs()
	if err != nil {
		log.Fatalf("Failed to get initial configs: %v", err)
	}

	fmt.Printf("🔍 Watching config changes...\n")
	fmt.Printf("📋 Initial config (%d models):\n", len(initialConfigs))
	printConfigs(initialConfigs)

	// 设置信号处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听配置变化
	configCh := make(chan []types.ModelConfig, 1)
	err = configCenter.WatchConfigChanges(func(configs []types.ModelConfig) {
		select {
		case configCh <- configs:
		default:
		}
	})
	if err != nil {
		log.Fatalf("Failed to watch config changes: %v", err)
	}

	// 启动配置变化处理协程
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case configs := <-configCh:
				fmt.Printf("\n🔄 Config changed at %s\n", time.Now().Format("2006-01-02 15:04:05"))
				printConfigs(configs)
			}
		}
	}()

	// 等待中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("\n⏳ Press Ctrl+C to stop watching...\n")
	<-sigCh

	fmt.Printf("\n👋 Stopping config watcher...\n")
}

// printConfigs 打印配置信息
func printConfigs(configs []types.ModelConfig) {
	for _, config := range configs {
		status := "❌ Disabled"
		if config.Enabled {
			status = "✅ Enabled"
		}
		fmt.Printf("   %s %s (priority: %d, weight: %d, url: %s)\n",
			status, config.Name, config.Priority, config.Weight, config.URL)
	}
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
	for i, c := range server {
		if c == ':' {
			ip := server[:i]
			port := server[i+1:]
			return ip, parseInt(port)
		}
	}
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
