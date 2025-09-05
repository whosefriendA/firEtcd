package main

import (
	"flag"
	"runtime"

	"github.com/whosefriendA/firEtcd/kvraft"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
	"github.com/whosefriendA/firEtcd/pkg/firlog"
)

var (
	ConfigPath = flag.String("c", "config.yml", "path fo config.yml folder")
)

func main() {
	runtime.GOMAXPROCS(1)
	flag.Parse()
	conf := firconfig.Kvserver{}
	firconfig.Init(*ConfigPath, &conf)
	if len(conf.Rafts.Endpoints)%2 == 0 {
		firlog.Logger.Fatalln("the number of nodes is not odd")
	}
	if len(conf.Rafts.Endpoints) < 3 {
		firlog.Logger.Fatalln("the number of nodes is less than 3")
	}

	firlog.InitLogger("kvserver", true, false, false)

	_ = kvraft.StartKVServer(conf, conf.Rafts.Me, conf.DataBasePath, conf.Maxraftstate)

	select {} // 阻塞主进程
}
