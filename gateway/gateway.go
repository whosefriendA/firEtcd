package gateway

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"

	"github.com/whosefriendA/firEtcd/client"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
	"github.com/whosefriendA/firEtcd/pkg/firlog"
)

type Gateway struct {
	ck   *client.Clerk
	conf firconfig.Gateway
}

func NewGateway(conf firconfig.Gateway) *Gateway {
	return &Gateway{
		ck:   client.MakeClerk(conf.Clerk),
		conf: conf,
	}
}

// realurl = baseUrl + '/keys' | '/put' act...
// default baseUrl = '/firEtcd'
func (g *Gateway) Run() error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET "+g.conf.BaseUrl+"/keys", g.keys)
	mux.HandleFunc("GET "+g.conf.BaseUrl+"/key", g.get)
	mux.HandleFunc("GET "+g.conf.BaseUrl+"/keysWithPrefix", g.getWithPrefix)
	mux.HandleFunc("GET "+g.conf.BaseUrl+"/kvs", g.kvs)
	mux.HandleFunc("GET "+g.conf.BaseUrl+"/watch", g.watch)
	mux.HandleFunc("POST "+g.conf.BaseUrl+"/put", g.put)
	mux.HandleFunc("POST "+g.conf.BaseUrl+"/putCAS", g.putCAS)
	mux.HandleFunc("DELETE "+g.conf.BaseUrl+"/key", g.del)
	mux.HandleFunc("DELETE "+g.conf.BaseUrl+"/keysWithPrefix", g.delWithPrefix)

	var handler http.Handler = mux
	handler = InstantLogger(handler)
	// 加载TLS证书
	certificate, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/gateway.crt", "/home/wanggang/firEtcd/pkg/tls/certs/gateway.key")
	if err != nil {
		firlog.Logger.Fatalf("无法加载网关证书: %v", err)
	}

	// 创建证书池并添加CA证书
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("/home/wanggang/firEtcd/pkg/tls/certs/ca.crt")
	if err != nil {
		firlog.Logger.Fatalf("无法读取CA证书: %v", err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		firlog.Logger.Fatal("无法将CA证书添加到证书池")
	}

	// 配置TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
	}

	// 创建TLS服务器
	server := &http.Server{
		Addr:      g.conf.Port,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	// 启动TLS服务器
	return server.ListenAndServeTLS("", "")
}
