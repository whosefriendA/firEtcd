package tls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"testing"
	"time"
)

func TestTLSEndToEnd(t *testing.T) {
	// 加载CA证书
	caCert, err := ioutil.ReadFile("/home/wanggang/firEtcd/pkg/tls/certs/ca.crt")
	if err != nil {
		t.Fatalf("无法读取CA证书: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		t.Fatalf("无法将CA证书添加到证书池")
	}

	// 测试服务器配置
	t.Run("测试服务器配置", func(t *testing.T) {
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

		// 验证服务器配置
		if len(serverConfig.Certificates) != 1 {
			t.Fatalf("服务器证书数量不正确")
		}
		if serverConfig.ClientAuth != tls.RequireAndVerifyClientCert {
			t.Fatalf("服务器未配置客户端证书验证")
		}
	})

	// 测试客户端配置
	t.Run("测试客户端配置", func(t *testing.T) {
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

		// 验证客户端配置
		if len(clientConfig.Certificates) != 1 {
			t.Fatalf("客户端证书数量不正确")
		}
		if clientConfig.RootCAs == nil {
			t.Fatalf("客户端未配置根证书")
		}
	})

	// 测试网关配置
	t.Run("测试网关配置", func(t *testing.T) {
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

		// 验证网关配置
		if len(gatewayConfig.Certificates) != 1 {
			t.Fatalf("网关证书数量不正确")
		}
		if gatewayConfig.ClientAuth != tls.RequireAndVerifyClientCert {
			t.Fatalf("网关未配置客户端证书验证")
		}
	})

	// 测试证书链验证
	t.Run("测试证书链验证", func(t *testing.T) {
		// 加载所有证书
		serverCert, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/server.crt", "/home/wanggang/firEtcd/pkg/tls/certs/server.key")
		if err != nil {
			t.Fatalf("无法加载服务器证书: %v", err)
		}

		clientCert, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/client.crt", "/home/wanggang/firEtcd/pkg/tls/certs/client.key")
		if err != nil {
			t.Fatalf("无法加载客户端证书: %v", err)
		}

		gatewayCert, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/gateway.crt", "/home/wanggang/firEtcd/pkg/tls/certs/gateway.key")
		if err != nil {
			t.Fatalf("无法加载网关证书: %v", err)
		}

		// 验证证书链
		certs := []tls.Certificate{serverCert, clientCert, gatewayCert}
		for _, cert := range certs {
			x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
			if err != nil {
				t.Fatalf("无法解析证书: %v", err)
			}

			// 验证证书是否由CA签名
			roots := x509.NewCertPool()
			roots.AddCert(x509Cert)
			opts := x509.VerifyOptions{
				Roots: roots,
			}

			if _, err := x509Cert.Verify(opts); err != nil {
				t.Fatalf("证书链验证失败: %v", err)
			}
		}
	})

	// 测试证书有效期
	t.Run("测试证书有效期", func(t *testing.T) {
		// 加载所有证书
		certs := []string{
			"/home/wanggang/firEtcd/pkg/tls/certs/server.crt",
			"/home/wanggang/firEtcd/pkg/tls/certs/client.crt",
			"/home/wanggang/firEtcd/pkg/tls/certs/gateway.crt",
			"/home/wanggang/firEtcd/pkg/tls/certs/ca.crt",
		}

		now := time.Now()
		for _, certPath := range certs {
			certPEM, err := ioutil.ReadFile(certPath)
			if err != nil {
				t.Fatalf("无法读取证书 %s: %v", certPath, err)
			}

			// 解码 PEM 到 DER
			block, _ := pem.Decode(certPEM)
			if block == nil {
				t.Fatalf("无法解码 PEM 格式证书: %s", certPath)
			}

			// 解析 DER 到 x509.Certificate
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				t.Fatalf("无法解析证书 %s: %v", certPath, err)
			}

			// 验证证书有效期
			if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
				t.Fatalf("证书 %s 已过期或尚未生效", certPath)
			}
		}
	})

	// 测试证书主题
	t.Run("测试证书主题", func(t *testing.T) {
		// 定义期望的主题
		expectedSubjects := map[string]string{
			"/home/wanggang/firEtcd/pkg/tls/certs/server.crt":  "server",
			"/home/wanggang/firEtcd/pkg/tls/certs/client.crt":  "client",
			"/home/wanggang/firEtcd/pkg/tls/certs/gateway.crt": "gateway",
			"/home/wanggang/firEtcd/pkg/tls/certs/ca.crt":      "CA",
		}

		for certPath, expectedSubject := range expectedSubjects {
			certPEM, err := ioutil.ReadFile(certPath)
			if err != nil {
				t.Fatalf("无法读取证书 %s: %v", certPath, err)
			}

			// 解码 PEM 到 DER
			block, _ := pem.Decode(certPEM)
			if block == nil {
				t.Fatalf("无法解码 PEM 格式证书: %s", certPath)
			}

			// 解析 DER 到 x509.Certificate
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				t.Fatalf("无法解析证书 %s: %v", certPath, err)
			}

			// 验证证书主题
			if cert.Subject.CommonName != expectedSubject {
				t.Fatalf("证书 %s 主题不匹配，期望: %s, 实际: %s",
					certPath, expectedSubject, cert.Subject.CommonName)
			}
		}
	})
}
