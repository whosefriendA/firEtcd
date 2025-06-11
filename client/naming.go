package client

import (
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
	// 使用 fmt.Sprintf 格式化字符串，确保所有部分都被包含
	return fmt.Sprintf("/%s/%s/%s:%s", n.Env, n.AppId, n.Name, n.Port)
}

func (n *Node) SetNode(ck KVStore, TTL time.Duration) error {
	err := ck.Put(n.Key(), n.Marshal(), TTL)
	if err != nil {
		return fmt.Errorf("database unavailable: %w", err) // 将错误返回
	}
	return nil
}

func (n *Node) SetNode_Watch(ck WatchDoger) (cancle func()) {
	return ck.WatchDog(n.Key(), n.Marshal())
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
