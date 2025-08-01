package firconfig

type Gateway struct {
	Clerk   Clerk
	Addr    string
	Port    string
	BaseUrl string `yaml:"baseUrl"`
}

func (g *Gateway) Default() {
	*g = DefaultGateway()
}

func DefaultGateway() Gateway {
	return Gateway{
		Clerk:   DefaultClerk(),
		Addr:    "127.0.0.1",
		Port:    ":51030",
		BaseUrl: "/firEtcd",
	}
}
