package client_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/whosefriendA/firEtcd/client"
	"github.com/whosefriendA/firEtcd/mocks"
)

func TestNode_Key(t *testing.T) {
	node := &client.Node{
		Env:   "production",
		AppId: "user-service",
		Name:  "user-service-01",
		Port:  "8080",
	}
	expectedKey := "/production/user-service/user-service-01:8080"
	if got := node.Key(); got != expectedKey {
		t.Errorf("Node.Key() = %v, want %v", got, expectedKey)
	}
}

func TestNode_MarshalUnmarshal(t *testing.T) {
	node := &client.Node{
		Name:	 "node-1",
		AppId:	 "app-1",
		Port:	 "9000",
		IPs:	 []string{"192.168.1.10", "10.0.0.1"},
		Location: "us-west-1",
		Connect:	 50,
		Weight:	 200,
		Env:	 "staging",
		MetaDate: map[string]string{"region": "ca", "version": "1.2.3"},
	}

	marshaledData, err := json.Marshal(node);
	if err != nil {
		t.Fatalf("Failed to marshal node for test setup: %v", err)
	}

	newNode := &client.Node{}
	err = json.Unmarshal(marshaledData, newNode)
	if err != nil {
		t.Fatalf("Failed to unmarshal data into new node: %v", err)
	}

	if !reflect.DeepEqual(newNode, node) {
		t.Errorf("Unmarshal() result is incorrect.\ngot:	%#v\nwant:	%#v", newNode, node)
	}
}

func TestNode_SetNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	node := &client.Node{Name: "test-node", AppId: "test-app", Env: "dev", Port: "80"}
	expectedKey := node.Key()
	expectedValue, _ := json.Marshal(node)
	ttl := 5 * time.Second

	t.Run("Successful SetNode", func(t *testing.T) {
		mockKVStore := mocks.NewMockKVStore(ctrl)

		mockKVStore.EXPECT().
			Put(expectedKey, expectedValue, ttl).
			Return(nil).
			Times(1)

		err := node.SetNode(mockKVStore, ttl)
		if err != nil {
			t.Errorf("SetNode() returned unexpected error: %v", err)
		}
	})

	t.Run("Failed SetNode", func(t *testing.T) {
		mockKVStore := mocks.NewMockKVStore(ctrl)
		expectedErr := errors.New("database unavailable")

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

	node := &client.Node{Name: "watch-node", AppId: "watch-app", Env: "dev", Port: "80"}
	expectedKey := node.Key()
	expectedValue, _ := json.Marshal(node)
	mockWatchDoger := mocks.NewMockWatchDoger(ctrl)

	var cancelFuncCalled bool
	dummyCancel := func() {
		cancelFuncCalled = true
	}

	mockWatchDoger.EXPECT().
		WatchDog(expectedKey, expectedValue).
		Return(dummyCancel).
		Times(1)

	returnedCancel := node.SetNode_Watch(mockWatchDoger)

	if returnedCancel == nil {
		t.Fatal("SetNode_Watch() returned a nil cancel function")
	}

	returnedCancel()
	if !cancelFuncCalled {
		t.Error("The cancel function returned by SetNode_Watch was not the one provided by the mock")
	}
}

func TestGetNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	node1 := &client.Node{Name: "node-1", AppId: "app-1", Env: "prod", Port: "80"}
	node2 := &client.Node{Name: "node-2", AppId: "app-1", Env: "prod", Port: "81"}
	expectedAppName := "app-1"
	expectedPrefixKey := "/prod/app-1/"

	node1Data, _ := json.Marshal(node1)
	node2Data, _ := json.Marshal(node2)
	mockReturnData := [][]byte{node1Data, node2Data}

	mockKVStore := mocks.NewMockKVStore(ctrl)

	t.Run("Successful GetNode with results", func(t *testing.T) {
		mockKVStore.EXPECT().
			GetWithPrefix(expectedPrefixKey).
			Return(mockReturnData, nil).
			Times(1)

		nodes, err := client.GetNode(mockKVStore, expectedAppName, "prod")
		if err != nil {
			t.Fatalf("GetNode() returned unexpected error: %v", err)
		}
		if len(nodes) != 2 {
			t.Fatalf("GetNode() expected 2 nodes, got %d", len(nodes))
		}
		if !reflect.DeepEqual(nodes[0], node1) {
			t.Errorf("First node did not match expectation")
		}
	})

	t.Run("GetNode with no results", func(t *testing.T) {
		mockKVStore.EXPECT().
			GetWithPrefix(gomock.Any()).
			Return([][]byte{}, nil).
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
