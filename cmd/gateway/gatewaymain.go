package main

import (
	"github.com/whosefriendA/firEtcd/gateway"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
	"github.com/whosefriendA/firEtcd/pkg/firlog"
)

var gate *gateway.Gateway

func init() {
	var conf firconfig.Gateway
	firconfig.Init("config.yml", &conf)
	firlog.Logger.Debugf("Gateway <UNK> %s %s", conf.Port, conf.BaseUrl)
	gate = gateway.NewGateway(conf)
}

func main() {
	gate.Run()
}
