package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/whosefriendA/firEtcd/client"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
)

// GatewayService API 网关服务
type GatewayService struct {
	discovery  *client.ServiceDiscoveryV3
	port       string
	instanceID string
}

// ServiceRoute 服务路由配置
type ServiceRoute struct {
	ServiceName string
	PathPrefix  string
	TargetPort  string
}

// NewGatewayService 创建网关服务
func NewGatewayService(port string) (*GatewayService, error) {
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

	// 生成实例ID
	instanceID := fmt.Sprintf("gateway-service-%s-%d", port, time.Now().Unix())

	return &GatewayService{
		discovery:  discovery,
		port:       port,
		instanceID: instanceID,
	}, nil
}

// Start 启动服务
func (s *GatewayService) Start() error {
	log.Printf("🚀 API 网关启动在端口 %s", s.port)

	// 启动 HTTP 服务器
	mux := http.NewServeMux()
	s.registerHandlers(mux)

	server := &http.Server{
		Addr:    ":" + s.port,
		Handler: mux,
	}

	// 启动服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 正在关闭 API 网关...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ 服务器关闭失败: %v", err)
	}

	return nil
}

// registerHandlers 注册 HTTP 处理器
func (s *GatewayService) registerHandlers(mux *http.ServeMux) {
	// 健康检查
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "gateway-service",
			"instance":  s.instanceID,
			"port":      s.port,
			"timestamp": time.Now().Unix(),
		})
	})

	// 服务发现状态
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		status := s.getServiceStatus()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	// 用户服务路由
	mux.HandleFunc("GET /api/users", s.handleUserService)
	mux.HandleFunc("GET /api/users/{id}", s.handleUserService)
	mux.HandleFunc("POST /api/users", s.handleUserService)

	// 订单服务路由
	mux.HandleFunc("GET /api/orders", s.handleOrderService)
	mux.HandleFunc("GET /api/orders/{id}", s.handleOrderService)
	mux.HandleFunc("POST /api/orders", s.handleOrderService)

	// 服务发现演示
	mux.HandleFunc("GET /api/discovery", s.handleDiscovery)
}

// getServiceStatus 获取服务状态
func (s *GatewayService) getServiceStatus() map[string]interface{} {
	// 发现用户服务
	userServices, err := s.discovery.Get(context.Background(), "user-service")
	userServiceCount := 0
	if err == nil {
		userServiceCount = len(userServices)
	}

	// 发现订单服务
	orderServices, err := s.discovery.Get(context.Background(), "order-service")
	orderServiceCount := 0
	if err == nil {
		orderServiceCount = len(orderServices)
	}

	return map[string]interface{}{
		"gateway": map[string]interface{}{
			"instance": s.instanceID,
			"port":     s.port,
			"status":   "running",
		},
		"services": map[string]interface{}{
			"user_service": map[string]interface{}{
				"count":  userServiceCount,
				"status": "healthy",
			},
			"order_service": map[string]interface{}{
				"count":  orderServiceCount,
				"status": "healthy",
			},
		},
		"timestamp": time.Now().Unix(),
	}
}

// handleUserService 处理用户服务请求
func (s *GatewayService) handleUserService(w http.ResponseWriter, r *http.Request) {
	// 发现用户服务
	services, err := s.discovery.Get(context.Background(), "user-service")
	if err != nil || len(services) == 0 {
		http.Error(w, "No user services available", http.StatusServiceUnavailable)
		return
	}

	// 负载均衡：选择服务实例
	selectedService := s.selectService(services)

	// 构建目标URL
	targetURL := fmt.Sprintf("http://%s%s", selectedService.Endpoint, r.URL.Path)

	// 转发请求
	s.forwardRequest(w, r, targetURL)
}

// handleOrderService 处理订单服务请求
func (s *GatewayService) handleOrderService(w http.ResponseWriter, r *http.Request) {
	// 发现订单服务
	services, err := s.discovery.Get(context.Background(), "order-service")
	if err != nil || len(services) == 0 {
		http.Error(w, "No order services available", http.StatusServiceUnavailable)
		return
	}

	// 负载均衡：选择服务实例
	selectedService := s.selectService(services)

	// 构建目标URL
	targetURL := fmt.Sprintf("http://%s%s", selectedService.Endpoint, r.URL.Path)

	// 转发请求
	s.forwardRequest(w, r, targetURL)
}

// handleDiscovery 处理服务发现请求
func (s *GatewayService) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	// 发现所有服务
	userServices, err := s.discovery.Get(context.Background(), "user-service")
	if err != nil {
		userServices = []*client.ServiceInfo{}
	}

	orderServices, err := s.discovery.Get(context.Background(), "order-service")
	if err != nil {
		orderServices = []*client.ServiceInfo{}
	}

	discoveryInfo := map[string]interface{}{
		"user_services":  userServices,
		"order_services": orderServices,
		"gateway": map[string]interface{}{
			"instance": s.instanceID,
			"port":     s.port,
		},
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(discoveryInfo)
}

// selectService 选择服务实例（负载均衡）
func (s *GatewayService) selectService(services []*client.ServiceInfo) *client.ServiceInfo {
	if len(services) == 0 {
		return nil
	}

	// 简单的轮询负载均衡
	// 实际应用中可以使用更复杂的算法（加权轮询、最少连接等）
	index := time.Now().UnixNano() % int64(len(services))
	return services[index]
}

// forwardRequest 转发请求到目标服务
func (s *GatewayService) forwardRequest(w http.ResponseWriter, r *http.Request, targetURL string) {
	// 解析目标URL
	target, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Invalid target URL", http.StatusInternalServerError)
		return
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(target)

	// 修改请求头
	r.URL.Host = target.Host
	r.URL.Scheme = target.Scheme
	r.Header.Set("X-Forwarded-For", r.RemoteAddr)
	r.Header.Set("X-Forwarded-By", s.instanceID)

	// 转发请求
	proxy.ServeHTTP(w, r)
}

// 简单的健康检查处理器
func (s *GatewayService) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "healthy",
		"service":  "gateway",
		"instance": s.instanceID,
	})
}

func main() {
	// 获取端口号
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 创建并启动网关服务
	service, err := NewGatewayService(port)
	if err != nil {
		log.Fatalf("❌ 创建网关服务失败: %v", err)
	}

	// 启动服务
	if err := service.Start(); err != nil {
		log.Fatalf("❌ 启动网关服务失败: %v", err)
	}
}
