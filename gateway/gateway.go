package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/whosefriendA/firEtcd/pkg/firlog"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whosefriendA/firEtcd/client"
	"github.com/whosefriendA/firEtcd/common"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
)

// func init() {
// 	conf := firConfig.Gateway{}
// 	firConfig.Init("config.yml", &conf)
// 	// firLog.Logger.Debugln("check conf", conf)
// 	ck = client.MakeClerk(conf.Clerk)
// }

// func main() {
// 	r := gin.Default()
// 	r.GET("/keys", getKeys)
// 	r.Run(":9292")
// }

type Gateway struct {
	ck   *client.Clerk
	conf firconfig.Gateway
}

func NewGateway(conf firconfig.Gateway) *Gateway {
	return &Gateway{
		ck:   client.MakeClerk(conf.Clerk),
		conf: conf,
	}
}

// realurl = baseUrl + '/keys' | '/put' act...
// default baseUrl = '/firEtcd'
func (g *Gateway) Run() error {
	r := gin.Default()

	r.Use(InstantLogger())

	r.GET(g.conf.BaseUrl+"/keys", g.keys)
	r.GET(g.conf.BaseUrl+"/key", g.get)
	r.GET(g.conf.BaseUrl+"/keysWithPrefix", g.getWithPrefix)
	r.GET(g.conf.BaseUrl+"/kvs", g.kvs)
	r.GET(g.conf.BaseUrl+"/watch", g.watch)
	r.POST(g.conf.BaseUrl+"/put", g.put)
	r.POST(g.conf.BaseUrl+"/putCAS", g.putCAS)
	r.DELETE(g.conf.BaseUrl+"/key", g.del)
	r.DELETE(g.conf.BaseUrl+"/keysWithPrefix", g.delWithPrefix)
	return r.Run(g.conf.Port)
}

func (g *Gateway) LoadRouter(r *gin.Engine) {
	r.GET(g.conf.BaseUrl+"/keys", g.keys)
	r.GET(g.conf.BaseUrl+"/key", g.get)
	r.GET(g.conf.BaseUrl+"/keysWithPrefix", g.getWithPrefix)
	r.GET(g.conf.BaseUrl+"/kvs", g.kvs)
	r.GET(g.conf.BaseUrl+"/watch", g.watch)
	r.POST(g.conf.BaseUrl+"/put", g.put)
	r.POST(g.conf.BaseUrl+"/putCAS", g.putCAS)
	r.DELETE(g.conf.BaseUrl+"/key", g.del)
	r.DELETE(g.conf.BaseUrl+"/keysWithPrefix", g.delWithPrefix)
}

func InstantLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求一进来就立即打印日志
		log.Printf("[InstantLogger] ==> %s %s", c.Request.Method, c.Request.RequestURI)

		// 调用 c.Next() 将控制权交给下一个中间件或处理器
		c.Next()

		// c.Next() 返回后（对于SSE永远不会），可以再记录一些信息
		// log.Printf("[InstantLogger] <== %s %s completed with status %d", c.Request.Method, c.Request.RequestURI, c.Writer.Status())
	}
}

// watch 是处理 /watch SSE 请求的 Gin处理器
func (g *Gateway) watch(c *gin.Context) {

	key := c.Query("key")
	isPrefixStr := c.DefaultQuery("isPrefix", "false")
	sendInitialStateStr := c.DefaultQuery("sendInitialState", "false") // 示例参数，实际实现可能更复杂

	//clientIP := c.ClientIP()
	//requestID := c.GetString("request_id") // 假设有中间件设置了 request_id

	//slog.Info("Watch request received", logAttrs...)

	if key == "" {
		//slog.Warn("Watch request rejected: key parameter is required", logAttrs...)
		c.JSON(http.StatusBadRequest, gin.H{"error": "key parameter is required"})
		return
	}

	isPrefix, err := strconv.ParseBool(isPrefixStr)
	if err != nil {
		//slog.Warn("Watch request rejected: invalid isPrefix value", append(logAttrs, slog.String("isPrefixStr", isPrefixStr))...)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid isPrefix value, must be true or false"})
		return
	}

	_, err = strconv.ParseBool(sendInitialStateStr) // 仅校验，具体实现 sendInitialState 较复杂，此处略
	if err != nil {
		//slog.Warn("Watch request rejected: invalid sendInitialState value", append(logAttrs, slog.String("sendInitialStateStr", sendInitialStateStr))...)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sendInitialState value, must be true or false"})
		return
	}
	// logAttrs = append(logAttrs, slog.Bool("sendInitialState", sendInitialState)) // 如果实现了

	// 设置 SSE 头部
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // 对于 Nginx 等代理

	// 立即刷新头部，确保客户端知道这是一个流式响应
	// Gin 的 c.Stream() 或 c.SSEvent() 会隐式处理 flush，但首次手动 flush 确保头部立即发送
	c.Writer.Flush()

	opts := []client.WatchOption{
		client.WithSendInitialState(true),
	}

	if isPrefix {
		opts = append(opts, client.WithPrefix())
	}

	// 调用后端 Watch 方法，传递请求上下文用于取消
	// 注意：实际的 WatchOption 处理（如 sendInitialState）可能需要更复杂的逻辑
	// 例如，如果 sendInitialState 为 true，可能需要先调用 Get，然后再 Watch
	eventCh, err := g.ck.Watch(c.Request.Context(), key, opts...)
	if err != nil {
		// 区分是客户端请求问题还是后端问题
		// 假设后端Watch方法在无法建立连接时返回特定错误类型
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			//slog.Warn("Backend watch initiation cancelled or timed out, possibly due to client disconnect before stream start", append(logAttrs, slog.Any("error", err))...)
			// 客户端可能已经断开，尝试写入JSON可能失败
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "watch initiation timed out or was cancelled"})
			return
		}
		//slog.Error("Failed to start backend watch stream", append(logAttrs, slog.Any("error", err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate watch on backend"})
		return
	}

	//slog.Info("SSE stream established with backend", logAttrs...)

	// 事件流循环
	for {
		select {
		case <-c.Request.Context().Done(): // HTTP 客户端断开连接
			// 当 c.Request.Context() 被取消时，g.ck.Watch() 内部应捕获此取消并清理其 gRPC 流
			//slog.Info("Client disconnected, watch stream closing", logAttrs...)
			return // 退出处理器函数，Gin 会处理后续清理

		case event, ok := <-eventCh:
			if !ok { // 后端事件通道已关闭
				//slog.Info("Backend event channel closed, watch stream ending", logAttrs...)
				// 可以选择向客户端发送一个流结束的信号
				firlog.Logger.Warnf("Gateway Watch: Backend event channel was closed.")
				c.SSEvent("stream_end", gin.H{"message": "Watch stream ended from backend"})
				//if errSSE!= nil {
				//// slog.Warn("Failed to send stream_end SSE event", append(logAttrs, slog.Any("error", errSSE))...)
				//}
				return
			}
			firlog.Logger.Infof("Gateway Watch: SUCCESSFULLY received an event from Clerk's channel. Preparing to send to client via SSE.")

			// 将 WatchEvent 格式化为 JSON
			// 注意：WatchEvent 结构体中的byte 字段在 json.Marshal 时会自动 base64 编码
			jsonData, err := json.Marshal(event)
			if err != nil {
				//slog.Error("Failed to marshal WatchEvent to JSON", append(logAttrs, slog.Any("error", err), slog.Any("event", event))...)
				// 决定是继续还是终止流
				// 如果错误是临时的或特定于事件的，可以 continue
				// 如果是严重错误，可以 return
				// 也可以尝试发送一个错误事件给客户端
				// c.SSEvent("error", gin.H{"type": "serialization_error", "message": err.Error()})
				continue
			}

			// 使用 Gin 的 c.SSEvent 发送事件。它会自动处理 SSE 格式和刷新。
			// 事件名可以是 "message" 或其他自定义名称。
			c.SSEvent("message", string(jsonData))
			//if err!= nil {
			//	// 写入失败通常意味着客户端已断开连接
			//	//slog.Warn("Error sending SSE event, client likely disconnected", append(logAttrs, slog.Any("error", err))...)
			//	// 循环将在下一次迭代中通过 c.Request.Context().Done() 捕获到断开
			//	// 或者，如果错误明确指示连接问题，可以直接 return
			//	return
			//}
			//slog.Debug("SSE event sent", append(logAttrs, slog.String("eventType", event.Type))...)
		}
	}
}

func (g *Gateway) keys(c *gin.Context) {
	var pairs []Pair
	switch c.Query("type") {
	case "", "all":
		cpairs, err := g.ck.Keys()
		if err != nil {
			c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": err.Error()}})
			return
		}
		pairs = make([]Pair, len(cpairs))
		for i := range cpairs {
			pairs[i].Key = cpairs[i].Key
			pairs[i].DeadTime = cpairs[i].Entry.DeadTime
		}
		c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: pairs})
	//TODO keys page
	case "page":

	default:
		c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": "wrong type"}})
	}
}

func (g *Gateway) kvs(c *gin.Context) {
	var pairs []Pair
	switch c.Query("type") {
	case "", "all":
		cpairs, err := g.ck.KVs()
		if err != nil {
			c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": err.Error()}})
			return
		}
		pairs = make([]Pair, len(cpairs))
		for i := range cpairs {
			pairs[i].Key = cpairs[i].Key
			pairs[i].Value = common.BytesToString(cpairs[i].Entry.Value)
			pairs[i].DeadTime = cpairs[i].Entry.DeadTime
		}
		c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: pairs})
	//TODO kvs page
	case "page":
		c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": "wrong type"}})

	default:
	}
}

type Pair struct {
	Key      string
	Value    string
	DeadTime int64
}

func (g *Gateway) put(c *gin.Context) {
	var pairs []Pair
	err := c.ShouldBind(&pairs)
	if err != nil {
		c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": err.Error()}})
		return
	}
	var ttl time.Duration
	pipe := g.ck.Pipeline()

	for i := range pairs {
		if pairs[i].DeadTime == 0 {
			ttl = 0
		} else {
			deadTimestamp := time.UnixMilli(pairs[i].DeadTime)
			if time.Now().After(deadTimestamp) {
				c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": "ok(ignore)"}})
				continue
			}
			ttl = time.Until(deadTimestamp)
		}
		pipe.Put(pairs[i].Key, common.StringToBytes(pairs[i].Value), ttl)
	}
	err = pipe.Exec()
	if err != nil {
		c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": "ok"}})

}

type PairCAS struct {
	Key      string
	Value    string
	OriValue string
	DeadTime int64
}

func (g *Gateway) putCAS(c *gin.Context) {
	var pair PairCAS
	err := c.ShouldBind(&pair)
	if err != nil {
		c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": err.Error()}})
		return
	}
	var ttl time.Duration
	if pair.DeadTime == 0 {
		ttl = 0
	} else {
		deadTimestamp := time.UnixMilli(pair.DeadTime)
		if time.Now().After(deadTimestamp) {
			c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": "ok(ignore)"}})
			return
		}
		ttl = time.Until(deadTimestamp)
	}
	ok, err := g.ck.CAS(pair.Key, common.StringToBytes(pair.OriValue), common.StringToBytes(pair.Value), ttl)
	if err != nil {
		c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": fmt.Sprintf("%t", ok)}})
}

func (g *Gateway) del(c *gin.Context) {
	key := c.Query("key")
	err := g.ck.Delete(key)
	if err != nil {
		c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": "ok"}})
}

func (g *Gateway) delWithPrefix(c *gin.Context) {
	prefix := c.Query("prefix")
	err := g.ck.DeleteWithPrefix(prefix)
	if err != nil {
		c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": "ok"}})
}

func (g *Gateway) get(c *gin.Context) {
	key := c.Query("key")
	value, err := g.ck.Get(key)
	if err != nil {
		c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"value": Pair{
		Value: common.BytesToString(value),
	}, "msg": "ok"}})
}

func (g *Gateway) getWithPrefix(c *gin.Context) {
	prefix := c.Query("prefix")
	value, err := g.ck.GetWithPrefix(prefix)
	if err != nil {
		c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"msg": err.Error()}})
		return
	}
	pairs := make([]Pair, len(value))
	for i := range value {
		pairs[i].Value = common.BytesToString(value[i])
	}
	c.JSON(http.StatusOK, common.Respond{Code: http.StatusOK, Data: gin.H{"value": pairs, "msg": "ok"}})
}
