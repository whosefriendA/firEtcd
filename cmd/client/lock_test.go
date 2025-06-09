package client_test

import (
	"sync"
	"testing"
	"time"

	"github.com/whosefriendA/firEtcd/pkg/firlog"
)

func TestLock(t *testing.T) {
	wait := sync.WaitGroup{}
	ck.Delete("lock")
	wait.Add(4)
	for i := range 4 {
		go func(routine int) {
			defer wait.Done()
			var (
				id  string
				err error
			)
			for {
				id, err = ck.Lock("lock", 0)
				if err != nil {
					firlog.Logger.Fatalln(err)
				}
				if id != "" {
					break
				}
				time.Sleep(500 * time.Millisecond)

			}
			firlog.Logger.Infof("routine[%d] gain lock", routine)
			firlog.Logger.Infof("routine[%d] do something (sleep 500ms)", routine)
			time.Sleep(time.Millisecond * 500)
			firlog.Logger.Infof("routine[%d] release the lock", routine)
			ok, err := ck.Unlock("lock", id)
			if err != nil {
				firlog.Logger.Fatalln(err)
			}
			if ok {
				firlog.Logger.Infof("routine[%d] success unlock", routine)
			} else {
				firlog.Logger.Infof("routine[%d] faild  unlock", routine)
			}
		}(i)
	}
	wait.Wait()
}
