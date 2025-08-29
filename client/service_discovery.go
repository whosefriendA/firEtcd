package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ServiceInfo 服务信息结构
type ServiceInfo struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Endpoint string            `json:"endpoint"`
	Version  string            `json:"version"`
	Metadata map[string]string `json:"metadata"`
	TTL      int64             `json:"ttl"`      // 剩余时间（毫秒）
	LeaseID  int64             `json:"lease_id"` // 租约ID
}

// ServiceInstance 服务实例
type ServiceInstance struct {
	*ServiceInfo
	Weight   int32  `json:"weight"`
	Location string `json:"location"`
	Env      string `json:"env"`
}

// ServiceRegistryV3 etcdv3 风格的服务注册器
type ServiceRegistryV3 struct {
	ck KVStore
}

// NewServiceRegistryV3 创建 etcdv3 风格的服务注册器
func NewServiceRegistryV3(ck KVStore) *ServiceRegistryV3 {
	return &ServiceRegistryV3{ck: ck}
}

// Register 注册服务实例
func (sr *ServiceRegistryV3) Register(ctx context.Context, serviceName, serviceID, endpoint string, TTL time.Duration, metadata map[string]string) (int64, error) {
	// 创建租约
	leaseID, err := sr.ck.LeaseGrant(TTL)
	if err != nil {
		return 0, fmt.Errorf("failed to grant lease: %w", err)
	}

	// 构建服务信息
	serviceInfo := &ServiceInfo{
		ID:       serviceID,
		Name:     serviceName,
		Endpoint: endpoint,
		Version:  metadata["version"],
		Metadata: metadata,
		TTL:      TTL.Milliseconds(),
		LeaseID:  leaseID,
	}

	// 使用 etcdv3 风格的键格式
	key := fmt.Sprintf("/services/%s/%s", serviceName, serviceID)

	// 序列化服务信息
	data, err := json.Marshal(serviceInfo)
	if err != nil {
		sr.ck.LeaseRevoke(leaseID)
		return 0, fmt.Errorf("failed to marshal service info: %w", err)
	}

	// 注册服务（无 TTL，使用租约）
	err = sr.ck.Put(key, data, 0)
	if err != nil {
		sr.ck.LeaseRevoke(leaseID)
		return 0, fmt.Errorf("failed to register service: %w", err)
	}

	return leaseID, nil
}

// RegisterWithLease 使用现有租约注册服务
func (sr *ServiceRegistryV3) RegisterWithLease(ctx context.Context, serviceName, serviceID, endpoint string, leaseID int64, metadata map[string]string) error {
	serviceInfo := &ServiceInfo{
		ID:       serviceID,
		Name:     serviceName,
		Endpoint: endpoint,
		Version:  metadata["version"],
		Metadata: metadata,
		LeaseID:  leaseID,
	}

	key := fmt.Sprintf("/services/%s/%s", serviceName, serviceID)
	data, err := json.Marshal(serviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	return sr.ck.Put(key, data, 0)
}

// Deregister 注销服务
func (sr *ServiceRegistryV3) Deregister(ctx context.Context, serviceName, serviceID string) error {
	key := fmt.Sprintf("/services/%s/%s", serviceName, serviceID)

	// 获取服务信息以获取租约ID
	data, err := sr.ck.Get(key)
	if err != nil {
		return fmt.Errorf("service not found: %w", err)
	}

	var serviceInfo ServiceInfo
	if err := json.Unmarshal(data, &serviceInfo); err != nil {
		return fmt.Errorf("failed to unmarshal service info: %w", err)
	}

	// 删除服务记录
	err = sr.ck.Put(key, nil, 0)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	// 撤销租约
	if serviceInfo.LeaseID > 0 {
		sr.ck.LeaseRevoke(serviceInfo.LeaseID)
	}

	return nil
}

// KeepAlive 续约服务
func (sr *ServiceRegistryV3) KeepAlive(ctx context.Context, leaseID int64) error {
	// 使用现有的 AutoKeepAlive 机制
	// 这里返回一个函数，调用者可以调用它来停止续约
	_ = sr.ck.AutoKeepAlive(leaseID, 30*time.Second)
	return nil
}

// ServiceDiscoveryV3 etcdv3 风格的服务发现器
type ServiceDiscoveryV3 struct {
	ck      KVStore
	watcher Watcher
}

// NewServiceDiscoveryV3 创建 etcdv3 风格的服务发现器
func NewServiceDiscoveryV3(ck KVStore, watcher Watcher) *ServiceDiscoveryV3 {
	return &ServiceDiscoveryV3{ck: ck, watcher: watcher}
}

// Get 获取服务实例列表
func (sd *ServiceDiscoveryV3) Get(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	prefixKey := fmt.Sprintf("/services/%s/", serviceName)
	datas, err := sd.ck.GetWithPrefix(prefixKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	services := make([]*ServiceInfo, 0, len(datas))
	for _, data := range datas {
		if len(data) == 0 {
			continue // 跳过空值（已删除的服务）
		}

		var serviceInfo ServiceInfo
		if err := json.Unmarshal(data, &serviceInfo); err != nil {
			continue // 跳过解析失败的数据
		}

		// 检查租约状态
		if serviceInfo.LeaseID > 0 {
			ttl, _, err := sd.ck.LeaseTimeToLive(serviceInfo.LeaseID, false)
			if err == nil {
				serviceInfo.TTL = ttl
			}
		}

		services = append(services, &serviceInfo)
	}

	return services, nil
}

// Watch 监控服务变化
func (sd *ServiceDiscoveryV3) Watch(ctx context.Context, serviceName string) (<-chan *WatchEvent, error) {
	prefixKey := fmt.Sprintf("/services/%s/", serviceName)
	return sd.watcher.Watch(ctx, prefixKey, WithPrefix())
}

// WatchPrefix 监控指定前缀的服务变化
func (sd *ServiceDiscoveryV3) WatchPrefix(ctx context.Context, prefix string) (<-chan *WatchEvent, error) {
	return sd.watcher.Watch(ctx, prefix, WithPrefix())
}

// HealthCheck 健康检查
func (sd *ServiceDiscoveryV3) HealthCheck(ctx context.Context, serviceName string) (map[string]bool, error) {
	services, err := sd.Get(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	health := make(map[string]bool)
	for _, service := range services {
		// 检查租约是否有效
		if service.LeaseID > 0 {
			ttl, _, err := sd.ck.LeaseTimeToLive(service.LeaseID, false)
			health[service.ID] = err == nil && ttl > 0
		} else {
			health[service.ID] = false
		}
	}

	return health, nil
}

// LoadBalancer 简单的负载均衡器
type LoadBalancer struct {
	discovery *ServiceDiscoveryV3
}

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(discovery *ServiceDiscoveryV3) *LoadBalancer {
	return &LoadBalancer{discovery: discovery}
}

// GetInstance 获取服务实例（轮询策略）
func (lb *LoadBalancer) GetInstance(ctx context.Context, serviceName string) (*ServiceInfo, error) {
	services, err := lb.discovery.Get(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no available instances for service: %s", serviceName)
	}

	// 简单的轮询策略（实际应用中可以使用更复杂的算法）
	// 这里使用时间戳作为简单的轮询依据
	index := time.Now().UnixNano() % int64(len(services))
	return services[index], nil
}

// GetAllInstances 获取所有可用实例
func (lb *LoadBalancer) GetAllInstances(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	return lb.discovery.Get(ctx, serviceName)
}
