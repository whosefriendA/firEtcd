package main

import (
	"github.com/whosefriendA/firEtcd/gateway"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
)

var gate *gateway.Gateway

func init() {
	var conf firconfig.Gateway
	firconfig.Init("config.yml", &conf)
	gate = gateway.NewGateway(conf)
}

func main() {
	gate.Run()
}
