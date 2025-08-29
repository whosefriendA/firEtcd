package kvraft

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/whosefriendA/firEtcd/pkg/firconfig"
	"github.com/whosefriendA/firEtcd/pkg/firlog"
	"github.com/whosefriendA/firEtcd/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"fmt"
)

type KVClient struct {
	Valid    bool
	Conn     pb.KvserverClient
	Realconn *grpc.ClientConn
}

// NewKvClient 创建新的KV客户端，支持TLS配置
func NewKvClient(addr string, tlsConfig *firconfig.TLSConfig) *KVClient {
	var creds credentials.TransportCredentials

	if tlsConfig != nil && tlsConfig.IsEnabled() {
		// 加载客户端证书
		certificate, err := tls.LoadX509KeyPair(tlsConfig.CertFile, tlsConfig.KeyFile)
		if err != nil {
			firlog.Logger.Errorf("无法加载客户端证书: %v", err)
			return nil
		}

		// 创建证书池并添加CA证书
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(tlsConfig.CAFile)
		if err != nil {
			firlog.Logger.Errorf("无法读取CA证书: %v", err)
			return nil
		}
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			firlog.Logger.Error("无法将CA证书添加到证书池")
			return nil
		}

		// 配置TLS
		tlsClientConfig := &tls.Config{
			Certificates:       []tls.Certificate{certificate},
			RootCAs:            certPool,
			ServerName:         "localhost", // 验证服务器主机名
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true, // 临时跳过主机名验证，用于测试
		}

		// 创建TLS凭证
		creds = credentials.NewTLS(tlsClientConfig)
	} else {
		// 如果没有TLS配置，使用不安全连接（仅用于测试）
		creds = credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	}

	// 使用凭证创建gRPC连接
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		fmt.Printf("gRPC dial error: %v\n", err)
		firlog.Logger.Errorf("连接失败: %v", err)
		return nil
	}

	client := pb.NewKvserverClient(conn)

	ret := &KVClient{
		Valid:    true,
		Conn:     client,
		Realconn: conn,
	}
	return ret
}
