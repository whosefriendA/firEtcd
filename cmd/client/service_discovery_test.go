package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/whosefriendA/firEtcd/client"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
)

func TestServiceDiscoveryV3(t *testing.T) {
	// 创建客户端
	conf := firconfig.Clerk{}
	firconfig.Init("config.yml", &conf)
	ck := client.MakeClerk(conf)

	// 创建服务注册器和发现器
	registry := client.NewServiceRegistryV3(ck)
	discovery := client.NewServiceDiscoveryV3(ck, ck)

	// 测试服务注册
	serviceName := "test-service"
	serviceID := "instance-1"
	endpoint := "localhost:8080"
	metadata := map[string]string{
		"version": "v1.0.0",
		"env":     "test",
	}

	// 注册服务
	leaseID, err := registry.Register(context.Background(), serviceName, serviceID, endpoint, 10*time.Second, metadata)
	if err != nil {
		t.Logf("Service registration failed (expected if no server running): %v", err)
		t.Skip("Skipping test - no server available")
	}
	t.Logf("Successfully registered service with lease ID: %d", leaseID)

	// 测试服务发现
	services, err := discovery.Get(context.Background(), serviceName)
	if err != nil {
		t.Logf("Service discovery failed: %v", err)
	} else {
		t.Logf("Found %d service instances", len(services))
		for _, service := range services {
			t.Logf("Service: %+v", service)
		}
	}

	// 测试健康检查
	health, err := discovery.HealthCheck(context.Background(), serviceName)
	if err != nil {
		t.Logf("Health check failed: %v", err)
	} else {
		t.Logf("Health status: %+v", health)
	}

	// 测试负载均衡
	lb := client.NewLoadBalancer(discovery)
	instance, err := lb.GetInstance(context.Background(), serviceName)
	if err != nil {
		t.Logf("Load balancer failed: %v", err)
	} else {
		t.Logf("Selected instance: %+v", instance)
	}

	// 等待一段时间让租约过期
	t.Logf("Waiting for lease expiration...")
	time.Sleep(12 * time.Second)

	// 再次检查服务状态
	servicesAfter, err := discovery.Get(context.Background(), serviceName)
	if err != nil {
		t.Logf("Service discovery after expiration failed: %v", err)
	} else {
		t.Logf("Services after expiration: %d", len(servicesAfter))
	}

	// 清理：注销服务
	err = registry.Deregister(context.Background(), serviceName, serviceID)
	if err != nil {
		t.Logf("Service deregistration failed: %v", err)
	} else {
		t.Logf("Successfully deregistered service")
	}
}

func TestServiceRegistryWithLease(t *testing.T) {
	// 创建客户端
	conf := firconfig.Clerk{}
	firconfig.Init("config.yml", &conf)
	ck := client.MakeClerk(conf)

	// 创建服务注册器
	registry := client.NewServiceRegistryV3(ck)

	// 测试手动租约管理
	serviceName := "test-service-lease"
	serviceID := "instance-lease"
	endpoint := "localhost:8081"
	metadata := map[string]string{
		"version": "v1.0.0",
		"env":     "test",
	}

	// 先创建租约
	leaseID, err := ck.LeaseGrant(15 * time.Second)
	if err != nil {
		t.Logf("Lease grant failed (expected if no server running): %v", err)
		t.Skip("Skipping test - no server available")
	}
	t.Logf("Granted lease with ID: %d", leaseID)

	// 使用现有租约注册服务
	err = registry.RegisterWithLease(context.Background(), serviceName, serviceID, endpoint, leaseID, metadata)
	if err != nil {
		t.Fatalf("Failed to register service with lease: %v", err)
	}
	t.Logf("Successfully registered service with existing lease")

	// 启动自动续约
	cancel := ck.AutoKeepAlive(leaseID, 5*time.Second)
	defer cancel()

	// 等待一段时间，观察续约效果
	time.Sleep(8 * time.Second)

	// 检查租约状态
	ttl, keys, err := ck.LeaseTimeToLive(leaseID, true)
	if err != nil {
		t.Logf("LeaseTimeToLive failed: %v", err)
	} else {
		t.Logf("Lease %d remaining TTL: %dms, keys: %v", leaseID, ttl, keys)
	}

	// 清理
	cancel()
	err = registry.Deregister(context.Background(), serviceName, serviceID)
	if err != nil {
		t.Logf("Service deregistration failed: %v", err)
	} else {
		t.Logf("Successfully deregistered service")
	}
}

func TestServiceWatch(t *testing.T) {
	// 创建客户端
	conf := firconfig.Clerk{}
	firconfig.Init("config.yml", &conf)
	ck := client.MakeClerk(conf)

	// 创建服务发现器
	discovery := client.NewServiceDiscoveryV3(ck, ck)

	// 启动服务监控
	serviceName := "test-watch-service"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	watchChan, err := discovery.Watch(ctx, serviceName)
	if err != nil {
		t.Logf("Service watch failed (expected if no server running): %v", err)
		t.Skip("Skipping test - no server available")
	}

	t.Logf("Started watching service: %s", serviceName)

	// 在另一个 goroutine 中处理 watch 事件
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-watchChan:
				if event != nil {
					t.Logf("Watch event: Type=%s, Key=%s, Value=%s",
						event.Type, event.Key, string(event.Value))
				}
			}
		}
	}()

	// 等待一段时间观察事件
	time.Sleep(5 * time.Second)
	t.Logf("Watch test completed")
}

func TestLegacyNamingMigration(t *testing.T) {
	// 创建客户端
	conf := firconfig.Clerk{}
	firconfig.Init("config.yml", &conf)
	ck := client.MakeClerk(conf)

	// 测试旧的命名系统（现在使用 Lease）
	node := &client.Node{
		Name:     "test-node",
		AppId:    "test-app",
		Port:     ":8080",
		IPs:      []string{"localhost"},
		Location: "test",
		Connect:  100,
		Weight:   1,
		Env:      "test",
		MetaDate: map[string]string{"version": "v1.0.0"},
	}

	// 使用新的 Lease 机制注册节点
	leaseID, cancel, err := node.SetNodeWithLease(ck, 10*time.Second)
	if err != nil {
		t.Logf("Node registration with lease failed (expected if no server running): %v", err)
		t.Skip("Skipping test - no server available")
	}
	t.Logf("Successfully registered node with lease ID: %d", leaseID)

	// 获取节点列表
	nodes, err := client.GetNode(ck, "test-app", "test")
	if err != nil {
		t.Logf("GetNode failed: %v", err)
	} else {
		t.Logf("Found %d nodes", len(nodes))
		for _, n := range nodes {
			t.Logf("Node: %+v", n)
		}
	}

	// 清理
	cancel()
	t.Logf("Node registration cancelled")
}
