package router

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"

	"github.com/UTC-Six/octopus/internal/models"
	"github.com/UTC-Six/octopus/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// DefaultModelRouter 默认大模型路由器
type DefaultModelRouter struct {
	// models 保存当前“生效”的模型实例集合。通过互斥锁保护，
	// UpdateConfigs 采用“整体替换”的方式原子更新，避免部分更新导致的不一致。
	models       map[string]types.ChatModel // modelName -> ChatModel
	configCenter types.ConfigCenter
	mu           sync.RWMutex
	stopCh       chan struct{}
}

// NewDefaultModelRouter 创建新的模型路由器
func NewDefaultModelRouter(configCenter types.ConfigCenter) *DefaultModelRouter {
	return &DefaultModelRouter{
		models:       make(map[string]types.ChatModel),
		configCenter: configCenter,
		stopCh:       make(chan struct{}),
	}
}

// Start 启动路由器
func (r *DefaultModelRouter) Start() error {
	// 初始化配置：从配置中心拉取一次完整模型配置，构建内存路由表。
	configs, err := r.configCenter.GetModelConfigs()
	if err != nil {
		return fmt.Errorf("failed to get initial configs: %w", err)
	}

	if err := r.UpdateConfigs(configs); err != nil {
		return fmt.Errorf("failed to update initial configs: %w", err)
	}

	// 监听配置变化：收到变更后调用 onConfigChange。
	// onConfigChange 会先尝试构建新的模型集合并整体替换；若失败则保持旧配置不变。
	if err := r.configCenter.WatchConfigChanges(r.onConfigChange); err != nil {
		return fmt.Errorf("failed to watch config changes: %w", err)
	}

	logx.Info("Model router started successfully")
	return nil
}

// Stop 停止路由器
func (r *DefaultModelRouter) Stop() error {
	close(r.stopCh)
	return r.configCenter.Close()
}

// onConfigChange 配置变化回调
func (r *DefaultModelRouter) onConfigChange(configs []types.ModelConfig) {
	logx.Infof("Config changed, updating %d models", len(configs))
	if err := r.UpdateConfigs(configs); err != nil {
		logx.Errorf("Failed to update configs: %v", err)
	}
}

// UpdateConfigs 更新模型配置
func (r *DefaultModelRouter) UpdateConfigs(configs []types.ModelConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 用新的临时表承接并验证配置，成功后整体替换，失败则保持旧配置
	newModels := make(map[string]types.ChatModel)

	// 按优先级排序（优先级越小越高）
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Priority < configs[j].Priority
	})

	// 创建新的模型实例
	for _, config := range configs {
		if !config.Enabled {
			continue
		}

		model := models.NewDefaultChatModel(config)
		newModels[config.Name] = model
		logx.Infof("Registered model: %s (priority: %d, weight: %d)",
			config.Name, config.Priority, config.Weight)
	}

	// 一次性替换
	r.models = newModels
	return nil
}

// GetModel 根据模型名称获取模型
func (r *DefaultModelRouter) GetModel(modelName string) (types.ChatModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	model, exists := r.models[modelName]
	if !exists {
		return nil, fmt.Errorf("model %s not found", modelName)
	}

	if !model.IsAvailable() {
		return nil, fmt.Errorf("model %s is not available", modelName)
	}

	return model, nil
}

// GetAvailableModel 根据优先级和权重获取可用模型
func (r *DefaultModelRouter) GetAvailableModel() (types.ChatModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.models) == 0 {
		return nil, fmt.Errorf("no models available")
	}

	// 获取所有可用模型
	var availableModels []types.ChatModel
	for _, model := range r.models {
		if model.IsAvailable() {
			availableModels = append(availableModels, model)
		}
	}

	if len(availableModels) == 0 {
		return nil, fmt.Errorf("no available models")
	}

	// 按优先级分组
	priorityGroups := make(map[int][]types.ChatModel)
	for _, model := range availableModels {
		priority := model.GetPriority()
		priorityGroups[priority] = append(priorityGroups[priority], model)
	}

	// 获取最高优先级（最小数值）
	var minPriority int
	for priority := range priorityGroups {
		if minPriority == 0 || priority < minPriority {
			minPriority = priority
		}
	}

	// 在最高优先级组中按权重选择
	models := priorityGroups[minPriority]
	return r.selectByWeight(models), nil
}

// selectByWeight 根据权重选择模型
func (r *DefaultModelRouter) selectByWeight(models []types.ChatModel) types.ChatModel {
	if len(models) == 1 {
		return models[0]
	}

	// 计算总权重
	totalWeight := 0
	for _, model := range models {
		totalWeight += model.GetWeight()
	}

	if totalWeight == 0 {
		// 如果所有权重都是0，随机选择
		return models[rand.Intn(len(models))]
	}

	// 按权重随机选择
	random := rand.Intn(totalWeight)
	currentWeight := 0

	for _, model := range models {
		currentWeight += model.GetWeight()
		if random < currentWeight {
			return model
		}
	}

	// 兜底：返回第一个模型
	return models[0]
}

// GetAllModels 获取所有模型
func (r *DefaultModelRouter) GetAllModels() []types.ChatModel {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []types.ChatModel
	for _, model := range r.models {
		models = append(models, model)
	}

	return models
}
