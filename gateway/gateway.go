package gateway

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/whosefriendA/firEtcd/pkg/firlog"

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

	// 加载TLS证书
	certificate, err := tls.LoadX509KeyPair("/home/wanggang/firEtcd/pkg/tls/certs/gateway.crt", "/home/wanggang/firEtcd/pkg/tls/certs/gateway.key")
	if err != nil {
		firlog.Logger.Fatalf("无法加载网关证书: %v", err)
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

	// 创建TLS服务器
	server := &http.Server{
		Addr:      g.conf.Port,
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	// 启动TLS服务器
	return server.ListenAndServeTLS("", "")
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

func (g *Gateway) watch(c *gin.Context) {
	// 1. 参数解析和头部设置
	key := c.Query("key")
	isPrefix, _ := strconv.ParseBool(c.DefaultQuery("isPrefix", "false"))
	sendInitialState, _ := strconv.ParseBool(c.DefaultQuery("sendInitialState", "false"))

	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key parameter is required"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	// 2. 创建独立的后端 Watch 上下文
	watchCtx, cancelWatch := context.WithCancel(context.Background())
	defer cancelWatch()

	// 3. 启动后端 Watch
	opts := []client.WatchOption{
		client.WithSendInitialState(sendInitialState),
	}
	if isPrefix {
		opts = append(opts, client.WithPrefix())
	}

	eventCh, err := g.ck.Watch(watchCtx, key, opts...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate watch on backend"})
		return
	}

	firlog.Logger.Infof("Gateway Watch: SSE stream established for key '%s'", key)

	// 4. 获取底层的 http.Flusher 接口，这是手动刷新能工作的关键
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		// 如果 Writer 不支持 Flusher，SSE 将无法工作
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	// 第一次手动刷新，确保头部立即发送
	flusher.Flush()

	// 5. 事件循环
	for {
		select {
		case <-c.Request.Context().Done():
			firlog.Logger.Infof("Gateway Watch: Client disconnected for key '%s'.", key)
			return

		case event, ok := <-eventCh:
			if !ok {
				firlog.Logger.Warnf("Gateway Watch: Backend event channel for key '%s' was closed.", key)
				// 发送一个流结束消息
				fmt.Fprintf(c.Writer, "event: stream_end\ndata: {\"message\": \"Watch stream ended from backend\"}\n\n")
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
			if _, err := fmt.Fprint(c.Writer, message); err != nil {
				// 写入失败，通常意味着客户端已断开
				firlog.Logger.Errorf("Gateway Watch: Failed to write SSE data. Error: %v", err)
				return
			}

			// **关键步骤：** 强制将缓冲区的数据刷新到网络连接上
			flusher.Flush()
			firlog.Logger.Infof("Gateway Watch: Successfully flushed data for key '%s'", event.Key)

		case <-time.After(30 * time.Second):
			// 心跳机制：为了防止代理或服务器因为连接空闲而切断连接，
			// 每隔 30 秒发送一个 SSE 注释作为心跳。
			firlog.Logger.Infof("Gateway Watch: No event for 30s on key '%s'. Sending keep-alive.", key)

			// SSE 注释以冒号开头，会被客户端忽略，但能保持连接活跃
			if _, err := fmt.Fprintf(c.Writer, ": keep-alive\n\n"); err != nil {
				firlog.Logger.Errorf("Gateway Watch: Failed to write keep-alive. Error: %v", err)
				return
			}
			flusher.Flush()
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
	firlog.Logger.Infof("asdfahsdljfkhaskldhflkasd")
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
