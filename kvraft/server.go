package kvraft

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/whosefriendA/firEtcd/pkg/firlog"

	"github.com/whosefriendA/firEtcd/common"
	bboltdb "github.com/whosefriendA/firEtcd/pkg/bboltdb"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
	"github.com/whosefriendA/firEtcd/pkg/lease"
	"github.com/whosefriendA/firEtcd/proto/pb"
	"github.com/whosefriendA/firEtcd/raft"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes" // 为 gRPC 状态添加
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status" // 为 gRPC 状态添加
)

type KVServer struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg
	dead    int32

	maxraftstate int

	lastAppliedIndex int
	lastIncludeIndex int
	db common.DB

	duplicateMap map[int64]duplicateType

	grpc *grpc.Server

	lastIndexCh chan int

	watcherManager *WatcherManager

	eventNotifier chan WatchEvent

	leaseMgr *lease.LeaseManager
}

// WatchEventType 映射 protobuf 枚举
type WatchEventType pb.EventType

const (
	WatchEventTypePut    = WatchEventType(pb.EventType_PUT_EVENT)
	WatchEventTypeDelete = WatchEventType(pb.EventType_DELETE_EVENT)
)

// WatchEvent 是变更事件的内部表示
type WatchEvent struct {
	Type  WatchEventType
	Key   string
	Value []byte
}

// Watcher 代表一个正在观察事件的客户端
type watcher struct {
	id        int64
	key       string
	isPrefix  bool
	eventChan chan<- WatchEvent
}

type WatcherManager struct {
	mu             sync.RWMutex
	nextWatcherID  int64
	exactWatchers  map[string]map[int64]*watcher
	prefixWatchers map[string]map[int64]*watcher
}

func NewWatcherManager() *WatcherManager {
	return &WatcherManager{
		nextWatcherID:  1,
		exactWatchers:  make(map[string]map[int64]*watcher),
		prefixWatchers: make(map[string]map[int64]*watcher),
	}
}

func (wm *WatcherManager) Register(key string, isPrefix bool, eventChan chan<- WatchEvent) (watchID int64) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	watchID = wm.nextWatcherID
	wm.nextWatcherID++

	w := &watcher{
		id:        watchID,
		key:       key,
		isPrefix:  isPrefix,
		eventChan: eventChan,
	}

	if isPrefix {
		if _, ok := wm.prefixWatchers[key]; !ok {
			wm.prefixWatchers[key] = make(map[int64]*watcher)
		}
		wm.prefixWatchers[key][w.id] = w
		firlog.Logger.Infof("WatcherManager: 为前缀 '%s' 注册了前缀观察者 ID %d", key, watchID)
	} else {
		if _, ok := wm.exactWatchers[key]; !ok {
			wm.exactWatchers[key] = make(map[int64]*watcher)
		}
		wm.exactWatchers[key][w.id] = w
		firlog.Logger.Infof("WatcherManager: 为键 '%s' 注册了精确观察者 ID %d", key, watchID)
	}
	return watchID
}

func (wm *WatcherManager) Deregister(watchID int64, key string, isPrefix bool) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if isPrefix {
		if watchersForKey, ok := wm.prefixWatchers[key]; ok {
			if _, watcherExists := watchersForKey[watchID]; watcherExists {
				delete(watchersForKey, watchID)
				if len(watchersForKey) == 0 {
					delete(wm.prefixWatchers, key)
				}
				firlog.Logger.Infof("WatcherManager: 为前缀 '%s' 注销了前缀观察者 ID %d", key, watchID)
			}
		}
	} else {
		if watchersForKey, ok := wm.exactWatchers[key]; ok {
			if _, watcherExists := watchersForKey[watchID]; watcherExists {
				delete(watchersForKey, watchID)
				if len(watchersForKey) == 0 {
					delete(wm.exactWatchers, key)
				}
				firlog.Logger.Infof("WatcherManager: 为键 '%s' 注销了精确观察者 ID %d", key, watchID)
			}
		}
	}
}

func (wm *WatcherManager) Notify(event WatchEvent) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if watchers, ok := wm.exactWatchers[event.Key]; ok {
		for _, w := range watchers {
			select {
			case w.eventChan <- event:
			default:
				firlog.Logger.Warnf("WatcherManager: 键 '%s' 的精确观察者 ID %d 事件通道已满或关闭。事件已丢弃。", event.Key, w.id)
			}
		}
	}

	for prefix, watchers := range wm.prefixWatchers {
		if strings.HasPrefix(event.Key, prefix) {
			for _, w := range watchers {
				select {
				case w.eventChan <- event:
				default:
					firlog.Logger.Warnf("WatcherManager: 前缀 '%s' 的前缀观察者 ID %d (事件键 '%s') 事件通道已满或关闭。事件已丢弃。", prefix, w.id, event.Key)
				}
			}
		}
	}
}

func (kv *KVServer) notifierLoop() {
	for event := range kv.eventNotifier {
		kv.watcherManager.Notify(event)
	}
}

type duplicateType struct {
	Offset int32
	// Reply     string
	CASResult bool
}

func (kv *KVServer) Get(_ context.Context, args *pb.GetArgs) (reply *pb.GetReply, err error) {
	reply = new(pb.GetReply)
	reply.Err = ErrWrongLeader
	reply.LeaderId = int32(kv.rf.GetleaderId())
	reply.ServerId = int32(kv.me)

	if _, ok := kv.rf.GetState(); ok {
	} else {
		return
	}

	if !kv.rf.IisBack {
		firlog.Logger.Infof("server [%d] [recovering] reject a [Get]🔰 args[%v]", kv.me, args)
		reply.Err = ErrWaitForRecover
		b := new(bytes.Buffer)
		e := gob.NewEncoder(b)
		e.Encode(raft.Op{
			OpType: int32(pb.OpType_EmptyT),
		})
		if err != nil {
			firlog.Logger.Fatalln(err)
		}
		kv.rf.Start(b.Bytes())
		return reply, nil
	}

	readLastIndex := kv.rf.GetCommitIndex()
	term := kv.rf.GetTerm()
	if kv.rf.CheckIfDepose() {
		reply.Err = ErrWrongLeader
		return
	}

	kv.mu.Lock()
	defer kv.mu.Unlock()

	if kv.lastAppliedIndex >= readLastIndex && kv.rf.GetLeader() && term == kv.rf.GetTerm() {
		var value [][]byte
		switch args.Op {
		case pb.OpType_GetT:
			if args.WithPrefix {
				entrys, err := kv.db.GetEntryWithPrefix(args.Key)
				if err != nil {
					firlog.Logger.Fatalf("database GetEntryWithPrefix faild:%s", err)
				}

				value = make([][]byte, 0, len(entrys))
				for _, e := range entrys {
					value = append(value, e.Value)
				}
			} else {
				v, err := kv.db.GetEntry(args.Key)
				if err != common.ErrNotFound {
					value = append(value, v.Value)
				}
			}
		case pb.OpType_GetKeys:
			ret, err := kv.db.Keys(int(args.PageSize), int(args.PageIndex))
			if err != nil {
				firlog.Logger.Fatalln(err)
			}
			value = make([][]byte, 0, len(ret))
			for i := range ret {
				var buf bytes.Buffer
				gob.NewEncoder(&buf).Encode(&ret[i])
				value = append(value, buf.Bytes())
				// buf.Reset()
			}
		case pb.OpType_GetKVs:
			firlog.Logger.Infof("KVServer %d: Handling a GET_ALL request.", kv.me)
			ret, err := kv.db.KVs(int(args.PageSize), int(args.PageIndex))
			if err != nil {
				firlog.Logger.Fatalln(err)
			}
			value = make([][]byte, 0, len(ret))
			for i := range ret {
				var buf bytes.Buffer
				gob.NewEncoder(&buf).Encode(&ret[i])
				value = append(value, buf.Bytes())
				// buf.Reset()
			}
		}

		if len(value) == 0 {
			reply.Err = ErrNoKey
			return
		}
		reply.Err = ErrOK
		reply.Value = value
		// firlog.Logger.Infof("server [%d] [Get] [ok] lastAppliedIndex[%d] readLastIndex[%d]", kv.me, kv.lastAppliedIndex, readLastIndex)
		// firlog.Logger.Infof("server [%d] [Get] [Ok] the get args[%v] reply[%v]", kv.me, args, reply)
	} else {
		reply.Err = ErrWaitForRecover
		// firlog.Logger.Infof("server [%d] [Get] [ErrWaitForRecover] kv.lastAppliedIndex < readLastIndex args[%v] reply[%v]", kv.me, *args, *reply)
	}

	// firlog.Logger.Infof("server [%d] [Get] [NoKey] the get args[%v] reply[%v]", kv.me, args, reply)
	// firlog.Logger.Infof("server [%d] [map] -> %v", kv.me, kv.db)

	return reply, nil
}

func (kv *KVServer) PutAppend(_ context.Context, args *pb.PutAppendArgs) (reply *pb.PutAppendReply, err error) {
	reply = new(pb.PutAppendReply)
	reply.LeaderId = int32(kv.rf.GetleaderId())
	reply.Err = ErrWrongLeader
	reply.ServerId = int32(kv.me)

	if _, ok := kv.rf.GetState(); ok {
	} else {
		return
	}
	op := raft.Op{
		ClientId: args.ClientId,
		Offset:   args.LatestOffset,
		OpType:   args.Op,
		Key:      args.Key,
		OriValue: args.OriValue,
		Entry: common.Entry{
			Value:    args.Value,
			DeadTime: args.DeadTime,
		},
		LeaseId: args.LeaseId,
	}

	kv.mu.Lock()
	if args.LatestOffset < kv.duplicateMap[args.ClientId].Offset {
		kv.mu.Unlock()
		return
	}
	if args.LatestOffset == kv.duplicateMap[args.ClientId].Offset {
		if op.OpType != int32(pb.OpType_CAST) {
			reply.Err = ErrOK
		} else {
			if kv.duplicateMap[args.ClientId].CASResult {
				reply.Err = ErrOK
			} else {
				reply.Err = ErrCasFaildInt
			}
		}
		kv.mu.Unlock()
		return
	}
	kv.mu.Unlock()

	index, term, isleader := kv.rf.Start(op.Marshal())

	if !isleader {
		return
	}

	kv.rf.SendAppendEntriesToAll()
	startWait := time.Now()
	for !kv.killed() {

		kv.mu.Lock()

		if index <= kv.lastAppliedIndex {
			if args.LatestOffset < kv.duplicateMap[args.ClientId].Offset {
				kv.mu.Unlock()
				return
			}
			if args.LatestOffset == kv.duplicateMap[args.ClientId].Offset {
				if op.OpType != int32(pb.OpType_CAST) {
					reply.Err = ErrOK
				} else {
					if kv.duplicateMap[args.ClientId].CASResult {
						reply.Err = ErrOK
					} else {
						reply.Err = ErrCasFaildInt
					}
				}
				kv.mu.Unlock()
				return
			}

			firlog.Logger.Infof("server [%d] [PutAppend] appliedIndex available :PutAppend index[%d] lastAppliedIndex[%d]", kv.me, index, kv.lastAppliedIndex)
			if term != kv.rf.GetTerm() {
				kv.mu.Unlock()
				return
			} //term匹配，说明本次提交一定是有效的

			reply.Err = ErrOK
			firlog.Logger.Infof("server [%d] [PutAppend] success args.index[%d]", kv.me, index)
			kv.mu.Unlock()
			if _, isleader := kv.rf.GetState(); !isleader {
				reply.Err = ErrWrongLeader
			}
			return
		}
		kv.mu.Unlock()
		select {
		case <-kv.lastIndexCh:
		case <-time.After(time.Millisecond * 500):
			firlog.Logger.Infof("server [%d] [PutAppend] fail [time out] args.index[%d]", kv.me, index)
			return
		}
		if time.Since(startWait).Milliseconds() > 500 {
			firlog.Logger.Infof("server [%d] [PutAppend] fail [time out] args.index[%d]", kv.me, index)
			return
		}
	}
	return reply, nil
}

func (kv *KVServer) Watch(req *pb.WatchRequest, stream pb.Kvserver_WatchServer) error {
	if kv.killed() {
		return status.Errorf(codes.Unavailable, "服务器正在关闭")
	}

	if _, ok := kv.rf.GetState(); !ok {
		leaderId := int32(kv.rf.GetleaderId())
		return status.Errorf(codes.FailedPrecondition, "不是 leader，当前 leader 是 %d", leaderId)
	}
	if !kv.rf.IisBack {
		return status.Errorf(codes.Unavailable, "服务器正在恢复中")
	}

	clientEventChan := make(chan WatchEvent, 10)
	watchKey := string(req.Key)
	isPrefix := req.IsPrefix

	watchID := kv.watcherManager.Register(watchKey, isPrefix, clientEventChan)
	firlog.Logger.Infof("KVServer %d: Watch RPC 为键/前缀 '%s' 注册了观察者 ID %d, isPrefix: %t", kv.me, watchID, watchKey, isPrefix)

	defer func() {
		kv.watcherManager.Deregister(watchID, watchKey, isPrefix)
		close(clientEventChan)
		firlog.Logger.Infof("KVServer %d: Watch RPC 为键/前缀 '%s' 注销了观察者 ID %d", kv.me, watchID, watchKey)
	}()

	if req.GetSendInitialState() {
		firlog.Logger.Infof("KVServer %d: Watch ID %d, sendInitialState=true. PREPARING to send initial state for key '%s'.", kv.me, watchID, watchKey)
		kv.mu.Lock()
		if isPrefix {
			pairs, _ := kv.db.GetPairsWithPrefix(watchKey)
			for _, pair := range pairs {
				protoResp := &pb.WatchResponse{
					Type:  pb.EventType_PUT_EVENT,
					Key:   []byte(pair.Key),
					Value: pair.Entry.Value,
				}
				if err := stream.Send(protoResp); err != nil {
					firlog.Logger.Errorf("KVServer %d: Watch ID %d, FAILED to send initial prefix value over gRPC stream. Error: %v", kv.me, watchID, err)
					kv.mu.Unlock()
					return err
				}
			}
			firlog.Logger.Infof("KVServer %d: Watch ID %d, SUCCESSFULLY SENT all initial prefix values.", kv.me, watchID)
		} else {
			entry, err := kv.db.GetEntry(watchKey)
			if err == nil {
				protoResp := &pb.WatchResponse{
					Type:  pb.EventType_PUT_EVENT,
					Key:   []byte(watchKey),
					Value: entry.Value,
				}
				firlog.Logger.Infof("KVServer %d: Watch ID %d, FOUND initial value. PREPARING TO SEND over gRPC stream.", kv.me, watchID)
				if sendErr := stream.Send(protoResp); sendErr != nil {
					firlog.Logger.Errorf("KVServer %d: Watch ID %d, FAILED to send initial value over gRPC stream. Error: %v", kv.me, watchID, sendErr)
					kv.mu.Unlock()
					return sendErr
				}
				firlog.Logger.Infof("KVServer %d: Watch ID %d, SUCCESSFULLY SENT initial value over gRPC stream.", kv.me, watchID)
			} else {
				firlog.Logger.Infof("KVServer %d: Watch ID %d, Initial value for key '%s' not found in db. Error: %v", kv.me, watchID, watchKey, err)
			}
		}
		kv.mu.Unlock()
	}

	for {
		select {
		case event, ok := <-clientEventChan:
			if !ok {
				firlog.Logger.Infof("KVServer %d: 观察者 ID %d 事件通道已关闭，结束流。", kv.me, watchID)
				return nil
			}
			protoResp := &pb.WatchResponse{
				Type:  pb.EventType(event.Type),
				Key:   []byte(event.Key),
				Value: event.Value,
			}
			if err := stream.Send(protoResp); err != nil {
				firlog.Logger.Warnf("KVServer %d: 向观察者 ID %d 发送事件时出错: %v。关闭流。", kv.me, watchID, err)
				return err
			}
			firlog.Logger.Debugf("KVServer %d: 已向观察者 ID %d 发送事件: %v", kv.me, watchID, event)

		case <-stream.Context().Done():
			firlog.Logger.Infof("KVServer %d: 观察者 ID %d 流上下文完成: %v。关闭流。", kv.me, watchID, stream.Context().Err())
			return stream.Context().Err()
		}
	}
}

func (kv *KVServer) HandleApplych() {
	for !kv.killed() {
		select {
		case raft_type := <-kv.applyCh:
			if kv.killed() {
				return
			}

			var eventsToNotify []WatchEvent

			kv.mu.Lock()

			if raft_type.CommandValid {
				if raft_type.CommandIndex > kv.lastAppliedIndex { // Prevent re-applying old commands
					eventsToNotify = kv.HandleApplychCommand(raft_type)
					kv.lastAppliedIndex = raft_type.CommandIndex
				}

				kv.checkifNeedSnapshot(raft_type.CommandIndex)

				for {
					select {
					case kv.lastIndexCh <- raft_type.CommandIndex:
					default:
						goto APPLYBREAK
					}
				}
			APPLYBREAK:
			} else if raft_type.SnapshotValid {
				firlog.Logger.Infof("📷 server [%d] receive raftSnapshotIndex[%d]", kv.me, raft_type.SnapshotIndex)
				kv.HandleApplychSnapshot(raft_type)
			} else {
				firlog.Logger.Fatalf("Unrecordnized applyArgs type")
			}

			kv.mu.Unlock()

			if len(eventsToNotify) > 0 {
				for _, event := range eventsToNotify {
					select {
					case kv.eventNotifier <- event:
					default:
						firlog.Logger.Warnf("KVServer %d: Notifier channel full. Discarding watch event for key %s.", kv.me, event.Key)
					}
				}
			}
		}
	}
}

func (kv *KVServer) HandleApplychCommand(raft_type raft.ApplyMsg) []WatchEvent {
	OP := new(raft.Op)
	OP.Unmarshal(raft_type.Command)
	if OP.OpType == int32(pb.OpType_EmptyT) {
		return nil
	}
	if OP.Offset <= kv.duplicateMap[OP.ClientId].Offset {
		firlog.Logger.Infof("⛔server [%d] [%v] lastapplied[%v]find in the cache and discard %v", kv.me, OP, kv.lastAppliedIndex, kv.db)
		return nil
	}
	kv.duplicateMap[OP.ClientId] = duplicateType{
		Offset: OP.Offset,
	}

	var eventsToNotify []WatchEvent

	switch OP.OpType {
	case int32(pb.OpType_PutT):
		err := kv.db.PutEntry(OP.Key, OP.Entry)
		if err != nil {
			firlog.Logger.Fatalf("database putEntry faild:%s", err)
		}
		if OP.LeaseId != 0 && kv.leaseMgr != nil {
			_ = kv.leaseMgr.AttachKey(OP.LeaseId, OP.Key)
		}

		eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypePut, Key: OP.Key, Value: OP.Entry.Value})
		firlog.Logger.Debugf("KVServer %d: 已应用 Put, Key: %s。已排队等待 watch 通知。", kv.me, OP.Key)

	case int32(pb.OpType_AppendT):

		ori, _ := kv.db.GetEntry(OP.Key)
		var buffer bytes.Buffer

		buffer.Write(ori.Value)
		buffer.Write(OP.Entry.Value)

		result := buffer.Bytes()
		err := kv.db.PutEntry(OP.Key, common.Entry{
			Value:    result,
			DeadTime: OP.Entry.DeadTime,
		})
		if err != nil {
			firlog.Logger.Fatalf("database putEntry faild:%s", err)
		}
		if OP.LeaseId != 0 && kv.leaseMgr != nil {
			_ = kv.leaseMgr.AttachKey(OP.LeaseId, OP.Key)
		}

		eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypePut, Key: OP.Key, Value: result})
		firlog.Logger.Debugf("KVServer %d: 已应用 Put, Key: %s。已排队等待 watch 通知。", kv.me, OP.Key)

	case int32(pb.OpType_DelT):
		_, err := kv.db.GetEntry(OP.Key)
		kv.db.Del(OP.Key)
		if OP.LeaseId != 0 && kv.leaseMgr != nil {
			_ = kv.leaseMgr.DetachKey(OP.LeaseId, OP.Key)
		}
		if err == nil {
			eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypeDelete, Key: OP.Key})
			firlog.Logger.Debugf("KVServer %d: 已应用 Del, Key: %s。已排队等待 watch 通知。", kv.me, OP.Key)
		}

	case int32(pb.OpType_DelWithPrefix):
		kv.db.DelWithPrefix(OP.Key)

		eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypeDelete, Key: OP.Key})
		firlog.Logger.Debugf("KVServer %d: 已应用 DelWithPrefix, Prefix: %s。已排队等待 watch 通知 (作为单个前缀删除)。", kv.me, OP.Key)

	case int32(pb.OpType_CAST):
		ori, _ := kv.db.GetEntry(OP.Key)
		if bytes.Equal(ori.Value, OP.OriValue) {
			if len(OP.Entry.Value) == 0 {
				kv.db.Del(OP.Key)
				if OP.LeaseId != 0 && kv.leaseMgr != nil {
					_ = kv.leaseMgr.DetachKey(OP.LeaseId, OP.Key)
				}
				eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypeDelete, Key: OP.Key})
				firlog.Logger.Debugf("KVServer %d: 已应用 DelWithPrefix, Prefix: %s。已排队等待 watch 通知 (作为单个前缀删除)。", kv.me, OP.Key)

			} else {
				kv.db.PutEntry(OP.Key, OP.Entry)
				if OP.LeaseId != 0 && kv.leaseMgr != nil {
					_ = kv.leaseMgr.AttachKey(OP.LeaseId, OP.Key)
				}
				eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypePut, Key: OP.Key, Value: OP.Entry.Value})
				firlog.Logger.Debugf("KVServer %d: 已应用 DelWithPrefix, Prefix: %s。已排队等待 watch 通知 (作为单个前缀删除)。", kv.me, OP.Key)

			}
			kv.duplicateMap[OP.ClientId] = duplicateType{
				Offset:    OP.Offset,
				CASResult: true,
			}
		}
	case int32(pb.OpType_GetT):

		firlog.Logger.Fatalf("日志中不应该出现getType")

	case int32(pb.OpType_BatchT):
		var ops []raft.Op
		b := bytes.NewBuffer([]byte(OP.Entry.Value))
		d := gob.NewDecoder(b)
		err := d.Decode(&ops)
		if err != nil {
			firlog.Logger.Fatalln("raw data:", []byte(OP.Entry.Value), err)
		}
		for _, op := range ops {
			switch op.OpType {
			case int32(pb.OpType_PutT):
				kv.db.PutEntry(op.Key, op.Entry)
				if op.LeaseId != 0 && kv.leaseMgr != nil {
					_ = kv.leaseMgr.AttachKey(op.LeaseId, op.Key)
				}
				eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypePut, Key: op.Key, Value: op.Entry.Value})
			case int32(pb.OpType_AppendT):
				ori, _ := kv.db.GetEntry(op.Key)

				var buffer bytes.Buffer

				buffer.Write(ori.Value)
				buffer.Write(op.Entry.Value)

				result := buffer.Bytes()
				kv.db.PutEntry(op.Key, common.Entry{
					Value:    result,
					DeadTime: op.Entry.DeadTime,
				})
				if op.LeaseId != 0 && kv.leaseMgr != nil {
					_ = kv.leaseMgr.AttachKey(op.LeaseId, op.Key)
				}
				eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypePut, Key: op.Key, Value: result})
			case int32(pb.OpType_DelT):
				kv.db.Del(op.Key)
				if op.LeaseId != 0 && kv.leaseMgr != nil {
					_ = kv.leaseMgr.DetachKey(op.LeaseId, op.Key)
				}
				eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypeDelete, Key: op.Key})
			case int32(pb.OpType_DelWithPrefix):
				kv.db.DelWithPrefix(op.Key)
				eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypeDelete, Key: op.Key})
			}
		}

	default:
		firlog.Logger.Fatalf("日志中出现未知optype = [%d]", OP.OpType)
	}

	return eventsToNotify
}

// 被动快照,follower接受从leader传来的snapshot
func (kv *KVServer) HandleApplychSnapshot(raft_type raft.ApplyMsg) {
	if raft_type.SnapshotIndex < kv.lastAppliedIndex {
		return
	}
	snapshot := raft_type.Snapshot
	kv.readPersist(snapshot)
	firlog.Logger.Infof("server [%d] passive📷 lastAppliedIndex[%d] -> [%d]", kv.me, kv.lastAppliedIndex, raft_type.SnapshotIndex)
	kv.lastAppliedIndex = raft_type.SnapshotIndex
	select {
	case kv.lastIndexCh <- raft_type.CommandIndex:
	default:
	}
}

func (kv *KVServer) checkifNeedSnapshot(spanshotindex int) {
	if kv.maxraftstate == -1 {
		return
	}
	if kv.rf.GetRaftStateSize() < kv.maxraftstate {
		return
	}

	firlog.Logger.Infof("server [%d] need snapshot limit[%d] curRaftStatesize[%d] snapshotIndex[%d]", kv.me, kv.maxraftstate, kv.rf.GetRaftStateSize(), spanshotindex)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(kv.duplicateMap); err != nil {
		firlog.Logger.Fatalf("snapshot duplicateMap encoder fail:%s", err)
	}
	data, err := kv.db.SnapshotData()
	if err != nil {
		firlog.Logger.Fatalf("database snapshotdata faild:%s", err)
	}
	enc.Encode(data)
	kv.rf.Snapshot(spanshotindex, buf.Bytes())
}

func (kv *KVServer) readPersist(data []byte) {

	if data == nil || len(data) < 1 {
		return
	}
	firlog.Logger.Infof("server [%d] passive 📷 len of snapshotdate[%d] ", kv.me, len(data))
	firlog.Logger.Infof("server [%d] before map[%v]", kv.me, kv.db)
	r := bytes.NewBuffer(data)
	d := gob.NewDecoder(r)

	duplicateMap := make(map[int64]duplicateType)
	if err := d.Decode(&duplicateMap); err != nil {
		firlog.Logger.Fatalf("decode err:%s", err)
	}

	newdb := bboltdb.NewDB()
	dbData := make([]byte, 0)
	err := d.Decode(&dbData)
	if err != nil {
		firlog.Logger.Fatalln("read persiset err", err)
	}
	newdb.InstallSnapshotData(dbData)
	kv.db = newdb
	kv.duplicateMap = duplicateMap

	firlog.Logger.Infof("server [%d] after map[%v]", kv.me, kv.db)

}

// the tester calls Kill() when a KVServer instance won't
// be needed again. for your convenience, we supply
// code to set rf.dead (without needing a lock),
// and a killed() method to test rf.dead in
// long-running loops. you can also add your own
// code to Kill(). you're not required to do anything
// about this, but it may be convenient (for example)
// to suppress debug output from a Kill()ed instance.
func (kv *KVServer) Kill() {
	atomic.StoreInt32(&kv.dead, 1)
	kv.rf.Kill()
	// Your code here, if desired.
}

func (kv *KVServer) killed() bool {
	z := atomic.LoadInt32(&kv.dead)
	return z == 1
}

// servers[] contains the ports of the set of
// servers that will cooperate via Raft to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
// the k/v server should store snapshots through the underlying Raft
// implementation, which should call persister.SaveStateAndSnapshot() to
// atomically save the Raft state along with the snapshot.
// the k/v server should snapshot when Raft's saved state exceeds maxraftstate bytes,
// in order to allow Raft to garbage-collect its log. if maxraftstate is -1,
// you don't need to snapshot.
// StartKVServer() must return quickly, so it should start goroutines
// for any long-running work.
func StartKVServer(conf firconfig.Kvserver, me int, dataDir string, maxraftstate int) *KVServer {
	// call labgob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	var err error
	gob.Register(raft.Op{})
	gob.Register(map[string]string{})
	gob.Register(map[int64]duplicateType{})
	kv := &KVServer{
		me:               me,
		maxraftstate:     maxraftstate,
		applyCh:          make(chan raft.ApplyMsg),
		lastAppliedIndex: 0,
		lastIncludeIndex: 0,
		db:               bboltdb.NewDB(),
		lastIndexCh:      make(chan int),
		duplicateMap:     make(map[int64]duplicateType),
		watcherManager:   NewWatcherManager(),
		eventNotifier:    make(chan WatchEvent, 1024),
	}

	// initialize lease manager with a sane minimal TTL (e.g., 1s)
	kv.leaseMgr = lease.NewLeaseManager(time.Second)

	go kv.notifierLoop()

	// background lease expiration loop (leader only)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			if kv.killed() {
				return
			}
			<-ticker.C
			// only leader and fully recovered can propose revokes
			if _, isLeader := kv.rf.GetState(); !isLeader || !kv.rf.IisBack {
				continue
			}
			for _, id := range kv.leaseMgr.ExpiredLeases(time.Now()) {
				kv.proposeLeaseRevoke(id)
			}
		}
	}()

	walDir := filepath.Join(dataDir, fmt.Sprintf("raft-%d", me))

	// WAL-MOD: 调用新的 raft.Make 函数，传入 WAL 目录路径
	kv.rf = raft.Make(me, walDir, kv.applyCh, conf.Rafts)
	go kv.HandleApplych()

	// 加载TLS证书
	certificate, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/server.crt", "/home/wanggang/firEtcd/pkg/tls/certs/server.key")
	if err != nil {
		firlog.Logger.Fatalf("无法加载服务器证书: %v", err)
	}

	// 创建证书池并添加CA证书
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("/home/wanggang/firEtcd/pkg/tls/certs/ca.crt")
	if err != nil {
		firlog.Logger.Fatalf("无法读取CA证书: %v", err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		firlog.Logger.Fatal("无法将CA证书添加到证书池")
	}

	// 配置TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
		NextProtos:   []string{"h2"}, // 为gRPC启用HTTP/2
	}

	creds := credentials.NewTLS(tlsConfig)

	lis, err := net.Listen("tcp", conf.Addr+conf.Port)
	if err != nil {
		firlog.Logger.Fatalln("error: etcd start failed", err)
	}

	// 创建gRPC服务器，启用TLS
	gServer := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterKvserverServer(gServer, kv)
	// register lease service
	pb.RegisterLeaseServer(gServer, NewLeaseService(kv))
	go func() {
		if err := gServer.Serve(lis); err != nil {
			firlog.Logger.Fatalln("failed to serve : ", err.Error())
		}
	}()

	firlog.Logger.Infoln("etcd service is running on addr:", conf.Addr+conf.Port)
	kv.grpc = gServer

	firlog.Logger.Infof("server [%d] restart", kv.me)
	return kv
}

func (kv *KVServer) proposeLeaseRevoke(leaseID int64) {
	// Get keys attached to this lease and propose deletes via Raft
	keys := kv.leaseMgr.GetLeaseKeys(leaseID)
	for _, k := range keys {
		op := raft.Op{
			ClientId: 0,
			Offset:   0,
			OpType:   int32(pb.OpType_DelT),
			Key:      k,
			LeaseId:  leaseID,
		}
		kv.rf.Start(op.Marshal())
	}
	// Finally, update local lease manager to drop the lease entry itself
	_ = kv.leaseMgr.Revoke(leaseID)
}
