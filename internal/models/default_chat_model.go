package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/UTC-Six/octopus/internal/types"
)

// DefaultChatModel 默认大模型实现
type DefaultChatModel struct {
	name     string
	url      string
	appKey   string
	priority int
	weight   int
	enabled  bool
	client   *http.Client
}

// NewDefaultChatModel 创建新的默认大模型
func NewDefaultChatModel(config types.ModelConfig) *DefaultChatModel {
	return &DefaultChatModel{
		name:     config.Name,
		url:      config.URL,
		appKey:   config.AppKey,
		priority: config.Priority,
		weight:   config.Weight,
		enabled:  config.Enabled,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName 获取模型名称
func (m *DefaultChatModel) GetName() string {
	return m.name
}

// GetURL 获取模型URL
func (m *DefaultChatModel) GetURL() string {
	return m.url
}

// GetAppKey 获取应用密钥
func (m *DefaultChatModel) GetAppKey() string {
	return m.appKey
}

// GetPriority 获取优先级
func (m *DefaultChatModel) GetPriority() int {
	return m.priority
}

// GetWeight 获取权重
func (m *DefaultChatModel) GetWeight() int {
	return m.weight
}

// IsAvailable 检查模型是否可用
func (m *DefaultChatModel) IsAvailable() bool {
	return m.enabled
	/*if !m.enabled {
		return false
	}

	// 简单的健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", m.url+"/health", nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+m.appKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK*/
}

// Chat 发送聊天请求
func (m *DefaultChatModel) Chat(ctx context.Context, messages []types.Message) (*types.ChatResponse, error) {
	if !m.IsAvailable() {
		return &types.ChatResponse{
			Error: "model is not available",
		}, fmt.Errorf("model %s is not available", m.name)
	}

	// 构建请求体
	requestBody := map[string]interface{}{
		"model":    m.name,
		"messages": messages,
		"stream":   false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return &types.ChatResponse{
			Error: "failed to marshal request",
		}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", m.url+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return &types.ChatResponse{
			Error: "failed to create request",
		}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.appKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := m.client.Do(req)
	if err != nil {
		return &types.ChatResponse{
			Error: "failed to send request",
		}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &types.ChatResponse{
			Error: "failed to read response",
		}, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &types.ChatResponse{
			Error: fmt.Sprintf("API returned status %d", resp.StatusCode),
		}, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return &types.ChatResponse{
			Error: "failed to parse response",
		}, fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Error.Message != "" {
		return &types.ChatResponse{
			Error: response.Error.Message,
		}, fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return &types.ChatResponse{
			Error: "no choices in response",
		}, fmt.Errorf("no choices in response")
	}

	return &types.ChatResponse{
		Content:   response.Choices[0].Message.Content,
		Model:     m.name,
		Timestamp: time.Now(),
	}, nil
}
