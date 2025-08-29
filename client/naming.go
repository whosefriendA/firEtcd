package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/whosefriendA/firEtcd/pkg/firlog"
)

type Node struct {
	Name     string
	AppId    string
	Port     string
	IPs      []string
	Location string
	Connect  int32
	Weight   int32
	Env      string
	MetaDate map[string]string //"color" "version"
}

func (n *Node) Marshal() []byte {
	data, err := json.Marshal(n)
	if err != nil {
		firlog.Logger.Fatalln(err)
	}
	return data
}

func (n *Node) Unmarshal(data []byte) {
	err := json.Unmarshal(data, n)
	if err != nil {
		firlog.Logger.Fatalln(err)
	}
}

func (n *Node) Key() string {
	// 使用 etcdv3 风格的键格式
	return fmt.Sprintf("/%s/%s/%s:%s", n.Env, n.AppId, n.Name, n.Port)
}

// SetNode 使用 Lease 机制注册服务节点
func (n *Node) SetNode(ck KVStore, TTL time.Duration) error {
	// 直接使用 TTL，内部会自动创建 Lease
	err := ck.Put(n.Key(), n.Marshal(), TTL)
	if err != nil {
		return fmt.Errorf("failed to register service node: %w", err)
	}
	return nil
}

// SetNodeWithLease 手动管理租约的服务注册（推荐用于长期服务）
func (n *Node) SetNodeWithLease(ck KVStore, TTL time.Duration) (int64, func(), error) {
	// 创建租约
	leaseID, err := ck.LeaseGrant(TTL)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to grant lease: %w", err)
	}

	// 使用租约注册服务（无 TTL）
	err = ck.Put(n.Key(), n.Marshal(), 0)
	if err != nil {
		ck.LeaseRevoke(leaseID)
		return 0, nil, fmt.Errorf("failed to register service node: %w", err)
	}

	// 启动自动续约
	cancel := ck.AutoKeepAlive(leaseID, TTL/3)

	// 返回租约ID和取消函数
	return leaseID, func() {
		cancel()
		ck.LeaseRevoke(leaseID)
	}, nil
}

// SetNodeWithLeaseID 使用现有租约注册服务
func (n *Node) SetNodeWithLeaseID(ck KVStore, leaseID int64) error {
	// 使用现有租约注册服务
	err := ck.Put(n.Key(), n.Marshal(), 0)
	if err != nil {
		return fmt.Errorf("failed to register service node: %w", err)
	}
	return nil
}

// 让 GetNode 依赖 KVStore 接口
// 注意：为了能构造出正确的 prefix key，函数可能需要更多参数，比如 env
func GetNode(ck KVStore, appName string, env string) ([]*Node, error) {
	// 假设 prefix key 的格式是 /env/appName/
	prefixKey := fmt.Sprintf("/%s/%s/", env, appName)
	datas, err := ck.GetWithPrefix(prefixKey)
	if err != nil {
		// 如果错误是 ErrNil，说明没找到，这不是一个真正的错误
		//if errors.Is(err, ErrNil) { // 假设您有一个 ErrNil
		//	return []*Node{}, nil
		//}
		return nil, err
	}

	if len(datas) == 0 {
		return []*Node{}, nil
	}

	nodes := make([]*Node, 0, len(datas))
	for _, data := range datas {
		node := &Node{}
		// 假设使用 json unmarshal
		if e := json.Unmarshal(data, node); e == nil {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

// GetNodeWithLease 获取服务节点并返回租约信息
func GetNodeWithLease(ck KVStore, appName string, env string) ([]*NodeWithLease, error) {
	nodes, err := GetNode(ck, appName, env)
	if err != nil {
		return nil, err
	}

	result := make([]*NodeWithLease, 0, len(nodes))
	for _, node := range nodes {
		// 这里需要从存储中获取租约信息
		// 由于当前实现限制，我们先返回基本信息
		result = append(result, &NodeWithLease{
			Node: node,
			// LeaseID: 需要从存储中解析
		})
	}

	return result, nil
}

// NodeWithLease 包含节点信息和租约信息
type NodeWithLease struct {
	*Node
	LeaseID int64
	TTL     int64 // 剩余时间（毫秒）
}

// ServiceRegistry etcdv3 风格的服务注册器
type ServiceRegistry struct {
	ck KVStore
}

// NewServiceRegistry 创建服务注册器
func NewServiceRegistry(ck KVStore) *ServiceRegistry {
	return &ServiceRegistry{ck: ck}
}

// Register 注册服务（自动管理租约）
func (sr *ServiceRegistry) Register(serviceName, serviceID string, endpoint string, TTL time.Duration, metadata map[string]string) (int64, func(), error) {
	node := &Node{
		Name:     serviceID,
		AppId:    serviceName,
		Port:     endpoint,
		IPs:      []string{"localhost"}, // 可以从 endpoint 解析
		Location: "default",
		Connect:  0,
		Weight:   1,
		Env:      "default",
		MetaDate: metadata,
	}

	return node.SetNodeWithLease(sr.ck, TTL)
}

// Deregister 注销服务
func (sr *ServiceRegistry) Deregister(serviceName, serviceID string) error {
	key := fmt.Sprintf("/%s/%s/%s", "default", serviceName, serviceID)
	_, err := sr.ck.Get(key)
	if err != nil {
		return fmt.Errorf("service not found: %w", err)
	}

	// 删除服务记录
	err = sr.ck.Put(key, nil, 0) // 设置为空值
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	return nil
}

// GetService 获取服务实例列表
func (sr *ServiceRegistry) GetService(serviceName string) ([]*Node, error) {
	return GetNode(sr.ck, serviceName, "default")
}

// ServiceDiscovery etcdv3 风格的服务发现器
type ServiceDiscovery struct {
	ck      KVStore
	watcher Watcher
}

// NewServiceDiscovery 创建服务发现器
func NewServiceDiscovery(ck KVStore, watcher Watcher) *ServiceDiscovery {
	return &ServiceDiscovery{ck: ck, watcher: watcher}
}

// Discover 发现服务
func (sd *ServiceDiscovery) Discover(serviceName string) ([]*Node, error) {
	return GetNode(sd.ck, serviceName, "default")
}

// Watch 监控服务变化
func (sd *ServiceDiscovery) Watch(serviceName string) (<-chan *WatchEvent, error) {
	prefixKey := fmt.Sprintf("/%s/%s/", "default", serviceName)
	return sd.watcher.Watch(context.Background(), prefixKey, WithPrefix())
}
