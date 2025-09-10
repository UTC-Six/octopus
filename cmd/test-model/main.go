package main

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()
	// 初始化模型 (以openai为例)
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: "https://open.bigmodel.cn/api/paas/v4",
		APIKey:  "8ebb6a00993c4a7a99be11990af9858c.QWTswoyOF7AFn52B",
		Model:   "glm-4.5",
	})
	if err != nil {
		panic(err)
	}

	// 准备输入消息
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: "帮我看看今日 A 股投资指南, 我想做短线交易，今日尾盘入手，明天早盘出掉。",
		},
		{
			Role:    schema.User,
			Content: "今天的市场情况如何？",
		},
	}

	// 生成响应
	//response, err := cm.Generate(ctx, messages, model.WithTemperature(0.8))

	// 响应处理
	//fmt.Println(response.Content)

	// 流式生成
	streamResult, err := cm.Stream(ctx, messages)

	defer streamResult.Close()

	for {
		chunk, err := streamResult.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			// 错误处理
		}
		// 响应片段处理
		fmt.Print(chunk.Content)
	}
}
