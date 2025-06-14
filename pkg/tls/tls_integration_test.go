package certs

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// 模拟gRPC服务
type testServer struct {
	*grpc.Server
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *testServer) Start(addr string, tlsConfig *tls.Config) error {
	// 创建带取消功能的上下文
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// 创建 gRPC 服务器，使用 TLS 凭证
	creds := credentials.NewTLS(tlsConfig)
	s.Server = grpc.NewServer(grpc.Creds(creds))

	// 创建监听器
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("创建监听器失败: %v", err)
	}

	// 在goroutine中启动服务器
	go func() {
		if err := s.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			panic(fmt.Sprintf("服务器启动失败: %v", err))
		}
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (s *testServer) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.Server != nil {
		s.Server.GracefulStop()
	}
}

// 模拟gRPC客户端
type testClient struct {
	conn *grpc.ClientConn
}

func (c *testClient) Connect(addr string, tlsConfig *tls.Config) error {
	// 创建 TLS 凭证
	creds := credentials.NewTLS(tlsConfig)

	// 设置连接选项
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 创建连接
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}

	c.conn = conn
	return nil
}

func (c *testClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func TestTLSIntegration(t *testing.T) {
	// 加载CA证书
	caPEM, err := ioutil.ReadFile("/home/wanggang/firEtcd/pkg/tls/certs/ca.crt")
	if err != nil {
		t.Fatalf("无法读取CA证书: %v", err)
	}

	// 解码 PEM 到 DER
	block, _ := pem.Decode(caPEM)
	if block == nil {
		t.Fatalf("无法解码 PEM 格式CA证书")
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPEM) {
		t.Fatalf("无法将CA证书添加到证书池")
	}

	// 测试服务器-客户端TLS连接
	t.Run("测试服务器-客户端TLS连接", func(t *testing.T) {
		// 加载服务器证书
		serverCert, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/server.crt", "/home/wanggang/firEtcd/pkg/tls/certs/server.key")
		if err != nil {
			t.Fatalf("无法加载服务器证书: %v", err)
		}

		// 创建服务器TLS配置
		serverConfig := &tls.Config{
			Certificates: []tls.Certificate{serverCert},
			ClientCAs:    certPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
			MinVersion:   tls.VersionTLS12,
		}

		// 启动测试服务器
		server := &testServer{}
		addr := "localhost:12345"
		if err := server.Start(addr, serverConfig); err != nil {
			t.Fatalf("服务器启动失败: %v", err)
		}
		defer server.Stop()

		// 等待服务器启动
		time.Sleep(200 * time.Millisecond)

		// 测试TCP连接
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err != nil {
			t.Fatalf("TCP连接失败: %v", err)
		}
		conn.Close()

		// 加载客户端证书
		clientCert, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/client.crt", "/home/wanggang/firEtcd/pkg/tls/certs/client.key")
		if err != nil {
			t.Fatalf("无法加载客户端证书: %v", err)
		}

		// 创建客户端TLS配置
		clientConfig := &tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
			MinVersion:   tls.VersionTLS12,
		}

		// 创建测试客户端
		client := &testClient{}
		if err := client.Connect(addr, clientConfig); err != nil {
			t.Fatalf("客户端连接失败: %v", err)
		}
		defer client.Close()
	})

	// 测试网关-客户端TLS连接
	t.Run("测试网关-客户端TLS连接", func(t *testing.T) {
		// 加载网关证书
		gatewayCert, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/gateway.crt", "/home/wanggang/firEtcd/pkg/tls/certs/gateway.key")
		if err != nil {
			t.Fatalf("无法加载网关证书: %v", err)
		}

		// 创建网关TLS配置
		gatewayConfig := &tls.Config{
			Certificates: []tls.Certificate{gatewayCert},
			ClientCAs:    certPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
			MinVersion:   tls.VersionTLS12,
		}

		// 启动测试网关
		gateway := &testServer{}
		addr := "localhost:12346"
		if err := gateway.Start(addr, gatewayConfig); err != nil {
			t.Fatalf("网关启动失败: %v", err)
		}
		defer gateway.Stop()

		// 等待网关启动
		time.Sleep(200 * time.Millisecond)

		// 测试TCP连接
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err != nil {
			t.Fatalf("TCP连接失败: %v", err)
		}
		conn.Close()

		// 加载客户端证书
		clientCert, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/client.crt", "/home/wanggang/firEtcd/pkg/tls/certs/client.key")
		if err != nil {
			t.Fatalf("无法加载客户端证书: %v", err)
		}

		// 创建客户端TLS配置
		clientConfig := &tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
			MinVersion:   tls.VersionTLS12,
		}

		// 创建测试客户端
		client := &testClient{}
		if err := client.Connect(addr, clientConfig); err != nil {
			t.Fatalf("客户端连接网关失败: %v", err)
		}
		defer client.Close()
	})

	// 测试无效证书
	t.Run("测试无效证书", func(t *testing.T) {
		// 启动一个正常的服务器
		serverCert, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/server.crt", "/home/wanggang/firEtcd/pkg/tls/certs/server.key")
		if err != nil {
			t.Fatalf("无法加载服务器证书: %v", err)
		}

		serverConfig := &tls.Config{
			Certificates: []tls.Certificate{serverCert},
			ClientCAs:    certPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
			MinVersion:   tls.VersionTLS12,
		}

		server := &testServer{}
		addr := "localhost:12347"
		if err := server.Start(addr, serverConfig); err != nil {
			t.Fatalf("服务器启动失败: %v", err)
		}
		defer server.Stop()

		// 等待服务器启动
		time.Sleep(200 * time.Millisecond)

		// 测试TCP连接
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err != nil {
			t.Fatalf("TCP连接失败: %v", err)
		}
		conn.Close()

		// 创建一个无效的TLS配置（不提供任何证书）
		invalidConfig := &tls.Config{
			RootCAs:    certPool,
			MinVersion: tls.VersionTLS12,
		}

		// 尝试连接
		client := &testClient{}
		err = client.Connect(addr, invalidConfig)
		if err == nil {
			t.Fatalf("使用无效证书连接成功，期望失败")
		}
	})
}
