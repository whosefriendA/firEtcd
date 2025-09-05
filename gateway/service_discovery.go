package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/whosefriendA/firEtcd/client"
)

// ServiceDiscoveryHandler 服务发现 API 处理器
type ServiceDiscoveryHandler struct {
	registry  *client.ServiceRegistryV3
	discovery *client.ServiceDiscoveryV3
}

// NewServiceDiscoveryHandler 创建服务发现处理器
func NewServiceDiscoveryHandler(ck client.KVStore) *ServiceDiscoveryHandler {
	registry := client.NewServiceRegistryV3(ck)

	watcher, ok := ck.(client.Watcher)
	if !ok {
		watcher = &noopWatcher{}
	}

	discovery := client.NewServiceDiscoveryV3(ck, watcher)
	return &ServiceDiscoveryHandler{
		registry:  registry,
		discovery: discovery,
	}
}

// noopWatcher 空实现的 Watcher 接口
type noopWatcher struct{}

func (nw *noopWatcher) Watch(ctx context.Context, key string, opts ...client.WatchOption) (<-chan *client.WatchEvent, error) {
	ch := make(chan *client.WatchEvent)
	close(ch)
	return ch, nil
}

// RegisterServiceRequest 服务注册请求
type RegisterServiceRequest struct {
	ServiceName string            `json:"service_name"`
	ServiceID   string            `json:"service_id"`
	Endpoint    string            `json:"endpoint"`
	TTL         int64             `json:"ttl"` // TTL in seconds
	Metadata    map[string]string `json:"metadata"`
}

// RegisterServiceResponse 服务注册响应
type RegisterServiceResponse struct {
	Success bool   `json:"success"`
	LeaseID int64  `json:"lease_id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ServiceListResponse 服务列表响应
type ServiceListResponse struct {
	Success  bool                  `json:"success"`
	Services []*client.ServiceInfo `json:"services,omitempty"`
	Error    string                `json:"error,omitempty"`
}

// HealthCheckResponse 健康检查响应
type HealthCheckResponse struct {
	Success bool            `json:"success"`
	Health  map[string]bool `json:"health,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// RegisterService 注册服务
func (h *ServiceDiscoveryHandler) RegisterService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ServiceName == "" || req.ServiceID == "" || req.Endpoint == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if req.TTL <= 0 {
		req.TTL = 30 // 默认 30 秒
	}

	leaseID, err := h.registry.Register(
		context.Background(),
		req.ServiceName,
		req.ServiceID,
		req.Endpoint,
		time.Duration(req.TTL)*time.Second,
		req.Metadata,
	)

	response := RegisterServiceResponse{}
	if err != nil {
		response.Success = false
		response.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Success = true
		response.LeaseID = leaseID
		response.Message = "Service registered successfully"
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeregisterService 注销服务
func (h *ServiceDiscoveryHandler) DeregisterService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serviceName := r.URL.Query().Get("service_name")
	serviceID := r.URL.Query().Get("service_id")

	if serviceName == "" || serviceID == "" {
		http.Error(w, "Missing service_name or service_id", http.StatusBadRequest)
		return
	}

	err := h.registry.Deregister(context.Background(), serviceName, serviceID)

	response := RegisterServiceResponse{}
	if err != nil {
		response.Success = false
		response.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Success = true
		response.Message = "Service deregistered successfully"
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetServices 获取服务列表
func (h *ServiceDiscoveryHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serviceName := r.URL.Query().Get("service_name")
	if serviceName == "" {
		http.Error(w, "Missing service_name", http.StatusBadRequest)
		return
	}

	services, err := h.discovery.Get(context.Background(), serviceName)

	response := ServiceListResponse{}
	if err != nil {
		response.Success = false
		response.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Success = true
		response.Services = services
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthCheck 健康检查
func (h *ServiceDiscoveryHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serviceName := r.URL.Query().Get("service_name")
	if serviceName == "" {
		http.Error(w, "Missing service_name", http.StatusBadRequest)
		return
	}

	health, err := h.discovery.HealthCheck(context.Background(), serviceName)

	response := HealthCheckResponse{}
	if err != nil {
		response.Success = false
		response.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Success = true
		response.Health = health
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LoadBalancer 负载均衡
func (h *ServiceDiscoveryHandler) LoadBalancer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serviceName := r.URL.Query().Get("service_name")
	if serviceName == "" {
		http.Error(w, "Missing service_name", http.StatusBadRequest)
		return
	}

	lb := client.NewLoadBalancer(h.discovery)

	instance, err := lb.GetInstance(context.Background(), serviceName)

	response := ServiceListResponse{}
	if err != nil {
		response.Success = false
		response.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Success = true
		response.Services = []*client.ServiceInfo{instance}
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// KeepAlive 续约服务
func (h *ServiceDiscoveryHandler) KeepAlive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	leaseIDStr := r.URL.Query().Get("lease_id")
	if leaseIDStr == "" {
		http.Error(w, "Missing lease_id", http.StatusBadRequest)
		return
	}

	leaseID, err := strconv.ParseInt(leaseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid lease_id", http.StatusBadRequest)
		return
	}

	err = h.registry.KeepAlive(context.Background(), leaseID)

	response := RegisterServiceResponse{}
	if err != nil {
		response.Success = false
		response.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Success = true
		response.Message = "Keep alive successful"
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes 注册服务发现路由
func (h *ServiceDiscoveryHandler) RegisterRoutes(mux *http.ServeMux, baseURL string) {
	mux.HandleFunc("POST "+baseURL+"/services/register", h.RegisterService)

	mux.HandleFunc("DELETE "+baseURL+"/services/deregister", h.DeregisterService)

	mux.HandleFunc("GET "+baseURL+"/services", h.GetServices)

	mux.HandleFunc("GET "+baseURL+"/services/health", h.HealthCheck)

	mux.HandleFunc("GET "+baseURL+"/services/loadbalancer", h.LoadBalancer)

	mux.HandleFunc("POST "+baseURL+"/services/keepalive", h.KeepAlive)
}
