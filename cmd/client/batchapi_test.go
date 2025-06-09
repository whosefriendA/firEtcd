package client_test

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"strconv"
	"testing"
	"text/tabwriter"
	"time"

	"github.com/whosefriendA/firEtcd/kvraft"
	"github.com/whosefriendA/firEtcd/pkg/firlog"
)

func TestBatchApi(t *testing.T) {
	start := time.Now()
	size := 1000
	pipe := ck.Pipeline()
	pipe.DeleteWithPrefix("")
	for i := range size {
		pipe.Put("key"+strconv.Itoa(i), []byte(strconv.Itoa(i)), 0)
	}
	err := pipe.Exec()
	if err != nil {
		firlog.Logger.Fatalln(err)
	}
	firlog.Logger.Infoln("batch 1000 spand time:", time.Since(start))
	get, err := ck.Get("key1")
	if err != nil {
		firlog.Logger.Fatalln(err)
	}
	firlog.Logger.Infoln("key1=", get)

	rets, err := ck.GetWithPrefix("key")
	if err != nil {
		firlog.Logger.Fatalln(err)
	}
	if len(rets) != size {
		firlog.Logger.Errorln("batch 1000 key not correct real size", len(rets))
	}

}

func TestGetWithPrefix(t *testing.T) {
	ck.DeleteWithPrefix("")
	err := ck.Put("key:1", []byte("1"), 0)
	if err != nil {
		firlog.Logger.Fatalln(err)
	}
	err = ck.Put("key:2", []byte("2"), 0)
	if err != nil {
		firlog.Logger.Fatalln(err)
	}
	pipe := ck.Pipeline()
	pipe.Put("key:3", []byte("3"), 0)
	err = pipe.Exec()
	if err != nil {
		firlog.Logger.Fatalln(err)
	}

	rowdata, err := ck.GetWithPrefix("key")
	if err != nil {
		firlog.Logger.Fatalln(err)
	}
	if len(rowdata) != 3 {
		firlog.Logger.Fatalln("wrong count")
	}

	ck.DeleteWithPrefix("key")
	_, err = ck.GetWithPrefix("key")
	if err != kvraft.ErrNil {
		firlog.Logger.Fatalln(err)
	}

}

func TestKvs(t *testing.T) {
	ck.DeleteWithPrefix("")
	err := ck.Put("key:1", []byte("1"), 0)
	if err != nil {
		firlog.Logger.Fatalln(err)
	}
	err = ck.Put("key:2", []byte("2"), time.Second)
	if err != nil {
		firlog.Logger.Fatalln(err)
	}
	pipe := ck.Pipeline()
	pipe.Put("key:3", []byte("3"), time.Hour)
	err = pipe.Exec()
	if err != nil {
		firlog.Logger.Fatalln(err)
	}
	pairs, err := ck.KVs()
	if err != nil {
		firlog.Logger.Fatalln(err)
	}
	if len(pairs) != 3 {
		firlog.Logger.Fatalln("wrong count")
	}
	tab := tabwriter.NewWriter(os.Stdout, 15, 0, 1, ' ', 0)
	fmt.Fprintln(tab, "key\tvalue\tdeadtime\t")
	for i := range pairs {
		key := pairs[i].Key
		value := pairs[i].Entry.Value
		deadtime := pairs[i].Entry.DeadTime
		if key != "key:"+strconv.Itoa(i+1) {
			firlog.Logger.Fatalln("wrong key:", key, "expect:", "key:"+strconv.Itoa(i+1))
		}
		fmt.Fprintf(tab, "%s\t%s\t%d\t\n", key, value, deadtime)
	}
	tab.Flush()
	ck.DeleteWithPrefix("key")
	_, err = ck.KVs()
	if err != kvraft.ErrNil {
		firlog.Logger.Fatalln(err)
	}

}

func TestSlice(t *testing.T) {
	value := make([][]byte, 0)
	var buf bytes.Buffer
	for i := range 3 {
		gob.NewEncoder(&buf).Encode(i)
		value = append(value, buf.Bytes())
		firlog.Logger.Infof("%p ", value[i])
		buf.Reset()
	}
	firlog.Logger.Infof("hack\n")
	value = make([][]byte, 0, 3)
	tmp := make([]byte, 0, 64)
	for i := range 3 {
		var buf = bytes.NewBuffer(tmp)
		gob.NewEncoder(buf).Encode(i)
		value = append(value, buf.Bytes())
		firlog.Logger.Infof("%p ", value[i])
		tmp = tmp[:0]
	}
}

func TestPrintKeys(t *testing.T) {
	pairs, err := ck.KVs()
	if err != nil {
		if err != kvraft.ErrNil {
			firlog.Logger.Fatalln(err)
		}
		firlog.Logger.Infoln("no kv exist")
	}
	tab := tabwriter.NewWriter(os.Stdout, 15, 0, 1, ' ', 0)
	fmt.Fprintln(tab, "Key\tValue\tDeadtime\t")
	for i := range pairs {
		fmt.Fprintf(tab, "%s\t%s\t%d\t\n", pairs[i].Key, pairs[i].Entry.Value, pairs[i].Entry.DeadTime)
	}
	tab.Flush()

}

// func BenchmarkSingleApi(t *testing.B) {
// 	start := time.Now()
// 	for i := range t.N {
// 		ck.Put("key"+strconv.Itoa(i), strconv.Itoa(i), 0)
// 	}
// 	firLog.Logger.Infof("single %f spand time:%v", t.N, time.Since(start))
// }
