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
)

type KVconn struct {
	Valid    bool
	Conn     pb.KvserverClient
	Realconn *grpc.ClientConn
}

// NewKvConn 创建新的KV客户端，支持TLS配置
func NewKvConn(addr string, tlsConfig *firconfig.TLSConfig) *KVconn {
	var creds credentials.TransportCredentials

	if tlsConfig != nil && tlsConfig.IsEnabled() {
		certificate, err := tls.LoadX509KeyPair(tlsConfig.CertFile, tlsConfig.KeyFile)
		if err != nil {
			firlog.Logger.Errorf("无法加载客户端证书: %v", err)
			return nil
		}

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

		tlsClientConfig := &tls.Config{
			Certificates:       []tls.Certificate{certificate},
			RootCAs:            certPool,
			ServerName:         "localhost",
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		}

		creds = credentials.NewTLS(tlsClientConfig)
	} else {
		creds = credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		firlog.Logger.Errorf("连接失败: %v", err)
		return nil
	}

	client := pb.NewKvserverClient(conn)

	ret := &KVconn{
		Valid:    true,
		Conn:     client,
		Realconn: conn,
	}
	return ret
}
