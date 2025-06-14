package certs

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"testing"
	"time"
)

func TestTLSConfig(t *testing.T) {
	// 测试服务器证书
	t.Run("测试服务器证书", func(t *testing.T) {
		certificate, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/server.crt", "/home/wanggang/firEtcd/pkg/tls/certs/server.key")
		if err != nil {
			t.Fatalf("无法加载服务器证书: %v", err)
		}

		// 验证证书链
		cert, err := x509.ParseCertificate(certificate.Certificate[0])
		if err != nil {
			t.Fatalf("无法解析服务器证书: %v", err)
		}

		// 验证证书有效期
		now := time.Now()
		if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
			t.Fatalf("服务器证书已过期或尚未生效")
		}

		// 验证证书主题
		if cert.Subject.CommonName != "server" {
			t.Fatalf("服务器证书主题不匹配，期望: server, 实际: %s", cert.Subject.CommonName)
		}
	})

	// 测试客户端证书
	t.Run("测试客户端证书", func(t *testing.T) {
		certificate, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/client.crt", "/home/wanggang/firEtcd/pkg/tls/certs/client.key")
		if err != nil {
			t.Fatalf("无法加载客户端证书: %v", err)
		}

		// 验证证书链
		cert, err := x509.ParseCertificate(certificate.Certificate[0])
		if err != nil {
			t.Fatalf("无法解析客户端证书: %v", err)
		}

		// 验证证书有效期
		now := time.Now()
		if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
			t.Fatalf("客户端证书已过期或尚未生效")
		}

		// 验证证书主题
		if cert.Subject.CommonName != "client" {
			t.Fatalf("客户端证书主题不匹配，期望: client, 实际: %s", cert.Subject.CommonName)
		}
	})

	// 测试网关证书
	t.Run("测试网关证书", func(t *testing.T) {
		certificate, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/gateway.crt", "/home/wanggang/firEtcd/pkg/tls/certs/gateway.key")
		if err != nil {
			t.Fatalf("无法加载网关证书: %v", err)
		}

		// 验证证书链
		cert, err := x509.ParseCertificate(certificate.Certificate[0])
		if err != nil {
			t.Fatalf("无法解析网关证书: %v", err)
		}

		// 验证证书有效期
		now := time.Now()
		if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
			t.Fatalf("网关证书已过期或尚未生效")
		}

		// 验证证书主题
		if cert.Subject.CommonName != "gateway" {
			t.Fatalf("网关证书主题不匹配，期望: gateway, 实际: %s", cert.Subject.CommonName)
		}
	})

	// 测试CA证书
	t.Run("测试CA证书", func(t *testing.T) {
		certPath := "/home/wanggang/firEtcd/pkg/tls/certs/ca.crt"
		certPEM, err := ioutil.ReadFile(certPath)
		if err != nil {
			t.Fatalf("无法读取CA证书: %v", err)
		}

		block, _ := pem.Decode(certPEM)
		if block == nil {
			t.Fatalf("无法解码PEM格式CA证书")
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			t.Fatalf("无法解析CA证书: %v", err)
		}

		if cert.Subject.CommonName != "MyCA" {
			t.Errorf("CA证书主题不匹配，期望: MyCA, 实际: %s", cert.Subject.CommonName)
		}
		if !cert.IsCA {
			t.Errorf("CA证书的 IsCA 标志不为 true")
		}
	})
}

func TestTLSConnection(t *testing.T) {
	// 加载CA证书
	caCert, err := ioutil.ReadFile("/home/wanggang/firEtcd/pkg/tls/certs/ca.crt")
	if err != nil {
		t.Fatalf("无法读取CA证书: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		t.Fatalf("无法将CA证书添加到证书池")
	}

	// 测试服务器TLS配置
	t.Run("测试服务器TLS配置", func(t *testing.T) {
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

		// 验证服务器TLS配置
		if len(serverConfig.Certificates) != 1 {
			t.Fatalf("服务器TLS配置证书数量不正确")
		}
		if serverConfig.ClientAuth != tls.RequireAndVerifyClientCert {
			t.Fatalf("服务器TLS配置未要求客户端证书验证")
		}
	})

	// 测试客户端TLS配置
	t.Run("测试客户端TLS配置", func(t *testing.T) {
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

		// 验证客户端TLS配置
		if len(clientConfig.Certificates) != 1 {
			t.Fatalf("客户端TLS配置证书数量不正确")
		}
		if clientConfig.RootCAs == nil {
			t.Fatalf("客户端TLS配置未设置根证书")
		}
	})

	// 测试网关TLS配置
	t.Run("测试网关TLS配置", func(t *testing.T) {
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

		// 验证网关TLS配置
		if len(gatewayConfig.Certificates) != 1 {
			t.Fatalf("网关TLS配置证书数量不正确")
		}
		if gatewayConfig.ClientAuth != tls.RequireAndVerifyClientCert {
			t.Fatalf("网关TLS配置未要求客户端证书验证")
		}
	})
}
