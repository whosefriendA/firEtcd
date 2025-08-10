package firconfig

import (
	"fmt"
)

type Kvserver struct {
	Addr         string
	Port         string
	Rafts        RaftEnds
	DataBasePath string
	Maxraftstate int
	TLS          *TLSConfig `yaml:"tls,omitempty"`
}

func (c *Kvserver) Default() {
	*c = DefaultKVServer()
}

func (c *Kvserver) Validate() error {
	// 验证TLS配置
	if c.TLS != nil && c.TLS.IsEnabled() {
		if err := c.TLS.Validate(); err != nil {
			return fmt.Errorf("kvserver TLS config validation failed: %w", err)
		}
	}
	return nil
}

func DefaultKVServer() Kvserver {
	return Kvserver{
		Addr:         "127.0.0.1",
		Port:         ":51242",
		Rafts:        DefaultRaftEnds(),
		DataBasePath: "data",
		Maxraftstate: 100000,
		TLS: &TLSConfig{
			CertFile:   "certs/server.crt",
			KeyFile:    "certs/server.key",
			CAFile:     "certs/ca.crt",
			ClientAuth: true,
		},
	}
}
