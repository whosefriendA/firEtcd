package firconfig

import (
	"fmt"
)

type Gateway struct {
	Clerk   Clerk
	Addr    string
	Port    string
	BaseUrl string     `yaml:"baseUrl"`
	TLS     *TLSConfig `yaml:"tls,omitempty"`
}

func (g *Gateway) Default() {
	*g = DefaultGateway()
}

func (g *Gateway) Validate() error {
	// 验证TLS配置
	if g.TLS != nil && g.TLS.IsEnabled() {
		if err := g.TLS.Validate(); err != nil {
			return fmt.Errorf("gateway TLS config validation failed: %w", err)
		}
	}
	return nil
}

func DefaultGateway() Gateway {
	return Gateway{
		Clerk:   DefaultClerk(),
		Addr:    "127.0.0.1",
		Port:    ":51030",
		BaseUrl: "/firEtcd",
		TLS: &TLSConfig{
			CertFile:   "certs/gateway.crt",
			KeyFile:    "certs/gateway.key",
			CAFile:     "certs/ca.crt",
			ClientAuth: true,
		},
	}
}
