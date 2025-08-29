package firconfig

type Clerk struct {
	EtcdAddrs []string
	TLS       *TLSConfig // 添加TLS配置支持
}

func (c *Clerk) Default() {
	*c = DefaultClerk()
}

func (c *Clerk) Validate() error {
	// 验证TLS配置
	if c.TLS != nil {
		if err := c.TLS.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func DefaultClerk() Clerk {
	return Clerk{
		EtcdAddrs: []string{
			"127.0.0.1:51240", "127.0.0.1:51241", "127.0.0.1:51242",
		},
		TLS: DefaultTLSConfig(), // 使用默认TLS配置
	}
}
