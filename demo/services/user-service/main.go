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

// UserService 用户服务
type UserService struct {
	registry   *client.ServiceRegistryV3
	discovery  *client.ServiceDiscoveryV3
	port       string
	instanceID string
}

// User 用户模型
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	CreateAt string `json:"create_at"`
}

// NewUserService 创建用户服务
func NewUserService(port string) (*UserService, error) {
	// 创建 firEtcd 客户端配置
	conf := firconfig.Clerk{
		EtcdAddrs: []string{
			"127.0.0.1:51240",
			"127.0.0.1:51241",
			"127.0.0.1:51242",
		},
		TLS: &firconfig.TLSConfig{
			CertFile: "../pkg/tls/certs/client.crt",
			KeyFile:  "../pkg/tls/certs/client.key",
			CAFile:   "../pkg/tls/certs/ca.crt",
		},
	}
	ck := client.MakeClerk(conf)

	// 创建服务注册器
	registry := client.NewServiceRegistryV3(ck)

	// 生成实例ID
	instanceID := fmt.Sprintf("user-service-%s-%d", port, time.Now().Unix())

	return &UserService{
		registry:   registry,
		port:       port,
		instanceID: instanceID,
	}, nil
}

// Start 启动服务
func (s *UserService) Start() error {
	// 注册服务
	leaseID, err := s.registry.Register(
		context.Background(),
		"user-service",
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

	log.Printf("✅ 用户服务注册成功，租约ID: %d", leaseID)

	// 启动自动续约 - 需要重新获取客户端
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
		log.Printf("🚀 用户服务启动在端口 %s", s.port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 正在关闭用户服务...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ 服务器关闭失败: %v", err)
	}

	// 注销服务
	if err := s.registry.Deregister(context.Background(), "user-service", s.instanceID); err != nil {
		log.Printf("⚠️  服务注销失败: %v", err)
	} else {
		log.Println("✅ 服务注销成功")
	}

	return nil
}

// registerHandlers 注册 HTTP 处理器
func (s *UserService) registerHandlers(mux *http.ServeMux) {
	// 健康检查
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "user-service",
			"instance":  s.instanceID,
			"port":      s.port,
			"timestamp": time.Now().Unix(),
		})
	})

	// 获取用户列表
	mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		users := []User{
			{
				ID:       "1",
				Name:     "张三",
				Email:    "zhangsan@example.com",
				Age:      25,
				CreateAt: "2024-01-01",
			},
			{
				ID:       "2",
				Name:     "李四",
				Email:    "lisi@example.com",
				Age:      30,
				CreateAt: "2024-01-02",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    users,
			"count":   len(users),
		})
	})

	// 获取单个用户
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		// 简单的路由解析
		path := r.URL.Path
		if len(path) < 8 {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}
		userID := path[7:]

		user := User{
			ID:       userID,
			Name:     fmt.Sprintf("用户%s", userID),
			Email:    fmt.Sprintf("user%s@example.com", userID),
			Age:      25,
			CreateAt: "2024-01-01",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    user,
		})
	})

	// 创建用户
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		user.ID = fmt.Sprintf("%d", time.Now().Unix())
		user.CreateAt = time.Now().Format("2006-01-02")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    user,
			"message": "用户创建成功",
		})
	})

	// 服务信息
	mux.HandleFunc("GET /info", func(w http.ResponseWriter, r *http.Request) {
		// 发现其他服务
		services, err := s.discovery.Get(context.Background(), "order-service")
		orderServiceCount := 0
		if err == nil {
			orderServiceCount = len(services)
		}

		info := map[string]interface{}{
			"service":        "user-service",
			"instance":       s.instanceID,
			"port":           s.port,
			"status":         "running",
			"order_services": orderServiceCount,
			"registered_at":  time.Now().Format("2006-01-02 15:04:05"),
			"uptime":         time.Since(time.Now()).String(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})
}

func main() {
	// 获取端口号
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	// 创建并启动服务
	service, err := NewUserService(port)
	if err != nil {
		log.Fatalf("❌ 创建用户服务失败: %v", err)
	}

	// 启动服务
	if err := service.Start(); err != nil {
		log.Fatalf("❌ 启动用户服务失败: %v", err)
	}
}
