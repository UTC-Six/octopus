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
	fmt.Println("🚀 大模型路由系统演示")
	fmt.Println("========================")

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

	// 等待配置加载
	time.Sleep(1 * time.Second)

	// 显示可用模型
	showAvailableModels(chatService)

	// 测试自动选择模型
	testAutoSelection(chatService)

	// 测试指定模型
	testSpecificModel(chatService, "glm")
	testSpecificModel(chatService, "doubao")

	// 测试不存在的模型
	testSpecificModel(chatService, "non-existent-model")

	// 演示配置热更新
	demonstrateHotReload(configCenter, chatService)
}

func showAvailableModels(chatService *service.ChatService) {
	fmt.Println("\n📋 可用模型列表:")
	fmt.Println("----------------")
	models := chatService.GetAvailableModels()
	if len(models) == 0 {
		fmt.Println("❌ 没有可用的模型")
		return
	}

	for _, model := range models {
		status := "❌ 禁用"
		if model.IsAvailable() {
			status = "✅ 启用"
		}
		fmt.Printf("%s %s (优先级: %d, 权重: %d)\n",
			status, model.GetName(), model.GetPriority(), model.GetWeight())
	}
}

func testAutoSelection(chatService *service.ChatService) {
	fmt.Println("\n🎯 测试自动选择模型:")
	fmt.Println("-------------------")

	ctx := context.Background()
	messages := []types.Message{
		{
			Role:    "user",
			Content: "你好，请介绍一下自己",
		},
	}

	response, err := chatService.Chat(ctx, "", messages)
	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
	} else {
		fmt.Printf("✅ 选择的模型: %s\n", response.Model)
		fmt.Printf("💬 回复: %s\n", response.Content)
	}
}

func testSpecificModel(chatService *service.ChatService, modelName string) {
	fmt.Printf("\n🎯 测试指定模型 '%s':\n", modelName)
	fmt.Println("-------------------")

	ctx := context.Background()
	messages := []types.Message{
		{
			Role:    "user",
			Content: "请简单介绍一下你的功能",
		},
	}

	response, err := chatService.Chat(ctx, modelName, messages)
	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
	} else {
		fmt.Printf("✅ 模型: %s\n", response.Model)
		fmt.Printf("💬 回复: %s\n", response.Content)
	}
}

func demonstrateHotReload(configCenter *config.FileConfigCenter, chatService *service.ChatService) {
	fmt.Println("\n🔄 配置热更新演示:")
	fmt.Println("-------------------")
	fmt.Println("💡 提示: 你可以编辑 configs/chat-models.json 文件来测试热更新功能")
	fmt.Println("💡 例如: 修改 enabled 字段、priority 或 weight 值")
	fmt.Println("💡 按 Ctrl+C 退出程序")

	// 监听配置变化
	configCenter.WatchConfigChanges(func(configs []types.ModelConfig) {
		fmt.Printf("\n🔄 配置已更新 (%s)\n", time.Now().Format("15:04:05"))
		fmt.Printf("📊 总模型数: %d\n", len(configs))
		for _, config := range configs {
			status := "❌ 禁用"
			if config.Enabled {
				status = "✅ 启用"
			}
			fmt.Printf("  %s %s (优先级: %d, 权重: %d)\n",
				status, config.Name, config.Priority, config.Weight)
		}
	})

	// 保持程序运行
	select {}
}
