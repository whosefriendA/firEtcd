package client_test

import (
	"testing"
	"time"

	"github.com/whosefriendA/firEtcd/pkg/firlog"
)

func TestWatchDog(t *testing.T) {

	cancel := ck.WatchDog("watchdog", []byte("exist"))

	go func() {
		for {
			ret, err := ck.Get("watchdog")
			if err != nil {
				firlog.Logger.Fatalln(err)
			}
			firlog.Logger.Infoln(string(ret))
			time.Sleep(time.Millisecond * 500)
		}
	}()
	time.Sleep(time.Second * 2)
	firlog.Logger.Infoln("after cancel")
	cancel()
	time.Sleep(time.Second * 5)
}
