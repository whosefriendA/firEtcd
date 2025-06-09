package client_test

import (
	"testing"

	"github.com/whosefriendA/firEtcd/pkg/firlog"
)

func TestCAS(t *testing.T) {
	key := "comet"
	ck.Delete(key)
	firlog.Logger.Infof("CAS set key[%s] ori[%v] dest[%s] ", key, nil, "A")
	ok, err := ck.CAS(key, nil, []byte("A"), 0)
	if err != nil {
		firlog.Logger.Infoln(err)
	}
	if ok {
		firlog.Logger.Infof("success")
	} else {
		firlog.Logger.Infof("fail")
	}

	rt, err := ck.Get(key)
	if err != nil {
		firlog.Logger.Infoln(err)
	}
	firlog.Logger.Infof("get key[%s] value[%s]", key, rt)

	firlog.Logger.Infof("CAS set again key[%s] ori[%v] dest[%s]", key, nil, "B")
	ok, err = ck.CAS(key, nil, []byte("B"), 0)
	if err != nil {
		firlog.Logger.Infoln(err)
	}
	if ok {
		firlog.Logger.Infof("success")
	} else {
		firlog.Logger.Infof("fail")
	}

	rt, err = ck.Get(key)
	if err != nil {
		firlog.Logger.Infoln(err)
	}
	firlog.Logger.Infof("get key[%s] value[%s]", key, rt)
	firlog.Logger.Infof("CAS set again key[%s] ori[%s] dest[%s]", key, "A", "B")
	ok, err = ck.CAS(key, []byte("A"), []byte("B"), 0)
	if err != nil {
		firlog.Logger.Infoln(err)
	}
	if ok {
		firlog.Logger.Infof("success")
	} else {
		firlog.Logger.Infof("fail")
	}

	rt, err = ck.Get(key)
	if err != nil {
		firlog.Logger.Infoln(err)
	}
	firlog.Logger.Infof("get key[%s] value[%s]", key, rt)

	err = ck.Delete(key)
	if err != nil {
		firlog.Logger.Infoln(err)
	}
}
