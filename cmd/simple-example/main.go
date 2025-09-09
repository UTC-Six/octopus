package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/UTC-Six/octopus/internal/router"
	"github.com/UTC-Six/octopus/internal/service"
	"github.com/UTC-Six/octopus/internal/types"
)

func main() {
	// 创建模拟配置中心（不依赖 Nacos）
	configCenter := &MockConfigCenter{
		configs: []types.ModelConfig{
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
		},
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

	// 测试配置更新
	fmt.Println("\n=== Test Config Update ===")
	testConfigUpdate(configCenter, modelRouter)
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

func testConfigUpdate(configCenter *MockConfigCenter, modelRouter *router.DefaultModelRouter) {
	fmt.Println("Updating config...")
	
	// 添加新模型
	newConfigs := append(configCenter.configs, types.ModelConfig{
		Name:     "new-model",
		URL:      "https://api.newmodel.com",
		AppKey:   "sk-new-key",
		Priority: 1,
		Weight:   10,
		Enabled:  true,
	})
	
	configCenter.configs = newConfigs
	
	// 更新路由器配置
	if err := modelRouter.UpdateConfigs(newConfigs); err != nil {
		fmt.Printf("Failed to update config: %v\n", err)
		return
	}
	
	fmt.Println("Config updated successfully!")
	
	// 显示更新后的模型列表
	models := modelRouter.GetAllModels()
	fmt.Printf("Total models: %d\n", len(models))
	for _, model := range models {
		status := "❌ Disabled"
		if model.IsAvailable() {
			status = "✅ Enabled"
		}
		fmt.Printf("  %s %s (priority: %d, weight: %d)\n",
			status, model.GetName(), model.GetPriority(), model.GetWeight())
	}
}

// MockConfigCenter 模拟配置中心
type MockConfigCenter struct {
	configs []types.ModelConfig
}

func (m *MockConfigCenter) GetModelConfigs() ([]types.ModelConfig, error) {
	return m.configs, nil
}

func (m *MockConfigCenter) WatchConfigChanges(callback func([]types.ModelConfig)) error {
	// 模拟配置中心，不实现监听
	return nil
}

func (m *MockConfigCenter) Close() error {
	return nil
}
