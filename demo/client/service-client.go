package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/whosefriendA/firEtcd/client"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
)

// ServiceClient 服务客户端示例
type ServiceClient struct {
	discovery  *client.ServiceDiscoveryV3
	httpClient *http.Client
}

// NewServiceClient 创建服务客户端
func NewServiceClient() (*ServiceClient, error) {
	// 创建 firEtcd 客户端配置
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

	// 创建服务发现器
	discovery := client.NewServiceDiscoveryV3(ck, ck)

	return &ServiceClient{
		discovery:  discovery,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// DiscoverAndCallUserService 发现并调用用户服务
func (sc *ServiceClient) DiscoverAndCallUserService() error {
	log.Println("🔍 发现用户服务...")

	// 发现用户服务
	services, err := sc.discovery.Get(context.Background(), "user-service")
	if err != nil {
		return fmt.Errorf("发现用户服务失败: %v", err)
	}

	if len(services) == 0 {
		return fmt.Errorf("未发现可用的用户服务")
	}

	log.Printf("✅ 发现 %d 个用户服务实例", len(services))

	// 选择第一个可用的服务实例
	selectedService := services[0]
	log.Printf("🎯 选择服务实例: %s at %s", selectedService.ID, selectedService.Endpoint)

	// 调用用户服务
	return sc.callUserService(selectedService)
}

// callUserService 调用用户服务
func (sc *ServiceClient) callUserService(service *client.ServiceInfo) error {
	// 获取用户列表
	url := fmt.Sprintf("http://%s/users", service.Endpoint)
	log.Printf("📡 调用用户服务: %s", url)

	resp, err := sc.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("调用用户服务失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("用户服务返回错误状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	log.Printf("✅ 用户服务调用成功，返回 %d 个用户", response["count"])
	return nil
}

// DiscoverAndCallOrderService 发现并调用订单服务
func (sc *ServiceClient) DiscoverAndCallOrderService() error {
	log.Println("🔍 发现订单服务...")

	// 发现订单服务
	services, err := sc.discovery.Get(context.Background(), "order-service")
	if err != nil {
		return fmt.Errorf("发现订单服务失败: %v", err)
	}

	if len(services) == 0 {
		return fmt.Errorf("未发现可用的订单服务")
	}

	log.Printf("✅ 发现 %d 个订单服务实例", len(services))

	// 选择第一个可用的服务实例
	selectedService := services[0]
	log.Printf("🎯 选择服务实例: %s at %s", selectedService.ID, selectedService.Endpoint)

	// 调用订单服务
	return sc.callOrderService(selectedService)
}

// callOrderService 调用订单服务
func (sc *ServiceClient) callOrderService(service *client.ServiceInfo) error {
	// 获取订单列表
	url := fmt.Sprintf("http://%s/orders", service.Endpoint)
	log.Printf("📡 调用订单服务: %s", url)

	resp, err := sc.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("调用订单服务失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("订单服务返回错误状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	log.Printf("✅ 订单服务调用成功，返回 %d 个订单", response["count"])
	return nil
}

// MonitorServices 监控服务变化
func (sc *ServiceClient) MonitorServices() error {
	log.Println("👀 开始监控服务变化...")

	// 监控用户服务
	userWatchChan, err := sc.discovery.Watch(context.Background(), "user-service")
	if err != nil {
		return fmt.Errorf("监控用户服务失败: %v", err)
	}

	// 监控订单服务
	orderWatchChan, err := sc.discovery.Watch(context.Background(), "order-service")
	if err != nil {
		return fmt.Errorf("监控订单服务失败: %v", err)
	}

	// 启动监控协程
	go sc.watchUserServices(userWatchChan)
	go sc.watchOrderServices(orderWatchChan)

	// 监控一段时间
	time.Sleep(30 * time.Second)
	log.Println("⏰ 监控时间结束")

	return nil
}

// watchUserServices 监控用户服务变化
func (sc *ServiceClient) watchUserServices(watchChan <-chan *client.WatchEvent) {
	for event := range watchChan {
		if event == nil {
			continue
		}

		switch event.Type {
		case 1: // PUT
			log.Printf("🆕 用户服务新增/更新: %s", string(event.Key))
		case 2: // DELETE
			log.Printf("🗑️  用户服务删除: %s", string(event.Key))
		default:
			log.Printf("📝 用户服务变化: %d - %s", event.Type, string(event.Key))
		}
	}
}

// watchOrderServices 监控订单服务变化
func (sc *ServiceClient) watchOrderServices(watchChan <-chan *client.WatchEvent) {
	for event := range watchChan {
		if event == nil {
			continue
		}

		switch event.Type {
		case 1: // PUT
			log.Printf("🆕 订单服务新增/更新: %s", string(event.Key))
		case 2: // DELETE
			log.Printf("🗑️  订单服务删除: %s", string(event.Key))
		default:
			log.Printf("📝 订单服务变化: %d - %s", event.Type, string(event.Key))
		}
	}
}

// HealthCheck 执行健康检查
func (sc *ServiceClient) HealthCheck() error {
	log.Println("🏥 执行健康检查...")

	// 检查用户服务健康状态
	userHealth, err := sc.discovery.HealthCheck(context.Background(), "user-service")
	if err != nil {
		return fmt.Errorf("用户服务健康检查失败: %v", err)
	}

	log.Println("📊 用户服务健康状态:")
	for serviceID, isHealthy := range userHealth {
		status := "✅ 健康"
		if !isHealthy {
			status = "❌ 不健康"
		}
		log.Printf("  - %s: %s", serviceID, status)
	}

	// 检查订单服务健康状态
	orderHealth, err := sc.discovery.HealthCheck(context.Background(), "order-service")
	if err != nil {
		return fmt.Errorf("订单服务健康检查失败: %v", err)
	}

	log.Println("📊 订单服务健康状态:")
	for serviceID, isHealthy := range orderHealth {
		status := "✅ 健康"
		if !isHealthy {
			status = "❌ 不健康"
		}
		log.Printf("  - %s: %s", serviceID, status)
	}

	return nil
}

// LoadBalancingDemo 负载均衡演示
func (sc *ServiceClient) LoadBalancingDemo() error {
	log.Println("⚖️ 负载均衡演示...")

	// 创建负载均衡器
	lb := client.NewLoadBalancer(sc.discovery)

	// 多次获取服务实例，演示负载均衡
	for i := 0; i < 5; i++ {
		// 获取用户服务实例
		userInstance, err := lb.GetInstance(context.Background(), "user-service")
		if err != nil {
			log.Printf("❌ 获取用户服务实例失败: %v", err)
			continue
		}
		log.Printf("🎯 第 %d 次选择用户服务: %s at %s", i+1, userInstance.ID, userInstance.Endpoint)

		// 获取订单服务实例
		orderInstance, err := lb.GetInstance(context.Background(), "order-service")
		if err != nil {
			log.Printf("❌ 获取订单服务实例失败: %v", err)
			continue
		}
		log.Printf("🎯 第 %d 次选择订单服务: %s at %s", i+1, orderInstance.ID, orderInstance.Endpoint)

		time.Sleep(1 * time.Second)
	}

	return nil
}

func main() {
	log.Println("🚀 启动 firEtcd 服务发现客户端演示...")

	// 创建服务客户端
	client, err := NewServiceClient()
	if err != nil {
		log.Fatalf("❌ 创建服务客户端失败: %v", err)
	}

	// 等待服务启动
	log.Println("⏳ 等待服务启动...")
	time.Sleep(5 * time.Second)

	// 执行演示
	log.Println("🎬 开始执行演示...")

	// 1. 健康检查
	if err := client.HealthCheck(); err != nil {
		log.Printf("⚠️  健康检查失败: %v", err)
	}

	// 2. 发现并调用用户服务
	if err := client.DiscoverAndCallUserService(); err != nil {
		log.Printf("⚠️  用户服务演示失败: %v", err)
	}

	// 3. 发现并调用订单服务
	if err := client.DiscoverAndCallOrderService(); err != nil {
		log.Printf("⚠️  订单服务演示失败: %v", err)
	}

	// 4. 负载均衡演示
	if err := client.LoadBalancingDemo(); err != nil {
		log.Printf("⚠️  负载均衡演示失败: %v", err)
	}

	// 5. 监控服务变化
	log.Println("👀 开始监控服务变化（30秒）...")
	if err := client.MonitorServices(); err != nil {
		log.Printf("⚠️  服务监控失败: %v", err)
	}

	log.Println("🎉 演示完成！")
}
