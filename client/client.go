package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"io"
	"math/big"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	"github.com/whosefriendA/firEtcd/common"
	"github.com/whosefriendA/firEtcd/kvraft"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
	"github.com/whosefriendA/firEtcd/pkg/firlog"
	"github.com/whosefriendA/firEtcd/proto/pb"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

// Clerk 定义了客户端的行为
type Clerk interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte, TTL time.Duration) error
	Delete(key string) error
	DeleteWithPrefix(prefix string) error
	GetWithPrefix(key string) ([][]byte, error)
	Keys() ([]common.Pair, error)
	KeysWithPage(pageSize, pageIndex int) ([]common.Pair, error)
	KVs() ([]common.Pair, error)
	KVsWithPage(pageSize, pageIndex int) ([]common.Pair, error)
	Lock(key string, TTL time.Duration) (string, error)
	Unlock(key string, id string) (bool, error)
	Pipeline() *Pipe
	Watch(ctx context.Context, key string, opts ...WatchOption) (<-chan *WatchEvent, error)
	WatchDog(key string, value []byte) func()
	CAS(key string, origin, dest []byte, TTL time.Duration) (bool, error)
}

var pipeLimit int = 1024 * 4000

type clerkImpl struct {
	servers []*kvraft.KVClient
	// You will have to modify this struct.
	nextSendLocalId int
	LatestOffset    int32
	clientId        int64
	cTos            []int
	sToc            []int
	conf            firconfig.Clerk
	mu              sync.Mutex
}

type WatchEventChan <-chan *pb.WatchResponse

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

type WatchEvent pb.WatchResponse

// WatchOption 定义了 Watch 请求的可选参数。
type WatchOption func(*pb.WatchRequest)

// WithSendInitialState 是一个 WatchOption，用于请求在 Watch 开始时发送初始状态。
func WithSendInitialState(send bool) WatchOption {
	return func(req *pb.WatchRequest) {
		req.SendInitialState = send
	}
}

// WithPrefix 是一个 WatchOption，用于指定 Watch 的键是否为前缀。
func WithPrefix() WatchOption {
	return func(req *pb.WatchRequest) {
		req.IsPrefix = true
	}
}

const (
	watchRetryInitialBackoff = 500 * time.Millisecond
	watchRetryMaxBackoff     = 30 * time.Second
	watchRetryMultiplier     = 2
	watchEventBufferSize     = 100 // Buffer size for the event channel returned to the application
)

func (ck *clerkImpl) Watch(ctx context.Context, key string, opts ...WatchOption) (<-chan *WatchEvent, error) {
	watchReq := &pb.WatchRequest{
		Key:              []byte(key),
		IsPrefix:         false, // Default, overridden by WithPrefix
		SendInitialState: true,  // Default, overridden by WithSendInitialState
	}
	for _, opt := range opts {
		opt(watchReq)
	}

	eventChan := make(chan *WatchEvent, watchEventBufferSize)

	go ck.manageWatchStream(ctx, key, watchReq, eventChan)

	// Note: This Watch() returns immediately. Errors during stream establishment
	// or subsequent errors will result in eventChan being closed.
	// If an immediate, unrecoverable error during the *very first* connection attempt
	// is desired to be returned directly from Watch(), the initial connection logic
	// would need to be partially synchronous here, which adds complexity.
	// The current model (async goroutine, errors close channel) is common for watches.
	return eventChan, nil
}

func (ck *clerkImpl) manageWatchStream(ctx context.Context, watchKeyStr string, req *pb.WatchRequest, eventChan chan<- *WatchEvent) {
	defer close(eventChan)

	currentBackoff := watchRetryInitialBackoff
	serverIdx := -1 // Start with no specific server, find one.

	for {
		select {
		case <-ctx.Done():
			firlog.Logger.Infof("Watch for key '%s' cancelled by context before stream attempt: %v", watchKeyStr, ctx.Err())
			return
		default:
		}

		var stream pb.Kvserver_WatchClient
		var err error
		var currentKvClient *kvraft.KVClient

		// --- Begin critical section for server selection ---
		ck.mu.Lock()
		if serverIdx == -1 { // First attempt or after a leaderless error
			serverIdx = ck.nextSendLocalId
		}

		// Try to find a valid server, similar to read/write, but adapted for streaming
		initialAttemptIdx := serverIdx
		for i := 0; i < len(ck.servers); i++ {
			targetIdx := (initialAttemptIdx + i) % len(ck.servers)
			if ck.servers[targetIdx].Valid && ck.servers[targetIdx].Conn != nil {
				currentKvClient = ck.servers[targetIdx]
				serverIdx = targetIdx // Remember the server we are trying
				break
			}
		}
		ck.mu.Unlock()
		// --- End critical section for server selection ---

		if currentKvClient == nil {
			//firlog.Logger.Warnf("Watch for key '%s': No valid servers found, retrying after backoff.", watchKeyStr)
			goto handle_error_and_retry
		}

		//firlog.Logger.Infof("Attempting to establish watch stream for key '%s' with server %d (local index %d)", watchKeyStr, currentKvClient.me, serverIdx)
		stream, err = currentKvClient.Conn.Watch(ctx, req) // Use the main context for the gRPC call

		if err != nil {
			//firlog.Logger.Warnf("Watch for key '%s': Failed to establish stream with server %s: %v", watchKeyStr, currentKvClient.Addr, err)
			// Handle gRPC error status
			if s, ok := status.FromError(err); ok {
				if s.Code() == codes.FailedPrecondition { // Likely ErrWrongLeader or server not ready
					// Try to parse leader hint if available (not standard in gRPC error details, but server might provide)
					// For now, assume it's a generic "try another server"
					//firlog.Logger.Infof("Watch for key '%s': Server %s indicated not leader or not ready. Trying next.", watchKeyStr, currentKvClient.Addr)
					ck.mu.Lock()
					ck.changeNextSendId()          // Rotate to next server for next attempt
					serverIdx = ck.nextSendLocalId // Update serverIdx for next loop
					ck.mu.Unlock()
					currentBackoff = watchRetryInitialBackoff // Reset backoff if it's a leader issue
					goto handle_error_and_retry_no_wait       // Retry immediately with new server
				}
			}
			// For other initial connection errors, proceed to backoff
			goto handle_error_and_retry
		}

		//firlog.Logger.Infof("Watch stream established for key '%s' with server %s.", watchKeyStr, currentKvClient.me)
		currentBackoff = watchRetryInitialBackoff // Reset backoff on successful connection

		// Receive loop
		for {
			resp, recvErr := stream.Recv()
			if recvErr != nil {
				if recvErr == io.EOF {
					firlog.Logger.Errorf("Clerk Watch: FAILED to receive from gRPC stream. Error: %v", recvErr)
					//firlog.Logger.Infof("Watch stream for key '%s' closed by server %s (EOF).", watchKeyStr, currentKvClient.Addr)
				} else {
					//firlog.Logger.Warnf("Watch stream for key '%s' on server %s Recv error: %v.", watchKeyStr, currentKvClient.Addr, recvErr)
				}
				// Any error from Recv() means the stream is broken, try to re-establish.
				// Update server validity if connection seems truly lost (e.g. Unavailable)
				if s, ok := status.FromError(recvErr); ok && s.Code() == codes.Unavailable {
					ck.mu.Lock()
					if ck.servers[serverIdx] == currentKvClient { // Check if it's still the same client instance
						ck.servers[serverIdx].Valid = false // Mark as potentially invalid
						if ck.servers[serverIdx].Realconn != nil {
							ck.servers[serverIdx].Realconn.Close() // Close the underlying connection
							ck.servers[serverIdx].Realconn = nil
						}
					}
					ck.mu.Unlock()
				}
				goto handle_error_and_retry // Break recv_loop and go to retry logic
			} else {
				firlog.Logger.Infof("Clerk Watch: SUCCESSFULLY received response from gRPC. Key: %s", string(resp.Key))
			}

			event := (*WatchEvent)(resp)
			select {
			case eventChan <- event:
				// Event sent
			case <-ctx.Done():
				firlog.Logger.Infof("Watch for key '%s' cancelled by context while sending event: %v.", watchKeyStr, ctx.Err())
				// stream.CloseSend() // Optional: attempt to signal server
				return // Exit goroutine
			}
		}

	handle_error_and_retry:
		// Non-blocking check for context cancellation before sleeping
		select {
		case <-ctx.Done():
			firlog.Logger.Infof("Watch for key '%s' cancelled by context during retry/backoff: %v.", watchKeyStr, ctx.Err())
			return
		default:
		}

		firlog.Logger.Infof("Watch for key '%s': Retrying after %v.", watchKeyStr, currentBackoff)
		time.Sleep(currentBackoff)
		currentBackoff *= watchRetryMultiplier
		if currentBackoff > watchRetryMaxBackoff {
			currentBackoff = watchRetryMaxBackoff
		}
		// serverIdx will be re-evaluated or kept for next attempt
		continue // Continue the outer loop to retry connection

	handle_error_and_retry_no_wait: // For cases like wrong leader where we want to try next server immediately
		select {
		case <-ctx.Done():
			firlog.Logger.Infof("Watch for key '%s' cancelled by context during immediate retry: %v.", watchKeyStr, ctx.Err())
			return
		default:
		}
		continue

	} // End of main retry loop
}

// Helper to check if a gRPC error code suggests a retry
// This is a simplified example; more sophisticated logic might be needed.
func shouldRetry(code codes.Code) bool {
	switch code {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Internal, codes.Unknown:
		return true
	case codes.FailedPrecondition: // This might be a leader issue, client should try another server
		return true
	default:
		return false
	}
}
func (c *clerkImpl) watchEtcd() {
	for {
		for i, kvclient := range c.servers {
			if !kvclient.Valid {
				if kvclient.Realconn != nil {
					kvclient.Realconn.Close()
				}
				k := kvraft.NewKvClient(c.conf.EtcdAddrs[i])
				if k != nil {
					c.servers[i] = k
					// firLog.Logger.Warnf("update etcd server[%d] addr[%s]", i, c.conf.EtcdAddrs[i])
				}
			}
		}
		time.Sleep(time.Millisecond * 500)
	}
}

func MakeClerk(conf firconfig.Clerk) Clerk {
	return &clerkImpl{
		servers:         make([]*kvraft.KVClient, len(conf.EtcdAddrs)),
		nextSendLocalId: 0,
		LatestOffset:    0,
		clientId:        nrand(),
		cTos:            make([]int, len(conf.EtcdAddrs)),
		sToc:            make([]int, len(conf.EtcdAddrs)),
		conf:            conf,
	}
}

func (ck *clerkImpl) doGetValue(key string, withPrefix bool) ([][]byte, error) {
	args := pb.GetArgs{
		Key:          key,
		ClientId:     ck.clientId,
		LatestOffset: ck.LatestOffset,
		WithPrefix:   withPrefix,
		Op:           pb.OpType_GetT,
	}
	return ck.read(&args)
}

func (ck *clerkImpl) doGetKV(key string, withPrefix bool, op pb.OpType, pageSize, pageIndex int) ([]common.Pair, error) {
	args := pb.GetArgs{
		Key:          key,
		ClientId:     ck.clientId,
		LatestOffset: ck.LatestOffset,
		WithPrefix:   withPrefix,
		PageSize:     int32(pageSize),
		PageIndex:    int32(pageIndex),
		Op:           op,
	}
	datas, err := ck.read(&args)
	if err != nil {
		return nil, err
	}
	rets := make([]common.Pair, len(datas))
	for i := range rets {
		d := gob.NewDecoder(bytes.NewBuffer(datas[i]))
		d.Decode(&rets[i])
	}
	return rets, nil
}

// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.Get", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.

func (ck *clerkImpl) read(args *pb.GetArgs) ([][]byte, error) {
	// You will have to modify this function.
	ck.mu.Lock()
	defer ck.mu.Unlock()
	totalCount := 0
	count := 0
	lastSendLocalId := -1
	for {
		totalCount++
		if totalCount > 2*len(ck.servers)*3 {
			return nil, kvraft.ErrFaild
		}
		if ck.nextSendLocalId == lastSendLocalId {
			count++
			if count > 3 {
				count = 0
				ck.changeNextSendId()
			}
		}

		// firLog.Logger.Infof("clinet [%d] [Get]:send[%d] args[%v]", ck.clientId, ck.nextSendLocalId, args)
		var validCount = 0
		for !ck.servers[ck.nextSendLocalId].Valid {
			ck.changeNextSendId()
			validCount++
			if validCount == len(ck.servers) {
				break
			}
		}
		if validCount == len(ck.servers) {
			firlog.Logger.Infoln("not exist valid etcd server")
			time.Sleep(time.Second)
			continue
		}
		reply, err := ck.servers[ck.nextSendLocalId].Conn.Get(context.Background(), args)

		//用reply初始化本地server表

		lastSendLocalId = ck.nextSendLocalId
		if err != nil {
			// firLog.Logger.Infof("clinet [%d] [Get]:[lost] args[%v]", ck.clientId, args)
			//对面失联，那就换下一个继续发
			ck.changeNextSendId()
			continue
		}

		ck.sToc[reply.ServerId] = ck.nextSendLocalId

		switch reply.Err {
		case kvraft.ErrOK:
			ck.LatestOffset++
			// firLog.Logger.Infof("clinet [%d] [Get]:[OK] get args[%v] reply[%v]", ck.clientId, args, reply)
			if len(reply.Value) == 0 {
				return nil, kvraft.ErrNil
			}

			return reply.Value, nil
		case kvraft.ErrNoKey:
			// firLog.Logger.Infof("clinet [%d] [Get]:[ErrNo key] get args[%v]", ck.clientId, args)
			ck.LatestOffset++
			return nil, kvraft.ErrNil
		case kvraft.ErrWrongLeader:
			// firLog.Logger.Infof("clinet [%d] [Get]:[ErrWrong LeaderId][%d] get args[%v] reply[%v]", ck.clientId, ck.nextSendLocalId, args, reply)
			//对方也不知道leader
			if reply.LeaderId == -1 {
				//寻找下一个
				ck.changeNextSendId()
			} else {
				//先记录对方返回的leaderid,再看看本地有没有这个id对应的记录
				if ck.sToc[reply.LeaderId] == -1 { //本地没初始化，往下一个发
					ck.changeNextSendId()
				} else { //本地知道，那下一个就发它所指定的localServerAddress
					ck.nextSendLocalId = ck.sToc[reply.LeaderId]
				}

			}
		case kvraft.ErrWaitForRecover:
			// firLog.Logger.Infof("client [%d] [Get]:[Wait for leader recover]", ck.clientId)
			time.Sleep(time.Millisecond * 200)
		default:
			firlog.Logger.Fatalf("Client [%d] Get reply unknown err [%s](probaly not init)", ck.clientId, reply.Err)
		}

	}
}

// shared by Put and Append.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer."+op, &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
func (ck *clerkImpl) write(key string, value, oriValue []byte, TTL time.Duration, op int32) error {
	// You will have to modify this function.

	ck.mu.Lock()
	defer ck.mu.Unlock()
	args := pb.PutAppendArgs{
		Key:          key,
		Value:        value,
		OriValue:     oriValue,
		Op:           op,
		ClientId:     ck.clientId,
		LatestOffset: ck.LatestOffset,
	}
	if TTL == 0 {
		args.DeadTime = 0
	} else {
		args.DeadTime = time.Now().Add(TTL).UnixMilli()
	}
	count := 0
	lastSendLocalId := -1
	totalCount := 0
	for {
		totalCount++
		if totalCount > 2*len(ck.servers)*3 {
			return kvraft.ErrFaild
		}
		if ck.nextSendLocalId == lastSendLocalId {
			count++
			if count > 5 {
				count = 0
				ck.changeNextSendId()
			}
		}

		var validCount = 0
		for !ck.servers[ck.nextSendLocalId].Valid {
			ck.changeNextSendId()
			validCount++
			if validCount == len(ck.servers) {
				break
			}
		}
		if validCount == len(ck.servers) {
			// firLog.Logger.Infoln("not exist valid etcd server")
			time.Sleep(time.Millisecond * 10)
			continue
		}

		// firLog.Logger.Infof("clinet [%d] [PutAppend]:send[%d] args[%v]", ck.clientId, ck.nextSendLocalId, args.String())
		reply, err := ck.servers[ck.nextSendLocalId].Conn.PutAppend(context.Background(), &args)
		// firLog.Logger.Debugln("receive etcd:", reply.String(), err)
		//根据reply初始化一下本地server表

		lastSendLocalId = ck.nextSendLocalId
		if err != nil {
			// firLog.Logger.Infof("clinet [%d] [PutAppend]:[lost] args[%v] err:", ck.clientId, args, err)
			//对面失联，那就换下一个继续发
			ck.changeNextSendId()
			continue
		}

		ck.sToc[reply.ServerId] = ck.nextSendLocalId

		switch reply.Err {
		case kvraft.ErrOK:
			ck.LatestOffset++
			// firLog.Logger.Infof("clinet [%d] [Get]:[OK] get args[%v] reply[%v]", ck.clientId, args, reply)
			return nil
		case kvraft.ErrNoKey:
			// firLog.Logger.Infof("clinet [%d] [Get]:[ErrNo key] get args[%v]", ck.clientId, args)
			ck.LatestOffset++
			return kvraft.ErrNil
		case kvraft.ErrCasFaildInt:
			ck.LatestOffset++
			return kvraft.ErrCASFaild
		case kvraft.ErrWrongLeader:
			// firLog.Logger.Infof("clinet [%d] [PutAppend]:[ErrWrong LeaderId][%d] get args[%v] reply[%v]", ck.clientId, ck.nextSendLocalId, args, reply)
			//对方也不知道leader
			if reply.LeaderId == -1 {
				//寻找下一个
				ck.changeNextSendId()
			} else {
				//记录对方返回的不可靠leaderId
				if ck.sToc[reply.LeaderId] == -1 { //但是本地还没初始化呢，那就往下一个发
					ck.changeNextSendId()
				} else { //本地还真知道，那下一个就发它所指定的localServerAddress
					ck.nextSendLocalId = ck.sToc[reply.LeaderId]
				}

			}

		default:
			firlog.Logger.Fatalf("Client [%d] [PutAppend]:reply unknown err [%s](probaly not init)", ck.clientId, reply.Err)
		}

	}
}

func (ck *clerkImpl) changeNextSendId() {
	ck.nextSendLocalId = (ck.nextSendLocalId + 1) % len(ck.servers)
}

func (ck *clerkImpl) Put(key string, value []byte, TTL time.Duration) error {
	return ck.write(key, value, nil, TTL, int32(pb.OpType_PutT))
}

func (ck *clerkImpl) Append(key string, value []byte, TTL time.Duration) error {
	return ck.write(key, value, nil, TTL, int32(pb.OpType_AppendT))
}

func (ck *clerkImpl) Delete(key string) error {
	return ck.write(key, nil, nil, 0, int32(pb.OpType_DelT))
}

func (ck *clerkImpl) DeleteWithPrefix(prefix string) error {
	return ck.write(prefix, nil, nil, 0, int32(pb.OpType_DelWithPrefix))
}

func (ck *clerkImpl) CAS(key string, origin, dest []byte, TTL time.Duration) (bool, error) {
	err := ck.write(key, dest, origin, TTL, int32(pb.OpType_CAST))
	if err != nil {
		if err == kvraft.ErrCASFaild {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (ck *clerkImpl) BatchWrite(p *Pipe) error {
	return ck.write("", p.Marshal(), nil, 0, int32(pb.OpType_BatchT))
}

func (ck *clerkImpl) Pipeline() *Pipe {
	return &Pipe{
		ck: ck,
	}
}

func (ck *clerkImpl) Get(key string) ([]byte, error) {
	r, err := ck.doGetValue(key, false)
	if err != nil {
		return nil, err
	}
	if len(r) == 1 {
		return r[0], nil
	}
	return nil, kvraft.ErrNil
}

func (ck *clerkImpl) GetWithPrefix(key string) ([][]byte, error) {
	r, err := ck.doGetValue(key, true)
	if err != nil {
		return nil, err
	}
	if len(r) != 0 {
		return r, nil
	}
	return nil, kvraft.ErrNil
}

func (ck *clerkImpl) Keys() ([]common.Pair, error) {
	return ck.doGetKV("", true, pb.OpType_GetKeys, 0, 0)
}

func (ck *clerkImpl) KVs() ([]common.Pair, error) {
	return ck.doGetKV("", true, pb.OpType_GetKVs, 0, 0)
}

func (ck *clerkImpl) KeysWithPage(pageSize, pageIndex int) ([]common.Pair, error) {
	return ck.doGetKV("", true, pb.OpType_GetKeys, pageSize, pageIndex)
}

func (ck *clerkImpl) KVsWithPage(pageSize, pageIndex int) ([]common.Pair, error) {
	return ck.doGetKV("", true, pb.OpType_GetKVs, pageSize, pageIndex)
}

// TODO 当TTL为零时，启动watchDog机制
func (ck *clerkImpl) Lock(key string, TTL time.Duration) (id string, err error) {
	r, err := uuid.NewRandom()
	if err != nil {
		firlog.Logger.Fatalln(err)
		return "", err
	}
	ok, err := ck.CAS(key, nil, []byte(r.String()), TTL)
	if err != nil {
		firlog.Logger.Fatalln(err)
		return "", err
	}
	if !ok {
		return "", nil
	}

	return r.String(), nil
}

func (ck *clerkImpl) Unlock(key, id string) (bool, error) {
	ok, err := ck.CAS(key, []byte(id), nil, 0)
	if err != nil {
		firlog.Logger.Fatalln(err)
		return false, err
	}
	if !ok {
		return false, nil
	}
	return true, nil
}

func (ck *clerkImpl) WatchDog(key string, value []byte) (cancel func()) {
	// 选择最小5s的生存周期
	var (
		TTL  = time.Second * 5
		flag = new(bool)
	)
	ck.Put(key, value, TTL)
	*flag = false
	go func(key string, ori []byte) { //每500ms续期一次WatchDog
		for {
			curData := make([]byte, len(ori))
			copy(curData, ori)
			if *flag {
				return
			}
			ck.Put(key, curData, TTL)
			time.Sleep(TTL)
		}
	}(key, value)
	return func() {
		*flag = true
	}
}
