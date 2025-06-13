package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/whosefriendA/firEtcd/client"
	mock_client "github.com/whosefriendA/firEtcd/mocks" // 导入生成的 mock 包
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
)

// setupTestRouterWithGateway 是一个辅助函数，用于设置测试环境
// 它返回一个 gin 引擎和一个 mock clerk 控制器，方便后续的模拟设置
func setupTestRouterWithGateway(t *testing.T) (*gin.Engine, *Gateway, *gomock.Controller) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	mockClerk := mock_client.NewMockClerk(ctrl)

	conf := firconfig.Gateway{
		ListenAddr: ":8080", // 使用一个测试地址
	}

	g := &Gateway{
		ck:   mockClerk,
		conf: conf,
	}

	r := gin.New()
	g.LoadRouter(r) // 加载所有路由

	return r, g, ctrl
}

func TestNewGateway(t *testing.T) {
	type args struct {
		conf firconfig.Gateway
	}
	tests := []struct {
		name string
		args args
		want *Gateway
	}{
		{
			name: "Default case",
			args: args{
				conf: firconfig.Gateway{
					ListenAddr: ":8080",
					EtcdAddrs:  []string{"localhost:2379"},
				},
			},
			// 我们无法直接比较 ck，因为它是通过 client.NewClerk 创建的
			// 所以在实际测试中，我们只验证 conf 是否正确传递
			// 并在断言时忽略 ck 字段或单独验证
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGateway(tt.args.conf)
			// 断言 conf 是否相等
			assert.Equal(t, tt.args.conf, got.conf, "NewGateway() conf = %v, want %v", got.conf, tt.args.conf)
			// 断言 ck 是否被初始化 (不为 nil)
			assert.NotNil(t, got.ck, "NewGateway() clerk should be initialized")
		})
	}
}

func TestGateway_Run(t *testing.T) {
	// 这个测试很难在不真正监听端口的情况下进行
	// 我们可以测试一个明确会失败的场景
	t.Run("Invalid Listen Address", func(t *testing.T) {
		g := &Gateway{
			conf: firconfig.Gateway{ListenAddr: "invalid-address"},
		}
		err := g.Run()
		assert.Error(t, err, "Expected an error for invalid listen address")
	})
	// 成功的测试需要启动一个 goroutine 并在之后停止它，会使测试复杂化
	// 通常我们信任 http.ListenAndServe 的行为，只测试我们的 handler
}

func TestGateway_LoadRouter(t *testing.T) {
	r, _, _ := setupTestRouterWithGateway(t)

	// 定义期望的路由
	expectedRoutes := map[string]string{
		"GET":    "/watch",
		"GET":    "/keys",
		"GET":    "/kvs",
		"PUT":    "/kv",
		"POST":   "/cas",
		"DELETE": "/del",
		"DELETE": "/del-prefix",
		"GET":    "/get",
		"GET":    "/get-prefix",
	}

	routes := r.Routes()
	for _, route := range routes {
		if _, ok := expectedRoutes[route.Method]; ok {
			assert.Equal(t, expectedRoutes[route.Method], route.Path)
			delete(expectedRoutes, route.Method) // 移除已找到的路由
		}
	}
	assert.Empty(t, expectedRoutes, "Not all expected routes were registered")
}

func TestGateway_put(t *testing.T) {
	r, g, ctrl := setupTestRouterWithGateway(t)
	defer ctrl.Finish()

	mockCk := g.ck.(*mock_client.MockClerk)

	t.Run("Success", func(t *testing.T) {
		key, value := "testkey", "testvalue"
		mockCk.EXPECT().Put(key, value).Return(nil).Times(1)

		reqBody, _ := json.Marshal(map[string]string{"key": key, "value": value})
		req, _ := http.NewRequest("PUT", "/kv", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"code":0, "msg":"success"}`, w.Body.String())
	})

	t.Run("Binding Error", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/kv", bytes.NewBufferString(`{"key": "test"`)) // 无效 JSON
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGateway_get(t *testing.T) {
	r, g, ctrl := setupTestRouterWithGateway(t)
	defer ctrl.Finish()

	mockCk := g.ck.(*mock_client.MockClerk)

	t.Run("Success", func(t *testing.T) {
		key, value := "testkey", "testvalue"
		mockCk.EXPECT().Get(key).Return(value, nil).Times(1)

		req, _ := http.NewRequest("GET", "/get?key="+key, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		expectedJSON := `{"code":0, "msg":"success", "data":{"key":"testkey", "value":"testvalue"}}`
		assert.JSONEq(t, expectedJSON, w.Body.String())
	})

	t.Run("Key Not Found", func(t *testing.T) {
		key := "nonexistentkey"
		mockCk.EXPECT().Get(key).Return("", client.ErrKeyNotFound).Times(1)

		req, _ := http.NewRequest("GET", "/get?key="+key, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), client.ErrKeyNotFound.Error())
	})
}

func TestGateway_del(t *testing.T) {
	r, g, ctrl := setupTestRouterWithGateway(t)
	defer ctrl.Finish()

	mockCk := g.ck.(*mock_client.MockClerk)

	t.Run("Success", func(t *testing.T) {
		key := "testkey"
		mockCk.EXPECT().Del(key).Return(nil).Times(1)

		reqBody, _ := json.Marshal(map[string]string{"key": key})
		req, _ := http.NewRequest("DELETE", "/del", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"code":0, "msg":"success"}`, w.Body.String())
	})

	t.Run("Client Error", func(t *testing.T) {
		key := "testkey"
		mockCk.EXPECT().Del(key).Return(errors.New("some etcd error")).Times(1)

		reqBody, _ := json.Marshal(map[string]string{"key": key})
		req, _ := http.NewRequest("DELETE", "/del", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "some etcd error")
	})
}

func TestGateway_getWithPrefix(t *testing.T) {
	r, g, ctrl := setupTestRouterWithGateway(t)
	defer ctrl.Finish()

	mockCk := g.ck.(*mock_client.MockClerk)

	t.Run("Success", func(t *testing.T) {
		prefix := "user/"
		kvs := map[string]string{
			"user/1": "Alice",
			"user/2": "Bob",
		}
		mockCk.EXPECT().GetWithPrefix(prefix).Return(kvs, nil).Times(1)

		req, _ := http.NewRequest("GET", "/get-prefix?prefix="+prefix, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		expectedBody, _ := json.Marshal(gin.H{
			"code": 0,
			"msg":  "success",
			"data": kvs,
		})
		assert.JSONEq(t, string(expectedBody), w.Body.String())
	})

	t.Run("No Prefix Query", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/get-prefix", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "prefix is required")
	})
}

// ... 你可以为 putCAS, delWithPrefix, kvs, keys, watch 等其他函数添加类似的测试 ...
