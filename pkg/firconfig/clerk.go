package firconfig

type Clerk struct {
	EtcdAddrs []string
}

func (c *Clerk) Default() {
	*c = DefaultClerk()
}

func (c *Clerk) Validate() error {
	// Clerk配置验证逻辑可以在这里添加
	return nil
}

func DefaultClerk() Clerk {
	return Clerk{
		EtcdAddrs: []string{
			"127.0.0.1:51240", "127.0.0.1:51241", "127.0.0.1:51242",
		},
	}
}
