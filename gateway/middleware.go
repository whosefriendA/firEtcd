package gateway

import (
	"log"
	"net/http"
)

// 新的 InstantLogger
func InstantLogger(next http.Handler) http.Handler {
	// http.HandlerFunc 是一个适配器，让普通函数可以作为 http.Handler 使用
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 请求进来时打印日志
		log.Printf("[InstantLogger] ==> %s %s", r.Method, r.RequestURI)

		// 调用链中的下一个处理器 (可能是另一个中间件，也可能是最终的路由处理器)
		next.ServeHTTP(w, r)

		// 响应返回后可以再做一些事，但需要 ResponseWriter Wrapper，这里先简化
	})
}
