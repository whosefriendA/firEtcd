package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/whosefriendA/firEtcd/common"
	"github.com/whosefriendA/firEtcd/kvraft"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
	"github.com/whosefriendA/firEtcd/pkg/firlog"
	"github.com/whosefriendA/firEtcd/proto/pb"
)

var pipeLimit int = 1024 * 4000

type Client struct {
	servers []*kvraft.KVconn

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

func (ck *Client) Watch(ctx context.Context, key string, opts ...WatchOption) (<-chan *WatchEvent, error) {
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

	return eventChan, nil
}

func (ck *Client) manageWatchStream(ctx context.Context, watchKeyStr string, req *pb.WatchRequest, eventChan chan<- *WatchEvent) {
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
		var currentKvClient *kvraft.KVconn

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
					firlog.Logger.Errorf("Client Watch: FAILED to receive from gRPC stream. Error: %v", recvErr)
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
				firlog.Logger.Infof("Client Watch: SUCCESSFULLY received response from gRPC. Key: %s", string(resp.Key))
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

// shouldRetry checks if a gRPC error code suggests a retry.
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
func (c *Client) watchEtcd() {
	for {
		for i, kvclient := range c.servers {
			if !kvclient.Valid {
				fmt.Printf("Client: try connect to etcd %s\n", c.conf.EtcdAddrs[i])
				if kvclient.Realconn != nil {
					kvclient.Realconn.Close()
				}
				k := kvraft.NewKvConn(c.conf.EtcdAddrs[i], c.conf.TLS)
				if k != nil {
					c.servers[i] = k
					// firLog.Logger.Warnf("update etcd server[%d] addr[%s]", i, c.conf.EtcdAddrs[i])
				}
			}
		}
		time.Sleep(time.Millisecond * 500)
	}
}

func MakeClerk(conf firconfig.Clerk) *Client {
	ck := new(Client)
	ck.conf = conf
	ck.servers = make([]*kvraft.KVconn, len(conf.EtcdAddrs))
	for i := range ck.servers {
		ck.servers[i] = new(kvraft.KVconn)
		ck.servers[i].Valid = false
	}
	fmt.Printf("Client etcd addrs: %+v\n", conf.EtcdAddrs)
	ck.nextSendLocalId = int(nrand() % int64(len(conf.EtcdAddrs)))
	ck.LatestOffset = 1
	ck.clientId = nrand()
	ck.cTos = make([]int, len(conf.EtcdAddrs))
	ck.sToc = make([]int, len(conf.EtcdAddrs))
	for i := range ck.cTos {
		ck.cTos[i] = -1
		ck.sToc[i] = -1
	}
	go ck.watchEtcd()
	firlog.Logger.Infof("client:::asdfahsdljfkhaskldhflkasd")
	return ck
}

func (ck *Client) doGetValue(key string, withPrefix bool) ([][]byte, error) {
	args := pb.GetArgs{
		Key:          key,
		ClientId:     ck.clientId,
		LatestOffset: ck.LatestOffset,
		WithPrefix:   withPrefix,
		Op:           pb.OpType_GetT,
	}
	return ck.read(&args)
}

func (ck *Client) doGetKV(key string, withPrefix bool, op pb.OpType, pageSize, pageIndex int) ([]common.Pair, error) {
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

func (ck *Client) read(args *pb.GetArgs) ([][]byte, error) {
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
			firlog.Logger.Warnf("gRPC Get for key '%s' failed with server %d: %v", args.Key, ck.nextSendLocalId, err)
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

func (ck *Client) write(key string, value, oriValue []byte, TTL time.Duration, op int32) error {
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

		// If TTL>0 and LeaseId not yet acquired, try to grant a lease on the target server
		if TTL > 0 && args.LeaseId == 0 {
			leaseClient := pb.NewLeaseClient(ck.servers[ck.nextSendLocalId].Realconn)
			lr, lerr := leaseClient.LeaseGrant(context.Background(), &pb.LeaseGrantRequest{TTLMs: int64(TTL / time.Millisecond)})
			if lerr != nil || lr == nil || lr.ID == 0 {
				// try next server
				ck.changeNextSendId()
				continue
			}
			args.LeaseId = lr.ID
		}

		// firLog.Logger.Infof("clinet [%d] [PutAppend]:send[%d] args[%v]", ck.clientId, ck.nextSendLocalId, args.String())
		reply, err := ck.servers[ck.nextSendLocalId].Conn.PutAppend(context.Background(), &args)
		// firLog.Logger.Debugln("receive etcd:", reply.String(), err)
		//根据reply初始化一下本地server表

		lastSendLocalId = ck.nextSendLocalId
		if err != nil {
			firlog.Logger.Warnf("gRPC PutAppend for key '%s' failed with server %d: %v", args.Key, ck.nextSendLocalId, err)
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
				} else { //本地知道，那下一个就发它所指定的localServerAddress
					ck.nextSendLocalId = ck.sToc[reply.LeaderId]
				}

			}

		default:
			firlog.Logger.Fatalf("Client [%d] [PutAppend]:reply unknown err [%s](probaly not init)", ck.clientId, reply.Err)
		}

	}
}

func (ck *Client) changeNextSendId() {
	ck.nextSendLocalId = (ck.nextSendLocalId + 1) % len(ck.servers)
}

func (ck *Client) Put(key string, value []byte, TTL time.Duration) error {
	return ck.write(key, value, nil, TTL, int32(pb.OpType_PutT))
}

func (ck *Client) Append(key string, value []byte, TTL time.Duration) error {
	return ck.write(key, value, nil, TTL, int32(pb.OpType_AppendT))
}

func (ck *Client) Delete(key string) error {
	return ck.write(key, nil, nil, 0, int32(pb.OpType_DelT))
}

func (ck *Client) DeleteWithPrefix(prefix string) error {
	return ck.write(prefix, nil, nil, 0, int32(pb.OpType_DelWithPrefix))
}

func (ck *Client) CAS(key string, origin, dest []byte, TTL time.Duration) (bool, error) {
	err := ck.write(key, dest, origin, TTL, int32(pb.OpType_CAST))
	if err != nil {
		if err == kvraft.ErrCASFaild {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (ck *Client) BatchWrite(p *Pipe) error {
	// Pre-grant leases for ops with DeadTime>0
	for i := range p.Ops {
		op := &p.Ops[i]
		if op.Entry.DeadTime != 0 && op.LeaseId == 0 {
			// compute remaining TTL
			ttlMs := op.Entry.DeadTime - time.Now().UnixMilli()
			if ttlMs <= 0 {
				continue
			}
			granted := false
			attempts := 0
			for attempts < 2*len(ck.servers) && !granted {
				// rotate to a valid server
				validCount := 0
				for !ck.servers[ck.nextSendLocalId].Valid {
					ck.changeNextSendId()
					validCount++
					if validCount == len(ck.servers) {
						break
					}
				}
				if validCount == len(ck.servers) {
					time.Sleep(10 * time.Millisecond)
					attempts++
					continue
				}
				leaseClient := pb.NewLeaseClient(ck.servers[ck.nextSendLocalId].Realconn)
				lr, lerr := leaseClient.LeaseGrant(context.Background(), &pb.LeaseGrantRequest{TTLMs: ttlMs})
				if lerr != nil || lr == nil || lr.ID == 0 {
					ck.changeNextSendId()
					attempts++
					continue
				}
				op.LeaseId = lr.ID
				granted = true
			}
		}
	}
	return ck.write("", p.Marshal(), nil, 0, int32(pb.OpType_BatchT))
}

func (ck *Client) Pipeline() *Pipe {
	return &Pipe{
		ck: ck,
	}
}

func (ck *Client) Get(key string) ([]byte, error) {
	r, err := ck.doGetValue(key, false)
	if err != nil {
		return nil, err
	}
	if len(r) == 1 {
		return r[0], nil
	}
	return nil, kvraft.ErrNil
}

func (ck *Client) GetWithPrefix(key string) ([][]byte, error) {
	r, err := ck.doGetValue(key, true)
	if err != nil {
		return nil, err
	}
	if len(r) != 0 {
		return r, nil
	}
	return nil, kvraft.ErrNil
}

func (ck *Client) Keys() ([]common.Pair, error) {
	return ck.doGetKV("", true, pb.OpType_GetKeys, 0, 0)
}

func (ck *Client) KVs() ([]common.Pair, error) {
	return ck.doGetKV("", true, pb.OpType_GetKVs, 0, 0)
}

func (ck *Client) KeysWithPage(pageSize, pageIndex int) ([]common.Pair, error) {
	return ck.doGetKV("", true, pb.OpType_GetKeys, pageSize, pageIndex)
}

func (ck *Client) KVsWithPage(pageSize, pageIndex int) ([]common.Pair, error) {
	return ck.doGetKV("", true, pb.OpType_GetKVs, pageSize, pageIndex)
}

func (ck *Client) Lock(key string, TTL time.Duration) (id string, err error) {
	leaseID, err := ck.LeaseGrant(TTL)
	if err != nil {
		return "", err
	}

	leaseIDStr := strconv.FormatInt(leaseID, 10)
	ok, err := ck.CAS(key, nil, []byte(leaseIDStr), 0)
	if err != nil || !ok {
		ck.LeaseRevoke(leaseID)
		return "", err
	}

	return leaseIDStr, nil
}

func (ck *Client) Unlock(key, id string) (bool, error) {
	ok, err := ck.CAS(key, []byte(id), nil, 0)
	if err != nil || !ok {
		return false, err
	}

	leaseID, _ := strconv.ParseInt(id, 10, 64)
	if leaseID > 0 {
		ck.LeaseRevoke(leaseID)
	}

	return true, nil
}

// LockWithKeepAlive 获取锁并自动续约，返回取消函数
func (ck *Client) LockWithKeepAlive(key string, TTL time.Duration) (id string, cancel func(), err error) {
	leaseID, err := ck.LeaseGrant(TTL)
	if err != nil {
		return "", nil, err
	}

	leaseIDStr := strconv.FormatInt(leaseID, 10)
	ok, err := ck.CAS(key, nil, []byte(leaseIDStr), 0)
	if err != nil || !ok {
		ck.LeaseRevoke(leaseID)
		return "", nil, err
	}

	cancelFunc := ck.AutoKeepAlive(leaseID, TTL/3)

	return leaseIDStr, func() {
		cancelFunc()
		ck.Unlock(key, leaseIDStr)
	}, nil
}

func (ck *Client) LeaseGrant(ttl time.Duration) (int64, error) {
	// find a valid server
	validCount := 0
	for !ck.servers[ck.nextSendLocalId].Valid {
		ck.changeNextSendId()
		validCount++
		if validCount == len(ck.servers) {
			return 0, fmt.Errorf("no valid servers")
		}
	}
	leaseClient := pb.NewLeaseClient(ck.servers[ck.nextSendLocalId].Realconn)
	lr, err := leaseClient.LeaseGrant(context.Background(), &pb.LeaseGrantRequest{TTLMs: int64(ttl / time.Millisecond)})
	if err != nil || lr == nil {
		return 0, err
	}
	return lr.ID, nil
}

func (ck *Client) LeaseRevoke(leaseID int64) error {
	validCount := 0
	for !ck.servers[ck.nextSendLocalId].Valid {
		ck.changeNextSendId()
		validCount++
		if validCount == len(ck.servers) {
			return fmt.Errorf("no valid servers")
		}
	}
	leaseClient := pb.NewLeaseClient(ck.servers[ck.nextSendLocalId].Realconn)
	_, err := leaseClient.LeaseRevoke(context.Background(), &pb.LeaseRevokeRequest{ID: leaseID})
	return err
}

func (ck *Client) LeaseTimeToLive(leaseID int64, withKeys bool) (int64, []string, error) {
	validCount := 0
	for !ck.servers[ck.nextSendLocalId].Valid {
		ck.changeNextSendId()
		validCount++
		if validCount == len(ck.servers) {
			return 0, nil, fmt.Errorf("no valid servers")
		}
	}
	leaseClient := pb.NewLeaseClient(ck.servers[ck.nextSendLocalId].Realconn)
	lr, err := leaseClient.LeaseTimeToLive(context.Background(), &pb.LeaseTimeToLiveRequest{ID: leaseID, Keys: withKeys})
	if err != nil || lr == nil {
		return 0, nil, err
	}
	return lr.TTLMs, lr.Keys, nil
}

// AutoKeepAlive starts background renewal for a lease, returns cancel function
func (ck *Client) AutoKeepAlive(leaseID int64, interval time.Duration) (cancel func()) {
	flag := new(bool)
	*flag = false
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			if *flag {
				return
			}
			select {
			case <-ticker.C:
				// try to keep alive
				validCount := 0
				for !ck.servers[ck.nextSendLocalId].Valid {
					ck.changeNextSendId()
					validCount++
					if validCount == len(ck.servers) {
						time.Sleep(interval)
						continue
					}
				}
				leaseClient := pb.NewLeaseClient(ck.servers[ck.nextSendLocalId].Realconn)
				stream, err := leaseClient.LeaseKeepAlive(context.Background())
				if err != nil {
					time.Sleep(interval)
					continue
				}
				if err := stream.Send(&pb.LeaseKeepAliveRequest{ID: leaseID}); err != nil {
					stream.CloseSend()
					time.Sleep(interval)
					continue
				}
				resp, err := stream.Recv()
				stream.CloseSend()
				if err != nil || resp == nil || resp.TTLMs == 0 {
					// lease expired or not found
					return
				}
			}
		}
	}()
	return func() {
		*flag = true
	}
}
