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
)

func main() {
	// 创建基于文件的配置中心
	configCenter := config.NewFileConfigCenter("configs/chat-models.json")

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
		status := "❌ Disabled"
		if model.IsAvailable() {
			status = "✅ Enabled"
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

	// 测试配置热更新
	fmt.Println("\n=== Test Config Hot Reload ===")
	testConfigHotReload(configCenter, modelRouter)
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

func testConfigHotReload(configCenter *config.FileConfigCenter, modelRouter *router.DefaultModelRouter) {
	fmt.Println("Testing config hot reload...")
	fmt.Println("You can edit configs/chat-models.json and see changes reflected here")

	// 监听配置变化
	configCenter.WatchConfigChanges(func(configs []types.ModelConfig) {
		fmt.Printf("\n🔄 Config reloaded at %s\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf("Total models: %d\n", len(configs))
		for _, config := range configs {
			status := "❌ Disabled"
			if config.Enabled {
				status = "✅ Enabled"
			}
			fmt.Printf("  %s %s (priority: %d, weight: %d)\n",
				status, config.Name, config.Priority, config.Weight)
		}
	})

	// 等待用户输入
	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()
}
