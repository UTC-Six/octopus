package main

import (
	"context"
	"log"
	"time"

	"github.com/UTC-Six/octopus/internal/config"
	"github.com/UTC-Six/octopus/internal/router"
	"github.com/UTC-Six/octopus/internal/service"
	"github.com/UTC-Six/octopus/internal/types"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
)

func main() {
	logx.Info("🚀 大模型路由系统")
	logx.Info("==================")

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

	// 等待配置加载完成
	time.Sleep(1 * time.Second)

	// 显示可用模型
	showAvailableModels(chatService)

	// 演示自动选择模型
	//demonstrateAutoSelection(chatService)

	// 演示指定模型
	demonstrateSpecificModel(chatService)
}

func showAvailableModels(chatService *service.ChatService) {
	logx.Info("\n📋 可用模型列表:")
	logx.Info("----------------")
	models := chatService.GetAvailableModels()
	if len(models) == 0 {
		logx.Error("❌ 没有可用的模型")
		return
	}

	for _, model := range models {
		status := "❌ 禁用"
		if model.IsAvailable() {
			status = "✅ 启用"
		}
		logx.Infof("%s %s (优先级: %d, 权重: %d)",
			status, model.GetName(), model.GetPriority(), model.GetWeight())
	}
}

func demonstrateAutoSelection(chatService *service.ChatService) {
	logx.Info("\n🎯 自动选择模型演示:")
	logx.Info("-------------------")

	ctx := context.Background()
	messages := []*schema.Message{
		{
			Role:    "user",
			Content: "你好，请介绍一下自己",
		},
	}

	response, err := chatService.Chat(ctx, "glm-4.5", messages)
	if err != nil {
		logx.Errorf("❌ 错误: %v", err)
	} else {
		logx.Infof("✅ 选择的模型: %s", response.Model)
		logx.Infof("💬 回复: %s", response.Content)
	}
}

func demonstrateSpecificModel(chatService *service.ChatService) {
	logx.Info("\n🎯 指定模型演示:")
	logx.Info("---------------")

	// 获取第一个可用模型进行演示
	models := chatService.GetAvailableModels()
	if len(models) == 0 {
		logx.Error("❌ 没有可用的模型")
		return
	}

	modelName := models[0].GetName()
	logx.Infof("使用模型: %s", modelName)

	ctx := context.Background()
	messages := []*schema.Message{
		{
			Role:    "user",
			Content: "请全面的介绍一下你自己",
		},
	}

	response, err := chatService.Chat(ctx, modelName, messages)
	if err != nil {
		logx.Errorf("❌ 错误: %v", err)
	} else {
		logx.Infof("✅ 模型: %s", response.Model)
		logx.Infof("💬 回复: %s", response.Content)
	}
}

// createNacosConfigCenter 创建 Nacos 配置中心
func createNacosConfigCenter() (types.ConfigCenter, error) {
	return config.NewNacosConfigCenter()
}
