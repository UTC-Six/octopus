package service

import (
	"context"

	"github.com/UTC-Six/octopus/internal/router"
	"github.com/UTC-Six/octopus/internal/types"
	"github.com/cloudwego/eino/schema"
)

// ChatService 聊天服务
type ChatService struct {
	router *router.DefaultModelRouter
}

// NewChatService 创建新的聊天服务
func NewChatService(router *router.DefaultModelRouter) *ChatService {
	return &ChatService{
		router: router,
	}
}

// Chat 发送聊天请求
func (s *ChatService) Chat(ctx context.Context, modelName string, messages []*schema.Message) (*types.ChatResponse, error) {
	var model types.ChatModel
	var err error

	// 根据是否指定模型名称选择模型
	if modelName != "" {
		model, err = s.router.GetModel(modelName)
		if err != nil {
			return &types.ChatResponse{
				Error: err.Error(),
			}, err
		}
	} else {
		model, err = s.router.GetAvailableModel()
		if err != nil {
			return &types.ChatResponse{
				Error: err.Error(),
			}, err
		}
	}

	// 调用模型进行聊天
	response, err := model.Chat(ctx, messages)
	if err != nil {
		return &types.ChatResponse{
			Error: err.Error(),
		}, err
	}

	return response, nil
}

// GetAvailableModels 获取可用模型列表
func (s *ChatService) GetAvailableModels() []types.ChatModel {
	return s.router.GetAllModels()
}

// GetModelInfo 获取模型信息
func (s *ChatService) GetModelInfo(modelName string) (*types.ModelConfig, error) {
	model, err := s.router.GetModel(modelName)
	if err != nil {
		return nil, err
	}

	return &types.ModelConfig{
		Name:     model.GetName(),
		URL:      model.GetURL(),
		AppKey:   model.GetAppKey(),
		Priority: model.GetPriority(),
		Weight:   model.GetWeight(),
		Enabled:  model.IsAvailable(),
	}, nil
}
