package kvraft

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

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

func NewKvClient(addr string) *KVClient {
	// 加载客户端证书
	certificate, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/client.crt", "/home/wanggang/firEtcd/pkg/tls/certs/client.key")
	if err != nil {
		firlog.Logger.Errorf("无法加载客户端证书: %v", err)
		return nil
	}

	// 创建证书池并添加CA证书
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("/home/wanggang/firEtcd/pkg/tls/certs/ca.crt")
	if err != nil {
		firlog.Logger.Errorf("无法读取CA证书: %v", err)
		return nil
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		firlog.Logger.Error("无法将CA证书添加到证书池")
		return nil
	}

	// 配置TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
		ServerName:   "localhost", // 验证服务器主机名
		MinVersion:   tls.VersionTLS12,
	}

	// 创建TLS凭证
	creds := credentials.NewTLS(tlsConfig)

	// 使用TLS凭证创建gRPC连接
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
