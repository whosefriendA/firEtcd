package client_test // 包名修改

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	// !!! 确保你导入了正确的 client 包路径
	"github.com/whosefriendA/firEtcd/client" // <-- 新增导入
	// !!! 确保你的 mocks 包路径正确
	"github.com/whosefriendA/firEtcd/mocks"
)

// --- Standard Unit Tests (No Mocks Needed) ---

func TestNode_Key(t *testing.T) {
	// 使用 client.Node 访问结构体
	node := &client.Node{
		Env:   "production",
		AppId: "user-service",
		Name:  "user-service-01",
		Port:  "8080",
	}
	// 假设 Key 的格式为 /env/appId/name:port
	expectedKey := "/production/user-service/user-service-01:8080"
	if got := node.Key(); got != expectedKey {
		t.Errorf("Node.Key() = %v, want %v", got, expectedKey)
	}
}

func TestNode_MarshalUnmarshal(t *testing.T) {
	// 此测试同时验证 Marshal 和 Unmarshal
	node := &client.Node{
		Name:     "node-1",
		AppId:    "app-1",
		Port:     "9000",
		IPs:      []string{"192.168.1.10", "10.0.0.1"},
		Location: "us-west-1",
		Connect:  50,
		Weight:   200,
		Env:      "staging",
		MetaDate: map[string]string{"region": "ca", "version": "1.2.3"},
	}

	// 1. 序列化节点
	marshaledData, err := json.Marshal(node) // 假设使用 JSON 序列化
	if err != nil {
		t.Fatalf("Failed to marshal node for test setup: %v", err)
	}

	// 2. 反序列化到一个新节点
	newNode := &client.Node{}
	err = json.Unmarshal(marshaledData, newNode) // 假设使用 JSON 反序列化
	if err != nil {
		t.Fatalf("Failed to unmarshal data into new node: %v", err)
	}

	// 3. 验证新节点与原始节点完全相同
	if !reflect.DeepEqual(newNode, node) {
		t.Errorf("Unmarshal() result is incorrect.\ngot:  %#v\nwant: %#v", newNode, node)
	}
}

func TestNode_SetNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 1. 设置
	node := &client.Node{Name: "test-node", AppId: "test-app", Env: "dev", Port: "80"}
	expectedKey := node.Key()
	expectedValue, _ := json.Marshal(node) // 假设使用 JSON
	ttl := 5 * time.Second

	t.Run("Successful SetNode", func(t *testing.T) {
		// 创建实现了 KVStore 接口的 mock 对象
		mockKVStore := mocks.NewMockKVStore(ctrl)

		// 设置期望：Put 方法应该被以正确的 key, value 和 TTL 调用一次
		mockKVStore.EXPECT().
			Put(expectedKey, expectedValue, ttl).
			Return(nil). // 成功场景下返回 nil
			Times(1)

		// 使用 mock 对象调用方法
		err := node.SetNode(mockKVStore, ttl)
		if err != nil {
			t.Errorf("SetNode() returned unexpected error: %v", err)
		}
	})

	t.Run("Failed SetNode", func(t *testing.T) {
		mockKVStore := mocks.NewMockKVStore(ctrl)
		expectedErr := errors.New("database unavailable")

		// 设置期望：Put 方法将被调用，但会返回一个错误
		mockKVStore.EXPECT().
			Put(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(expectedErr).
			Times(1)

		err := node.SetNode(mockKVStore, ttl)
		if !errors.Is(err, expectedErr) {
			t.Errorf("SetNode() did not return the expected error. got: %v, want: %v", err, expectedErr)
		}
	})
}

func TestNode_SetNode_Watch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 1. 设置
	node := &client.Node{Name: "watch-node", AppId: "watch-app", Env: "dev", Port: "80"}
	expectedKey := node.Key()
	expectedValue, _ := json.Marshal(node)
	mockWatchDoger := mocks.NewMockWatchDoger(ctrl)

	// 一个虚拟的 cancel 函数，供 mock 返回
	var cancelFuncCalled bool
	dummyCancel := func() {
		cancelFuncCalled = true
	}

	// 2. 设置期望
	mockWatchDoger.EXPECT().
		WatchDog(expectedKey, expectedValue).
		Return(dummyCancel).
		Times(1)

	// 3. 执行
	returnedCancel := node.SetNode_Watch(mockWatchDoger)

	// 4. 断言
	if returnedCancel == nil {
		t.Fatal("SetNode_Watch() returned a nil cancel function")
	}

	// 验证返回的函数是我们 mock 的那个
	returnedCancel()
	if !cancelFuncCalled {
		t.Error("The cancel function returned by SetNode_Watch was not the one provided by the mock")
	}
}

func TestGetNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 1. 设置
	node1 := &client.Node{Name: "node-1", AppId: "app-1", Env: "prod", Port: "80"}
	node2 := &client.Node{Name: "node-2", AppId: "app-1", Env: "prod", Port: "81"}
	expectedAppName := "app-1"
	// 假设 GetNode 内部会创建像 "/env/appId/" 这样的前缀 key
	expectedPrefixKey := "/prod/app-1/"

	// 准备 mock 将要返回的数据
	node1Data, _ := json.Marshal(node1)
	node2Data, _ := json.Marshal(node2)
	mockReturnData := [][]byte{node1Data, node2Data}

	mockKVStore := mocks.NewMockKVStore(ctrl)

	t.Run("Successful GetNode with results", func(t *testing.T) {
		// 设置期望：GetWithPrefix 将被调用并返回我们准备好的数据
		mockKVStore.EXPECT().
			GetWithPrefix(expectedPrefixKey).
			Return(mockReturnData, nil).
			Times(1)

		// 调用 client 包的 GetNode 函数
		nodes, err := client.GetNode(mockKVStore, expectedAppName, "prod") // 假设 GetNode 接收 env 参数
		if err != nil {
			t.Fatalf("GetNode() returned unexpected error: %v", err)
		}
		if len(nodes) != 2 {
			t.Fatalf("GetNode() expected 2 nodes, got %d", len(nodes))
		}
		// 这里需要比较 client.Node 指针
		if !reflect.DeepEqual(nodes[0], node1) {
			t.Errorf("First node did not match expectation")
		}
	})

	t.Run("GetNode with no results", func(t *testing.T) {
		// 设置期望：GetWithPrefix 将返回一个空切片
		mockKVStore.EXPECT().
			GetWithPrefix(gomock.Any()).
			Return([][]byte{}, nil). // 返回空切片
			Times(1)

		nodes, err := client.GetNode(mockKVStore, "non-existent-app", "prod")
		if err != nil {
			t.Fatalf("GetNode() returned unexpected error: %v", err)
		}
		if len(nodes) != 0 {
			t.Errorf("GetNode() should return an empty slice when no nodes are found, got %d", len(nodes))
		}
	})

	t.Run("GetNode with underlying error", func(t *testing.T) {
		expectedErr := errors.New("connection refused")
		// 设置期望：GetWithPrefix 将返回一个错误
		mockKVStore.EXPECT().
			GetWithPrefix(gomock.Any()).
			Return(nil, expectedErr).
			Times(1)

		_, err := client.GetNode(mockKVStore, "any-app", "prod")
		if !errors.Is(err, expectedErr) {
			t.Errorf("GetNode() did not return the expected error. got: %v, want: %v", err, expectedErr)
		}
	})
}
