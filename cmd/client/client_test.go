package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/whosefriendA/firEtcd/client"
	"github.com/whosefriendA/firEtcd/kvraft"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
	"github.com/whosefriendA/firEtcd/pkg/firlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var ck *client.Clerk

func init() {
	conf := firconfig.Clerk{}
	firconfig.Init("config.yml", &conf)
	// firLog.Logger.Debugln("check conf", conf)
	ck = client.MakeClerk(conf)
}

// 测试TTL功能

func TestTimeOut(t *testing.T) {
	key := "comet"
	value := []byte("localhost")
	ttl := time.Millisecond * 300
	ck.Put(key, value, ttl)
	firlog.Logger.Infof("set key [%s] value [%s] TTL[%v]", key, value, ttl)
	firlog.Logger.Infoln("time.Sleep for 200ms")
	time.Sleep(time.Millisecond * 200)
	v, err := ck.Get("comet")
	if err == kvraft.ErrNil {
		firlog.Logger.Infoln("err:", err)
	} else {
		firlog.Logger.Infof("get value [%s]", v)
	}
	firlog.Logger.Infoln("time.Sleep for 200ms")
	time.Sleep(time.Millisecond * 200)
	v, err = ck.Get("comet")
	if err == kvraft.ErrNil {
		firlog.Logger.Infoln("err:", err)
	} else {
		firlog.Logger.Infof("get value [%s]", v)
	}
	firlog.Logger.Infof("reset key [%s] value [%s] TTL[%v]", key, value, ttl)
	ck.Put("comet", value, time.Millisecond*300)
	v, err = ck.Get("comet")
	if err == kvraft.ErrNil {
		firlog.Logger.Infoln("err:", err)
	} else {
		firlog.Logger.Infof("get value [%s]", v)
	}
}

func BenchmarkFirEtcdPut(b *testing.B) {
	key := "logic"
	value := []byte("test")
	for range b.N {
		err := ck.Put(key, value, 0)
		if err != nil {
			b.Error(err)
		}
	}
}

func TestFirEtcdPut(t *testing.T) {
	key := "logic"
	value := []byte("test")
	for range 4 {
		start := time.Now()
		err := ck.Put(key, value, 0)
		firlog.Logger.Warnln("client finish put key[%s] spand time:", "logic", time.Since(start))
		if err != nil {
			t.Error(err)
		}
	}
}
func BenchmarkFirEtcdGet(b *testing.B) {
	for range b.N {
		_, err := ck.Get("logic")
		if err != nil && err != kvraft.ErrNil {
			b.Error(err)
		}
	}
}

func TestFirEtcdGet(t *testing.T) {
	for range 4 {
		start := time.Now()
		_, err := ck.Get("logic")
		firlog.Logger.Warnln("client finish put key[%s] spand time:", "logic", time.Since(start))
		if err != nil && err != kvraft.ErrNil {
			t.Error(err)
		}
	}
}

func BenchmarkFirEtcdGetWithPrefix(b *testing.B) {
	for range b.N {
		_, err := ck.GetWithPrefix("logic")
		if err != nil && err != kvraft.ErrNil {
			b.Error(err)
		}
	}

}

func BenchmarkFirEtcdDelete(b *testing.B) {
	for range b.N {
		err := ck.Delete("logic")
		if err != nil {
			b.Error(err)
		}
	}

}

var etcd = NewEtcd()

func NewEtcd() *clientv3.Client {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379", "127.0.0.1:22379", "127.0.0.1:32379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		firlog.Logger.Fatalln(err)
	}
	firlog.Logger.Infoln("success connect etcd")
	return c
}

func BenchmarkEtcdPut(b *testing.B) {

	for range b.N {
		_, err := etcd.Put(context.Background(), "logic", "test")
		if err != nil {
			b.Error(err)
		}
	}
}
func BenchmarkEtcdGet(b *testing.B) {
	for range b.N {
		_, err := etcd.Get(context.Background(), "logic")
		if err != nil {
			b.Error(err)
		}
	}
}
func BenchmarkEtcdGetWithPrefix(b *testing.B) {

	for range b.N {
		_, err := etcd.Get(context.Background(), "logic", clientv3.WithPrefix())
		if err != nil {
			b.Error(err)
		}
	}
}
func BenchmarkEtcdDelete(b *testing.B) {

	for range b.N {
		_, err := etcd.Delete(context.Background(), "logic")
		if err != nil {
			b.Error(err)
		}
	}
}

func TestEtcdPut(t *testing.T) {

	for range 4 {
		start := time.Now()
		_, err := etcd.Put(context.Background(), "logic", "test")
		firlog.Logger.Warnln("client finish put key[%s] spand time:", "logic", time.Since(start))
		if err != nil {
			t.Error(err)
		}
	}
}
