package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/whosefriendA/firEtcd/pkg/firconfig"
)

// LoadTLSConfig 根据配置加载TLS配置
func LoadTLSConfig(config *firconfig.TLSConfig, basePath string) (*tls.Config, error) {
	if config == nil {
		return nil, nil
	}

	// 解析路径（支持相对路径和绝对路径）
	certFile := resolvePath(config.CertFile, basePath)
	keyFile := resolvePath(config.KeyFile, basePath)
	caFile := resolvePath(config.CAFile, basePath)

	// 加载证书和私钥
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	// 加载CA证书
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	// 创建证书池
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to append CA certificate to pool")
	}

	// 配置客户端认证
	clientAuth := tls.NoClientCert
	if config.ClientAuth {
		clientAuth = tls.RequireAndVerifyClientCert
	}

	// 解析TLS版本
	minVersion, err := parseTLSVersion(config.MinVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid min TLS version: %w", err)
	}

	maxVersion, err := parseTLSVersion(config.MaxVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid max TLS version: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{certificate},
		ClientCAs:          certPool,
		ClientAuth:         clientAuth,
		MinVersion:         minVersion,
		MaxVersion:         maxVersion,
		InsecureSkipVerify: config.InsecureSkipVerify,
	}

	return tlsConfig, nil
}

// LoadClientTLSConfig 加载客户端TLS配置
func LoadClientTLSConfig(config *firconfig.TLSConfig, basePath string) (*tls.Config, error) {
	if config == nil {
		return nil, nil
	}

	// 解析路径
	certFile := resolvePath(config.CertFile, basePath)
	keyFile := resolvePath(config.KeyFile, basePath)
	caFile := resolvePath(config.CAFile, basePath)

	// 加载客户端证书和私钥
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	// 加载CA证书
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	// 创建证书池
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to append CA certificate to pool")
	}

	// 解析TLS版本
	minVersion, err := parseTLSVersion(config.MinVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid min TLS version: %w", err)
	}

	maxVersion, err := parseTLSVersion(config.MaxVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid max TLS version: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{certificate},
		RootCAs:            certPool,
		MinVersion:         minVersion,
		MaxVersion:         maxVersion,
		InsecureSkipVerify: config.InsecureSkipVerify,
	}

	return tlsConfig, nil
}

// parseTLSVersion 解析TLS版本字符串为tls.VersionTLS常量
func parseTLSVersion(version string) (uint16, error) {
	if version == "" {
		return tls.VersionTLS12, nil // 默认使用TLS 1.2
	}

	switch version {
	case "1.0":
		return tls.VersionTLS10, nil
	case "1.1":
		return tls.VersionTLS11, nil
	case "1.2":
		return tls.VersionTLS12, nil
	case "1.3":
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("unsupported TLS version: %s", version)
	}
}

// resolvePath 解析证书文件路径
func resolvePath(certPath, basePath string) string {
	// 如果是绝对路径，直接返回
	if filepath.IsAbs(certPath) {
		return certPath
	}

	// 如果是相对路径，相对于basePath
	return filepath.Join(basePath, certPath)
}
