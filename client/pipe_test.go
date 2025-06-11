package client

import (
	"reflect"
	"testing"
	"time"

	"github.com/whosefriendA/firEtcd/raft"
)

func TestPipe_Size(t *testing.T) {
	type fields struct {
		ops  []raft.Op
		size int
		ck   *Clerk
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipe{
				ops:  tt.fields.ops,
				size: tt.fields.size,
				ck:   tt.fields.ck,
			}
			if got := p.Size(); got != tt.want {
				t.Errorf("Pipe.Size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipe_Marshal(t *testing.T) {
	type fields struct {
		ops  []raft.Op
		size int
		ck   *Clerk
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipe{
				ops:  tt.fields.ops,
				size: tt.fields.size,
				ck:   tt.fields.ck,
			}
			if got := p.Marshal(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pipe.Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipe_UnMarshal(t *testing.T) {
	type fields struct {
		ops  []raft.Op
		size int
		ck   *Clerk
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipe{
				ops:  tt.fields.ops,
				size: tt.fields.size,
				ck:   tt.fields.ck,
			}
			p.UnMarshal(tt.args.data)
		})
	}
}

func TestPipe_Delete(t *testing.T) {
	type fields struct {
		ops  []raft.Op
		size int
		ck   *Clerk
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipe{
				ops:  tt.fields.ops,
				size: tt.fields.size,
				ck:   tt.fields.ck,
			}
			if err := p.Delete(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Pipe.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPipe_DeleteWithPrefix(t *testing.T) {
	type fields struct {
		ops  []raft.Op
		size int
		ck   *Clerk
	}
	type args struct {
		prefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipe{
				ops:  tt.fields.ops,
				size: tt.fields.size,
				ck:   tt.fields.ck,
			}
			if err := p.DeleteWithPrefix(tt.args.prefix); (err != nil) != tt.wantErr {
				t.Errorf("Pipe.DeleteWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPipe_Put(t *testing.T) {
	type fields struct {
		ops  []raft.Op
		size int
		ck   *Clerk
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
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipe{
				ops:  tt.fields.ops,
				size: tt.fields.size,
				ck:   tt.fields.ck,
			}
			if err := p.Put(tt.args.key, tt.args.value, tt.args.TTL); (err != nil) != tt.wantErr {
				t.Errorf("Pipe.Put() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPipe_Append(t *testing.T) {
	type fields struct {
		ops  []raft.Op
		size int
		ck   *Clerk
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
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipe{
				ops:  tt.fields.ops,
				size: tt.fields.size,
				ck:   tt.fields.ck,
			}
			if err := p.Append(tt.args.key, tt.args.value, tt.args.TTL); (err != nil) != tt.wantErr {
				t.Errorf("Pipe.Append() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPipe_append(t *testing.T) {
	type fields struct {
		ops  []raft.Op
		size int
		ck   *Clerk
	}
	type args struct {
		op raft.Op
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipe{
				ops:  tt.fields.ops,
				size: tt.fields.size,
				ck:   tt.fields.ck,
			}
			if err := p.append(tt.args.op); (err != nil) != tt.wantErr {
				t.Errorf("Pipe.append() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPipe_Exec(t *testing.T) {
	type fields struct {
		ops  []raft.Op
		size int
		ck   *Clerk
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipe{
				ops:  tt.fields.ops,
				size: tt.fields.size,
				ck:   tt.fields.ck,
			}
			if err := p.Exec(); (err != nil) != tt.wantErr {
				t.Errorf("Pipe.Exec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
