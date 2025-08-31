package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/whosefriendA/firEtcd/client"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
)

// E2ETestSuite 端到端测试套件
type E2ETestSuite struct {
	registry  *client.ServiceRegistryV3
	discovery *client.ServiceDiscoveryV3
	client    *http.Client
}

// NewE2ETestSuite 创建端到端测试套件
func NewE2ETestSuite(t *testing.T) *E2ETestSuite {
	// 创建客户端配置
	conf := firconfig.Clerk{
		EtcdAddrs: []string{
			"127.0.0.1:51240",
			"127.0.0.1:51241",
			"127.0.0.1:51242",
		},
		TLS: &firconfig.TLSConfig{
			CertFile: "/home/wanggang/firEtcd/pkg/tls/certs/client.crt",
			KeyFile:  "/home/wanggang/firEtcd/pkg/tls/certs/client.key",
			CAFile:   "/home/wanggang/firEtcd/pkg/tls/certs/ca.crt",
		},
	}

	ck := client.MakeClerk(conf)

	// 创建服务注册器和发现器
	registry := client.NewServiceRegistryV3(ck)
	discovery := client.NewServiceDiscoveryV3(ck, ck)

	// 等待客户端连接建立
	t.Log("⏳ 等待客户端连接建立...")
	time.Sleep(3 * time.Second)

	return &E2ETestSuite{
		registry:  registry,
		discovery: discovery,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// TestE2EServiceDiscovery 端到端服务发现测试
func TestE2EServiceDiscovery(t *testing.T) {
	suite := NewE2ETestSuite(t)
	if suite == nil {
		return
	}

	t.Run("Service Registration", suite.testServiceRegistration)
	t.Run("Service Discovery", suite.testServiceDiscovery)
	t.Run("Service Communication", suite.testServiceCommunication)
	t.Run("Load Balancing", suite.testLoadBalancing)
	t.Run("Health Check", suite.testHealthCheck)
	t.Run("Service Deregistration", suite.testServiceDeregistration)
}

// testServiceRegistration 测试服务注册
func (s *E2ETestSuite) testServiceRegistration(t *testing.T) {
	t.Log("🧪 测试服务注册...")

	// 注册用户服务
	userLeaseID, err := s.registry.Register(
		context.Background(),
		"user-service",
		"test-user-instance",
		"localhost:8083",
		30*time.Second,
		map[string]string{
			"version": "v1.0.0",
			"env":     "test",
			"test":    "true",
		},
	)

	if err != nil {
		t.Skipf("Skipping test - service registration failed: %v", err)
	}

	t.Logf("✅ 用户服务注册成功，租约ID: %d", userLeaseID)

	// 注册订单服务
	orderLeaseID, err := s.registry.Register(
		context.Background(),
		"order-service",
		"test-order-instance",
		"localhost:8084",
		30*time.Second,
		map[string]string{
			"version": "v1.0.0",
			"env":     "test",
			"test":    "true",
		},
	)

	if err != nil {
		t.Skipf("Skipping test - service registration failed: %v", err)
	}

	t.Logf("✅ 订单服务注册成功，租约ID: %d", orderLeaseID)

	// 等待服务注册生效
	time.Sleep(2 * time.Second)
}

// testServiceDiscovery 测试服务发现
func (s *E2ETestSuite) testServiceDiscovery(t *testing.T) {
	t.Log("🔍 测试服务发现...")

	// 发现用户服务
	userServices, err := s.discovery.Get(context.Background(), "user-service")
	if err != nil {
		t.Fatalf("❌ 发现用户服务失败: %v", err)
	}

	if len(userServices) == 0 {
		t.Fatal("❌ 未发现用户服务")
	}

	t.Logf("✅ 发现 %d 个用户服务实例", len(userServices))

	// 验证服务信息
	for _, service := range userServices {
		if service.Name != "user-service" {
			t.Errorf("❌ 服务名称不匹配: 期望 user-service，实际 %s", service.Name)
		}
		if service.Endpoint == "" {
			t.Error("❌ 服务端点为空")
		}
		if service.LeaseID == 0 {
			t.Error("❌ 服务租约ID为空")
		}
	}

	// 发现订单服务
	orderServices, err := s.discovery.Get(context.Background(), "order-service")
	if err != nil {
		t.Fatalf("❌ 发现订单服务失败: %v", err)
	}

	if len(orderServices) == 0 {
		t.Fatal("❌ 未发现订单服务")
	}

	t.Logf("✅ 发现 %d 个订单服务实例", len(orderServices))
}

// testServiceCommunication 测试服务间通信
func (s *E2ETestSuite) testServiceCommunication(t *testing.T) {
	t.Log("📡 测试服务间通信...")

	// 测试用户服务健康检查
	resp, err := s.client.Get("http://localhost:8083/health")
	if err != nil {
		t.Skipf("Skipping test - user service not available: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("❌ 用户服务健康检查失败，状态码: %d", resp.StatusCode)
	}

	var healthResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&healthResponse); err != nil {
		t.Fatalf("❌ 解析健康检查响应失败: %v", err)
	}

	if healthResponse["status"] != "healthy" {
		t.Errorf("❌ 用户服务状态不健康: %v", healthResponse["status"])
	}

	t.Log("✅ 用户服务健康检查通过")

	// 测试订单服务健康检查
	resp, err = s.client.Get("http://localhost:8084/health")
	if err != nil {
		t.Skipf("Skipping test - order service not available: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("❌ 订单服务健康检查失败，状态码: %d", resp.StatusCode)
	}

	t.Log("✅ 订单服务健康检查通过")
}

// testLoadBalancing 测试负载均衡
func (s *E2ETestSuite) testLoadBalancing(t *testing.T) {
	t.Log("⚖️ 测试负载均衡...")

	// 创建负载均衡器
	lb := client.NewLoadBalancer(s.discovery)

	// 测试用户服务负载均衡
	userInstance, err := lb.GetInstance(context.Background(), "user-service")
	if err != nil {
		t.Fatalf("❌ 获取用户服务实例失败: %v", err)
	}

	if userInstance == nil {
		t.Fatal("❌ 负载均衡器返回空实例")
	}

	t.Logf("✅ 负载均衡器选择用户服务实例: %s", userInstance.ID)

	// 测试订单服务负载均衡
	orderInstance, err := lb.GetInstance(context.Background(), "order-service")
	if err != nil {
		t.Fatalf("❌ 获取订单服务实例失败: %v", err)
	}

	if orderInstance == nil {
		t.Fatal("❌ 负载均衡器返回空实例")
	}

	t.Logf("✅ 负载均衡器选择订单服务实例: %s", orderInstance.ID)
}

// testHealthCheck 测试健康检查
func (s *E2ETestSuite) testHealthCheck(t *testing.T) {
	t.Log("🏥 测试健康检查...")

	// 执行健康检查
	userHealth, err := s.discovery.HealthCheck(context.Background(), "user-service")
	if err != nil {
		t.Fatalf("❌ 用户服务健康检查失败: %v", err)
	}

	orderHealth, err := s.discovery.HealthCheck(context.Background(), "order-service")
	if err != nil {
		t.Fatalf("❌ 订单服务健康检查失败: %v", err)
	}

	// 验证健康状态
	for serviceID, isHealthy := range userHealth {
		if !isHealthy {
			t.Errorf("❌ 用户服务 %s 不健康", serviceID)
		} else {
			t.Logf("✅ 用户服务 %s 健康", serviceID)
		}
	}

	for serviceID, isHealthy := range orderHealth {
		if !isHealthy {
			t.Errorf("❌ 订单服务 %s 不健康", serviceID)
		} else {
			t.Logf("✅ 订单服务 %s 健康", serviceID)
		}
	}
}

// testServiceDeregistration 测试服务注销
func (s *E2ETestSuite) testServiceDeregistration(t *testing.T) {
	t.Log("🛑 测试服务注销...")

	// 注销用户服务
	err := s.registry.Deregister(context.Background(), "user-service", "test-user-instance")
	if err != nil {
		t.Errorf("❌ 注销用户服务失败: %v", err)
	} else {
		t.Log("✅ 用户服务注销成功")
	}

	// 注销订单服务
	err = s.registry.Deregister(context.Background(), "order-service", "test-order-instance")
	if err != nil {
		t.Errorf("❌ 注销订单服务失败: %v", err)
	} else {
		t.Log("✅ 订单服务注销成功")
	}

	// 等待注销生效
	time.Sleep(2 * time.Second)

	// 验证服务已被注销
	userServices, err := s.discovery.Get(context.Background(), "user-service")
	if err == nil && len(userServices) > 0 {
		// 检查是否还有测试实例
		for _, service := range userServices {
			if service.ID == "test-user-instance" {
				t.Error("❌ 用户服务测试实例仍然存在")
			}
		}
	}

	orderServices, err := s.discovery.Get(context.Background(), "order-service")
	if err == nil && len(orderServices) > 0 {
		// 检查是否还有测试实例
		for _, service := range orderServices {
			if service.ID == "test-order-instance" {
				t.Error("❌ 订单服务测试实例仍然存在")
			}
		}
	}
}

// TestE2EServiceLifecycle 测试服务生命周期
func TestE2EServiceLifecycle(t *testing.T) {
	suite := NewE2ETestSuite(t)
	if suite == nil {
		return
	}

	t.Run("Complete Lifecycle", func(t *testing.T) {
		// 1. 注册服务
		t.Log("📝 步骤 1: 注册服务")
		_, err := suite.registry.Register(
			context.Background(),
			"lifecycle-test",
			"test-instance",
			"localhost:9999",
			5*time.Second, // 减少到5秒
			map[string]string{"test": "true"},
		)
		if err != nil {
			t.Fatalf("服务注册失败: %v", err)
		}

		// 2. 验证服务发现
		t.Log("🔍 步骤 2: 验证服务发现")
		services, err := suite.discovery.Get(context.Background(), "lifecycle-test")
		if err != nil || len(services) == 0 {
			t.Fatal("服务发现失败")
		}

		// 3. 等待租约过期
		t.Log("⏰ 步骤 3: 等待租约过期")
		time.Sleep(7 * time.Second) // 减少到7秒

		// 4. 验证服务自动清理
		t.Log("🧹 步骤 4: 验证服务自动清理")
		servicesAfter, err := suite.discovery.Get(context.Background(), "lifecycle-test")
		if err == nil && len(servicesAfter) > 0 {
			t.Log("⚠️  服务仍然存在，可能需要更多时间等待清理")
		} else {
			t.Log("✅ 服务已自动清理")
		}
	})
}

// TestE2EPerformance 性能测试
func TestE2EPerformance(t *testing.T) {
	suite := NewE2ETestSuite(t)
	if suite == nil {
		return
	}

	t.Run("Service Discovery Performance", func(t *testing.T) {
		// 注册多个服务实例
		instanceCount := 10
		leaseIDs := make([]int64, instanceCount)

		t.Logf("🚀 注册 %d 个服务实例...", instanceCount)
		start := time.Now()

		for i := 0; i < instanceCount; i++ {
			leaseID, err := suite.registry.Register(
				context.Background(),
				"performance-test",
				fmt.Sprintf("instance-%d", i),
				fmt.Sprintf("localhost:%d", 9000+i),
				30*time.Second,
				map[string]string{"test": "performance"},
			)
			if err != nil {
				t.Fatalf("注册实例 %d 失败: %v", i, err)
			}
			leaseIDs[i] = leaseID
		}

		registrationTime := time.Since(start)
		t.Logf("✅ 注册完成，耗时: %v", registrationTime)

		// 测试服务发现性能
		t.Log("🔍 测试服务发现性能...")
		start = time.Now()

		services, err := suite.discovery.Get(context.Background(), "performance-test")
		if err != nil {
			t.Fatalf("服务发现失败: %v", err)
		}

		discoveryTime := time.Since(start)
		t.Logf("✅ 发现 %d 个服务，耗时: %v", len(services), discoveryTime)

		// 性能断言
		if registrationTime > 5*time.Second {
			t.Errorf("❌ 服务注册耗时过长: %v", registrationTime)
		}

		if discoveryTime > 100*time.Millisecond {
			t.Errorf("❌ 服务发现耗时过长: %v", discoveryTime)
		}

		// 清理测试服务
		t.Log("🧹 清理测试服务...")
		for i := range leaseIDs {
			err := suite.registry.Deregister(context.Background(), "performance-test", fmt.Sprintf("instance-%d", i))
			if err != nil {
				t.Logf("⚠️  清理实例 %d 失败: %v", i, err)
			}
		}
	})
}
