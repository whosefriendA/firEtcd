package kvraft

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"io/ioutil"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/whosefriendA/firEtcd/pkg/firlog"

	"github.com/whosefriendA/firEtcd/common"
	buntdbx "github.com/whosefriendA/firEtcd/pkg/buntdb"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
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
	dead    int32 // set by Kill()

	maxraftstate int // snapshot if log grows this big

	persister *raft.Persister
	// Your definitions here.

	//duplicateMap: use to handle mulity RPC request
	// duplicateMap map[int64]duplicateType

	lastAppliedIndex int //最近添加到状态机中的raft层的log的index
	//lastInclude
	lastIncludeIndex int
	//log state machine
	db common.DB

	//缓存的log, seq->index,reply
	duplicateMap map[int64]duplicateType

	grpc *grpc.Server

	lastIndexCh chan int

	watcherManager *WatcherManager

	eventNotifier chan WatchEvent
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
	Value []byte // 用于 PUT_EVENT
	// OldValue []byte // 可选
}

// Watcher 代表一个正在观察事件的客户端
type watcher struct {
	id        int64
	key       string
	isPrefix  bool
	eventChan chan<- WatchEvent // 指向客户端 gRPC 流的只写通道
	// sendInitialState bool // 可选
}

// WatcherManager 管理所有活跃的 watcher
type WatcherManager struct {
	mu             sync.RWMutex
	nextWatcherID  int64
	exactWatchers  map[string]map[int64]*watcher // key -> watcherID -> watcher
	prefixWatchers map[string]map[int64]*watcher // prefix -> watcherID -> watcher
}

func NewWatcherManager() *WatcherManager {
	return &WatcherManager{
		nextWatcherID:  1,
		exactWatchers:  make(map[string]map[int64]*watcher),
		prefixWatchers: make(map[string]map[int64]*watcher),
	}
}

// Register 注册一个新的 watcher
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
		wm.prefixWatchers[key][w.id] = w // 修正：应使用 w.id 作为 key
		firlog.Logger.Infof("WatcherManager: 为前缀 '%s' 注册了前缀观察者 ID %d", key, watchID)
	} else {
		if _, ok := wm.exactWatchers[key]; !ok {
			wm.exactWatchers[key] = make(map[int64]*watcher)
		}
		wm.exactWatchers[key][w.id] = w // 修正：应使用 w.id 作为 key
		firlog.Logger.Infof("WatcherManager: 为键 '%s' 注册了精确观察者 ID %d", key, watchID)
	}
	return watchID
}

// Deregister 注销一个 watcher
func (wm *WatcherManager) Deregister(watchID int64, key string, isPrefix bool) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if isPrefix {
		if watchersForKey, ok := wm.prefixWatchers[key]; ok {
			if _, watcherExists := watchersForKey[watchID]; watcherExists { // 修正：应检查 watchID
				delete(watchersForKey, watchID)
				if len(watchersForKey) == 0 {
					delete(wm.prefixWatchers, key)
				}
				firlog.Logger.Infof("WatcherManager: 为前缀 '%s' 注销了前缀观察者 ID %d", key, watchID)
			}
		}
	} else {
		if watchersForKey, ok := wm.exactWatchers[key]; ok {
			if _, watcherExists := watchersForKey[watchID]; watcherExists { // 修正：应检查 watchID
				delete(watchersForKey, watchID)
				if len(watchersForKey) == 0 {
					delete(wm.exactWatchers, key)
				}
				firlog.Logger.Infof("WatcherManager: 为键 '%s' 注销了精确观察者 ID %d", key, watchID)
			}
		}
	}
}

// Notify 通知相关 watcher 一个事件。此方法不应阻塞。
func (wm *WatcherManager) Notify(event WatchEvent) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	// 通知精确匹配的 watcher
	if watchers, ok := wm.exactWatchers[event.Key]; ok {
		for _, w := range watchers {
			select {
			case w.eventChan <- event:
			default:
				// 客户端通道已满或已关闭，记录此情况。
				// 如果某个 watcher 持续缓慢，可以考虑关闭它。
				firlog.Logger.Warnf("WatcherManager: 键 '%s' 的精确观察者 ID %d 事件通道已满或关闭。事件已丢弃。", event.Key, w.id)
			}
		}
	}

	// 通知前缀匹配的 watcher
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

// 新增: 这是专门处理通知的 goroutine
func (kv *KVServer) notifierLoop() {
	// 它会不断地从通知通道中读取事件
	for event := range kv.eventNotifier {
		// 这个调用在独立的 goroutine 中，不会持有 kv.mu，因此是安全的
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

	//判断自己是不是leader
	if _, ok := kv.rf.GetState(); ok {
		// firlog.Logger.Infof("server [%d] [info] i am leader", kv.me)
	} else {
		// firlog.Logger.Infof("server [%d] [info] i am not leader ,leader is [%d]", kv.me, reply.LeaderId)
		return
	}

	//判断自己有没有从重启中恢复完毕状态机
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
	//需要发送一轮心跳获得大多数回复，只是为了确定没有一个任期更加新的leader，保证自己的数据不是脏的
	if kv.rf.CheckIfDepose() {
		reply.Err = ErrWrongLeader
		return
	}

	//return false ,但是进入下面代码段的时候，发现自己又不是leader了，捏麻麻的
	kv.mu.Lock()
	defer kv.mu.Unlock()
	//跟raft层之间的同步问题，raft刚当选leader的时候，还没有

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
	// start := time.Now()
	// firlog.Logger.Infof("server [%d] [PutAppend] 📨receive a args[%v]", kv.me, args.String())
	// defer func() {
	// 	firlog.Logger.Infof("server [%d] [PutAppend] 📨complete a args[%v] spand time:%v", kv.me, args.String(), time.Since(start))
	// }()
	reply = new(pb.PutAppendReply)
	// Your code here.
	reply.LeaderId = int32(kv.rf.GetleaderId())
	reply.Err = ErrWrongLeader
	reply.ServerId = int32(kv.me)

	if _, ok := kv.rf.GetState(); ok {
		// firlog.Logger.Infof("server [%d] [info] i am leader", kv.me)
	} else {
		// firlog.Logger.Infof("server [%d] [info] i am not leader ,leader is [%d]", kv.me, reply.LeaderId)
		return
	}
	// v := DateToValue(args.Value)

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
	}

	//start前需要查看本地log缓存是否有seq

	//这里通过缓存提交，一方面提高了kvserver应对网络错误的回复速度，另一方面进行了第一层的重复检测
	//但是注意可能同时有两个相同的getDuplicateMap通过这里
	kv.mu.Lock()
	if args.LatestOffset < kv.duplicateMap[args.ClientId].Offset {
		kv.mu.Unlock()
		//firlog.Logger.Debugln("pass", kv.me)
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
		//firlog.Logger.Debugln("pass", kv.me)
		kv.mu.Unlock()
		return
	}
	kv.mu.Unlock()

	//没有在本地缓存发现过seq
	//向raft提交操作
	// firlog.Logger.Debugln("raw data:", []byte(op.Value))
	// data, err := json.Marshal(op)

	index, term, isleader := kv.rf.Start(op.Marshal())

	if !isleader {
		return
	}

	kv.rf.SendAppendEntriesToAll()
	// firlog.Logger.Infof("server [%d] submit to raft key[%v] value[%v]", kv.me, op.Key, op.Value)
	//提交后阻塞等待
	//等待applyCh拿到对应的index，比对seq是否正确
	startWait := time.Now()
	for !kv.killed() {

		kv.mu.Lock()

		if index <= kv.lastAppliedIndex {
			//双重防重复
			if args.LatestOffset < kv.duplicateMap[args.ClientId].Offset {
				//firlog.Logger.Debugln("pass", kv.me)
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
				//firlog.Logger.Debugln("pass", kv.me)
				kv.mu.Unlock()
				return
			}

			firlog.Logger.Infof("server [%d] [PutAppend] appliedIndex available :PutAppend index[%d] lastAppliedIndex[%d]", kv.me, index, kv.lastAppliedIndex)
			if term != kv.rf.GetTerm() {
				//term不匹配了，说明本次提交失效
				kv.mu.Unlock()
				//firlog.Logger.Debugln("pass", kv.me)
				return
			} //term匹配，说明本次提交一定是有效的

			reply.Err = ErrOK
			firlog.Logger.Infof("server [%d] [PutAppend] success args.index[%d]", kv.me, index)
			kv.mu.Unlock()
			if _, isleader := kv.rf.GetState(); !isleader {
				reply.Err = ErrWrongLeader
			}
			//firlog.Logger.Debugln("pass", kv.me)
			return
		}
		kv.mu.Unlock()
		select {
		case <-kv.lastIndexCh:
			// 阻塞等待
		case <-time.After(time.Millisecond * 500):
			firlog.Logger.Infof("server [%d] [PutAppend] fail [time out] args.index[%d]", kv.me, index)
			return
		}
		// 因为time.After可能会在超时前多次被重置，所以还需要在外层额外做保证
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

	// 判断自己是不是 leader
	if _, ok := kv.rf.GetState(); !ok {
		leaderId := int32(kv.rf.GetleaderId())
		// 你可能想返回一个更具体的错误或头部信息来指明 leader
		return status.Errorf(codes.FailedPrecondition, "不是 leader，当前 leader 是 %d", leaderId)
	}
	// 判断自己有没有从重启中恢复完毕状态机
	if !kv.rf.IisBack {
		return status.Errorf(codes.Unavailable, "服务器正在恢复中")
	}

	clientEventChan := make(chan WatchEvent, 10) // 为此客户端设置的带缓冲的通道
	watchKey := string(req.Key)
	isPrefix := req.IsPrefix

	watchID := kv.watcherManager.Register(watchKey, isPrefix, clientEventChan)
	firlog.Logger.Infof("KVServer %d: Watch RPC 为键/前缀 '%s' 注册了观察者 ID %d, isPrefix: %t", kv.me, watchID, watchKey, isPrefix)

	defer func() {
		kv.watcherManager.Deregister(watchID, watchKey, isPrefix)
		close(clientEventChan) // 很重要，如果下面的循环没有退出，需要关闭通道以停止它
		firlog.Logger.Infof("KVServer %d: Watch RPC 为键/前缀 '%s' 注销了观察者 ID %d", kv.me, watchID, watchKey)
	}()

	// 可选：如果 req.SendInitialState 为 true，发送初始状态
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
					return err // Exit on send error
				}
			}
			firlog.Logger.Infof("KVServer %d: Watch ID %d, SUCCESSFULLY SENT all initial prefix values.", kv.me, watchID)
		} else {
			// ======================= FIX START =======================
			// 这是修正的核心：为精确匹配添加数据库查询逻辑
			entry, err := kv.db.GetEntry(watchKey)
			if err == nil {
				// 如果找到了 key
				protoResp := &pb.WatchResponse{
					Type:  pb.EventType_PUT_EVENT,
					Key:   []byte(watchKey),
					Value: entry.Value,
				}
				firlog.Logger.Infof("KVServer %d: Watch ID %d, FOUND initial value. PREPARING TO SEND over gRPC stream.", kv.me, watchID)
				if sendErr := stream.Send(protoResp); sendErr != nil {
					firlog.Logger.Errorf("KVServer %d: Watch ID %d, FAILED to send initial value over gRPC stream. Error: %v", kv.me, watchID, sendErr)
					kv.mu.Unlock()
					return sendErr // Exit on send error
				}
				firlog.Logger.Infof("KVServer %d: Watch ID %d, SUCCESSFULLY SENT initial value over gRPC stream.", kv.me, watchID)
			} else {
				// 只有在 kv.db.GetEntry 确实返回错误时，才打印这个日志
				firlog.Logger.Infof("KVServer %d: Watch ID %d, Initial value for key '%s' not found in db. Error: %v", kv.me, watchID, watchKey, err)
			}
			// ======================= FIX END =========================
		}
		kv.mu.Unlock()
	}

	// 循环向客户端发送事件或检测客户端断开连接
	for {
		select {
		case event, ok := <-clientEventChan:
			if !ok {
				// 通道被 Deregister 关闭，流应该结束
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
				return err // 客户端断开连接或其他流错误
			}
			firlog.Logger.Debugf("KVServer %d: 已向观察者 ID %d 发送事件: %v", kv.me, watchID, event)

		case <-stream.Context().Done():
			// 客户端关闭了连接或上下文超时
			firlog.Logger.Infof("KVServer %d: 观察者 ID %d 流上下文完成: %v。关闭流。", kv.me, watchID, stream.Context().Err())
			return stream.Context().Err()
		}
	}
}

// state machine
// 将value重新转换为 Op，添加到本地db中
func (kv *KVServer) HandleApplych() {
	for !kv.killed() {
		select {
		case raft_type := <-kv.applyCh:
			if kv.killed() {
				return
			}

			// NEW: Declare a slice to hold events outside the lock
			var eventsToNotify []WatchEvent

			// --- Critical Section Starts ---
			kv.mu.Lock()

			if raft_type.CommandValid {
				// MODIFIED: Call the refactored HandleApplychCommand and receive the events
				// The rest of the logic inside the lock remains the same
				if raft_type.CommandIndex > kv.lastAppliedIndex { // Prevent re-applying old commands
					eventsToNotify = kv.HandleApplychCommand(raft_type)
					kv.lastAppliedIndex = raft_type.CommandIndex
				}

				kv.checkifNeedSnapshot(raft_type.CommandIndex)

				// This non-blocking send to unblock client RPCs is fine to keep inside the lock
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
				kv.HandleApplychSnapshot(raft_type) // Assumes this function correctly handles locking
			} else {
				firlog.Logger.Fatalf("Unrecordnized applyArgs type")
			}

			kv.mu.Unlock()
			// --- Critical Section Ends ---

			// **MOVED OUTSIDE LOCK**: The notification logic is now here.
			// This part runs without holding the server's main lock, so it can't freeze the state machine.
			if len(eventsToNotify) > 0 {
				for _, event := range eventsToNotify {
					// Non-blockingly send the event to the dedicated notifier goroutine
					select {
					case kv.eventNotifier <- event:
						// Event successfully queued for notification
					default:
						// This case is a safeguard against a full notifier channel
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

	var eventsToNotify []WatchEvent // 收集事件，在数据库操作后进行通知

	switch OP.OpType {
	case int32(pb.OpType_PutT):
		//更新状态机
		//有可能有多个start重复执行，所以这一步要检验重复

		err := kv.db.PutEntry(OP.Key, OP.Entry)
		if err != nil {
			firlog.Logger.Fatalf("database putEntry faild:%s", err)
		}

		eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypePut, Key: OP.Key, Value: OP.Entry.Value})
		firlog.Logger.Debugf("KVServer %d: 已应用 Put, Key: %s。已排队等待 watch 通知。", kv.me, OP.Key)

		// firlog.Logger.Infof("server [%d] [Update] [Put]->[%s,%s] [map] -> %v", kv.me, op_type.Key, op_type.Value, kv.db)
		// firlog.Logger.Infof("server [%d] [Update] [Put]->[%s : %s] ", kv.me, op_type.Key, op_type.Value)
	case int32(pb.OpType_AppendT):

		ori, _ := kv.db.GetEntry(OP.Key)
		var buffer bytes.Buffer

		// 写入数据
		buffer.Write(ori.Value)
		buffer.Write(OP.Entry.Value)

		// 获取拼接结果
		result := buffer.Bytes()
		err := kv.db.PutEntry(OP.Key, common.Entry{
			Value:    result,
			DeadTime: OP.Entry.DeadTime,
		})
		if err != nil {
			firlog.Logger.Fatalf("database putEntry faild:%s", err)
		}

		eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypePut, Key: OP.Key, Value: result})
		firlog.Logger.Debugf("KVServer %d: 已应用 Put, Key: %s。已排队等待 watch 通知。", kv.me, OP.Key)

		// firlog.Logger.Infof("server [%d] [Update] [Append]->[%s : %s]", kv.me, op_type.Key, op_type.Value)

	case int32(pb.OpType_DelT):
		_, err := kv.db.GetEntry(OP.Key) // 检查键是否存在
		kv.db.Del(OP.Key)
		if err == nil { // 仅当键实际存在时才通知
			eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypeDelete, Key: OP.Key})
			firlog.Logger.Debugf("KVServer %d: 已应用 Del, Key: %s。已排队等待 watch 通知。", kv.me, OP.Key)
		}

	case int32(pb.OpType_DelWithPrefix):
		kv.db.DelWithPrefix(OP.Key)

		eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypeDelete, Key: OP.Key /* 此事件的键是前缀 */})
		firlog.Logger.Debugf("KVServer %d: 已应用 DelWithPrefix, Prefix: %s。已排队等待 watch 通知 (作为单个前缀删除)。", kv.me, OP.Key)

	case int32(pb.OpType_CAST):
		ori, _ := kv.db.GetEntry(OP.Key)
		if bytes.Equal(ori.Value, OP.OriValue) {
			if len(OP.Entry.Value) == 0 {
				kv.db.Del(OP.Key)
				eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypeDelete, Key: OP.Key})
				firlog.Logger.Debugf("KVServer %d: 已应用 DelWithPrefix, Prefix: %s。已排队等待 watch 通知 (作为单个前缀删除)。", kv.me, OP.Key)

			} else {
				kv.db.PutEntry(OP.Key, OP.Entry)
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
		// firlog.Logger.Debugln("receive batch data:", b.Bytes())
		d := gob.NewDecoder(b)
		err := d.Decode(&ops)
		if err != nil {
			firlog.Logger.Fatalln("raw data:", []byte(OP.Entry.Value), err)
		}
		for _, op := range ops {
			switch op.OpType {
			case int32(pb.OpType_PutT):
				kv.db.PutEntry(op.Key, op.Entry)
				eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypePut, Key: op.Key, Value: op.Entry.Value})
			case int32(pb.OpType_AppendT):
				ori, _ := kv.db.GetEntry(op.Key)

				var buffer bytes.Buffer

				// 写入数据
				buffer.Write(ori.Value)
				buffer.Write(op.Entry.Value)

				// 获取拼接结果
				result := buffer.Bytes()
				kv.db.PutEntry(op.Key, common.Entry{
					Value:    result,
					DeadTime: op.Entry.DeadTime,
				})
				eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypePut, Key: op.Key, Value: result})
			case int32(pb.OpType_DelT):
				kv.db.Del(op.Key)
				eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypeDelete, Key: op.Key})
			case int32(pb.OpType_DelWithPrefix):
				kv.db.DelWithPrefix(op.Key)
				eventsToNotify = append(eventsToNotify, WatchEvent{Type: WatchEventTypeDelete, Key: op.Key})
			}
			// firlog.Logger.Infof("exec batch op: %+v", op)
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

// 主动快照,每一个服务器都在自己log超标的时候启动快照
func (kv *KVServer) checkifNeedSnapshot(spanshotindex int) {
	if kv.maxraftstate == -1 {
		return
	}
	if !kv.rf.IfNeedExceedLog(kv.maxraftstate) {
		return
	} //需要进行快照了

	firlog.Logger.Infof("server [%d] need snapshot limit[%d] curRaftStatesize[%d] snapshotIndex[%d]", kv.me, kv.maxraftstate, kv.persister.RaftStateSize(), spanshotindex)
	//首先查看一下自己的状态机应用到了那一步

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
	//将状态机传了进去
	kv.rf.Snapshot(spanshotindex, buf.Bytes())

}

// 被动快照
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

	newdb := buntdbx.NewDB()
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
func StartKVServer(conf firconfig.Kvserver, me int, persister *raft.Persister, maxraftstate int) *KVServer {
	// call labgob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	var err error
	gob.Register(raft.Op{})
	gob.Register(map[string]string{})
	gob.Register(map[int64]duplicateType{})
	kv := &KVServer{
		me:               me,
		maxraftstate:     maxraftstate,
		persister:        persister,
		applyCh:          make(chan raft.ApplyMsg),
		lastAppliedIndex: 0,
		lastIncludeIndex: 0,
		db:               buntdbx.NewDB(),
		lastIndexCh:      make(chan int),
		duplicateMap:     make(map[int64]duplicateType),
		watcherManager:   NewWatcherManager(),
		eventNotifier:    make(chan WatchEvent, 1024),
	}

	go kv.notifierLoop()

	kv.rf = raft.Make(me, persister, kv.applyCh, conf.Rafts)

	kv.readPersist(persister.ReadSnapshot())
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
	}

	// 创建TLS监听器
	lis, err := tls.Listen("tcp", conf.Addr+conf.Port, tlsConfig)
	if err != nil {
		firlog.Logger.Fatalln("error: etcd start failed", err)
	}

	// 创建gRPC服务器，启用TLS
	gServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	pb.RegisterKvserverServer(gServer, kv)
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
