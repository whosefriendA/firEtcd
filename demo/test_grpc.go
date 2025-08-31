package main

import (
	"context"
	"fmt"
	"time"

	"github.com/whosefriendA/firEtcd/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	fmt.Println("🔍 测试 gRPC 连接...")

	// 测试不安全连接
	fmt.Println("\n📋 测试不安全连接...")

	addresses := []string{
		"127.0.0.1:51240",
		"127.0.0.1:51241",
		"127.0.0.1:51242",
	}

	for i, addr := range addresses {
		fmt.Printf("\n🔗 测试地址 %d: %s\n", i+1, addr)

		// 尝试不安全连接
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Printf("❌ 不安全连接失败: %v\n", err)
			continue
		}
		defer conn.Close()

		fmt.Printf("✅ 不安全连接成功\n")

		// 尝试调用服务
		client := pb.NewKvserverClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 尝试一个简单的 Get 请求
		args := &pb.GetArgs{
			Key:          "test-key",
			ClientId:     12345,
			LatestOffset: 1,
			Op:           pb.OpType_GetT,
		}

		reply, err := client.Get(ctx, args)
		if err != nil {
			fmt.Printf("❌ gRPC 调用失败: %v\n", err)
		} else {
			fmt.Printf("✅ gRPC 调用成功: %+v\n", reply)
		}
	}

	fmt.Println("\n🏁 测试完成")
}
