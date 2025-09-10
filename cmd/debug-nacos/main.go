package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func main() {
	fmt.Println("🔍 测试 Nacos 连接...")

	// 1. 测试网络连接
	fmt.Println("\n1. 测试网络连接...")
	conn, err := net.DialTimeout("tcp", "43.137.73.8:8080", 5*time.Second)
	if err != nil {
		log.Printf("❌ 无法连接到 43.137.73.8:8080: %v", err)
		return
	}
	fmt.Println("✅ 网络连接正常")
	conn.Close()

	// 2. 测试 Nacos 客户端
	fmt.Println("\n2. 测试 Nacos 客户端...")
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:   "82.157.212.7",
			Port:     8848,
			GrpcPort: 9848,
		},
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         "18b01bb8-ac70-4527-89d0-01079f30ca09",
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}

	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		log.Fatalf("❌ 创建 Nacos 客户端失败: %v", err)
	}
	fmt.Println("✅ Nacos 客户端创建成功")

	// 3. 测试获取配置
	fmt.Println("\n3. 测试获取配置...")
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: "chat-models",
		Group:  "DEFAULT_GROUP",
	})
	if err != nil {
		log.Printf("❌ 获取配置失败: %v", err)
	} else {
		fmt.Printf("✅ 获取配置成功: %s\n", content)
	}
}
