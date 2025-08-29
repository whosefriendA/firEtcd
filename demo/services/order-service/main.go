package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/whosefriendA/firEtcd/client"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
)

// OrderService 订单服务
type OrderService struct {
	registry   *client.ServiceRegistryV3
	discovery  *client.ServiceDiscoveryV3
	port       string
	instanceID string
}

// Order 订单模型
type Order struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	ProductID  string    `json:"product_id"`
	Quantity   int       `json:"quantity"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	CreateAt   string    `json:"create_at"`
	UserInfo   *UserInfo `json:"user_info,omitempty"`
}

// UserInfo 用户信息（从用户服务获取）
type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NewOrderService 创建订单服务
func NewOrderService(port string) (*OrderService, error) {
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

	// 创建服务注册器
	registry := client.NewServiceRegistryV3(ck)

	// 生成实例ID
	instanceID := fmt.Sprintf("order-service-%s-%d", port, time.Now().Unix())

	return &OrderService{
		registry:   registry,
		port:       port,
		instanceID: instanceID,
	}, nil
}

// Start 启动服务
func (s *OrderService) Start() error {
	// 注册服务
	leaseID, err := s.registry.Register(
		context.Background(),
		"order-service",
		s.instanceID,
		fmt.Sprintf("localhost:%s", s.port),
		30*time.Second,
		map[string]string{
			"version": "v1.0.0",
			"env":     "demo",
			"port":    s.port,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to register service: %v", err)
	}

	log.Printf("✅ 订单服务注册成功，租约ID: %d", leaseID)

	// 启动自动续约
	conf := firconfig.Clerk{}
	firconfig.Init("../../cluster_config/node1/config.yml", &conf)
	ck := client.MakeClerk(conf)

	cancel := ck.AutoKeepAlive(leaseID, 10*time.Second)
	defer cancel()

	// 启动 HTTP 服务器
	mux := http.NewServeMux()
	s.registerHandlers(mux)

	server := &http.Server{
		Addr:    ":" + s.port,
		Handler: mux,
	}

	// 启动服务器
	go func() {
		log.Printf("🚀 订单服务启动在端口 %s", s.port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 正在关闭订单服务...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ 服务器关闭失败: %v", err)
	}

	// 注销服务
	if err := s.registry.Deregister(context.Background(), "order-service", s.instanceID); err != nil {
		log.Printf("⚠️  服务注销失败: %v", err)
	} else {
		log.Println("✅ 服务注销成功")
	}

	return nil
}

// registerHandlers 注册 HTTP 处理器
func (s *OrderService) registerHandlers(mux *http.ServeMux) {
	// 健康检查
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "order-service",
			"instance":  s.instanceID,
			"port":      s.port,
			"timestamp": time.Now().Unix(),
		})
	})

	// 获取订单列表
	mux.HandleFunc("GET /orders", func(w http.ResponseWriter, r *http.Request) {
		orders := []Order{
			{
				ID:         "1",
				UserID:     "1",
				ProductID:  "product-1",
				Quantity:   2,
				TotalPrice: 199.99,
				Status:     "pending",
				CreateAt:   "2024-01-01",
			},
			{
				ID:         "2",
				UserID:     "2",
				ProductID:  "product-2",
				Quantity:   1,
				TotalPrice: 99.99,
				Status:     "completed",
				CreateAt:   "2024-01-02",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    orders,
			"count":   len(orders),
		})
	})

	// 获取单个订单（包含用户信息）
	mux.HandleFunc("GET /orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		// 简单的路由解析
		path := r.URL.Path
		if len(path) < 9 {
			http.Error(w, "Invalid order ID", http.StatusBadRequest)
			return
		}
		orderID := path[9:]

		// 模拟订单数据
		order := Order{
			ID:         orderID,
			UserID:     "1",
			ProductID:  fmt.Sprintf("product-%s", orderID),
			Quantity:   2,
			TotalPrice: 199.99,
			Status:     "pending",
			CreateAt:   "2024-01-01",
		}

		// 尝试获取用户信息（服务发现演示）
		userInfo, err := s.getUserInfo(order.UserID)
		if err == nil {
			order.UserInfo = userInfo
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    order,
		})
	})

	// 创建订单
	mux.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {
		var order Order
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		order.ID = fmt.Sprintf("%d", time.Now().Unix())
		order.CreateAt = time.Now().Format("2006-01-02")
		order.Status = "pending"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    order,
			"message": "订单创建成功",
		})
	})

	// 服务信息
	mux.HandleFunc("GET /info", func(w http.ResponseWriter, r *http.Request) {
		// 发现其他服务
		userServices, err := s.discovery.Get(context.Background(), "user-service")
		userServiceCount := 0
		if err == nil {
			userServiceCount = len(userServices)
		}

		info := map[string]interface{}{
			"service":       "order-service",
			"instance":      s.instanceID,
			"port":          s.port,
			"status":        "running",
			"user_services": userServiceCount,
			"registered_at": time.Now().Format("2006-01-02 15:04:05"),
			"uptime":        time.Since(time.Now()).String(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})

	// 服务发现演示
	mux.HandleFunc("GET /discovery", func(w http.ResponseWriter, r *http.Request) {
		// 发现用户服务
		userServices, err := s.discovery.Get(context.Background(), "user-service")
		if err != nil {
			http.Error(w, "Failed to discover user services", http.StatusInternalServerError)
			return
		}

		// 发现订单服务
		orderServices, err := s.discovery.Get(context.Background(), "order-service")
		if err != nil {
			http.Error(w, "Failed to discover order services", http.StatusInternalServerError)
			return
		}

		discoveryInfo := map[string]interface{}{
			"user_services":  userServices,
			"order_services": orderServices,
			"timestamp":      time.Now().Unix(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(discoveryInfo)
	})
}

// getUserInfo 获取用户信息（演示服务间通信）
func (s *OrderService) getUserInfo(userID string) (*UserInfo, error) {
	// 发现用户服务
	services, err := s.discovery.Get(context.Background(), "user-service")
	if err != nil || len(services) == 0 {
		return nil, fmt.Errorf("no user services available")
	}

	// 选择第一个可用的用户服务
	userService := services[0]

	// 构建请求URL
	url := fmt.Sprintf("http://%s/users/%s", userService.Endpoint, userID)

	// 发送HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call user service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status: %d", resp.StatusCode)
	}

	// 解析响应
	var response struct {
		Success bool     `json:"success"`
		Data    UserInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("user service request failed")
	}

	return &response.Data, nil
}

func main() {
	// 获取端口号
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	// 创建并启动服务
	service, err := NewOrderService(port)
	if err != nil {
		log.Fatalf("❌ 创建订单服务失败: %v", err)
	}

	// 启动服务
	if err := service.Start(); err != nil {
		log.Fatalf("❌ 启动订单服务失败: %v", err)
	}
}
