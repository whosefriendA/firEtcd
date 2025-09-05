package client

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/whosefriendA/firEtcd/common"
	"github.com/whosefriendA/firEtcd/kvraft"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
	"github.com/whosefriendA/firEtcd/proto/pb"
	"google.golang.org/grpc/codes"
)

func Test_nrand(t *testing.T) {
	tests := []struct {
		name string
		want int64
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nrand(); got != tt.want {
				t.Errorf("nrand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithSendInitialState(t *testing.T) {
	type args struct {
		send bool
	}
	tests := []struct {
		name string
		args args
		want WatchOption
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithSendInitialState(tt.args.send); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithSendInitialState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithPrefix(t *testing.T) {
	tests := []struct {
		name string
		want WatchOption
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithPrefix(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_Watch(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		ctx  context.Context
		key  string
		opts []WatchOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    <-chan *WatchEvent
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			got, err := ck.Watch(tt.args.ctx, tt.args.key, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Watch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Watch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_manageWatchStream(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		ctx         context.Context
		watchKeyStr string
		req         *pb.WatchRequest
		eventChan   chan<- *WatchEvent
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			ck.manageWatchStream(tt.args.ctx, tt.args.watchKeyStr, tt.args.req, tt.args.eventChan)
		})
	}
}

func Test_shouldRetry(t *testing.T) {
	type args struct {
		code codes.Code
	}
	tests := []struct {
		name string
		args args
		want bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRetry(tt.args.code); got != tt.want {
				t.Errorf("shouldRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_watchEtcd(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			c.watchEtcd()
		})
	}
}

func TestMakeClerk(t *testing.T) {
	type args struct {
		conf firconfig.Clerk
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeClerk(tt.args.conf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeClerk() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_doGetValue(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		key        string
		withPrefix bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]byte
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			got, err := ck.doGetValue(tt.args.key, tt.args.withPrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.doGetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.doGetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_doGetKV(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		key        string
		withPrefix bool
		op         pb.OpType
		pageSize   int
		pageIndex  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []common.Pair
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			got, err := ck.doGetKV(tt.args.key, tt.args.withPrefix, tt.args.op, tt.args.pageSize, tt.args.pageIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.doGetKV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.doGetKV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_read(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		args *pb.GetArgs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]byte
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			got, err := ck.read(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_write(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		key      string
		value    []byte
		oriValue []byte
		TTL      time.Duration
		op       int32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			if err := ck.write(tt.args.key, tt.args.value, tt.args.oriValue, tt.args.TTL, tt.args.op); (err != nil) != tt.wantErr {
				t.Errorf("Client.write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClerk_changeNextSendId(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			ck.changeNextSendId()
		})
	}
}

func TestClerk_Put(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		key   string
		value []byte
		TTL   time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			if err := ck.Put(tt.args.key, tt.args.value, tt.args.TTL); (err != nil) != tt.wantErr {
				t.Errorf("Client.Put() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClerk_Append(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		key   string
		value []byte
		TTL   time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			if err := ck.Append(tt.args.key, tt.args.value, tt.args.TTL); (err != nil) != tt.wantErr {
				t.Errorf("Client.Append() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClerk_Delete(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			if err := ck.Delete(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Client.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClerk_DeleteWithPrefix(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		prefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			if err := ck.DeleteWithPrefix(tt.args.prefix); (err != nil) != tt.wantErr {
				t.Errorf("Client.DeleteWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClerk_CAS(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		key    string
		origin []byte
		dest   []byte
		TTL    time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			got, err := ck.CAS(tt.args.key, tt.args.origin, tt.args.dest, tt.args.TTL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.CAS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Client.CAS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_batchWrite(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		p *Pipe
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			if err := ck.BatchWrite(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("Client.BatchWrite() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClerk_Pipeline(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   *Pipe
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			if got := ck.Pipeline(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Pipeline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_Get(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			got, err := ck.Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_GetWithPrefix(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]byte
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			got, err := ck.GetWithPrefix(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetWithPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_Keys(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	tests := []struct {
		name    string
		fields  fields
		want    []common.Pair
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			got, err := ck.Keys()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Keys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Keys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_KVs(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	tests := []struct {
		name    string
		fields  fields
		want    []common.Pair
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			got, err := ck.KVs()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.KVs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.KVs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_KeysWithPage(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		pageSize  int
		pageIndex int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []common.Pair
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			got, err := ck.KeysWithPage(tt.args.pageSize, tt.args.pageIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.KeysWithPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.KeysWithPage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_KVsWithPage(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		pageSize  int
		pageIndex int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []common.Pair
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			got, err := ck.KVsWithPage(tt.args.pageSize, tt.args.pageIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.KVsWithPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.KVsWithPage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_Lock(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		key string
		TTL time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantId  string
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			gotId, err := ck.Lock(tt.args.key, tt.args.TTL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Lock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotId != tt.wantId {
				t.Errorf("Client.Lock() = %v, want %v", gotId, tt.wantId)
			}
		})
	}
}

func TestClerk_Unlock(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		key string
		id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}
			got, err := ck.Unlock(tt.args.key, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Unlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Client.Unlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClerk_WatchDog(t *testing.T) {
	type fields struct {
		servers         []*kvraft.KVconn
		nextSendLocalId int
		LatestOffset    int32
		clientId        int64
		cTos            []int
		sToc            []int
		conf            firconfig.Clerk
		mu              sync.Mutex
	}
	type args struct {
		key   string
		value []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "basic watchdog test",
			fields: fields{
				// 根据你真实需要初始化 fields，也可以传空值做 smoke test
				conf: firconfig.Clerk{}, // 示例
			},
			args: args{
				key:   "exampleKey",
				value: []byte("exampleValue"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := &Client{
				servers:         tt.fields.servers,
				nextSendLocalId: tt.fields.nextSendLocalId,
				LatestOffset:    tt.fields.LatestOffset,
				clientId:        tt.fields.clientId,
				cTos:            tt.fields.cTos,
				sToc:            tt.fields.sToc,
				conf:            tt.fields.conf,
				mu:              tt.fields.mu,
			}

			cancelFunc := ck.WatchDog(tt.args.key, tt.args.value)
			if cancelFunc == nil {
				t.Errorf("Client.WatchDog() returned nil cancelFunc")
				return
			}

			// 安全地调用一下，确认不会 panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Client.WatchDog() cancelFunc panicked: %v", r)
				}
			}()
			cancelFunc() // 调用返回的 cancel 函数
		})
	}
}
