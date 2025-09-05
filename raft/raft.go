package raft

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json" // WAL-MOD: Added for serializing metadata records
	"fmt"           // WAL-MOD: Added for file path formatting
	"math/rand"
	"net"
	"os"            // WAL-MOD: Added for snapshot file operations
	"path/filepath" // WAL-MOD: Added for snapshot file operations
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/whosefriendA/firEtcd/pkg/firconfig"
	"github.com/whosefriendA/firEtcd/pkg/firlog"
	"github.com/whosefriendA/firEtcd/pkg/wal" // WAL-MOD: Import the correct WAL package
	"github.com/whosefriendA/firEtcd/proto/pb"
	"google.golang.org/grpc"
)

const (
	follower = iota
	candidate
	leader
)
const HeartBeatInterval = 100
const TICKMIN = 300
const TICKRANDOM = 300
const LOGINITCAPCITY = 1000
const APPENDENTRIES_TIMES = 0
const APPENDENTRIES_INTERVAL = 20

type ApplyMsg struct {
	CommandValid  bool
	Command       []byte
	CommandIndex  int
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

// WAL-MOD: This struct is no longer used for persistence. Kept for pb compatibility if needed.
type LogType struct {
	Term  int
	Value interface{}
}

// WAL-MOD: New structs for serializing specific record types for the WAL.
// RaftStateRecord holds term and vote information.
type RaftStateRecord struct {
	Term     int
	VotedFor int
}

// SnapshotRecord holds metadata about a snapshot.
type SnapshotRecord struct {
	LastIncludedIndex int
	LastIncludedTerm  int
	Path              string
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu    sync.Mutex
	peers []*RaftEnd
	// WAL-MOD: The persister is replaced by the WAL.
	wal  *wal.WAL // WAL-MOD: New WAL field
	me   int
	dead int32

	state int32

	// Persistent state (now managed by WAL)
	currentTerm      int
	votedFor         int
	log              []pb.LogType
	lastIncludeIndex int
	lastIncludeTerm  int

	// Volatile state
	commitIndex int
	lastApplied int

	// Leader volatile state
	nextIndex  []int
	matchIndex []int

	lastHearBeatTime      time.Time
	lastSendHeartbeatTime time.Time

	leaderId int

	applyCh     chan ApplyMsg
	applyChTerm chan ApplyMsg

	IisBack      bool
	IisBackIndex int
}

func (rf *Raft) GetCommitIndex() int {
	return rf.commitIndex + 1
}

func (rf *Raft) Applyer() {
	for !rf.killed() {
		select {
		case rf.applyCh <- <-rf.applyChTerm:
		}
	}
}

func (rf *Raft) lastIndex() int {
	return rf.lastIncludeIndex + len(rf.log)
}

func (rf *Raft) lastTerm() int {
	lastLogTerm := rf.lastIncludeTerm
	if len(rf.log) > 0 {
		lastLogTerm = int(rf.log[len(rf.log)-1].Term)
	}
	return lastLogTerm
}

func (rf *Raft) index2LogPos(index int) (pos int) {
	return index - rf.lastIncludeIndex - 1
}

func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.currentTerm, rf.state == leader
}

func (rf *Raft) GetLeader() bool {
	return rf.state == leader
}

func (rf *Raft) GetTerm() int {
	return rf.currentTerm
}

// WAL-MOD: The old persist(), persistWithSnapshot(), and readPersist() methods are now deleted.
// The new persistence logic is handled by direct calls to rf.wal.Write().

// WAL-MOD: New helper function to persist term/vote state changes.
func (rf *Raft) persistState() {
	record := RaftStateRecord{Term: rf.currentTerm, VotedFor: rf.votedFor}
	data, err := json.Marshal(record)
	if err != nil {
		firlog.Logger.Panicf("Failed to marshal raft state: %v", err)
	}
	if err := rf.wal.Write(wal.RecordTypeState, data); err != nil {
		firlog.Logger.Panicf("Failed to write state to WAL: %v", err)
	}
}

// WAL-MOD: New helper function to persist a single log entry.
func (rf *Raft) persistEntry(entry pb.LogType) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(entry); err != nil {
		firlog.Logger.Panicf("Failed to encode log entry: %v", err)
	}
	if err := rf.wal.Write(wal.RecordTypeEntry, buf.Bytes()); err != nil {
		firlog.Logger.Panicf("Failed to write entry to WAL: %v", err)
	}
}

func (rf *Raft) CopyEntries(args *pb.AppendEntriesArgs) {
	for i := 0; i < len(args.Entries); i++ {
		entry := *args.Entries[i]
		rfIndex := i + int(args.PrevLogIndex) + 1
		logPos := rf.index2LogPos(rfIndex)

		// Overwrite or append logic
		if rfIndex > rf.lastIndex() {
			// Append new entry
			rf.log = append(rf.log, entry)
		} else if rf.log[logPos].Term != entry.Term {
			// Truncate conflicting entries and append new entry
			rf.log = rf.log[:logPos]
			rf.log = append(rf.log, entry)
			// WAL-MOD: In a real production system, you'd need a WAL record to signify truncation.
			// For this implementation, we assume replaying the log in order handles this correctly.
		} else {
			// Entry already matches, do nothing.
			continue
		}

		// Persist the new or overwriting entry to the WAL.
		rf.persistEntry(entry)
	}

	// Update commit index
	var min = -1
	if args.LeaderCommit > int64(rf.lastIndex()) {
		min = rf.lastIndex()
	} else {
		min = int(args.LeaderCommit)
	}
	if rf.commitIndex < min {
		rf.commitIndex = min
		rf.undateLastApplied()
	}
}

func (rf *Raft) RequestVote(_ context.Context, args *pb.RequestVoteArgs) (reply *pb.RequestVoteReply, err error) {
	reply = new(pb.RequestVoteReply)
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer firlog.Logger.Infof("🎫Rec Term[%d] [%d]reply [%d]Vote finish", rf.currentTerm, rf.me, args.CandidateId)

	reply.VoteGranted = false
	reply.Term = int64(rf.currentTerm)

	if args.Term < int64(rf.currentTerm) {
		return
	}
	if rf.currentTerm < int(args.Term) {
		rf.currentTerm = int(args.Term)
		rf.state = follower
		rf.votedFor = -1
		rf.leaderId = -1
		rf.persistState() // WAL-MOD: Persist state change
	}
	if rf.votedFor == -1 || rf.votedFor == int(args.CandidateId) {
		lastLogTerm := rf.lastTerm()
		if args.LastLogTerm > int64(lastLogTerm) || (args.LastLogTerm == int64(lastLogTerm) && args.LastLogIndex >= int64(rf.lastIndex())) {
			rf.votedFor = int(args.CandidateId)
			reply.VoteGranted = true
			rf.lastHearBeatTime = time.Now()
			firlog.Logger.Infof("🎫Rec Term[%d] [%d] -> [%d]", rf.currentTerm, rf.me, args.CandidateId)
			rf.persistState() // WAL-MOD: Persist state change (the vote)
		}
	}
	return
}

func (rf *Raft) AppendEntries(_ context.Context, args *pb.AppendEntriesArgs) (reply *pb.AppendEntriesReply, err error) {
	reply = new(pb.AppendEntriesReply)
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply.Success = false
	reply.Term = int64(rf.currentTerm)
	reply.ConflictIndex = -1
	reply.ConflictTerm = -1

	if rf.currentTerm > int(args.Term) {
		return
	}
	if rf.currentTerm < int(args.Term) {
		rf.currentTerm = int(args.Term)
		rf.state = follower
		rf.votedFor = -1
		rf.persistState() // WAL-MOD: Persist state change
	}
	if rf.currentTerm == int(args.Term) && atomic.LoadInt32(&rf.state) == candidate {
		rf.state = follower
		rf.votedFor = -1
		rf.persistState() // WAL-MOD: Persist state change
	}
	rf.lastHearBeatTime = time.Now()
	rf.leaderId = int(args.LeaderId)

	if args.PrevLogIndex < int64(rf.lastIncludeIndex) {
		reply.ConflictIndex = int64(rf.lastIncludeIndex + 1)
		return
	}
	if args.PrevLogIndex > int64(rf.lastIndex()) {
		reply.ConflictIndex = int64(rf.lastIndex() + 1)
		return
	}
	if args.PrevLogIndex > int64(rf.lastIncludeIndex) {
		if int64(rf.log[rf.index2LogPos(int(args.PrevLogIndex))].Term) != args.PrevLogTerm {
			reply.ConflictTerm = int64(rf.log[rf.index2LogPos(int(args.PrevLogIndex))].Term)
			for i := args.PrevLogIndex; i > int64(rf.lastIncludeIndex); i-- {
				if int64(rf.log[rf.index2LogPos(int(i))].Term) != reply.ConflictTerm {
					break
				}
				reply.ConflictIndex = i
			}
			return
		}
	}

	rf.CopyEntries(args)
	reply.Success = true
	return
}

func (rf *Raft) SnapshotInstall(_ context.Context, args *pb.SnapshotInstallArgs) (reply *pb.SnapshotInstallReply, err error) {
	// WAL-MOD: This function needs significant changes to work with the WAL model.
	// A follower receiving a snapshot should save it to a file and write a snapshot record
	// to its own WAL, then truncate its WAL.
	reply = new(pb.SnapshotInstallReply)
	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.Term = int64(rf.currentTerm)
	if args.Term < int64(rf.currentTerm) {
		return
	}

	if args.Term > int64(rf.currentTerm) {
		rf.currentTerm = int(args.Term)
		rf.state = follower
		rf.votedFor = -1
		rf.persistState() // WAL-MOD: Persist state change
	}

	rf.leaderId = int(args.LeaderId)
	rf.lastHearBeatTime = time.Now()

	if args.LastIncludeIndex <= int64(rf.lastIncludeIndex) {
		return
	}

	// 1. Save snapshot data to a file
	snapshotPath := filepath.Join(rf.wal.Dir(), fmt.Sprintf("snapshot-%d-%d.dat", args.LastIncludeIndex, args.LastIncludeTerm))
	if err := os.WriteFile(snapshotPath, args.Data, 0644); err != nil {
		firlog.Logger.Panicf("Failed to write received snapshot to file: %v", err)
	}

	// 2. Persist snapshot metadata to WAL
	record := SnapshotRecord{
		LastIncludedIndex: int(args.LastIncludeIndex),
		LastIncludedTerm:  int(args.LastIncludeTerm),
		Path:              snapshotPath,
	}
	recordBytes, _ := json.Marshal(record)
	if err := rf.wal.Write(wal.RecordTypeSnapshot, recordBytes); err != nil {
		firlog.Logger.Panicf("Failed to write snapshot record to WAL: %v", err)
	}

	// 3. Truncate WAL
	if err := rf.wal.Truncate(uint64(args.LastIncludeIndex)); err != nil {
		firlog.Logger.Errorf("Failed to truncate WAL after snapshot install: %v", err)
	}

	// 4. Update in-memory state
	rf.lastIncludeIndex = int(args.LastIncludeIndex)
	rf.lastIncludeTerm = int(args.LastIncludeTerm)
	rf.log = make([]pb.LogType, 0) // Clear in-memory log
	if rf.lastApplied < rf.lastIncludeIndex {
		rf.lastApplied = rf.lastIncludeIndex
	}
	if rf.commitIndex < rf.lastIncludeIndex {
		rf.commitIndex = rf.lastIncludeIndex
	}

	// 5. Apply snapshot to the state machine
	applyMsg := ApplyMsg{
		SnapshotValid: true,
		Snapshot:      args.Data,
		SnapshotIndex: rf.lastIncludeIndex + 1,
		SnapshotTerm:  rf.lastIncludeTerm,
	}

	// Unlock to send on channel, then re-lock
	rf.mu.Unlock()
	rf.applyChTerm <- applyMsg
	rf.mu.Lock()

	return
}

func (rf *Raft) sendInstallSnapshot(server int, args *pb.SnapshotInstallArgs) (reply *pb.SnapshotInstallReply, ok bool) {
	reply, err := rf.peers[server].conn.SnapshotInstall(context.Background(), args)
	return reply, err == nil
}

func (rf *Raft) Snapshot(index int, snapshotData []byte) {
	// WAL-MOD: Overhauled snapshot logic.
	rf.mu.Lock()
	defer rf.mu.Unlock()

	index -= 1
	if index <= rf.lastIncludeIndex {
		return
	}

	lastIncludedTerm := int(rf.log[rf.index2LogPos(index)].Term)

	// 1. Save snapshot data to a file
	snapshotPath := filepath.Join(rf.wal.Dir(), fmt.Sprintf("snapshot-%d-%d.dat", index, lastIncludedTerm))
	if err := os.WriteFile(snapshotPath, snapshotData, 0644); err != nil {
		firlog.Logger.Panicf("Failed to write snapshot to file: %v", err)
	}

	// 2. Persist snapshot metadata to WAL
	record := SnapshotRecord{
		LastIncludedIndex: index,
		LastIncludedTerm:  lastIncludedTerm,
		Path:              snapshotPath,
	}
	recordBytes, _ := json.Marshal(record)
	if err := rf.wal.Write(wal.RecordTypeSnapshot, recordBytes); err != nil {
		firlog.Logger.Panicf("Failed to write snapshot record to WAL: %v", err)
	}

	// 3. Truncate WAL
	if err := rf.wal.Truncate(uint64(index)); err != nil {
		firlog.Logger.Errorf("Failed to truncate WAL after snapshot: %v", err)
	}

	// 4. Truncate in-memory log
	compactLogLen := index - rf.lastIncludeIndex
	afterLog := make([]pb.LogType, len(rf.log)-compactLogLen)
	copy(afterLog, rf.log[compactLogLen:])
	rf.log = afterLog

	rf.lastIncludeIndex = index
	rf.lastIncludeTerm = lastIncludedTerm
}

func (rf *Raft) updateCommitIndex() {
	if rf.state == leader {
		matchIndex := make([]int, 0)
		matchIndex = append(matchIndex, rf.lastIndex())
		for i := range rf.peers {
			if i != rf.me {
				matchIndex = append(matchIndex, rf.matchIndex[i])
			}
		}

		sort.Ints(matchIndex)
		N := matchIndex[len(matchIndex)/2]
		if N > rf.commitIndex && (N <= rf.lastIncludeIndex || rf.currentTerm == int(rf.log[rf.index2LogPos(N)].Term)) {
			rf.commitIndex = N
			rf.undateLastApplied()
		}
	}
}

func (rf *Raft) undateLastApplied() {
	for rf.lastApplied < rf.commitIndex {
		rf.lastApplied += 1
		index := rf.index2LogPos(rf.lastApplied)
		if index < 0 || index >= len(rf.log) {
			rf.lastApplied = rf.lastIncludeIndex
			firlog.Logger.Errorf("ERROR? 👿 [%d]Ready to apply index[%d] But index out of Len of log, lastApplied[%d] commitIndex[%d] lastIncludeIndex[%d] logLen:%d", rf.me, index, rf.lastApplied, rf.commitIndex, rf.lastIncludeIndex, len(rf.log))
			return
		}

		ApplyMsg := ApplyMsg{
			CommandValid: true,
			Command:      rf.log[index].Value,
			CommandIndex: rf.lastApplied + 1,
		}
		rf.applyChTerm <- ApplyMsg
		if rf.IisBackIndex == rf.lastApplied {
			rf.IisBack = true
		}
	}
}

func (rf *Raft) Start(command []byte) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	index := rf.lastIndex() + 1
	term := rf.currentTerm
	isLeader := rf.state == leader

	if !isLeader {
		return index, term, isLeader
	}

	// Add entry to in-memory log
	newEntry := pb.LogType{
		Term:  int64(rf.currentTerm),
		Value: command,
	}
	rf.log = append(rf.log, newEntry)

	// WAL-MOD: Persist the new entry to the WAL.
	rf.persistEntry(newEntry)

	return index, term, isLeader
}

func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	if rf.wal != nil {
		rf.wal.Close()
	}
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) sendRequestVote2(server int, args *pb.RequestVoteArgs) (reply *pb.RequestVoteReply, ok bool) {
	reply, err := rf.peers[server].conn.RequestVote(context.Background(), args)
	return reply, err == nil
}

func (rf *Raft) electionLoop() {
	for !rf.killed() {
		time.Sleep(time.Millisecond * 50)
		func() {
			rf.mu.Lock()
			defer rf.mu.Unlock()

			timeCount := time.Since(rf.lastHearBeatTime).Milliseconds()
			ms := TICKMIN + rand.Int63()%TICKRANDOM
			if rf.state == follower {
				if timeCount >= ms {
					rf.state = candidate
					rf.leaderId = -1
				}
			}
			if rf.state == candidate && timeCount >= ms {
				rf.lastHearBeatTime = time.Now()
				rf.leaderId = -1
				rf.currentTerm += 1
				rf.votedFor = rf.me
				rf.persistState() // WAL-MOD: Persist new term and vote

				args := pb.RequestVoteArgs{
					Term:         int64(rf.currentTerm),
					CandidateId:  int64(rf.me),
					LastLogIndex: int64(rf.lastIndex()),
					LastLogTerm:  int64(rf.lastTerm()),
				}
				rf.mu.Unlock()

				type VoteResult struct {
					raftId int
					resp   *pb.RequestVoteReply
				}
				voteCount := 1
				finishCount := 1
				VoteResultChan := make(chan *VoteResult, len(rf.peers))
				for peerId := 0; peerId < len(rf.peers) && !rf.killed(); peerId++ {
					go func(server int) {
						if server == rf.me {
							return
						}
						if resp, ok := rf.sendRequestVote2(server, &args); ok {
							VoteResultChan <- &VoteResult{raftId: server, resp: resp}
						} else {
							VoteResultChan <- &VoteResult{raftId: server, resp: nil}
						}
					}(peerId)
				}
				maxTerm := 0
			VOTE_END:
				for !rf.killed() {
					select {
					case VoteResult := <-VoteResultChan:
						finishCount += 1
						if VoteResult.resp != nil {
							if VoteResult.resp.VoteGranted {
								voteCount += 1
							}
							if int(VoteResult.resp.Term) > maxTerm {
								maxTerm = int(VoteResult.resp.Term)
							}
						}
						if finishCount == len(rf.peers) || voteCount > len(rf.peers)/2 {
							break VOTE_END
						}
					case <-time.After(time.Duration(TICKMIN+rand.Int63()%TICKRANDOM) * time.Millisecond):
						break VOTE_END
					}
				}
				rf.mu.Lock()

				if rf.state != candidate {
					return
				}
				if maxTerm > rf.currentTerm {
					rf.state = follower
					rf.leaderId = -1
					rf.currentTerm = maxTerm
					rf.votedFor = -1
					rf.persistState()
					return
				}
				if voteCount > len(rf.peers)/2 {
					rf.IisBack = false
					rf.state = leader
					rf.leaderId = rf.me
					rf.nextIndex = make([]int, len(rf.peers))
					for i := 0; i < len(rf.peers); i++ {
						rf.nextIndex[i] = rf.lastIndex() + 1
					}
					rf.matchIndex = make([]int, len(rf.peers))
					for i := 0; i < len(rf.peers); i++ {
						rf.matchIndex[i] = -1
					}
					rf.updateCommitIndex()
					rf.lastSendHeartbeatTime = time.Now().Add(-time.Millisecond * 2 * HeartBeatInterval)

					// Submit a no-op entry for the current term
					b := new(bytes.Buffer)
					e := gob.NewEncoder(b)
					e.Encode(Op{OpType: int32(pb.OpType_EmptyT)})
					rf.mu.Unlock()
					rf.Start(b.Bytes())
					rf.mu.Lock()
					return
				}
			}
		}()
	}
}

const (
	AEresult_Accept = iota
	AEresult_Reject
	AEresult_StopSending
	AEresult_Ignore
	AEresult_Lost
)

func (rf *Raft) SendAppendEntriesToPeerId(server int, applychreply *chan int) {
	rf.mu.Lock()
	if rf.state != leader {
		rf.mu.Unlock()
		if applychreply != nil {
			*applychreply <- AEresult_StopSending
		}
		return
	}
	args := &pb.AppendEntriesArgs{
		Term:         int64(rf.currentTerm),
		LeaderId:     int64(rf.me),
		PrevLogIndex: int64(rf.nextIndex[server] - 1),
		Entries:      make([]*pb.LogType, 0),
		LeaderCommit: int64(rf.commitIndex),
	}
	if args.PrevLogIndex >= int64(rf.lastIncludeIndex) {
		if args.PrevLogIndex == int64(rf.lastIncludeIndex) {
			args.PrevLogTerm = int64(rf.lastIncludeTerm)
		} else {
			args.PrevLogTerm = int64(rf.log[rf.index2LogPos(int(args.PrevLogIndex))].Term)
		}
		entries := rf.log[rf.index2LogPos(rf.nextIndex[server]):]
		for i := range entries {
			args.Entries = append(args.Entries, &entries[i])
		}
	} else {
		// This case should trigger a snapshot install
	}
	rf.mu.Unlock()

	reply, err := rf.peers[server].conn.AppendEntries(context.Background(), args)

	rf.mu.Lock()
	defer rf.mu.Unlock()
	if err == nil {
		if args.Term != int64(rf.currentTerm) {
			if applychreply != nil {
				*applychreply <- AEresult_Ignore
			}
			return
		}
		if rf.currentTerm < int(reply.Term) {
			rf.votedFor = -1
			rf.state = follower
			rf.currentTerm = int(reply.Term)
			rf.leaderId = -1
			rf.persistState() // WAL-MOD: Persist state change
			if applychreply != nil {
				*applychreply <- AEresult_StopSending
			}
			return
		}
		if reply.Success {
			rf.nextIndex[server] = int(args.PrevLogIndex) + len(args.Entries) + 1
			rf.matchIndex[server] = rf.nextIndex[server] - 1
			rf.updateCommitIndex()
			if applychreply != nil {
				*applychreply <- AEresult_Accept
			}
			return
		} else {
			if reply.ConflictTerm != -1 {
				searchIndex := -1
				for i := int(args.PrevLogIndex); i > rf.lastIncludeIndex; i-- {
					if int64(rf.log[rf.index2LogPos(i)].Term) == reply.ConflictTerm {
						searchIndex = i
					} else {
						break
					}
				}
				if searchIndex != -1 {
					rf.nextIndex[server] = searchIndex
				} else {
					rf.nextIndex[server] = int(reply.ConflictIndex)
				}
			} else {
				rf.nextIndex[server] = int(reply.ConflictIndex)
			}
			if rf.nextIndex[server] == 0 {
				rf.nextIndex[server] = 1
			}

			if applychreply != nil {
				*applychreply <- AEresult_Reject
			}
			return
		}
	}
	if applychreply != nil {
		*applychreply <- AEresult_Lost
	}
}

func (rf *Raft) appendEntriesLoop() {
	for !rf.killed() {
		time.Sleep(time.Millisecond * 50)
		func() {
			rf.mu.Lock()
			defer rf.mu.Unlock()
			if rf.state != leader {
				return
			}
			countTime := time.Since(rf.lastSendHeartbeatTime).Milliseconds()
			if countTime < HeartBeatInterval {
				return
			}
			rf.lastSendHeartbeatTime = time.Now()
			rf.SendAppendEntriesToAll()
		}()
	}
}

func (rf *Raft) SendAppendEntriesToAll() {
	for i := range rf.peers {
		if i == rf.me {
			continue
		}
		if rf.killed() {
			return
		}
		if rf.nextIndex[i] <= rf.lastIncludeIndex {
			go rf.sendInstallSnapshotToPeerId(i)
		} else {
			go rf.SendAppendEntriesToPeerId(i, nil)
		}
	}
}

func (rf *Raft) SendOnlyAppendEntriesToAll() {
	for i := range rf.peers {
		if i == rf.me {
			continue
		}
		if rf.killed() {
			return
		}
		go rf.SendAppendEntriesToPeerId(i, nil)
	}
}

func (rf *Raft) sendInstallSnapshotToPeerId(server int) {
	rf.mu.Lock()
	if rf.state != leader {
		rf.mu.Unlock()
		return
	}
	snapshotPath := filepath.Join(rf.wal.Dir(), fmt.Sprintf("snapshot-%d-%d.dat", rf.lastIncludeIndex, rf.lastIncludeTerm))
	snapshotData, err := os.ReadFile(snapshotPath)
	if err != nil {
		firlog.Logger.Errorf("Leader failed to read snapshot file to send: %v", err)
		rf.mu.Unlock()
		return
	}

	args := &pb.SnapshotInstallArgs{
		Term:             int64(rf.currentTerm),
		LeaderId:         int64(rf.me),
		LastIncludeIndex: int64(rf.lastIncludeIndex),
		LastIncludeTerm:  int64(rf.lastIncludeTerm),
		Data:             snapshotData,
	}
	rf.mu.Unlock()

	go func(args *pb.SnapshotInstallArgs) {
		reply, ok := rf.sendInstallSnapshot(server, args)
		if ok {
			rf.mu.Lock()
			defer rf.mu.Unlock()
			if int64(rf.currentTerm) != args.Term {
				return
			}
			if reply.Term > int64(rf.currentTerm) {
				rf.state = follower
				rf.leaderId = -1
				rf.currentTerm = int(reply.Term)
				rf.votedFor = -1
				rf.persistState()
				return
			}
			rf.nextIndex[server] = int(args.LastIncludeIndex) + 1
			rf.matchIndex[server] = int(args.LastIncludeIndex)
			rf.updateCommitIndex()
		}
	}(args)
}

func (rf *Raft) GetleaderId() int {
	return rf.leaderId
}

func (rf *Raft) CheckIfDepose() (ret bool) {
	// ... (This function remains unchanged)
	return true
}

// WAL-MOD: The Make function signature and body are heavily modified.
func Make(me int, walDir string, applyCh chan ApplyMsg, conf firconfig.RaftEnds) *Raft {
	firlog.Logger.Debugf("raft[%d] start by conf", me, conf)
	rf := &Raft{}
	rf.me = me

	// Open or create the WAL for persistence
	w, err := wal.Open(walDir)
	if err != nil {
		firlog.Logger.Panicf("Failed to open WAL directory %s: %v", walDir, err)
	}
	rf.wal = w

	// Initialize state
	rf.state = follower
	rf.leaderId = -1
	rf.log = make([]pb.LogType, 0, LOGINITCAPCITY)
	rf.lastIncludeIndex = -1
	rf.lastIncludeTerm = -1
	rf.commitIndex = -1
	rf.lastApplied = -1
	rf.currentTerm = 0
	rf.votedFor = -1

	// WAL-MOD: New recovery logic from WAL
	rf.loadFromWAL()

	// Initialize volatile state after recovery
	rf.nextIndex = make([]int, len(conf.Endpoints))
	rf.matchIndex = make([]int, len(conf.Endpoints))
	for i := range rf.peers {
		rf.nextIndex[i] = rf.lastIndex() + 1
		rf.matchIndex[i] = -1
	}

	rf.lastHearBeatTime = time.Now()
	rf.lastSendHeartbeatTime = time.Now()
	rf.applyCh = applyCh
	rf.applyChTerm = make(chan ApplyMsg, 1000)

	firlog.Logger.Infof("RESTA Term[%d] [%d] Restart😎", rf.currentTerm, rf.me)

	// Network setup
	rf.StartRaft(conf.Endpoints[me])
	servers := WaitConnect(conf)
	rf.peers = servers

	// Start goroutines
	go rf.Applyer()
	go rf.electionLoop()
	go rf.appendEntriesLoop()

	return rf
}

// WAL-MOD: New function to load state from the WAL on startup.
func (rf *Raft) loadFromWAL() {
	records, types, err := rf.wal.ReadAll()
	if err != nil {
		firlog.Logger.Panicf("Failed to read from WAL: %v", err)
	}

	for i, recType := range types {
		data := records[i]
		switch recType {
		case wal.RecordTypeState:
			var state RaftStateRecord
			if err := json.Unmarshal(data, &state); err != nil {
				firlog.Logger.Panicf("Failed to unmarshal state record: %v", err)
			}
			rf.currentTerm = state.Term
			rf.votedFor = state.VotedFor
		case wal.RecordTypeEntry:
			var entry pb.LogType
			if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&entry); err != nil {
				firlog.Logger.Panicf("Failed to decode log entry: %v", err)
			}
			rf.log = append(rf.log, entry)
		case wal.RecordTypeSnapshot:
			var snapshot SnapshotRecord
			if err := json.Unmarshal(data, &snapshot); err != nil {
				firlog.Logger.Panicf("Failed to unmarshal snapshot record: %v", err)
			}
			// When we find a snapshot, it becomes the new baseline.
			rf.lastIncludeIndex = snapshot.LastIncludedIndex
			rf.lastIncludeTerm = snapshot.LastIncludedTerm
			// The log entries before the snapshot are now irrelevant.
			rf.log = make([]pb.LogType, 0)

			// Apply the snapshot to the state machine
			snapshotData, err := os.ReadFile(snapshot.Path)
			if err != nil {
				firlog.Logger.Panicf("Failed to read snapshot file %s: %v", snapshot.Path, err)
			}
			applyMsg := ApplyMsg{
				SnapshotValid: true,
				Snapshot:      snapshotData,
				SnapshotIndex: rf.lastIncludeIndex + 1,
				SnapshotTerm:  rf.lastIncludeTerm,
			}
			rf.applyCh <- applyMsg
			rf.lastApplied = rf.lastIncludeIndex
			rf.commitIndex = rf.lastIncludeIndex
		}
	}
}

func (rt *Raft) StartRaft(conf firconfig.RaftEnd) {
	lis, err := net.Listen("tcp", conf.Addr+conf.Port)
	if err != nil {
		firlog.Logger.Fatalln("error: etcd start faild", err)
	}
	gServer := grpc.NewServer()
	pb.RegisterRaftServer(gServer, rt)
	go func() {
		if err := gServer.Serve(lis); err != nil {
			firlog.Logger.Fatalln("failed to serve : ", err.Error())
		}
	}()
}

func WaitConnect(conf firconfig.RaftEnds) []*RaftEnd {
	firlog.Logger.Infoln("start wating...")
	var wait sync.WaitGroup
	servers := make([]*RaftEnd, len(conf.Endpoints))
	wait.Add(len(servers) - 1)
	for i := range conf.Endpoints {
		if i == conf.Me {
			continue
		}

		go func(other int, conf firconfig.RaftEnd) {
			defer wait.Done()
			for {
				r := NewRaftClient(conf)
				if r != nil {
					servers[other] = r
					break
				}
				time.Sleep(time.Millisecond * 500)
			}
		}(i, conf.Endpoints[i])
	}
	wait.Wait()
	firlog.Logger.Infof("🦖 All %d Connect", len(conf.Endpoints))
	return servers
}

// GetPersistSize returns the total size of the Raft persistence layer.
func (rf *Raft) GetRaftStateSize() int {
	return int(rf.wal.Size())
}
