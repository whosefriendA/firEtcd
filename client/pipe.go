package client

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/whosefriendA/firEtcd/common"
	"github.com/whosefriendA/firEtcd/pkg/firlog"
	"github.com/whosefriendA/firEtcd/proto/pb"
	"github.com/whosefriendA/firEtcd/raft"
)

type Pipe struct {
	Ops  []raft.Op
	size int
	//ck   *Clerk
	ck BatchWriter
}

func (p *Pipe) Size() int {
	return p.size
}

func (p *Pipe) Marshal() []byte {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	e.Encode(p.Ops)
	return b.Bytes()
}

func (p *Pipe) UnMarshal(data []byte) {
	var ops []raft.Op
	b := bytes.NewBuffer(data)
	d := gob.NewDecoder(b)
	err := d.Decode(&ops)
	if err != nil {
		firlog.Logger.Fatalln("raw data:", data, err)
	}
	p.Ops = ops
}

func (p *Pipe) Delete(key string) error {
	op := raft.Op{
		Key:    key,
		OpType: int32(pb.OpType_DelT),
	}
	return p.append(op)
}

func (p *Pipe) DeleteWithPrefix(prefix string) error {
	op := raft.Op{
		Key:    prefix,
		OpType: int32(pb.OpType_DelWithPrefix),
	}
	return p.append(op)
}
func (p *Pipe) Put(key string, value []byte, TTL time.Duration) error {
	op := raft.Op{
		Key: key,
		Entry: common.Entry{
			Value:    value,
			DeadTime: 0,
		},
		OpType: int32(pb.OpType_PutT),
	}
	if TTL != 0 {
		op.Entry.DeadTime = time.Now().Add(TTL).UnixMilli()
		// naive per-op lease grant: the actual client-side plumping may batch later
		// This requires the concrete Clerk behind BatchWriter to handle lease creation when executing.
	}
	return p.append(op)
}

func (p *Pipe) Append(key string, value []byte, TTL time.Duration) error {
	op := raft.Op{
		Key: key,
		Entry: common.Entry{
			Value:    value,
			DeadTime: time.Now().Add(TTL).UnixMilli(),
		},
		OpType: int32(pb.OpType_AppendT),
	}
	return p.append(op)
}

func (p *Pipe) append(op raft.Op) error {
	p.Ops = append(p.Ops, op)
	p.size += op.Size()
	if p.size > pipeLimit {
		return fmt.Errorf("too many pipeline data maxLimit:%d", pipeLimit)
	}
	return nil
}

func (p *Pipe) Exec() error {
	if len(p.Ops) == 0 { // 假设字段是公开的 Ops
		return nil
	}

	return p.ck.BatchWrite(p)
}

func NewPipe(ck BatchWriter) *Pipe {
	return &Pipe{
		ck:   ck,
		Ops:  make([]raft.Op, 0),
		size: 0, // 根据你的实际实现初始化
	}
}
