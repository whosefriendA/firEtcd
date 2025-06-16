package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/whosefriendA/firEtcd/client"
	"github.com/whosefriendA/firEtcd/common"
	"github.com/whosefriendA/firEtcd/pkg/firlog"
)

type Pair struct {
	Key      string
	Value    string
	DeadTime int64
}

type PairCAS struct {
	Key      string
	Value    string
	OriValue string
	DeadTime int64
}

func (g *Gateway) keys(w http.ResponseWriter, r *http.Request) {
	cpairs, err := g.ck.Keys()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	pairs := make([]Pair, len(cpairs))
	for i := range cpairs {
		pairs[i].Key = cpairs[i].Key
		pairs[i].DeadTime = cpairs[i].Entry.DeadTime
	}
	writeJSON(w, http.StatusOK, pairs)
}

func (g *Gateway) kvs(w http.ResponseWriter, r *http.Request) {
	cpairs, err := g.ck.KVs()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	pairs := make([]Pair, len(cpairs))
	for i := range cpairs {
		pairs[i].Key = cpairs[i].Key
		pairs[i].Value = common.BytesToString(cpairs[i].Entry.Value)
		pairs[i].DeadTime = cpairs[i].Entry.DeadTime
	}
	writeJSON(w, http.StatusOK, pairs)
}

func (g *Gateway) watch(w http.ResponseWriter, r *http.Request) {
	// 1. 参数解析 (标准库版本)
	query := r.URL.Query()
	key := query.Get("key")

	isPrefix, _ := strconv.ParseBool(query.Get("isPrefix"))
	sendInitialState, _ := strconv.ParseBool(query.Get("sendInitialState"))

	if key == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("query parameter 'key' is required"))
		return
	}

	// 2. 设置SSE响应头
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")

	// 3. 获取 http.Flusher 接口
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("streaming unsupported"))
		return
	}

	// 4. 创建独立的后端 Watch 上下文
	watchCtx, cancelWatch := context.WithCancel(context.Background())
	defer cancelWatch()

	// 5. 启动后端 Watch (这部分业务逻辑不变)
	opts := []client.WatchOption{
		client.WithSendInitialState(sendInitialState),
	}
	if isPrefix {
		opts = append(opts, client.WithPrefix())
	}

	eventCh, err := g.ck.Watch(watchCtx, key, opts...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to initiate watch on backend: %w", err))
		return
	}

	firlog.Logger.Infof("Gateway Watch: SSE stream established for key '%s'", key)

	// 第一次手动刷新，确保头部立即发送给客户端
	flusher.Flush()

	// 6. 事件循环 (使用 r.Context() 替代 c.Request.Context())
	for {
		select {
		case <-r.Context().Done():
			// 当客户端关闭连接时，r.Context().Done() 会被关闭
			firlog.Logger.Infof("Gateway Watch: Client disconnected for key '%s'.", key)
			return

		case event, ok := <-eventCh:
			if !ok {
				// 后端 Watch 通道已关闭
				firlog.Logger.Warnf("Gateway Watch: Backend event channel for key '%s' was closed.", key)
				fmt.Fprintf(w, "event: stream_end\ndata: {\"message\": \"Watch stream ended from backend\"}\n\n")
				flusher.Flush()
				return
			}

			jsonData, err := json.Marshal(event)
			if err != nil {
				firlog.Logger.Errorf("Gateway Watch: Failed to marshal event for key '%s'. Error: %v", key, err)
				continue
			}

			// 手动构建完全符合 SSE 规范的消息
			message := fmt.Sprintf("data: %s\n\n", string(jsonData))

			// 将消息写入响应
			if _, err := fmt.Fprint(w, message); err != nil {
				// 写入失败，通常意味着客户端已断开
				firlog.Logger.Errorf("Gateway Watch: Failed to write SSE data. Error: %v", err)
				return
			}

			// 关键步骤：强制将缓冲区的数据刷新到网络连接上
			flusher.Flush()
			firlog.Logger.Infof("Gateway Watch: Successfully flushed data for key '%s'", event.Key)

		case <-time.After(30 * time.Second):
			// 心跳机制：防止代理或服务器因为连接空闲而切断连接
			firlog.Logger.Infof("Gateway Watch: No event for 30s on key '%s'. Sending keep-alive.", key)

			// SSE 注释以冒号开头，会被客户端忽略，但能保持连接活跃
			if _, err := fmt.Fprintf(w, ": keep-alive\n\n"); err != nil {
				firlog.Logger.Errorf("Gateway Watch: Failed to write keep-alive. Error: %v", err)
				return
			}
			flusher.Flush()
		}
	}
}

func (g *Gateway) put(w http.ResponseWriter, r *http.Request) {
	var pairs []Pair
	if err := json.NewDecoder(r.Body).Decode(&pairs); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()
	var ttl time.Duration
	pipe := g.ck.Pipeline()

	for i := range pairs {
		if pairs[i].DeadTime == 0 {
			ttl = 0
		} else {
			deadTimestamp := time.UnixMilli(pairs[i].DeadTime)
			if time.Now().After(deadTimestamp) {
				continue
			}
			ttl = time.Until(deadTimestamp)
		}
		pipe.Put(pairs[i].Key, common.StringToBytes(pairs[i].Value), ttl)
	}

	if err := pipe.Exec(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"msg": "ok"})
}

func (g *Gateway) putCAS(w http.ResponseWriter, r *http.Request) {
	var p PairCAS
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	var ttl time.Duration
	if p.DeadTime != 0 {
		deadTimestamp := time.UnixMilli(p.DeadTime)
		if time.Now().After(deadTimestamp) {
			writeJSON(w, http.StatusOK, map[string]string{"msg": "ok(ignore)"})
			return
		}
		ttl = time.Until(deadTimestamp)
	}

	ok, err := g.ck.CAS(p.Key, common.StringToBytes(p.OriValue), common.StringToBytes(p.Value), ttl)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": ok})
}

func (g *Gateway) del(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("query parameter 'key' is required"))
		return
	}
	if err := g.ck.Delete(key); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"msg": "ok"})
}

func (g *Gateway) delWithPrefix(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("query parameter 'prefix' is required"))
		return
	}
	if err := g.ck.DeleteWithPrefix(prefix); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"msg": "ok"})
}

func (g *Gateway) get(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("query parameter 'key' is required"))
		return
	}

	value, err := g.ck.Get(key)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	if value == nil {
		writeError(w, http.StatusNotFound, fmt.Errorf("key '%s' not found", key))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{key: common.BytesToString(value)})
}

func (g *Gateway) getWithPrefix(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("query parameter 'prefix' is required"))
		return
	}
	kvs, err := g.ck.GetWithPrefix(prefix)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, kvs)
}
