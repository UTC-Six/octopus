package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/UTC-Six/octopus/internal/config"
	"github.com/UTC-Six/octopus/internal/router"
	"github.com/UTC-Six/octopus/internal/service"
	"github.com/UTC-Six/octopus/internal/types"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
)

func main() {
	// 创建 Nacos 配置中心
	configCenter, err := createNacosConfigCenter()
	if err != nil {
		log.Fatalf("Failed to create config center: %v", err)
	}

	// 创建模型路由器
	modelRouter := router.NewDefaultModelRouter(configCenter)

	// 启动路由器
	if err := modelRouter.Start(); err != nil {
		log.Fatalf("Failed to start model router: %v", err)
	}
	defer modelRouter.Stop()

	// 创建聊天服务
	chatService := service.NewChatService(modelRouter)

	// 等待一下让配置加载完成
	time.Sleep(1 * time.Second)

	// 测试获取模型列表
	fmt.Println("=== Available Models ===")
	models := chatService.GetAvailableModels()
	for _, model := range models {
		status := "❌"
		if model.IsAvailable() {
			status = "✅"
		}
		fmt.Printf("%s %s (priority: %d, weight: %d)\n",
			status, model.GetName(), model.GetPriority(), model.GetWeight())
	}

	// 测试自动选择模型
	fmt.Println("\n=== Test Auto Selection ===")
	testAutoSelection(chatService)

	// 测试指定模型
	fmt.Println("\n=== Test Specific Model ===")
	testSpecificModel(chatService, "gpt-3.5-turbo")

	// 测试不存在的模型
	fmt.Println("\n=== Test Non-existent Model ===")
	testSpecificModel(chatService, "non-existent-model")
}

func testAutoSelection(chatService *service.ChatService) {
	ctx := context.Background()
	messages := []types.Message{
		{
			Role:    "user",
			Content: "Hello, how are you?",
		},
	}

	response, err := chatService.Chat(ctx, "", messages)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Model: %s\n", response.Model)
		fmt.Printf("Response: %s\n", response.Content)
	}
}

func testSpecificModel(chatService *service.ChatService, modelName string) {
	ctx := context.Background()
	messages := []types.Message{
		{
			Role:    "user",
			Content: "What is the capital of France?",
		},
	}

	response, err := chatService.Chat(ctx, modelName, messages)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Model: %s\n", response.Model)
		fmt.Printf("Response: %s\n", response.Content)
	}
}

// createNacosConfigCenter 创建 Nacos 配置中心
func createNacosConfigCenter() (types.ConfigCenter, error) {
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

	return config.NewNacosConfigCenter(nacosConfig)
}
