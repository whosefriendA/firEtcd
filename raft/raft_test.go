package raft

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/whosefriendA/firEtcd/pkg/firconfig"
	"github.com/whosefriendA/firEtcd/proto/pb"
)

func TestRaft_GetCommitIndex(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
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
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			if got := rf.GetCommitIndex(); got != tt.want {
				t.Errorf("Raft.GetCommitIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaft_Applyer(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.Applyer()
		})
	}
}

func TestRaft_lastIndex(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
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
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			if got := rf.lastIndex(); got != tt.want {
				t.Errorf("Raft.lastIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaft_lastTerm(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
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
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			if got := rf.lastTerm(); got != tt.want {
				t.Errorf("Raft.lastTerm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaft_index2LogPos(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	type args struct {
		index int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantPos int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			if gotPos := rf.index2LogPos(tt.args.index); gotPos != tt.wantPos {
				t.Errorf("Raft.index2LogPos() = %v, want %v", gotPos, tt.wantPos)
			}
		})
	}
}

func TestRaft_GetState(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
		want1  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			got, got1 := rf.GetState()
			if got != tt.want {
				t.Errorf("Raft.GetState() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Raft.GetState() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestRaft_GetLeader(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			if got := rf.GetLeader(); got != tt.want {
				t.Errorf("Raft.GetLeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaft_GetTerm(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
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
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			if got := rf.GetTerm(); got != tt.want {
				t.Errorf("Raft.GetTerm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaft_persist(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.persist()
		})
	}
}

func TestRaft_persistWithSnapshot(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
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
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			if got := rf.persistWithSnapshot(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Raft.persistWithSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaft_readPersist(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
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
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.readPersist(tt.args.data)
		})
	}
}

func TestRaft_CopyEntries(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	type args struct {
		args *pb.AppendEntriesArgs
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
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.CopyEntries(tt.args.args)
		})
	}
}

func TestRaft_RequestVote(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	type args struct {
		in0  context.Context
		args *pb.RequestVoteArgs
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantReply *pb.RequestVoteReply
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			gotReply, err := rf.RequestVote(tt.args.in0, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Raft.RequestVote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotReply, tt.wantReply) {
				t.Errorf("Raft.RequestVote() = %v, want %v", gotReply, tt.wantReply)
			}
		})
	}
}

func TestRaft_AppendEntries(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	type args struct {
		in0  context.Context
		args *pb.AppendEntriesArgs
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantReply *pb.AppendEntriesReply
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			gotReply, err := rf.AppendEntries(tt.args.in0, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Raft.AppendEntries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotReply, tt.wantReply) {
				t.Errorf("Raft.AppendEntries() = %v, want %v", gotReply, tt.wantReply)
			}
		})
	}
}

func TestRaft_SnapshotInstall(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	type args struct {
		in0  context.Context
		args *pb.SnapshotInstallArgs
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantReply *pb.SnapshotInstallReply
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			gotReply, err := rf.SnapshotInstall(tt.args.in0, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Raft.SnapshotInstall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotReply, tt.wantReply) {
				t.Errorf("Raft.SnapshotInstall() = %v, want %v", gotReply, tt.wantReply)
			}
		})
	}
}

func TestRaft_installSnapshotToApplication(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.installSnapshotToApplication()
		})
	}
}

func TestRaft_sendInstallSnapshot(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	type args struct {
		server int
		args   *pb.SnapshotInstallArgs
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantReply *pb.SnapshotInstallReply
		wantOk    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			gotReply, gotOk := rf.sendInstallSnapshot(tt.args.server, tt.args.args)
			if !reflect.DeepEqual(gotReply, tt.wantReply) {
				t.Errorf("Raft.sendInstallSnapshot() gotReply = %v, want %v", gotReply, tt.wantReply)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Raft.sendInstallSnapshot() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestRaft_Snapshot(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	type args struct {
		index    int
		snapshot []byte
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
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.Snapshot(tt.args.index, tt.args.snapshot)
		})
	}
}

func TestRaft_updateCommitIndex(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.updateCommitIndex()
		})
	}
}

func TestRaft_undateLastApplied(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.undateLastApplied()
		})
	}
}

func TestRaft_Start(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	type args struct {
		command []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
		want1  int
		want2  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			got, got1, got2 := rf.Start(tt.args.command)
			if got != tt.want {
				t.Errorf("Raft.Start() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Raft.Start() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("Raft.Start() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestRaft_Kill(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.Kill()
		})
	}
}

func TestRaft_killed(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			if got := rf.killed(); got != tt.want {
				t.Errorf("Raft.killed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaft_sendRequestVote2(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	type args struct {
		server int
		args   *pb.RequestVoteArgs
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantReply *pb.RequestVoteReply
		wantOk    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			gotReply, gotOk := rf.sendRequestVote2(tt.args.server, tt.args.args)
			if !reflect.DeepEqual(gotReply, tt.wantReply) {
				t.Errorf("Raft.sendRequestVote2() gotReply = %v, want %v", gotReply, tt.wantReply)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Raft.sendRequestVote2() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestRaft_electionLoop(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.electionLoop()
		})
	}
}

func TestRaft_SendAppendEntriesToPeerId(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	type args struct {
		server       int
		applychreply *chan int
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
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.SendAppendEntriesToPeerId(tt.args.server, tt.args.applychreply)
		})
	}
}

func TestRaft_appendEntriesLoop(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.appendEntriesLoop()
		})
	}
}

func TestRaft_SendAppendEntriesToAll(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.SendAppendEntriesToAll()
		})
	}
}

func TestRaft_SendOnlyAppendEntriesToAll(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.SendOnlyAppendEntriesToAll()
		})
	}
}

func TestRaft_sendInstallSnapshotToPeerId(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	type args struct {
		server int
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
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rf.sendInstallSnapshotToPeerId(tt.args.server)
		})
	}
}

func TestRaft_IfNeedExceedLog(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	type args struct {
		logSize int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			if got := rf.IfNeedExceedLog(tt.args.logSize); got != tt.want {
				t.Errorf("Raft.IfNeedExceedLog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaft_GetleaderId(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
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
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			if got := rf.GetleaderId(); got != tt.want {
				t.Errorf("Raft.GetleaderId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaft_CheckIfDepose(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	tests := []struct {
		name    string
		fields  fields
		wantRet bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			if gotRet := rf.CheckIfDepose(); gotRet != tt.wantRet {
				t.Errorf("Raft.CheckIfDepose() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}

func TestMake(t *testing.T) {
	type args struct {
		me        int
		persister *Persister
		applyCh   chan ApplyMsg
		conf      firconfig.RaftEnds
	}
	tests := []struct {
		name string
		args args
		want *Raft
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Make(tt.args.me, tt.args.persister, tt.args.applyCh, tt.args.conf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Make() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaft_StartRaft(t *testing.T) {
	type fields struct {
		mu                    sync.Mutex
		peers                 []*RaftEnd
		persister             *Persister
		me                    int
		dead                  int32
		state                 int32
		currentTerm           int
		votedFor              int
		log                   []pb.LogType
		lastIncludeIndex      int
		lastIncludeTerm       int
		commitIndex           int
		lastApplied           int
		nextIndex             []int
		matchIndex            []int
		lastHearBeatTime      time.Time
		lastSendHeartbeatTime time.Time
		leaderId              int
		applyCh               chan ApplyMsg
		applyChTerm           chan ApplyMsg
		SnapshotDate          []byte
		IisBack               bool
		IisBackIndex          int
	}
	type args struct {
		conf firconfig.RaftEnd
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
			rt := &Raft{
				mu:                    tt.fields.mu,
				peers:                 tt.fields.peers,
				persister:             tt.fields.persister,
				me:                    tt.fields.me,
				dead:                  tt.fields.dead,
				state:                 tt.fields.state,
				currentTerm:           tt.fields.currentTerm,
				votedFor:              tt.fields.votedFor,
				log:                   tt.fields.log,
				lastIncludeIndex:      tt.fields.lastIncludeIndex,
				lastIncludeTerm:       tt.fields.lastIncludeTerm,
				commitIndex:           tt.fields.commitIndex,
				lastApplied:           tt.fields.lastApplied,
				nextIndex:             tt.fields.nextIndex,
				matchIndex:            tt.fields.matchIndex,
				lastHearBeatTime:      tt.fields.lastHearBeatTime,
				lastSendHeartbeatTime: tt.fields.lastSendHeartbeatTime,
				leaderId:              tt.fields.leaderId,
				applyCh:               tt.fields.applyCh,
				applyChTerm:           tt.fields.applyChTerm,
				SnapshotDate:          tt.fields.SnapshotDate,
				IisBack:               tt.fields.IisBack,
				IisBackIndex:          tt.fields.IisBackIndex,
			}
			rt.StartRaft(tt.args.conf)
		})
	}
}

func TestWaitConnect(t *testing.T) {
	type args struct {
		conf firconfig.RaftEnds
	}
	tests := []struct {
		name string
		args args
		want []*RaftEnd
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WaitConnect(tt.args.conf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WaitConnect() = %v, want %v", got, tt.want)
			}
		})
	}
}
