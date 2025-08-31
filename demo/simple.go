package main

import (
	"fmt"
	"log"
	"time"

	"github.com/whosefriendA/firEtcd/client"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
)

func main() {
	fmt.Println("🔍 简单测试 etcd 基本功能...")

	// 创建客户端配置
	conf := firconfig.Clerk{
		EtcdAddrs: []string{
			"127.0.0.1:51240",
		},
		TLS: &firconfig.TLSConfig{
			CertFile: "/home/wanggang/firEtcd/pkg/tls/certs/client.crt",
			KeyFile:  "/home/wanggang/firEtcd/pkg/tls/certs/client.key",
			CAFile:   "/home/wanggang/firEtcd/pkg/tls/certs/ca.crt",
		},
	}

	fmt.Printf("📋 配置: %+v\n", conf)

	// 创建客户端
	ck := client.MakeClerk(conf)
	if ck == nil {
		log.Fatal("❌ 创建客户端失败")
	}

	fmt.Println("✅ 客户端创建成功")

	// 等待连接建立
	fmt.Println("⏳ 等待连接建立...")
	time.Sleep(3 * time.Second)

	// 测试基本操作
	fmt.Println("🧪 测试基本操作...")

	// 1. 测试 Put
	err := ck.Put("test-key", []byte("test-value"), 0)
	if err != nil {
		fmt.Printf("❌ Put 失败: %v\n", err)
	} else {
		fmt.Printf("✅ Put 成功\n")
	}

	// 2. 测试 Get
	value, err := ck.Get("test-key")
	if err != nil {
		fmt.Printf("❌ Get 失败: %v\n", err)
	} else {
		fmt.Printf("✅ Get 成功: %s\n", value)
	}

	// 3. 测试租约
	fmt.Println("🔑 测试租约...")
	leaseID, err := ck.LeaseGrant(10 * time.Second)
	if err != nil {
		fmt.Printf("❌ 租约创建失败: %v\n", err)
	} else {
		fmt.Printf("✅ 租约创建成功，ID: %d\n", leaseID)
	}

	fmt.Println("🏁 测试完成")
}
